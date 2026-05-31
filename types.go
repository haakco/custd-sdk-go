package custd

import (
	"encoding/json"
)

// PageContext describes the page from which an event originated.
type PageContext struct {
	URL      string `json:"url,omitempty"`
	Path     string `json:"path,omitempty"`
	Title    string `json:"title,omitempty"`
	Referrer string `json:"referrer,omitempty"`
}

// DeviceContext describes the device that generated the event.
type DeviceContext struct {
	Type    string `json:"type"`
	OS      string `json:"os,omitempty"`
	Browser string `json:"browser,omitempty"`
}

// EventContext holds contextual metadata for an event.
type EventContext struct {
	Page     *PageContext   `json:"page,omitempty"`
	Device   *DeviceContext `json:"device,omitempty"`
	Locale   string         `json:"locale,omitempty"`
	Timezone string         `json:"timezone,omitempty"`
	IP       string         `json:"ip,omitempty"`
}

// EventEnvelope is the event structure accepted by the ingest API.
type EventEnvelope struct {
	EventUUID     string       `json:"eventUuid"`
	EventTypeSlug string       `json:"eventTypeSlug"`
	SchemaVersion string       `json:"schemaVersion"`
	Timestamp     string       `json:"timestamp"`
	SessionID     string       `json:"sessionId"`
	AnonymousID   string       `json:"anonymousId"`
	UserUUID      string       `json:"userUuid,omitempty"`
	CompanySlug   string       `json:"companySlug,omitempty"`
	Context       EventContext `json:"context"`
	// Payload holds the event-specific data. It uses json.RawMessage to avoid
	// unnecessary deserialization/reserialization of pass-through payloads.
	Payload json.RawMessage `json:"payload"`
}

type eventBatchRequest struct {
	Events []EventEnvelope `json:"events"`
}

type eventBatchResponse struct {
	Success bool `json:"success"`
}
