package custd

import (
	"time"
)

// RetryConfig controls retry and backoff behavior.
type RetryConfig struct {
	MaxAttempts     int
	BaseDelay       time.Duration
	MaxDelay        time.Duration
	Jitter          float64
	RetryOnStatuses []int
}

// ClientConfig holds all configuration for a CustdClient.
type ClientConfig struct {
	BaseURL       string
	APIKey        string
	ClientID      string
	ClientSecret  string
	TokenURL      string
	Audience      string
	Scopes        []string
	BatchSize     int
	FlushInterval time.Duration
	MaxRetries    int
	MaxQueueSize  int
	Retry         RetryConfig
	HTTPClient    HTTPDoer

	// CompressionEnabled gzip-compresses the batch request body when it meets
	// CompressionThreshold. Nil defaults to enabled; set to a pointer to false
	// to disable. The ingest API decodes Content-Encoding: gzip.
	CompressionEnabled *bool
	// CompressionThreshold is the minimum serialized body size in bytes before
	// gzip compression is applied. Zero falls back to the default.
	CompressionThreshold int
}

// HTTPDoer abstracts HTTP requests for testing.
type HTTPDoer interface {
	Do(req *HTTPRequest) (*HTTPResponse, error)
}

// HTTPRequest is a simplified HTTP request for the doer interface.
type HTTPRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
}

// HTTPResponse is a simplified HTTP response from the doer interface.
type HTTPResponse struct {
	StatusCode int
	Body       []byte
}

// DefaultRetryConfig returns sensible retry defaults.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     3,
		BaseDelay:       200 * time.Millisecond,
		MaxDelay:        2 * time.Second,
		Jitter:          0.2,
		RetryOnStatuses: []int{408, 429, 500, 502, 503, 504},
	}
}

// defaultCompressionThreshold is the body size in bytes at or above which the
// batch request body is gzip-compressed.
const defaultCompressionThreshold = 1024

// DefaultClientConfig returns sensible client defaults.
func DefaultClientConfig() ClientConfig {
	enabled := true
	return ClientConfig{
		BatchSize:            20,
		FlushInterval:        30 * time.Second,
		MaxQueueSize:         1000,
		Retry:                DefaultRetryConfig(),
		CompressionEnabled:   &enabled,
		CompressionThreshold: defaultCompressionThreshold,
	}
}
