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

// eventResult is the per-event outcome the ingest API returns in a batch
// response. On a partial failure the top-level success is false and the
// rejected events are identifiable here. Error carries the RFC 9457 problem
// detail for a rejected event and is absent for accepted events.
type eventResult struct {
	EventUUID string   `json:"eventUuid"`
	Success   bool     `json:"success"`
	Status    int      `json:"status"`
	Error     *Problem `json:"error,omitempty"`
}

type eventBatchResponse struct {
	Success bool          `json:"success"`
	Results []eventResult `json:"results"`
}

type DataSpaceCreate struct {
	Slug        string `json:"slug"`
	CompanyName string `json:"companyName"`
}

type DataSpace struct {
	Slug              string `json:"slug"`
	CompanyName       string `json:"companyName"`
	ParentCompanySlug string `json:"parentCompanySlug"`
	Enabled           bool   `json:"enabled"`
}

type DataSpaceEntitlementState struct {
	Enabled                        bool `json:"enabled"`
	ActiveDataSpaces               int  `json:"activeDataSpaces"`
	MaxActiveDataSpaces            int  `json:"maxActiveDataSpaces"`
	MaxActiveProducersPerDataSpace int  `json:"maxActiveProducersPerDataSpace"`
}

type DataSpaceList struct {
	DataSpaces  []DataSpace               `json:"dataSpaces"`
	Entitlement DataSpaceEntitlementState `json:"entitlement"`
}

type ProducerProvisionRequest struct {
	CompanySlug   string            `json:"companySlug"`
	ProducerSlug  string            `json:"producerSlug"`
	DisplayName   string            `json:"displayName,omitempty"`
	Environment   string            `json:"environment,omitempty"`
	Scopes        []string          `json:"scopes,omitempty"`
	ScopeTemplate string            `json:"scopeTemplate,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

type ProducerProvisionResponse struct {
	BaseURL      string            `json:"baseUrl,omitempty"`
	TokenURL     string            `json:"tokenUrl,omitempty"`
	Audience     string            `json:"audience,omitempty"`
	ClientID     string            `json:"clientId"`
	ClientSecret string            `json:"clientSecret"`
	CompanySlug  string            `json:"companySlug"`
	ProducerSlug string            `json:"producerSlug"`
	Scopes       []string          `json:"scopes"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type ProducerProvisionPublicClient struct {
	ClientID     string            `json:"clientId"`
	CompanySlug  string            `json:"companySlug"`
	ProducerSlug string            `json:"producerSlug"`
	Scopes       []string          `json:"scopes"`
	Environment  string            `json:"environment,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}
