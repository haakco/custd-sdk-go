package custd

type AdminSchemaRegister struct {
	TenantSlug    string `json:"tenantSlug"`
	EventTypeSlug string `json:"eventTypeSlug"`
	Version       string `json:"version"`
	// SchemaJSON is the schema document as text. The server validates and stores the
	// text, so sending a decoded object would be a different contract.
	SchemaJSON string `json:"schemaJson"`
	Dialect    string `json:"dialect"`
	Checksum   string `json:"checksum"`
	Enabled    bool   `json:"enabled"`
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

// The schema audit answers a compatibility question, not an audit log: for every
// registered schema it reports how the stored JSON Schema compiles under both dialects and
// how the recorded examples validate, then summarises the counts. The SDK cannot import the
// reporting library, so these mirror libs/go/operatorapi field for field.
type SchemaCompileResult struct {
	Valid  bool     `json:"valid"`
	Issues []string `json:"issues"`
}

type SchemaExampleValidationSummary struct {
	Discovered int      `json:"discovered"`
	Run        int      `json:"run"`
	SameAccept int      `json:"sameAccept"`
	SameReject int      `json:"sameReject"`
	Deltas     int      `json:"deltas"`
	IssueCodes []string `json:"issueCodes"`
}

// SchemaAuditResult is one registered schema's audit outcome. Class names the verdict and
// RecommendedAction says what an operator should do about it.
type SchemaAuditResult struct {
	TenantSlug        string                         `json:"tenantSlug"`
	EventTypeSlug     string                         `json:"eventTypeSlug"`
	Version           string                         `json:"version"`
	SchemaChecksum    string                         `json:"schemaChecksum"`
	Dialect           string                         `json:"dialect"`
	LegacyCompile     SchemaCompileResult            `json:"legacyCompile"`
	SharedCompile     SchemaCompileResult            `json:"sharedCompile"`
	ExampleValidation SchemaExampleValidationSummary `json:"exampleValidation"`
	Class             string                         `json:"class"`
	IssueCodes        []string                       `json:"issueCodes"`
	RecommendedAction string                         `json:"recommendedAction"`
}

type SchemaAuditSummary struct {
	Total          int `json:"total"`
	Compatible     int `json:"compatible"`
	NeedsMigration int `json:"needsMigration"`
	Blocked        int `json:"blocked"`
}

type SchemaAuditResponse struct {
	Schemas []SchemaAuditResult `json:"schemas"`
	Summary SchemaAuditSummary  `json:"summary"`
}
