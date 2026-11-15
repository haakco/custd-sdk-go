package custd

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"
)

type ErrorCaptureMode string

const (
	ErrorCaptureHandled   ErrorCaptureMode = "handled"
	ErrorCaptureUnhandled ErrorCaptureMode = "unhandled"
)

type ErrorPlatform string

const (
	ErrorPlatformJavaScript ErrorPlatform = "javascript"
	ErrorPlatformGo         ErrorPlatform = "go"
)

type ErrorCaptureOutcome string

const (
	ErrorCaptureAccepted    ErrorCaptureOutcome = "accepted"
	ErrorCaptureRejected    ErrorCaptureOutcome = "rejected"
	ErrorCaptureUnavailable ErrorCaptureOutcome = "unavailable"
	ErrorCaptureLimited     ErrorCaptureOutcome = "limited"
)

// ErrorCaptureFrame contains code locations only, never an absolute path or source content.
type ErrorCaptureFrame struct {
	Filename string `json:"filename"`
	Lineno   int    `json:"lineno"`
	Colno    int    `json:"colno,omitempty"`
}

// ErrorCaptureEvent is the allowlisted mobile-relay contract. Decode with unknown fields refused at the HTTP boundary.
type ErrorCaptureEvent struct {
	EventID       string              `json:"eventId,omitempty"`
	Platform      ErrorPlatform       `json:"platform"`
	ExceptionType string              `json:"exceptionType"`
	Frames        []ErrorCaptureFrame `json:"frames"`
	Mode          ErrorCaptureMode    `json:"mode"`
	Environment   string              `json:"environment"`
	Release       string              `json:"release,omitempty"`
	CorrelationID string              `json:"correlationId,omitempty"`
}

// CaptureExceptionRequest never serializes Error; only its type and the current code stack are inspected.
type CaptureExceptionRequest struct {
	Error         any
	Mode          ErrorCaptureMode
	CorrelationID string
}
type ErrorCaptureResult struct {
	Outcome ErrorCaptureOutcome `json:"outcome"`
	Status  int                 `json:"status,omitempty"`
}
type ErrorCaptureConfig struct {
	BaseURL     string
	WriteKey    string
	Environment string
	Release     string
	HTTPClient  *http.Client
}

var (
	captureKeyPattern         = regexp.MustCompile(`^site_pk_[A-Za-z0-9_-]{32,128}$`)
	captureEnvironmentPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,64}$`)
	captureReleasePattern     = regexp.MustCompile(`^[a-zA-Z0-9@._+-]{1,200}$`)
	captureTypePattern        = regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$.]{0,127}$`)
	captureFilenamePattern    = regexp.MustCompile(`^[A-Za-z0-9_.+\[\]-]{1,128}\.(go|[cm]?js|jsx|tsx?|bundle)$`)
	captureIDPattern          = regexp.MustCompile(`^[a-f0-9]{32}$`)
	captureCorrelationPattern = regexp.MustCompile(`^[0-7][0-9A-HJKMNP-TV-Z]{25}$`)
)

// ErrorCaptureClient is request-scoped, bounded to 20 captures and four concurrent sends. It never retries.
type ErrorCaptureClient struct {
	config     ErrorCaptureConfig
	httpClient *http.Client
	mu         sync.Mutex
	captures   int
	concurrent int
}

func NewErrorCaptureClient(config ErrorCaptureConfig) (*ErrorCaptureClient, error) {
	parsed, err := url.Parse(config.BaseURL)
	if err != nil || parsed == nil {
		return nil, fmt.Errorf("custd: invalid error capture configuration")
	}
	local := parsed.Hostname() == "localhost" || parsed.Hostname() == "127.0.0.1" || parsed.Hostname() == "::1" || parsed.Hostname() == "host.docker.internal"
	secureOrLocal := parsed.Scheme == "https" || (parsed.Scheme == "http" && local)
	if parsed.Host == "" || !secureOrLocal || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") || !captureKeyPattern.MatchString(config.WriteKey) || !captureEnvironmentPattern.MatchString(config.Environment) || (config.Release != "" && !captureReleasePattern.MatchString(config.Release)) {
		return nil, fmt.Errorf("custd: invalid error capture configuration")
	}
	config.BaseURL = parsed.Scheme + "://" + parsed.Host
	transport := http.Client{Timeout: 2 * time.Second}
	if config.HTTPClient != nil {
		transport = *config.HTTPClient
	}
	transport.Jar = nil
	transport.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	if transport.Timeout <= 0 || transport.Timeout > 2*time.Second {
		transport.Timeout = 2 * time.Second
	}
	return &ErrorCaptureClient{config: config, httpClient: &transport}, nil
}
func (c *ErrorCaptureClient) CaptureException(ctx context.Context, request CaptureExceptionRequest) ErrorCaptureResult {
	event := SnapshotError(request)
	event.Environment = c.config.Environment
	event.Release = c.config.Release
	return c.CaptureEvent(ctx, event)
}

// CaptureEvent revalidates relay metadata and builds the collector envelope with this client's private Site key.
func (c *ErrorCaptureClient) CaptureEvent(ctx context.Context, event ErrorCaptureEvent) ErrorCaptureResult {
	if !validErrorCaptureEvent(event) {
		return ErrorCaptureResult{Outcome: ErrorCaptureRejected}
	}
	c.mu.Lock()
	if c.captures >= 20 || c.concurrent >= 4 {
		c.mu.Unlock()
		return ErrorCaptureResult{Outcome: ErrorCaptureLimited}
	}
	c.captures++
	c.concurrent++
	c.mu.Unlock()
	defer func() { c.mu.Lock(); c.concurrent--; c.mu.Unlock() }()
	body, err := c.envelope(event)
	if err != nil {
		return ErrorCaptureResult{Outcome: ErrorCaptureUnavailable}
	}
	if len(body) > 16384 {
		return ErrorCaptureResult{Outcome: ErrorCaptureLimited}
	}
	bounded, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(bounded, http.MethodPost, c.config.BaseURL+"/api/v1/errors/api/1/envelope/", bytes.NewReader(body))
	if err != nil {
		return ErrorCaptureResult{Outcome: ErrorCaptureUnavailable}
	}
	req.Header.Set("Content-Type", "text/plain")
	response, err := c.httpClient.Do(req)
	if err != nil {
		return ErrorCaptureResult{Outcome: ErrorCaptureUnavailable}
	}
	defer func() { _ = response.Body.Close() }()
	outcome := ErrorCaptureRejected
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		outcome = ErrorCaptureAccepted
	}
	return ErrorCaptureResult{Outcome: outcome, Status: response.StatusCode}
}
func validErrorCaptureEvent(event ErrorCaptureEvent) bool {
	if (event.Platform != ErrorPlatformGo && event.Platform != ErrorPlatformJavaScript) || !captureTypePattern.MatchString(event.ExceptionType) || (event.Mode != ErrorCaptureHandled && event.Mode != ErrorCaptureUnhandled) || !captureEnvironmentPattern.MatchString(event.Environment) || (event.Release != "" && !captureReleasePattern.MatchString(event.Release)) || (event.EventID != "" && !captureIDPattern.MatchString(event.EventID)) || (event.CorrelationID != "" && !captureCorrelationPattern.MatchString(event.CorrelationID)) || len(event.Frames) > 24 {
		return false
	}
	for _, frame := range event.Frames {
		if !captureFilenamePattern.MatchString(frame.Filename) || frame.Lineno < 1 || frame.Lineno > 1000000 || frame.Colno < 0 || frame.Colno > math.MaxInt32 {
			return false
		}
	}
	return true
}
func (c *ErrorCaptureClient) envelope(event ErrorCaptureEvent) ([]byte, error) {
	id := event.EventID
	if id == "" {
		var random [16]byte
		if _, err := rand.Read(random[:]); err != nil {
			return nil, err
		}
		id = hex.EncodeToString(random[:])
	}
	origin, _ := url.Parse(c.config.BaseURL)
	origin.User = url.User(c.config.WriteKey)
	origin.Path = "/api/v1/errors/1"
	header, err := json.Marshal(struct {
		ID  string `json:"event_id"`
		DSN string `json:"dsn"`
	}{id, origin.String()})
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(newErrorEnvelopeEvent(id, event))
	if err != nil {
		return nil, err
	}
	return bytes.Join([][]byte{header, []byte(`{"type":"event"}`), encoded}, []byte("\n")), nil
}
