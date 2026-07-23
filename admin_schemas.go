package custd

type AdminSchemaRegister struct {
	EventTypeSlug string         `json:"eventTypeSlug"`
	Version       string         `json:"version"`
	JSONSchema    map[string]any `json:"jsonSchema"`
}

type AdminSchemaVersionCreate struct {
	Version    string         `json:"version"`
	JSONSchema map[string]any `json:"jsonSchema"`
}

type AdminSchema struct {
	EventTypeSlug string         `json:"eventTypeSlug"`
	Version       string         `json:"version"`
	JSONSchema    map[string]any `json:"jsonSchema,omitempty"`
}

type AdminSchemaList struct {
	Schemas []AdminSchema `json:"schemas"`
}

// AdminSchemaValidateRequest matches the Stage 1 contract: dialect is jsonschema
// or avro. The action does not persist.
type AdminSchemaValidateRequest struct {
	TenantSlug    string `json:"tenantSlug"`
	EventTypeSlug string `json:"eventTypeSlug"`
	Dialect       string `json:"dialect"`
	SchemaJSON    string `json:"schemaJson"`
}

type AdminSchemaIssue struct {
	Code    string `json:"code,omitempty"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}

type AdminSchemaValidateResult struct {
	Valid           bool               `json:"valid"`
	Issues          []AdminSchemaIssue `json:"issues"`
	SuggestedAction string             `json:"suggestedAction"`
}

// AdminSchemaEnableRequest matches Stage 1. enabledAt before the schema's
// registration time returns 400.
type AdminSchemaEnableRequest struct {
	TenantSlug    string `json:"tenantSlug"`
	EventTypeSlug string `json:"eventTypeSlug"`
	Version       string `json:"version"`
	EnabledAt     string `json:"enabledAt,omitempty"`
}

// AdminSchemaDryRunRequest matches Stage 1. dialect is jsonschema or avro.
type AdminSchemaDryRunRequest struct {
	TenantSlug    string `json:"tenantSlug"`
	EventTypeSlug string `json:"eventTypeSlug"`
	Dialect       string `json:"dialect"`
	SchemaJSON    string `json:"schemaJson"`
	Samples       []any  `json:"samples"`
}

type AdminSchemaDryRunResult struct {
	Passed int64              `json:"passed"`
	Failed int64              `json:"failed"`
	Issues []AdminSchemaIssue `json:"issues"`
}

// AdminSchemaAudit is the company-scoped schema audit result returned by
// auditSchemas (Stage 1 contract). One row per version registration.
type AdminSchemaAuditEntry struct {
	EventID       string `json:"eventId"`
	EventTypeSlug string `json:"eventTypeSlug"`
	Version       string `json:"version"`
	Action        string `json:"action"`
	RegisteredAt  string `json:"registeredAt"`
	RegisteredBy  string `json:"registeredBy"`
}

type AdminSchemaAuditListResponse struct {
	Entries []AdminSchemaAuditEntry `json:"entries"`
}
