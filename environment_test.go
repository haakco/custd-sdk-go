package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// TestEnvironmentDeclarationIsCarriedAsReservedLabel proves the whole client
// path: the field is validated locally, the client default applies when the
// event declares nothing, and the envelope on the wire carries the reserved
// label rather than a field of its own.
func TestEnvironmentDeclarationIsCarriedAsReservedLabel(t *testing.T) {
	capture := newCaptureDoer(http.StatusAccepted, `{"success":true}`)
	client := NewClient(&ClientConfig{
		BaseURL: "http://localhost:8080", APIKey: "token", Environment: "development",
		BatchSize: 1, HTTPClient: capture,
	})
	event := &EventEnvelope{
		CompanySlug: "acme", EventTypeSlug: "page-view", SchemaVersion: "1.0.0",
		Timestamp: "2026-09-12T10:00:00Z", Payload: json.RawMessage(`{}`),
		Context: EventContext{Device: &DeviceContext{Type: "desktop"}},
	}
	if err := client.Track(context.Background(), event); err != nil {
		t.Fatalf("track with client default: %v", err)
	}
	if got := event.Labels[EnvironmentLabelKey]; got != "development" {
		t.Fatalf("client default label = %q, want development", got)
	}
	if got := sentLabel(t, capture.requests[0].Body); got != "development" {
		t.Fatalf("wire label = %q, want development", got)
	}

	// An event-level declaration wins over the process default.
	override := &EventEnvelope{
		CompanySlug: "acme", EventTypeSlug: "page-view", SchemaVersion: "1.0.0",
		Timestamp: "2026-09-12T10:00:00Z", Payload: json.RawMessage(`{}`),
		Environment: "preview-pr-9",
		Context:     EventContext{Device: &DeviceContext{Type: "desktop"}},
	}
	if err := client.Track(context.Background(), override); err != nil {
		t.Fatalf("track with event override: %v", err)
	}
	if got := override.Labels[EnvironmentLabelKey]; got != "preview-pr-9" {
		t.Fatalf("event label = %q, want preview-pr-9", got)
	}
}

func TestEnvironmentDeclarationIsValidatedLocally(t *testing.T) {
	cases := map[string]string{
		"reserved value": UnclassifiedEnvironment,
		"uppercase":      "Production",
		"leading digit":  "1st-env",
		"underscore":     "staging_env",
		"too long":       "a23456789012345678901234567890123",
	}
	for name, value := range cases {
		t.Run(name, func(t *testing.T) {
			err := validateEnvironmentValue(value)
			if err == nil {
				t.Fatalf("environment %q was accepted", value)
			}
		})
	}
	if err := validateEnvironmentValue("preview-pr-9"); err != nil {
		t.Fatalf("canonical value rejected: %v", err)
	}
}

// TestReservedLabelKeysStayRejected keeps one obvious way to declare an
// environment: the field, not a hand-written reserved label.
func TestReservedLabelKeysStayRejected(t *testing.T) {
	err := validateEventLabels(map[string]string{EnvironmentLabelKey: "production"})
	if err == nil {
		t.Fatal("a hand-written custd.environment label was accepted")
	}
}

func sentLabel(t *testing.T, body []byte) string {
	t.Helper()
	var batch eventBatchRequest
	if err := json.Unmarshal(body, &batch); err != nil {
		t.Fatalf("decode batch: %v", err)
	}
	if len(batch.Events) != 1 {
		t.Fatalf("batch size = %d, want 1", len(batch.Events))
	}
	return batch.Events[0].Labels[EnvironmentLabelKey]
}
