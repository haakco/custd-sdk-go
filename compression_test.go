package custd

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// compressionTestEvent returns a payload large enough to exceed the default
// compression threshold so the gzip path is exercised.
func compressionTestEvent(t *testing.T) *EventEnvelope {
	t.Helper()
	event := validEvent()
	event.Payload = json.RawMessage(`{"page":"` + repeatStr("x", 2048) + `"}`)
	return event
}

func repeatStr(s string, n int) string {
	b := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		b = append(b, s...)
	}
	return string(b)
}

func newCompressionClient(t *testing.T, doer *captureDoer, cfg *ClientConfig) *CustdClient {
	t.Helper()
	cfg.BaseURL = "http://localhost:8080/"
	cfg.HTTPClient = doer
	cfg.FlushInterval = -1
	client := NewClient(cfg)
	t.Cleanup(func() { _ = client.Close(context.Background()) })
	return client
}

func gunzip(t *testing.T, data []byte) []byte {
	t.Helper()
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("gunzip read: %v", err)
	}
	return out
}

// rawBatchBody returns the JSON the SDK would have sent for the given event,
// for byte-exact comparison against the decompressed request body.
func rawBatchBody(t *testing.T, event *EventEnvelope) []byte {
	t.Helper()
	fillEnvelopeDefaults(event)
	body, err := json.Marshal(eventBatchRequest{Events: []EventEnvelope{*event}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return body
}

func TestSendBatchCompressesBodyOverThreshold(t *testing.T) {
	doer := newCaptureDoer(http.StatusAccepted, `{"success":true}`)
	client := newCompressionClient(t, doer, &ClientConfig{})

	event := compressionTestEvent(t)
	expected := rawBatchBody(t, compressionTestEvent(t))

	if err := client.Track(context.Background(), event); err != nil {
		t.Fatalf("Track: %v", err)
	}
	if err := client.Flush(context.Background()); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	req := doer.requests[0]
	if req.Headers["Content-Encoding"] != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", req.Headers["Content-Encoding"])
	}
	if req.Headers["Content-Type"] != "application/json" {
		t.Fatalf("Content-Type = %q", req.Headers["Content-Type"])
	}
	if got := gunzip(t, req.Body); !bytes.Equal(got, expected) {
		t.Fatalf("decompressed body mismatch:\n got %s\nwant %s", got, expected)
	}
}

func TestSendBatchSkipsCompressionUnderThreshold(t *testing.T) {
	doer := newCaptureDoer(http.StatusAccepted, `{"success":true}`)
	client := newCompressionClient(t, doer, &ClientConfig{})

	event := validEvent()
	expected := rawBatchBody(t, validEvent())

	if err := client.Track(context.Background(), event); err != nil {
		t.Fatalf("Track: %v", err)
	}
	if err := client.Flush(context.Background()); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	req := doer.requests[0]
	if _, ok := req.Headers["Content-Encoding"]; ok {
		t.Fatalf("Content-Encoding present for small body")
	}
	if !bytes.Equal(req.Body, expected) {
		t.Fatalf("body mismatch:\n got %s\nwant %s", req.Body, expected)
	}
}

func TestSendBatchDisabledCompressionNeverCompresses(t *testing.T) {
	doer := newCaptureDoer(http.StatusAccepted, `{"success":true}`)
	disabled := false
	client := newCompressionClient(t, doer, &ClientConfig{CompressionEnabled: &disabled})

	event := compressionTestEvent(t)
	expected := rawBatchBody(t, compressionTestEvent(t))

	if err := client.Track(context.Background(), event); err != nil {
		t.Fatalf("Track: %v", err)
	}
	if err := client.Flush(context.Background()); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	req := doer.requests[0]
	if _, ok := req.Headers["Content-Encoding"]; ok {
		t.Fatalf("Content-Encoding present when compression disabled")
	}
	if !bytes.Equal(req.Body, expected) {
		t.Fatalf("body mismatch:\n got %s\nwant %s", req.Body, expected)
	}
}
