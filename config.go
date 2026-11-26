package custd

import (
	"net/http"
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
	BaseURL string

	// ActingCompanySlug names the client company a system administrator's admin requests
	// target, sent as the `companySlug` query parameter the admin routes read and
	// validate. It is ordinary tenant selection: it never requests grant-backed support
	// context, so a caller using it still needs only the authority it already has.
	// Ordinary tenant commands leave it empty and the server uses the credential's own
	// company. The one exception is ingest-api's cleanup jobs, which read the tenant only
	// from `X-Custd-Effective-Company`: those calls send it there, and the server then
	// requires support access and an active grant for that company, even the caller's own.
	ActingCompanySlug string

	// CommandLabel names the operation the caller is performing, sent as
	// `X-Custd-Command`. The server uses it, together with `X-Custd-Client`, to label an
	// operator action's audit row with the command that caused it. Empty leaves the
	// header off, so a non-CLI caller never claims to be one.
	CommandLabel string

	// ClientLabel identifies the calling program, sent as `X-Custd-Client`. The server
	// only credits a command label from a client it recognises.
	ClientLabel string

	APIKey       string
	ClientID     string
	ClientSecret string
	TokenURL     string
	Audience     string
	Scopes       []string
	// Environment is the environment this process sends from. It is the
	// authenticated default for events that do not declare their own; a
	// credential is expected to carry every environment unless an operator
	// deliberately restricts it.
	Environment   string
	BatchSize     int
	FlushInterval time.Duration
	MaxRetries    int
	MaxQueueSize  int
	Retry         RetryConfig
	// HTTPClient sends every Custd request in place of the SDK's built-in client, so it owns
	// timeouts, redirects and any credential beyond APIKey: the OAuth client-credentials
	// fields are not used with it. A *http.Client satisfies it. It answers with the unread
	// response, so a download streams through it.
	HTTPClient HTTPDoer

	// CompressionEnabled gzip-compresses the batch request body when it meets
	// CompressionThreshold. Nil defaults to enabled; set to a pointer to false
	// to disable. The ingest API decodes Content-Encoding: gzip.
	CompressionEnabled *bool
	// CompressionThreshold is the minimum serialized body size in bytes before
	// gzip compression is applied. Zero falls back to the default.
	CompressionThreshold int
}

// HTTPDoer sends one HTTP request and returns the unread response: the shape *http.Client has.
// The request carries the caller's context, and the SDK closes the response body.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
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
