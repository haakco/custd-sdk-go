package custd

import (
	"bytes"
	"compress/gzip"
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
	body, gzipped, err := c.maybeCompress(body)
	if err != nil {
		return err
	}
	if c.config.HTTPClient != nil {
		return c.sendBatchViaDoer(body, gzipped, len(events))
	}
	return c.sendBatchViaHTTP(ctx, body, gzipped, len(events))
}

// maybeCompress gzip-compresses the body when compression is enabled and the
// body meets the configured threshold. It returns the body to send and whether
// it was compressed.
func (c *CustdClient) maybeCompress(body []byte) ([]byte, bool, error) {
	enabled := c.config.CompressionEnabled != nil && *c.config.CompressionEnabled
	if !enabled || len(body) < c.config.CompressionThreshold {
		return body, false, nil
	}
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(body); err != nil {
		return nil, false, fmt.Errorf("custd: gzip batch body: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, false, fmt.Errorf("custd: gzip batch body: %w", err)
	}
	return buf.Bytes(), true, nil
}

func (c *CustdClient) sendBatchViaDoer(body []byte, gzipped bool, eventCount int) error {
	resp, err := c.config.HTTPClient.Do(&HTTPRequest{
		Method:  http.MethodPost,
		URL:     c.batchEndpoint(),
		Headers: c.headers(gzipped),
		Body:    body,
	})
	if err != nil {
		return newRetryableTransportError(err)
	}
	return c.checkBatchResponse(resp.StatusCode, resp.Body, eventCount)
}

func (c *CustdClient) sendBatchViaHTTP(ctx context.Context, body []byte, gzipped bool, eventCount int) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.batchEndpoint(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("custd: create request: %w", err)
	}
	for k, v := range c.headers(gzipped) {
		req.Header.Set(k, v)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return newRetryableTransportError(err)
	}
	// nolint:errcheck // response body fully read below; a close error cannot affect the already-read batch response
	defer func() { _ = resp.Body.Close() }()
	responseBody, _ := io.ReadAll(resp.Body)
	return c.checkBatchResponse(resp.StatusCode, responseBody, eventCount)
}

// checkStatus returns a retryable or non-retryable error for non-2xx status
// codes. It decodes the RFC 9457 problem+json body when present so the
// surfaced error carries the server's type, title, detail, code, and fields.
func (c *CustdClient) checkStatus(statusCode int, body []byte) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}
	retryable := isRetryableStatus(statusCode, c.retrySet)
	if problem := parseProblem(body); problem != nil {
		return newProblemError(statusCode, retryable, problem)
	}
	if retryable {
		return newRetryableError(statusCode)
	}
	return newNonRetryableError(statusCode)
}

func (c *CustdClient) checkBatchResponse(statusCode int, body []byte, sentEventCount int) error {
	if err := c.checkStatus(statusCode, body); err != nil {
		return err
	}
	var response eventBatchResponse
	if len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("custd: decode batch response: %w", err)
	}
	if len(response.Results) != 0 && len(response.Results) != sentEventCount {
		return &sendError{
			StatusCode: statusCode,
			Message: fmt.Sprintf(
				"custd: batch response result count %d does not match sent event count %d",
				len(response.Results), sentEventCount,
			),
			Retryable: false,
		}
	}
	for i, result := range response.Results {
		if result.EventUUID == "" {
			return &sendError{
				StatusCode: statusCode,
				Message:    fmt.Sprintf("custd: batch response result %d missing eventUuid", i),
				Retryable:  false,
			}
		}
	}
	if !response.Success || hasFailedBatchResult(response.Results) {
		return newBatchRejectionError(statusCode, response.Results)
	}
	return nil
}

func hasFailedBatchResult(results []eventResult) bool {
	for _, result := range results {
		if !result.Success {
			return true
		}
	}
	return false
}

// newBatchRejectionError builds a non-retryable error that names the rejected
// events (uuid, status, reason) so a partial batch failure is diagnosable
// without re-probing the API. The list is capped to keep the message bounded.
func newBatchRejectionError(statusCode int, results []eventResult) *sendError {
	failed := make([]eventResult, 0, len(results))
	for _, r := range results {
		if !r.Success {
			failed = append(failed, r)
		}
	}
	if len(failed) == 0 {
		return &sendError{
			StatusCode: statusCode,
			Message:    fmt.Sprintf("custd: batch request failed with status %d (no per-event results)", statusCode),
			Retryable:  false,
		}
	}
	const maxList = 10
	parts := make([]string, 0, maxList+1)
	for i, r := range failed {
		if i >= maxList {
			break
		}
		parts = append(parts, fmt.Sprintf("%s [status %d] %s", r.EventUUID, r.Status, eventResultDetail(r)))
	}
	if len(failed) > maxList {
		parts = append(parts, fmt.Sprintf("+%d more", len(failed)-maxList))
	}
	return &sendError{
		StatusCode: statusCode,
		Message: fmt.Sprintf("custd: batch rejected %d of %d event(s): %s",
			len(failed), len(results), strings.Join(parts, "; ")),
		Retryable: false,
	}
}

// eventResultDetail renders the problem detail for a rejected per-event result,
// falling back to the problem title or a placeholder when no message is present.
func eventResultDetail(r eventResult) string {
	if r.Error == nil {
		return "no error detail"
	}
	if r.Error.Detail != "" {
		return r.Error.Detail
	}
	if r.Error.Title != "" {
		return r.Error.Title
	}
	if r.Error.Code != "" {
		return "code " + r.Error.Code
	}
	return "no error detail"
}

func (c *CustdClient) batchEndpoint() string {
	return strings.TrimRight(c.config.BaseURL, "/") + ingestBatchEndpoint
}

// headers returns the default request headers. When gzipped is true it sets
// Content-Encoding: gzip to signal a compressed body to the ingest API.
func (c *CustdClient) headers(gzipped bool) map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if gzipped {
		headers["Content-Encoding"] = "gzip"
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
	if cfg.CompressionEnabled == nil {
		cfg.CompressionEnabled = defaults.CompressionEnabled
	}
	if cfg.CompressionThreshold <= 0 {
		cfg.CompressionThreshold = defaults.CompressionThreshold
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
