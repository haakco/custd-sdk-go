package custd

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// validEvent returns a fully populated event for testing.
func validEvent() *EventEnvelope {
	return &EventEnvelope{
		EventUUID:     "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		EventTypeSlug: "page_view",
		SchemaVersion: "1.0.0",
		Timestamp:     "2025-01-01T00:00:00Z",
		SessionID:     "b2c3d4e5-f6a7-8901-bcde-f12345678901",
		AnonymousID:   "c3d4e5f6-a7b8-9012-cdef-123456789012",
		CompanySlug:   "test-company",
		Context:       EventContext{Device: &DeviceContext{Type: "desktop"}},
		Payload:       json.RawMessage(`{"page":"/home"}`),
	}
}

func TestOAuthClientCredentialsAttachBearerToken(t *testing.T) {
	var tokenForm url.Values
	tokenServer := oauthTokenServer(t, &tokenForm)
	defer tokenServer.Close()

	var gotAuth string
	ingestServer := authCaptureServer(t, &gotAuth)
	defer ingestServer.Close()
	client := oauthTestClient(ingestServer.URL, tokenServer.URL)
	defer func() { _ = client.Close(context.Background()) }()

	enqueueN(client, validEvent(), 1)
	flushClient(t, client)

	if gotAuth != "Bearer minted-sdk-token" {
		t.Fatalf("Authorization = %q, want minted OAuth2 token", gotAuth)
	}
	assertOAuthTokenRequest(t, tokenForm)
}

func flushClient(t *testing.T, client *CustdClient) {
	t.Helper()
	if err := client.Flush(context.Background()); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
}

func assertOAuthTokenRequest(t *testing.T, tokenForm url.Values) {
	t.Helper()
	if tokenForm.Get("grant_type") != "client_credentials" {
		t.Fatalf("grant_type = %q, want client_credentials", tokenForm.Get("grant_type"))
	}
	if tokenForm.Get("audience") != "custd" {
		t.Fatalf("audience = %q, want custd", tokenForm.Get("audience"))
	}
	if tokenForm.Get("scope") != "events.write" {
		t.Fatalf("scope = %q, want events.write", tokenForm.Get("scope"))
	}
}

func oauthTokenServer(t *testing.T, tokenForm *url.Values) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		*tokenForm = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"access_token":"minted-sdk-token","token_type":"Bearer","expires_in":3600}`)
	}))
}

func authCaptureServer(t *testing.T, gotAuth *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*gotAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(eventBatchResponse{Success: true})
	}))
}

func oauthTestClient(baseURL, tokenURL string) *CustdClient {
	return NewClient(&ClientConfig{
		BaseURL:       baseURL,
		ClientID:      "vorrent-media-cache",
		ClientSecret:  "secret",
		TokenURL:      tokenURL,
		Audience:      "custd",
		Scopes:        []string{"events.write"},
		BatchSize:     5,
		FlushInterval: time.Hour,
	})
}

func TestOAuthRejectsPlaintextNonLocalURLs(t *testing.T) {
	client := NewClient(&ClientConfig{
		BaseURL:       "http://custd.example.com",
		ClientID:      "vorrent-media-cache",
		ClientSecret:  "secret",
		TokenURL:      "https://auth.example.com/oauth2/token",
		BatchSize:     5,
		FlushInterval: time.Hour,
	})
	defer func() { _ = client.Close(context.Background()) }()

	enqueueN(client, validEvent(), 1)
	err := client.Flush(context.Background())
	if err == nil {
		t.Fatal("expected insecure URL error")
	}
	if !strings.Contains(err.Error(), "must use https") {
		t.Fatalf("expected https error, got %v", err)
	}
}

func loadFixture(t *testing.T, name string) *EventEnvelope {
	t.Helper()
	data := readContractFixture(t, name)
	var event EventEnvelope
	if err := json.Unmarshal(data, &event); err != nil {
		t.Fatalf("decode fixture %s: %v", name, err)
	}
	return &event
}

// testClientWithServer creates a client connected to an httptest server.
func testClientWithServer(t *testing.T, handler http.HandlerFunc) (*CustdClient, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	client := NewClient(&ClientConfig{
		BaseURL:       srv.URL,
		APIKey:        "test-key",
		BatchSize:     5,
		FlushInterval: time.Hour, // disable auto-flush in tests
		Retry: RetryConfig{
			MaxAttempts: 1,
			BaseDelay:   time.Millisecond,
			MaxDelay:    time.Millisecond,
		},
	})
	return client, srv
}

func TestTrackEnqueuesEvent(t *testing.T) {
	client := NewClient(&ClientConfig{
		BaseURL:       "http://localhost",
		APIKey:        "key",
		BatchSize:     100,
		FlushInterval: time.Hour,
	})
	defer func() { _ = client.Close(context.Background()) }()

	err := client.Track(context.Background(), validEvent())
	if err != nil {
		t.Fatalf("Track returned error: %v", err)
	}
	if client.q.len() != 1 {
		t.Fatalf("expected 1 event in queue, got %d", client.q.len())
	}
}

func TestTrackGeneratesMissingEnvelopeIdentities(t *testing.T) {
	client := NewClient(&ClientConfig{
		BaseURL:       "http://localhost",
		APIKey:        "key",
		BatchSize:     100,
		FlushInterval: time.Hour,
	})
	defer func() { _ = client.Close(context.Background()) }()

	event := validEvent()
	event.EventUUID = ""
	event.SessionID = ""
	event.AnonymousID = ""

	if err := client.Track(context.Background(), event); err != nil {
		t.Fatalf("Track returned error: %v", err)
	}

	if event.EventUUID == "" || event.SessionID == "" || event.AnonymousID == "" {
		t.Fatalf("expected generated identities, got eventUuid=%q sessionId=%q anonymousId=%q",
			event.EventUUID, event.SessionID, event.AnonymousID)
	}
}

func TestNewClientNilUsesDefaults(t *testing.T) {
	client := NewClient(nil)
	defer func() { _ = client.Close(context.Background()) }()

	if client.config.BatchSize == 0 {
		t.Fatal("expected default batch size")
	}
}

func TestValidateEventNilReturnsError(t *testing.T) {
	err := ValidateEvent(nil)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "event") {
		t.Fatalf("expected clear nil event error, got %v", err)
	}
}

func TestValidateEventAcceptsCanonicalValidFixture(t *testing.T) {
	event := loadFixture(t, "valid-event.json")

	if err := ValidateEvent(event); err != nil {
		t.Fatalf("expected valid fixture to pass, got %v", err)
	}
}

func TestValidateEventRejectsCanonicalMissingDeviceTypeFixture(t *testing.T) {
	event := loadFixture(t, "invalid-missing-device-type.json")

	err := ValidateEvent(event)
	if err == nil {
		t.Fatal("expected missing device type fixture to fail")
	}
	if !strings.Contains(err.Error(), "context.device.type") {
		t.Fatalf("expected context.device.type error, got %v", err)
	}
}

func TestTrackRejectsInvalidEvent(t *testing.T) {
	client := NewClient(&ClientConfig{
		BaseURL:       "http://localhost",
		APIKey:        "key",
		BatchSize:     100,
		FlushInterval: time.Hour,
	})
	defer func() { _ = client.Close(context.Background()) }()

	err := client.Track(context.Background(), &EventEnvelope{})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestAutoFlushOnBatchSize(t *testing.T) {
	var received atomic.Int32
	client, srv := testClientWithServer(t, func(w http.ResponseWriter, _ *http.Request) {
		received.Add(1)
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()
	defer func() { _ = client.Close(context.Background()) }()

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		if err := client.Track(ctx, validEvent()); err != nil {
			t.Fatalf("Track %d failed: %v", i, err)
		}
	}

	if received.Load() != 1 {
		t.Fatalf("expected 1 batch request, got %d", received.Load())
	}
	if client.q.len() != 0 {
		t.Fatalf("expected empty queue, got %d", client.q.len())
	}
}

func TestFlushSendsOneBatchRequest(t *testing.T) {
	var requests atomic.Int32
	client, srv := testClientWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path != "/api/v1/events/batch" {
			t.Fatalf("path = %s, want /api/v1/events/batch", r.URL.Path)
		}
		var body eventBatchRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode batch request: %v", err)
		}
		if len(body.Events) != 2 {
			t.Fatalf("batch events = %d, want 2", len(body.Events))
		}
		_ = json.NewEncoder(w).Encode(eventBatchResponse{Success: true})
	})
	defer srv.Close()
	defer func() { _ = client.Close(context.Background()) }()

	enqueueN(client, validEvent(), 2)
	if err := client.Flush(context.Background()); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	if requests.Load() != 1 {
		t.Fatalf("requests = %d, want 1", requests.Load())
	}
}

func TestFlushBatchRejectionNamesFailedEvents(t *testing.T) {
	client, srv := testClientWithServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"success":false,"results":[` +
			`{"eventUuid":"evt-ok","success":true,"status":202},` +
			`{"eventUuid":"evt-bad","success":false,"status":400,` +
			`"error":{"type":"validation_failed","title":"Validation Failed","status":400,"detail":"validation failed"}}` +
			`]}`))
	})
	defer srv.Close()
	defer func() { _ = client.Close(context.Background()) }()

	enqueueN(client, validEvent(), 2)
	err := client.Flush(context.Background())
	if err == nil {
		t.Fatal("expected batch rejection error, got nil")
	}
	for _, want := range []string{"evt-bad", "400", "validation failed", "1 of 2"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q must contain %q", err.Error(), want)
		}
	}
	if strings.Contains(err.Error(), "evt-ok") {
		t.Fatalf("error must not name succeeded events: %q", err.Error())
	}
}

func TestFlushSurfacesProblemDetailOnNon2xx(t *testing.T) {
	client, srv := testClientWithServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"type":"validation_failed","title":"Validation Failed",` +
			`"status":400,"detail":"event payload invalid","code":"invalid_payload",` +
			`"fields":{"payload":"required"}}`))
	})
	defer srv.Close()
	defer func() { _ = client.Close(context.Background()) }()

	enqueueN(client, validEvent(), 1)
	err := client.Flush(context.Background())
	if err == nil {
		t.Fatal("expected problem error, got nil")
	}
	for _, want := range []string{"event payload invalid", "400", "invalid_payload", "payload=required"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q must contain %q", err.Error(), want)
		}
	}
}

func TestFlushSendsEvents(t *testing.T) {
	var received []EventEnvelope
	var mu sync.Mutex

	client, srv := testClientWithServer(t, captureEventHandler(t, &mu, &received))
	defer srv.Close()
	defer func() { _ = client.Close(context.Background()) }()

	enqueueN(client, validEvent(), 2)
	if err := client.Flush(context.Background()); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 2 {
		t.Fatalf("expected 2 events, got %d", len(received))
	}
}

func enqueueN(client *CustdClient, event *EventEnvelope, n int) {
	for i := 0; i < n; i++ {
		client.q.enqueue(event)
	}
}

func captureEventHandler(t *testing.T, mu *sync.Mutex, received *[]EventEnvelope) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/events/batch" {
			t.Errorf("expected /api/v1/events/batch path, got %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var batch eventBatchRequest
		if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
			t.Errorf("decode error: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		mu.Lock()
		*received = append(*received, batch.Events...)
		mu.Unlock()
		_ = json.NewEncoder(w).Encode(eventBatchResponse{Success: true})
	}
}

func TestRetryOnServerError(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := attempts.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(&ClientConfig{
		BaseURL:       srv.URL,
		APIKey:        "test-key",
		BatchSize:     5,
		FlushInterval: time.Hour,
		Retry: RetryConfig{
			MaxAttempts: 5,
			BaseDelay:   time.Millisecond,
			MaxDelay:    5 * time.Millisecond,
			Jitter:      0,
		},
	})
	defer func() { _ = client.Close(context.Background()) }()

	client.q.enqueue(validEvent())
	err := client.Flush(context.Background())
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if attempts.Load() != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestNonRetryableErrorStopsImmediately(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	client := NewClient(&ClientConfig{
		BaseURL:       srv.URL,
		APIKey:        "test-key",
		BatchSize:     5,
		FlushInterval: time.Hour,
		Retry: RetryConfig{
			MaxAttempts: 3,
			BaseDelay:   time.Millisecond,
			MaxDelay:    time.Millisecond,
		},
	})
	defer func() { _ = client.Close(context.Background()) }()

	client.q.enqueue(validEvent())
	err := client.Flush(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if attempts.Load() != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts.Load())
	}
}

func TestCloseFlushesRemaining(t *testing.T) {
	var received atomic.Int32
	client, srv := testClientWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		var batch eventBatchRequest
		if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
			t.Fatalf("decode batch request: %v", err)
		}
		received.Add(int32(len(batch.Events)))
		_ = json.NewEncoder(w).Encode(eventBatchResponse{Success: true})
	})
	defer srv.Close()

	client.q.enqueue(validEvent())
	client.q.enqueue(validEvent())

	err := client.Close(context.Background())
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if received.Load() != 2 {
		t.Fatalf("expected 2 events flushed on close, got %d", received.Load())
	}
}

func TestContextCancellationStopsRetry(t *testing.T) {
	var attempts atomic.Int32
	srv := newCountingFailureServer(&attempts)
	defer srv.Close()

	client := newRetryClient(srv.URL)
	defer func() { _ = client.Close(context.Background()) }()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	client.q.enqueue(validEvent())
	if err := client.Flush(ctx); err == nil {
		t.Fatal("expected error from cancelled context")
	}
	if attempts.Load() < 1 {
		t.Fatalf("expected at least 1 attempt, got %d", attempts.Load())
	}
}

func newCountingFailureServer(attempts *atomic.Int32) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
}

func newRetryClient(baseURL string) *CustdClient {
	return NewClient(&ClientConfig{
		BaseURL:       baseURL,
		APIKey:        "test-key",
		BatchSize:     5,
		FlushInterval: time.Hour,
		Retry: RetryConfig{
			MaxAttempts: 10,
			BaseDelay:   50 * time.Millisecond,
			MaxDelay:    100 * time.Millisecond,
			Jitter:      0,
		},
	})
}

func TestRequestIncludesAuthHeader(t *testing.T) {
	var authHeader string
	client, srv := testClientWithServer(t, func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()
	defer func() { _ = client.Close(context.Background()) }()

	client.q.enqueue(validEvent())
	if err := client.Flush(context.Background()); err != nil {
		t.Fatalf("Flush failed: %v", err)
	}
	expected := "Bearer test-key"
	if authHeader != expected {
		t.Fatalf("expected auth header %q, got %q", expected, authHeader)
	}
}

type validateEventCase struct {
	name    string
	event   *EventEnvelope
	wantErr bool
}

func validateEventCases() []validateEventCase {
	return []validateEventCase{
		{name: "valid event", event: validEvent(), wantErr: false},
		{name: "empty event", event: &EventEnvelope{}, wantErr: true},
		{
			name: "missing device type",
			event: &EventEnvelope{
				EventUUID:     "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
				EventTypeSlug: "page_view",
				SchemaVersion: "1.0.0",
				Timestamp:     "2025-01-01T00:00:00Z",
				SessionID:     "b2c3d4e5-f6a7-8901-bcde-f12345678901",
				AnonymousID:   "c3d4e5f6-a7b8-9012-cdef-123456789012",
				Context:       EventContext{},
				Payload:       json.RawMessage(`{"page":"/home"}`),
			},
			wantErr: true,
		},
		{
			name: "missing company slug",
			event: &EventEnvelope{
				EventUUID:     "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
				EventTypeSlug: "page_view",
				SchemaVersion: "1.0.0",
				Timestamp:     "2025-01-01T00:00:00Z",
				SessionID:     "b2c3d4e5-f6a7-8901-bcde-f12345678901",
				AnonymousID:   "c3d4e5f6-a7b8-9012-cdef-123456789012",
				Context:       EventContext{Device: &DeviceContext{Type: "desktop"}},
				Payload:       json.RawMessage(`{"page":"/home"}`),
			},
			wantErr: true,
		},
		{
			name: "missing payload",
			event: &EventEnvelope{
				EventUUID:     "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
				EventTypeSlug: "page_view",
				SchemaVersion: "1.0.0",
				Timestamp:     "2025-01-01T00:00:00Z",
				SessionID:     "b2c3d4e5-f6a7-8901-bcde-f12345678901",
				AnonymousID:   "c3d4e5f6-a7b8-9012-cdef-123456789012",
				Context:       EventContext{Device: &DeviceContext{Type: "mobile"}},
			},
			wantErr: true,
		},
	}
}

func TestValidateEventCases(t *testing.T) {
	for _, tc := range validateEventCases() {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateEvent(tc.event)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateEvent() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestQueueDropsOldestAtCapacity(t *testing.T) {
	q := newQueue(2)
	enqueueLabeledEvents(q, "first", "second", "third")

	if q.len() != 2 {
		t.Fatalf("expected queue length 2, got %d", q.len())
	}

	batch := q.dequeue(2)
	if batch[0].EventTypeSlug != "second" {
		t.Fatalf("expected oldest dropped, got %q", batch[0].EventTypeSlug)
	}
	if batch[1].EventTypeSlug != "third" {
		t.Fatalf("expected newest kept, got %q", batch[1].EventTypeSlug)
	}
}

func enqueueLabeledEvents(q *queue, labels ...string) {
	for _, label := range labels {
		e := validEvent()
		e.EventTypeSlug = label
		q.enqueue(e)
	}
}

func TestRetryStatusCodes(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		retryable bool
	}{
		{"408 retryable", 408, true},
		{"429 retryable", 429, true},
		{"500 retryable", 500, true},
		{"502 retryable", 502, true},
		{"503 retryable", 503, true},
		{"504 retryable", 504, true},
		{"400 not retryable", 400, false},
		{"401 not retryable", 401, false},
		{"403 not retryable", 403, false},
		{"404 not retryable", 404, false},
	}

	retrySet := retryableStatusSet(DefaultRetryConfig().RetryOnStatuses)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isRetryableStatus(tc.status, retrySet)
			if got != tc.retryable {
				t.Errorf("isRetryableStatus(%d) = %v, want %v", tc.status, got, tc.retryable)
			}
		})
	}
}

func TestBackoffDelayIncreasesExponentially(t *testing.T) {
	cfg := RetryConfig{
		BaseDelay: 100 * time.Millisecond,
		MaxDelay:  10 * time.Second,
		Jitter:    0,
	}

	rng := newSecureRand()
	var mu sync.Mutex

	d1 := backoffDelay(cfg, 1, rng, &mu)
	d2 := backoffDelay(cfg, 2, rng, &mu)
	d3 := backoffDelay(cfg, 3, rng, &mu)

	if d2 <= d1 {
		t.Errorf("expected d2 (%v) > d1 (%v)", d2, d1)
	}
	if d3 <= d2 {
		t.Errorf("expected d3 (%v) > d2 (%v)", d3, d2)
	}
}

func TestBackoffDelayCapsAtMax(t *testing.T) {
	cfg := RetryConfig{
		BaseDelay: 100 * time.Millisecond,
		MaxDelay:  500 * time.Millisecond,
		Jitter:    0,
	}

	rng := newSecureRand()
	var mu sync.Mutex

	d10 := backoffDelay(cfg, 10, rng, &mu)
	if d10 > cfg.MaxDelay {
		t.Errorf("expected delay <= %v, got %v", cfg.MaxDelay, d10)
	}
}

func TestCloseWaitsFlusherGoroutine(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(&ClientConfig{
		BaseURL:       srv.URL,
		APIKey:        "test-key",
		BatchSize:     100,
		FlushInterval: 10 * time.Millisecond,
	})

	// Let the flusher tick a few times
	time.Sleep(50 * time.Millisecond)

	// Close should wait for the flusher goroutine to exit
	err := client.Close(context.Background())
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// After Close, calling Close again should be safe (idempotent)
	err = client.Close(context.Background())
	if err != nil {
		t.Fatalf("second Close failed: %v", err)
	}
}
