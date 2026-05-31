package custd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	mrand "math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2/clientcredentials"
)

const (
	ingestBatchEndpoint = "/api/v1/events/batch"
)

// defaultHTTPTimeout is the timeout applied to the built-in HTTP client
// when no custom HTTPClient is provided.
const defaultHTTPTimeout = 10 * time.Second

// CustdClient sends analytics events to the ingest API.
// custd-structure: allow-methods public SDK facade; ingestion, flush, and admin helpers share client config/lifecycle
// custd-structure: allow-fields public SDK facade; fields keep queue, retry, HTTP, and lifecycle state together
type CustdClient struct {
	Admin      *AdminClient
	config     ClientConfig
	q          *queue
	retrySet   map[int]bool
	mu         sync.Mutex
	ticker     *time.Ticker
	done       chan struct{}
	wg         sync.WaitGroup
	closeOnce  sync.Once
	httpClient *http.Client
	rng        *mrand.Rand
	rngMu      sync.Mutex
}

// NewClient creates a new CustdClient with the given configuration.
// Missing fields are filled from DefaultClientConfig.
func NewClient(config *ClientConfig) *CustdClient {
	cfg := applyDefaults(config)
	c := &CustdClient{
		config:     cfg,
		q:          newQueue(cfg.MaxQueueSize),
		retrySet:   retryableStatusSet(cfg.Retry.RetryOnStatuses),
		done:       make(chan struct{}),
		httpClient: buildHTTPClient(&cfg),
		rng:        newSecureRand(),
	}
	c.Admin = newAdminClient(c)
	c.startFlusher()
	return c
}

// Track validates and enqueues an event. Triggers a flush when the batch is full.
func (c *CustdClient) Track(ctx context.Context, event *EventEnvelope) error {
	fillEnvelopeDefaults(event)
	if err := ValidateEvent(event); err != nil {
		return err
	}
	queueLen := c.q.enqueue(event)
	if queueLen >= c.config.BatchSize {
		return c.Flush(ctx)
	}
	return nil
}

// Flush sends all queued events to the ingest API.
// The mutex serializes flush operations. The lock is released before network
// I/O so that Track can continue to enqueue events during sends.
func (c *CustdClient) Flush(ctx context.Context) error {
	c.mu.Lock()
	batch := c.q.dequeue(c.config.BatchSize)
	c.mu.Unlock()

	if len(batch) == 0 {
		return nil
	}
	if err := c.sendBatchWithRetry(ctx, batch); err != nil {
		c.q.requeue(batch)
		return err
	}
	return nil
}

// Close flushes remaining events and shuts down the background flusher.
func (c *CustdClient) Close(ctx context.Context) error {
	var flushErr error
	c.closeOnce.Do(func() {
		close(c.done)
		c.wg.Wait()
		if c.ticker != nil {
			c.ticker.Stop()
		}
		flushErr = c.Flush(ctx)
	})
	return flushErr
}

func (c *CustdClient) sendBatchWithRetry(ctx context.Context, events []EventEnvelope) error {
	if err := c.validateTransport(); err != nil {
		return err
	}
	return withRetry(ctx, c.config.Retry, c.rng, &c.rngMu, func() error {
		return c.sendBatch(ctx, events)
	})
}

func (c *CustdClient) validateTransport() error {
	if !isSecureOrLocalHTTP(c.config.BaseURL) {
		return fmt.Errorf("custd: base URL must use https unless it targets localhost")
	}
	if c.config.TokenURL != "" && !isSecureOrLocalHTTP(c.config.TokenURL) {
		return fmt.Errorf("custd: token URL must use https unless it targets localhost")
	}
	return nil
}

func (c *CustdClient) sendBatch(ctx context.Context, events []EventEnvelope) error {
	body, err := json.Marshal(eventBatchRequest{Events: events})
	if err != nil {
		return fmt.Errorf("custd: marshal event batch: %w", err)
	}
	if c.config.HTTPClient != nil {
		return c.sendBatchViaDoer(body)
	}
	return c.sendBatchViaHTTP(ctx, body)
}

func (c *CustdClient) sendBatchViaDoer(body []byte) error {
	resp, err := c.config.HTTPClient.Do(&HTTPRequest{
		Method:  http.MethodPost,
		URL:     c.batchEndpoint(),
		Headers: c.headers(),
		Body:    body,
	})
	if err != nil {
		return fmt.Errorf("custd: request failed: %w", err)
	}
	return c.checkBatchResponse(resp.StatusCode, resp.Body)
}

func (c *CustdClient) sendBatchViaHTTP(ctx context.Context, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.batchEndpoint(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("custd: create request: %w", err)
	}
	for k, v := range c.headers() {
		req.Header.Set(k, v)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("custd: request failed: %w", err)
	}
	// nolint:errcheck // response body fully read below; a close error cannot affect the already-read batch response
	defer func() { _ = resp.Body.Close() }()
	responseBody, _ := io.ReadAll(resp.Body)
	return c.checkBatchResponse(resp.StatusCode, responseBody)
}

// checkStatus returns a retryable or non-retryable error for non-2xx status codes.
func (c *CustdClient) checkStatus(statusCode int) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}
	if isRetryableStatus(statusCode, c.retrySet) {
		return newRetryableError(statusCode)
	}
	return newNonRetryableError(statusCode)
}

func (c *CustdClient) checkBatchResponse(statusCode int, body []byte) error {
	if err := c.checkStatus(statusCode); err != nil {
		return err
	}
	var response eventBatchResponse
	if len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("custd: decode batch response: %w", err)
	}
	if !response.Success {
		return newNonRetryableError(statusCode)
	}
	return nil
}

func (c *CustdClient) batchEndpoint() string {
	return strings.TrimRight(c.config.BaseURL, "/") + ingestBatchEndpoint
}

// headers returns the default request headers.
func (c *CustdClient) headers() map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if c.config.APIKey != "" {
		headers["Authorization"] = "Bearer " + c.config.APIKey
	}
	return headers
}

// startFlusher launches a background goroutine that flushes on interval.
func (c *CustdClient) startFlusher() {
	if c.config.FlushInterval <= 0 {
		return
	}
	c.ticker = time.NewTicker(c.config.FlushInterval)
	c.wg.Add(1)
	go c.runFlusher()
}

// runFlusher is the background flush loop.
func (c *CustdClient) runFlusher() {
	defer c.wg.Done()
	for {
		select {
		case <-c.done:
			return
		case <-c.ticker.C:
			// nolint:errcheck // periodic flush; a transient error is retried next tick and surfaced by explicit Flush/Close
			_ = c.Flush(context.Background())
		}
	}
}

// applyDefaults fills in zero-value fields from DefaultClientConfig.
func applyDefaults(cfg *ClientConfig) ClientConfig {
	defaults := DefaultClientConfig()
	if cfg == nil {
		return defaults
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = defaults.BatchSize
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = defaults.FlushInterval
	}
	if cfg.MaxQueueSize <= 0 {
		cfg.MaxQueueSize = defaults.MaxQueueSize
	}
	cfg.Retry = applyRetryDefaults(&cfg.Retry, &defaults.Retry)
	return *cfg
}

func buildHTTPClient(cfg *ClientConfig) *http.Client {
	if cfg.ClientID == "" && cfg.ClientSecret == "" && cfg.TokenURL == "" {
		return &http.Client{Timeout: defaultHTTPTimeout}
	}
	oauthCfg := clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.TokenURL,
		Scopes:       cfg.Scopes,
	}
	if cfg.Audience != "" {
		oauthCfg.EndpointParams = url.Values{"audience": {cfg.Audience}}
	}
	client := oauthCfg.Client(context.Background())
	client.Timeout = defaultHTTPTimeout
	return client
}

func isSecureOrLocalHTTP(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return false
	}
	if u.Scheme == "https" {
		return true
	}
	if u.Scheme != "http" {
		return false
	}
	host := u.Hostname()
	return host == "localhost" || net.ParseIP(host).IsLoopback()
}

// applyRetryDefaults fills in zero-value retry fields.
func applyRetryDefaults(cfg, defaults *RetryConfig) RetryConfig {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = defaults.MaxAttempts
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = defaults.BaseDelay
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = defaults.MaxDelay
	}
	if cfg.Jitter <= 0 {
		cfg.Jitter = defaults.Jitter
	}
	if len(cfg.RetryOnStatuses) == 0 {
		cfg.RetryOnStatuses = defaults.RetryOnStatuses
	}
	return *cfg
}
