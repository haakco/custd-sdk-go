package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// schemaRegistryEndpoint is the tenant's event schema collection. The schema
// registry serves it behind the authenticated tenant, beside the parser routes,
// so the credential names the tenant and the caller never supplies a slug.
const schemaRegistryEndpoint = "/api/v1/schemas"

// SchemaRegistryClient registers, reads and retires the tenant's event schemas.
type SchemaRegistryClient struct {
	registry *ParserClient
}

func newSchemaRegistryClient(client *CustdClient) *SchemaRegistryClient {
	return &SchemaRegistryClient{registry: newParserClient(client)}
}

// EventType is one event type the tenant can send.
type EventType struct {
	Slug    string `json:"slug"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// EventTypeList lists the tenant's event types.
type EventTypeList struct {
	EventTypes []EventType `json:"eventTypes"`
}

// EventSchemaVersion is one stored version of an event type's schema.
// JSONSchema is the schema document as text, exactly as registered.
type EventSchemaVersion struct {
	Version    string `json:"version"`
	JSONSchema string `json:"jsonSchema"`
	IsActive   bool   `json:"isActive"`
	Enabled    bool   `json:"enabled"`
}

// EventSchemaVersionList lists an event type's schema versions.
type EventSchemaVersionList struct {
	Versions []EventSchemaVersion `json:"versions"`
}

// EventSchemaDocument is one version's JSON Schema document as stored.
type EventSchemaDocument struct {
	Schema json.RawMessage `json:"schema"`
}

// EventSchemaRegisterRequest registers a schema version, creating the event
// type when it is new. JSONSchema is the schema document as text.
type EventSchemaRegisterRequest struct {
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	Version    string `json:"version"`
	JSONSchema string `json:"jsonSchema"`
	Enabled    bool   `json:"enabled"`
	IsActive   bool   `json:"isActive"`
}

// EventSchemaCheckRequest is a schema and optional example events to check
// without registering anything.
type EventSchemaCheckRequest struct {
	Slug       string            `json:"slug"`
	Name       string            `json:"name,omitempty"`
	Version    string            `json:"version"`
	JSONSchema string            `json:"jsonSchema"`
	Examples   []json.RawMessage `json:"examples,omitempty"`
}

// EventSchemaIssue is one problem the validator found.
type EventSchemaIssue struct {
	Path             string `json:"path"`
	Keyword          string `json:"keyword"`
	Message          string `json:"message"`
	SchemaLocation   string `json:"schemaLocation"`
	InstanceLocation string `json:"instanceLocation"`
	Severity         string `json:"severity"`
}

// EventSchemaExampleResult says whether one example event matches the schema.
type EventSchemaExampleResult struct {
	Index      int                `json:"index"`
	Valid      bool               `json:"valid"`
	IssueCodes []string           `json:"issueCodes"`
	Issues     []EventSchemaIssue `json:"issues"`
}

// EventSchemaConflict is a field a registration would change incompatibly.
type EventSchemaConflict struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// EventSchemaCheckResult is the outcome of validating or dry-running a schema.
type EventSchemaCheckResult struct {
	Valid                bool                       `json:"valid"`
	SchemaValid          bool                       `json:"schemaValid"`
	Issues               []EventSchemaIssue         `json:"issues"`
	Warnings             []EventSchemaIssue         `json:"warnings"`
	ExampleResults       []EventSchemaExampleResult `json:"exampleResults"`
	NormalizedSchema     string                     `json:"normalizedSchema"`
	Dialect              string                     `json:"dialect"`
	Checksum             string                     `json:"checksum"`
	WouldCreateEventType bool                       `json:"wouldCreateEventType"`
	WouldCreateVersion   bool                       `json:"wouldCreateVersion"`
	Conflicts            []EventSchemaConflict      `json:"conflicts"`
	ValidatorEngine      string                     `json:"validatorEngine"`
	ValidatorVersion     string                     `json:"validatorVersion"`
	SchemaChecksum       string                     `json:"schemaChecksum"`
	SchemaDialect        string                     `json:"schemaDialect"`
}

// EventSchemaInferRequest is the sample events a candidate schema is inferred from.
type EventSchemaInferRequest struct {
	Samples []json.RawMessage `json:"samples"`
}

// EventSchemaInferResult is a candidate schema inferred from samples.
type EventSchemaInferResult struct {
	Valid             bool               `json:"valid"`
	Issues            []EventSchemaIssue `json:"issues"`
	InferenceWarnings []EventSchemaIssue `json:"inferenceWarnings"`
	CandidateSchema   string             `json:"candidateSchema"`
	ValidatorEngine   string             `json:"validatorEngine"`
	ValidatorVersion  string             `json:"validatorVersion"`
	SchemaChecksum    string             `json:"schemaChecksum"`
	SchemaDialect     string             `json:"schemaDialect"`
}

// EventSchemaStateResult confirms a state change.
type EventSchemaStateResult struct {
	Updated bool `json:"updated"`
}

// List returns the tenant's event types.
func (c *SchemaRegistryClient) List(ctx context.Context) (*EventTypeList, error) {
	var eventTypes []EventType
	if err := c.registry.request(ctx, http.MethodGet, schemaRegistryEndpoint, nil, &eventTypes); err != nil {
		return nil, err
	}
	return &EventTypeList{EventTypes: emptyIfNil(eventTypes)}, nil
}

// ListVersions returns an event type's schema versions.
func (c *SchemaRegistryClient) ListVersions(ctx context.Context, eventType string) (*EventSchemaVersionList, error) {
	var versions []EventSchemaVersion
	path := eventSchemaPath(eventType) + "/versions"
	if err := c.registry.request(ctx, http.MethodGet, path, nil, &versions); err != nil {
		return nil, err
	}
	return &EventSchemaVersionList{Versions: emptyIfNil(versions)}, nil
}

// Get returns one version's JSON Schema document.
func (c *SchemaRegistryClient) Get(ctx context.Context, eventType, version string) (*EventSchemaDocument, error) {
	var document EventSchemaDocument
	if err := c.registry.request(ctx, http.MethodGet, eventSchemaVersionPath(eventType, version), nil, &document.Schema); err != nil {
		return nil, err
	}
	return &document, nil
}

// Register stores a schema version and returns its event type.
func (c *SchemaRegistryClient) Register(ctx context.Context, req EventSchemaRegisterRequest) (*EventType, error) {
	var out EventType
	if err := c.registry.request(ctx, http.MethodPost, schemaRegistryEndpoint, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Validate checks a schema and its examples without registering anything.
func (c *SchemaRegistryClient) Validate(ctx context.Context, req EventSchemaCheckRequest) (*EventSchemaCheckResult, error) {
	return c.check(ctx, "/validate", req)
}

// DryRun reports what registering the schema would do, without registering it.
func (c *SchemaRegistryClient) DryRun(ctx context.Context, req EventSchemaCheckRequest) (*EventSchemaCheckResult, error) {
	return c.check(ctx, "/dry-run", req)
}

func (c *SchemaRegistryClient) check(ctx context.Context, action string, req EventSchemaCheckRequest) (*EventSchemaCheckResult, error) {
	var out EventSchemaCheckResult
	if err := c.registry.request(ctx, http.MethodPost, schemaRegistryEndpoint+action, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Infer proposes a schema from sample events.
func (c *SchemaRegistryClient) Infer(ctx context.Context, req EventSchemaInferRequest) (*EventSchemaInferResult, error) {
	var out EventSchemaInferResult
	if err := c.registry.request(ctx, http.MethodPost, schemaRegistryEndpoint+"/infer", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Activate makes one version the event type's active schema.
func (c *SchemaRegistryClient) Activate(ctx context.Context, eventType, version string) (*EventSchemaStateResult, error) {
	return c.state(ctx, http.MethodPost, eventSchemaVersionPath(eventType, version)+"/active")
}

// Enable lets events validate against one version again.
func (c *SchemaRegistryClient) Enable(ctx context.Context, eventType, version string) (*EventSchemaStateResult, error) {
	return c.state(ctx, http.MethodPut, eventSchemaVersionPath(eventType, version)+"/enabled")
}

// Disable deprecates one version so new events no longer validate against it.
func (c *SchemaRegistryClient) Disable(ctx context.Context, eventType, version string) (*EventSchemaStateResult, error) {
	return c.state(ctx, http.MethodDelete, eventSchemaVersionPath(eventType, version)+"/enabled")
}

// DisableEventType retires the tenant's event type. Its schemas and stored
// events stay readable.
func (c *SchemaRegistryClient) DisableEventType(ctx context.Context, eventType string) (*EventSchemaStateResult, error) {
	return c.state(ctx, http.MethodDelete, eventSchemaPath(eventType))
}

func (c *SchemaRegistryClient) state(ctx context.Context, method, path string) (*EventSchemaStateResult, error) {
	var out EventSchemaStateResult
	if err := c.registry.request(ctx, method, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func eventSchemaPath(eventType string) string {
	return schemaRegistryEndpoint + "/" + url.PathEscape(eventType)
}

func eventSchemaVersionPath(eventType, version string) string {
	return eventSchemaPath(eventType) + "/" + url.PathEscape(version)
}

// emptyIfNil turns a JSON null list into an empty one.
func emptyIfNil[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}
