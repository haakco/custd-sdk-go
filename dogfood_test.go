package custd

import (
	"encoding/json"
	"testing"
)

func TestNewDogfoodEventBuildsValidEnvelopeWithCommonPayload(t *testing.T) {
	event, err := NewDogfoodEvent(&DogfoodEventInput{
		EventTypeSlug: "ingest-accepted",
		SchemaVersion: "1.0.0",
		CompanySlug:   "haakco",
		SourceSystem:  "custd",
		SourceCompany: "haakco",
		Environment:   "test",
		CorrelationID: "corr-123",
		Payload: map[string]any{
			"service": "ingest-api",
			"status":  "accepted",
		},
	})
	if err != nil {
		t.Fatalf("NewDogfoodEvent returned error: %v", err)
	}
	if err := ValidateEvent(event); err != nil {
		t.Fatalf("expected dogfood event to validate: %v", err)
	}
	if event.EventTypeSlug != "ingest-accepted" {
		t.Fatalf("expected event type slug, got %q", event.EventTypeSlug)
	}
	if event.SchemaVersion != "1.0.0" {
		t.Fatalf("expected schema version, got %q", event.SchemaVersion)
	}
	if event.Context.Device == nil || event.Context.Device.Type == "" {
		t.Fatal("expected minimal device context")
	}

	assertDogfoodPayload(t, decodeDogfoodPayload(t, event), map[string]string{
		"sourceSystem":  "custd",
		"sourceCompany": "haakco",
		"environment":   "test",
		"schemaVersion": "1.0.0",
		"correlationId": "corr-123",
		"service":       "ingest-api",
		"status":        "accepted",
	})
}

func TestNewDogfoodEventProtectsCommonPayloadFields(t *testing.T) {
	event, err := NewDogfoodEvent(&DogfoodEventInput{
		EventTypeSlug: "metric-snapshot",
		SchemaVersion: "1.0.0",
		CompanySlug:   "haakco",
		SourceSystem:  "custd",
		SourceCompany: "haakco",
		Environment:   "prod",
		CorrelationID: "corr-456",
		Payload: map[string]any{
			"sourceSystem":  "vorrent",
			"sourceCompany": "other",
			"environment":   "dev",
			"schemaVersion": "9.9.9",
			"correlationId": "wrong",
			"metricName":    "ingest_requests_total",
			"token":         "secret",
			"signedUrl":     "https://example.invalid/signed",
			"rawPrompt":     "prompt",
			"authorization": "bearer token",
			"nested": map[string]any{
				"apiKey": "secret",
				"safe":   "value",
			},
		},
	})
	if err != nil {
		t.Fatalf("NewDogfoodEvent returned error: %v", err)
	}

	payload := decodeDogfoodPayload(t, event)
	assertPayloadValue(t, payload, "sourceSystem", "custd")
	assertPayloadValue(t, payload, "sourceCompany", "haakco")
	assertPayloadValue(t, payload, "environment", "prod")
	assertPayloadValue(t, payload, "schemaVersion", "1.0.0")
	assertPayloadValue(t, payload, "correlationId", "corr-456")
	assertPayloadValue(t, payload, "metricName", "ingest_requests_total")
	assertPayloadMissingKeys(t, payload, "token", "signedUrl", "rawPrompt", "authorization")
	nested := payload["nested"].(map[string]any)
	assertPayloadMissing(t, nested, "apiKey")
	assertPayloadValue(t, nested, "safe", "value")
}

func TestNewDogfoodEventRejectsMissingCommonFields(t *testing.T) {
	_, err := NewDogfoodEvent(&DogfoodEventInput{
		EventTypeSlug: "metric-snapshot",
		SchemaVersion: "1.0.0",
		CompanySlug:   "haakco",
		SourceSystem:  "custd",
		Environment:   "prod",
	})
	if err == nil {
		t.Fatal("expected missing source company to fail")
	}
	if got := err.Error(); got != "custd: missing dogfood fields: sourceCompany" {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestNewDogfoodEventOmitsEmptyCorrelationID(t *testing.T) {
	event, err := NewDogfoodEvent(&DogfoodEventInput{
		EventTypeSlug: "metric-snapshot",
		SchemaVersion: "1.0.0",
		CompanySlug:   "haakco",
		SourceSystem:  "custd",
		SourceCompany: "haakco",
		Environment:   "prod",
	})
	if err != nil {
		t.Fatalf("NewDogfoodEvent returned error: %v", err)
	}
	assertPayloadMissing(t, decodeDogfoodPayload(t, event), "correlationId")
}

func decodeDogfoodPayload(t *testing.T, event *EventEnvelope) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	return payload
}

func assertPayloadValue(t *testing.T, payload map[string]any, key string, want string) {
	t.Helper()
	if got, ok := payload[key].(string); !ok || got != want {
		t.Fatalf("expected payload[%q] = %q, got %#v", key, want, payload[key])
	}
}

func assertDogfoodPayload(t *testing.T, payload map[string]any, values map[string]string) {
	t.Helper()
	for key, want := range values {
		assertPayloadValue(t, payload, key, want)
	}
}

func assertPayloadMissing(t *testing.T, payload map[string]any, key string) {
	t.Helper()
	if _, ok := payload[key]; ok {
		t.Fatalf("expected payload[%q] to be removed, got %#v", key, payload[key])
	}
}

func assertPayloadMissingKeys(t *testing.T, payload map[string]any, keys ...string) {
	t.Helper()
	for _, key := range keys {
		assertPayloadMissing(t, payload, key)
	}
}
