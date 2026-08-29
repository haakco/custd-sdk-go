package custd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateEventAcceptsCanonicalLabelledFixture(t *testing.T) {
	var event EventEnvelope
	if err := json.Unmarshal(readContractFixture(t, "valid-labelled-event.json"), &event); err != nil {
		t.Fatal(err)
	}
	if err := ValidateEvent(&event); err != nil {
		t.Fatalf("ValidateEvent() error = %v", err)
	}
}

func TestValidateEventRejectsInvalidLabels(t *testing.T) {
	event := validLabelTestEvent()
	cases := []map[string]string{
		manyLabels(17), {"App Plan": "paid"}, {"a." + strings.Repeat("b", 63): "paid"},
		{"custd.internal": "reserved"},
		{"app.plan": ""}, {"app.plan": " paid "}, {"app.plan": strings.Repeat("é", 65)},
	}
	for _, labels := range cases {
		event.Labels = labels
		if err := ValidateEvent(&event); err == nil || !strings.Contains(err.Error(), "labels") {
			t.Fatalf("ValidateEvent(%v) error = %v, want labels error", labels, err)
		}
	}
}

func validLabelTestEvent() EventEnvelope {
	return EventEnvelope{EventUUID: "evt", EventTypeSlug: "page-view", SchemaVersion: "1.0.0", Timestamp: "2026-01-01T00:00:00Z", SessionID: "session", AnonymousID: "anonymous", CompanySlug: "test", Context: EventContext{Device: &DeviceContext{Type: "desktop"}}, Payload: json.RawMessage(`{"ok":true}`)}
}

func manyLabels(count int) map[string]string {
	labels := make(map[string]string, count)
	for i := range count {
		labels["app.key"+string(rune('a'+i))] = "ok"
	}
	return labels
}
