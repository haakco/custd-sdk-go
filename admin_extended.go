package custd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// TenantStorageAdminClient owns tenant-scoped storage locations. Custd
// generates each location's identity, bucket prefix and write state; the
// tenant is derived from the auth context, and wrong-tenant reads collapse to
// an empty list indistinguishable from "no locations".
type TenantStorageAdminClient struct {
	admin *AdminClient
}

// TenantStorageLocation is the tenant-scoped view of one storage location: its
// routing identity and lifecycle state, never S3 credentials.
type TenantStorageLocation struct {
	ID               int64  `json:"id"`
	UUID             string `json:"uuid"`
	CompanyID        int64  `json:"companyId"`
	Code             string `json:"code"`
	IsolationTier    string `json:"isolationTier"`
	Provider         string `json:"provider"`
	Bucket           string `json:"bucket"`
	Region           string `json:"region"`
	Endpoint         string `json:"endpoint"`
	RootPrefix       string `json:"rootPrefix"`
	WriteEnabled     bool   `json:"writeEnabled"`
	ValidationStatus string `json:"validationStatus"`
	ValidatedAt      string `json:"validatedAt,omitempty"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

// TenantStorageListResponse is the body for GET /tenant-storage-locations.
type TenantStorageListResponse struct {
	Locations []TenantStorageLocation `json:"locations"`
}

// TenantStorageCreateRequest is the body for POST /tenant-storage-locations.
// It is empty: Custd generates the routing, identity, prefix and write state.
type TenantStorageCreateRequest struct{}

func (c *TenantStorageAdminClient) List(ctx context.Context) (*TenantStorageListResponse, error) {
	var out TenantStorageListResponse
	err := c.admin.requestNonAdmin(ctx, http.MethodGet, "/tenant-storage-locations", nil, &out)
	return &out, err
}

func (c *TenantStorageAdminClient) Create(
	ctx context.Context,
	req TenantStorageCreateRequest,
) (*TenantStorageLocation, error) {
	var out TenantStorageLocation
	err := c.admin.requestNonAdmin(ctx, http.MethodPost, "/tenant-storage-locations", req, &out)
	return &out, err
}

func (c *TenantStorageAdminClient) Get(ctx context.Context, id string) (*TenantStorageLocation, error) {
	var out TenantStorageLocation
	err := c.admin.requestNonAdmin(ctx, http.MethodGet, "/tenant-storage-locations/"+url.PathEscape(id), nil, &out)
	return &out, err
}

// Revoke removes a tenant storage location. The server is the authority for
// whether the prefix is immediately unusable; the SDK must not assume
// partial deletes are atomic.
func (c *TenantStorageAdminClient) Revoke(ctx context.Context, id string) error {
	return c.admin.requestNonAdmin(
		ctx,
		http.MethodDelete,
		"/tenant-storage-locations/"+url.PathEscape(id),
		nil,
		nil,
	)
}

// SubjectExportAdminClient owns per-tenant subject export requests. The
// download surface returns artifact metadata the SDK surfaces only to the
// caller; it must not be logged or echoed into error messages.
type SubjectExportAdminClient struct {
	admin *AdminClient
}

// SubjectExportStoreProgress is one store's progress for a subject export.
type SubjectExportStoreProgress struct {
	Store              string `json:"store"`
	DeletedCount       int64  `json:"deletedCount"`
	RetainedExceptions int64  `json:"retainedExceptions"`
	LastError          string `json:"lastError,omitempty"`
	LastCompletedAt    string `json:"lastCompletedAt,omitempty"`
}

// SubjectExportScope is what a subject export covers. A request may ask for both; each downloadable artifact is access or portability.
type SubjectExportScope string

// Subject export scopes. A request may ask for both; each downloadable artifact
// is one of access or portability.
const (
	SubjectExportScopeAccess      SubjectExportScope = "access"
	SubjectExportScopePortability SubjectExportScope = "portability"
	SubjectExportScopeBoth        SubjectExportScope = "both"
)

// SubjectExportStatus is where a subject export request stands.
type SubjectExportStatus string

const (
	SubjectExportStatusQueued    SubjectExportStatus = "queued"
	SubjectExportStatusRunning   SubjectExportStatus = "running"
	SubjectExportStatusComplete  SubjectExportStatus = "complete"
	SubjectExportStatusFailed    SubjectExportStatus = "failed"
	SubjectExportStatusCancelled SubjectExportStatus = "cancelled"
	SubjectExportStatusExpired   SubjectExportStatus = "expired"
)

// SubjectExport is the status of one subject export request. Scope is access,
// portability or both; State is queued, running, complete, failed, cancelled
// or expired.
type SubjectExport struct {
	ID                 string                       `json:"id"`
	OpaqueSubjectID    string                       `json:"opaqueSubjectId"`
	Scope              SubjectExportScope           `json:"scope"`
	State              SubjectExportStatus          `json:"state"`
	IdempotencyKey     string                       `json:"idempotencyKey"`
	RequestedAt        string                       `json:"requestedAt"`
	ExpiresAt          string                       `json:"expiresAt"`
	CompletedAt        string                       `json:"completedAt,omitempty"`
	LastProgressAt     string                       `json:"lastProgressAt,omitempty"`
	LastActionSafeNext string                       `json:"lastActionSafeNext,omitempty"`
	Error              string                       `json:"error,omitempty"`
	PerStoreProgress   []SubjectExportStoreProgress `json:"perStoreProgress"`
}

// SubjectExportListResponse is the body for GET /admin/subject-exports.
type SubjectExportListResponse struct {
	Requests []SubjectExport `json:"requests"`
}

// SubjectExportCreateRequest is the body for POST /admin/subject-exports.
// IdempotencyKey is required for safe retries.
type SubjectExportCreateRequest struct {
	OpaqueSubjectID string             `json:"opaqueSubjectId"`
	Scope           SubjectExportScope `json:"scope"`
	IdempotencyKey  string             `json:"idempotencyKey"`
}

// SubjectExportDownloadResponse is the body for GET
// /admin/subject-exports/{requestId}/download: the stored artifact's identity,
// location and integrity metadata. It names an object key, not a signed URL.
type SubjectExportDownloadResponse struct {
	ID              string             `json:"id"`
	RequestID       string             `json:"requestId"`
	Scope           SubjectExportScope `json:"scope"`
	ObjectKey       string             `json:"objectKey"`
	S3Key           string             `json:"s3Key"`
	EncryptionKeyID string             `json:"encryptionKeyId"`
	ChecksumSHA256  string             `json:"checksumSha256"`
	ByteSize        int64              `json:"byteSize"`
	RecordCount     int64              `json:"recordCount"`
	GeneratedAt     string             `json:"generatedAt"`
	ExpiresAt       string             `json:"expiresAt"`
}

func (c *SubjectExportAdminClient) Create(
	ctx context.Context,
	req SubjectExportCreateRequest,
) (*SubjectExport, error) {
	var out SubjectExport
	err := c.admin.request(ctx, http.MethodPost, "/subject-exports", req, &out)
	return &out, err
}

// List returns the tenant's subject exports, newest first. A limit of zero uses the
// server default (100); the server accepts 1 to 1000.
func (c *SubjectExportAdminClient) List(ctx context.Context, limit int) (*SubjectExportListResponse, error) {
	var out SubjectExportListResponse
	path := siteQueryPath("/subject-exports", url.Values{"limit": {sitePositiveInt(limit)}})
	err := c.admin.request(ctx, http.MethodGet, path, nil, &out)
	return &out, err
}

func (c *SubjectExportAdminClient) Get(ctx context.Context, requestID string) (*SubjectExport, error) {
	return adminGetByID[SubjectExport](ctx, c.admin, "/subject-exports/", requestID)
}

// Cancel stops a queued or running export and returns its cancelled status.
func (c *SubjectExportAdminClient) Cancel(ctx context.Context, requestID string) (*SubjectExport, error) {
	var out SubjectExport
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/subject-exports/"+url.PathEscape(requestID)+"/cancel",
		nil,
		&out,
	)
	return &out, err
}

// Download returns one stored export artifact's identity, location and
// integrity metadata. Scope names the artifact: SubjectExportScopeAccess or
// SubjectExportScopePortability. The object keys are tenant data; do not log them.
func (c *SubjectExportAdminClient) Download(
	ctx context.Context,
	requestID string,
	scope SubjectExportScope,
) (*SubjectExportDownloadResponse, error) {
	var out SubjectExportDownloadResponse
	path := "/subject-exports/" + url.PathEscape(requestID) + "/download?" + url.Values{"scope": {string(scope)}}.Encode()
	err := c.admin.request(ctx, http.MethodGet, path, nil, &out)
	return &out, err
}

// Force runs a stalled export now. The route requires companySlug, which it
// takes from ClientConfig.ActingCompanySlug.
func (c *SubjectExportAdminClient) Force(ctx context.Context, requestID string) (*SubjectExport, error) {
	var out SubjectExport
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/subject-exports/"+url.PathEscape(requestID)+"/force",
		nil,
		&out,
	)
	return &out, err
}

// PrivacyErasureAdminClient owns per-tenant subject erasure requests.
// Erasures are forward-only: there is no Cancel or Retry surface because
// the server contract has none. Force is the bounded operator action.
type PrivacyErasureAdminClient struct {
	admin *AdminClient
}

// PrivacyErasureSelector is the typed selector the SDK submits to identify
// a subject: one Value or a list of Values. Do not log it.
type PrivacyErasureSelector struct {
	Type   string   `json:"type"`
	Value  string   `json:"value,omitempty"`
	Values []string `json:"values,omitempty"`
}

// PrivacyErasure is the read model of one erasure request: its state plus
// the rewrite, SLA and archive-propagation rollups.
type PrivacyErasure struct {
	RequestUUID             string `json:"requestUuid"`
	Status                  string `json:"status"`
	SelectorType            string `json:"selectorType"`
	SelectorDisplay         string `json:"selectorDisplay"`
	Source                  string `json:"source,omitempty"`
	RequestedBy             string `json:"requestedBy,omitempty"`
	MatchedEventCount       int64  `json:"matchedEventCount"`
	RedactedEventCount      int64  `json:"redactedEventCount"`
	AffectedSegmentCount    int64  `json:"affectedSegmentCount"`
	RewriteCount            int64  `json:"rewriteCount,omitempty"`
	PendingRewriteCount     int64  `json:"pendingRewriteCount,omitempty"`
	RewritingRewriteCount   int64  `json:"rewritingRewriteCount,omitempty"`
	RewrittenRewriteCount   int64  `json:"rewrittenRewriteCount,omitempty"`
	RecompactedRewriteCount int64  `json:"recompactedRewriteCount,omitempty"`
	FailedRewriteCount      int64  `json:"failedRewriteCount,omitempty"`
	OldestPendingRewriteAt  string `json:"oldestPendingRewriteAt,omitempty"`
	LatestRecompactedAt     string `json:"latestRecompactedAt,omitempty"`
	ArchiveRecompactedAt    string `json:"archiveRecompactedAt,omitempty"`
	S3PropagationLagSeconds int64  `json:"s3PropagationLagSeconds,omitempty"`
	LatestFailureReason     string `json:"latestFailureReason,omitempty"`
	SLADeadlineAt           string `json:"slaDeadlineAt,omitempty"`
	SLAState                string `json:"slaState,omitempty"`
	ForceRequestedAt        string `json:"forceRequestedAt,omitempty"`
	CreatedAt               string `json:"createdAt,omitempty"`
	UpdatedAt               string `json:"updatedAt,omitempty"`
}

// PrivacyErasureCreateRequest is the body for POST /admin/privacy/erasures.
type PrivacyErasureCreateRequest struct {
	CompanySlug string                 `json:"companySlug"`
	Selector    PrivacyErasureSelector `json:"selector"`
	Reason      string                 `json:"reason"`
	RedactionID string                 `json:"redactionId"`
	ForceNow    bool                   `json:"forceNow"`
}

// PrivacyErasureListResponse is the body for GET /admin/privacy/erasures.
type PrivacyErasureListResponse struct {
	Requests []PrivacyErasure `json:"requests"`
}

// PrivacyErasureForceResponse is the body for POST
// /admin/privacy/erasures/{requestUuid}/force: the request after forcing and
// the safe next action.
type PrivacyErasureForceResponse struct {
	Request            *PrivacyErasure `json:"request"`
	SafeNextAction     string          `json:"safe_next_action"`
	SafeNextActionCode string          `json:"safe_next_action_code"`
}

func (c *PrivacyErasureAdminClient) Create(
	ctx context.Context,
	req PrivacyErasureCreateRequest,
) (*PrivacyErasure, error) {
	var out PrivacyErasure
	err := c.admin.request(ctx, http.MethodPost, "/privacy/erasures", req, &out)
	return &out, err
}

// List returns the tenant's erasure requests. An empty companySlug reads the
// credential's own tenant; a system administrator names another one.
func (c *PrivacyErasureAdminClient) List(ctx context.Context, companySlug string) (*PrivacyErasureListResponse, error) {
	var out PrivacyErasureListResponse
	err := c.admin.request(ctx, http.MethodGet, "/privacy/erasures"+optionalCompanySlug(companySlug), nil, &out)
	return &out, err
}

// Get returns one erasure request. An empty companySlug reads the credential's
// own tenant; a system administrator names another one.
func (c *PrivacyErasureAdminClient) Get(
	ctx context.Context,
	requestUUID string,
	companySlug string,
) (*PrivacyErasure, error) {
	var out PrivacyErasure
	path := "/privacy/erasures/" + url.PathEscape(requestUUID) + optionalCompanySlug(companySlug)
	err := c.admin.request(ctx, http.MethodGet, path, nil, &out)
	return &out, err
}

func (c *PrivacyErasureAdminClient) Force(ctx context.Context, requestUUID string) (*PrivacyErasureForceResponse, error) {
	var out PrivacyErasureForceResponse
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/privacy/erasures/"+url.PathEscape(requestUUID)+"/force",
		nil,
		&out,
	)
	return &out, err
}

// PrivacyAdminClient owns the privacy subtrack: rules (closed-purpose) and
// tenant identifier mappings. The identifier surfaces only ever return the
// truncated HMAC hash prefix; the plaintext externalId is consumed once on the
// request boundary.
type PrivacyAdminClient struct {
	admin *AdminClient
}

// PrivacyRule is one field's ingestion-time treatment. Action is one of keep, hash, truncate, remove,
// strip_query or origin; TruncateLength is required and positive for truncate.
type PrivacyRule struct {
	FieldPath      string `json:"fieldPath"`
	Action         string `json:"action"`
	TruncateLength *int   `json:"truncateLength,omitempty"`
	Enabled        bool   `json:"enabled"`
}

// PrivacyRules is a tenant's ingestion privacy configuration.
//
// The route this models takes `{companySlug, applyMode, enabled, rules`. An earlier version of this
// type described purposes and a hard-delete window, which the route does not accept or return: a
// caller could set both fields, have them silently ignored, and be told the update succeeded.
type PrivacyRules struct {
	CompanySlug string        `json:"companySlug"`
	ApplyMode   string        `json:"applyMode"`
	Enabled     bool          `json:"enabled"`
	Rules       []PrivacyRule `json:"rules"`
}

type PrivacyIdentifierMapRequest struct {
	// ExternalID is the plain identifier the SDK consumer already knows. It is
	// consumed once on the wire; do not log or echo the response payload for
	// this request back to a place where ExternalID could appear.
	ExternalID string `json:"externalId"`
}

// PrivacyIdentifierListResponse is the effective tenant's mapping list.
type PrivacyIdentifierListResponse struct {
	CompanySlug string                     `json:"companySlug"`
	Mappings    []PrivacyIdentifierMapping `json:"mappings"`
}

type PrivacyIdentifierMapping struct {
	IdentifierID         string `json:"identifierId"`
	InternalIDHash       string `json:"internalIdHash"`
	InternalIDHashPrefix string `json:"internalIdHashPrefix"`
	SaltVersion          int    `json:"saltVersion"`
	CreatedAt            string `json:"createdAt,omitempty"`
}

// GetRules returns the tenant's ingestion privacy configuration. An empty companySlug reads the
// configuration for the caller's own tenant.
func (c *PrivacyAdminClient) GetRules(ctx context.Context, companySlug string) (*PrivacyRules, error) {
	var out PrivacyRules
	err := c.admin.request(
		ctx, http.MethodGet, "/privacy/rules"+optionalCompanySlug(companySlug), nil, &out,
	)
	return &out, err
}

// SetRules replaces the tenant's ingestion privacy configuration. The server requires applyMode to be
// at_ingestion and refuses a truncate rule without a positive TruncateLength.
func (c *PrivacyAdminClient) SetRules(ctx context.Context, req PrivacyRules) (*PrivacyRules, error) {
	var out PrivacyRules
	err := c.admin.request(ctx, http.MethodPut, "/privacy/rules", req, &out)
	return &out, err
}

// MapIdentifier consumes the plaintext identifier once and returns the hash
// metadata. The SDK never stores or logs ExternalID; callers must surface it to
// their own capture-only recipient and drop it from their process memory.
func (c *PrivacyAdminClient) MapIdentifier(
	ctx context.Context,
	companySlug string,
	req PrivacyIdentifierMapRequest,
) (*PrivacyIdentifierMapping, error) {
	var out PrivacyIdentifierMapping
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/privacy/identifiers/"+url.PathEscape(companySlug)+"/map",
		req,
		&out,
	)
	return &out, err
}

// ListIdentifierMappings returns the hashed mappings for the effective tenant.
// Wrong-tenant requests collapse to a 404 indistinguishable from not-found.
func (c *PrivacyAdminClient) ListIdentifierMappings(
	ctx context.Context,
	companySlug string,
) (*PrivacyIdentifierListResponse, error) {
	var out PrivacyIdentifierListResponse
	err := c.admin.request(
		ctx,
		http.MethodGet,
		"/privacy/identifiers/"+url.PathEscape(companySlug),
		nil,
		&out,
	)
	return &out, err
}

// RetentionAdminClient owns per-tenant retention policies. Effective-tenant
// authority is enforced server-side; wrong-tenant requests return 404.
type RetentionAdminClient struct {
	admin *AdminClient
}

// RetentionClass is the class of stored data a retention policy governs.
type RetentionClass string

const (
	RetentionClassRaw            RetentionClass = "raw"
	RetentionClassArchive        RetentionClass = "archive"
	RetentionClassMaterialized   RetentionClass = "materialized"
	RetentionClassCache          RetentionClass = "cache"
	RetentionClassObject         RetentionClass = "object"
	RetentionClassAuditEvent     RetentionClass = "audit_event"
	RetentionClassAuditIPAddress RetentionClass = "audit_ip_address"
	RetentionClassAuditUserAgent RetentionClass = "audit_user_agent"
)

// RetentionScope is what a retention policy applies to.
type RetentionScope string

const (
	RetentionScopeTenant    RetentionScope = "tenant"
	RetentionScopeDataSpace RetentionScope = "data_space"
	RetentionScopeEventType RetentionScope = "event_type"
	RetentionScopeGlobal    RetentionScope = "global"
)

// RetentionPolicy is one bounded retention policy row. Scope is tenant,
// data_space, event_type or global; RetentionClass names the classified store.
type RetentionPolicy struct {
	ID                 int64            `json:"id"`
	CompanyID          int64            `json:"companyId,omitempty"`
	Scope              RetentionScope   `json:"scope"`
	RetentionClass     RetentionClass   `json:"retentionClass"`
	MaxAgeSeconds      int64            `json:"maxAgeSeconds"`
	Precedence         int64            `json:"precedence"`
	LegalHold          bool             `json:"legalHold"`
	LegalHoldUntil     string           `json:"legalHoldUntil,omitempty"`
	EffectiveAt        string           `json:"effectiveAt"`
	ExpiresAt          string           `json:"expiresAt,omitempty"`
	NextRunAt          string           `json:"nextRunAt,omitempty"`
	LastResult         string           `json:"lastResult,omitempty"`
	LastAffectedCounts map[string]int64 `json:"lastAffectedCounts"`
	LastError          string           `json:"lastError,omitempty"`
	AuditID            int64            `json:"auditId,omitempty"`
}

// RetentionPolicyUpsertRequest is the body for PUT
// /admin/retention/policies/{tenantSlug}. LegalHold is operator-only and needs
// a future LegalHoldUntil plus an opaque LegalHoldReference.
type RetentionPolicyUpsertRequest struct {
	Scope              RetentionScope `json:"scope"`
	RetentionClass     RetentionClass `json:"retentionClass"`
	MaxAgeSeconds      int64          `json:"maxAgeSeconds"`
	Precedence         int64          `json:"precedence"`
	LegalHold          bool           `json:"legalHold"`
	LegalHoldUntil     string         `json:"legalHoldUntil,omitempty"`
	LegalHoldReference string         `json:"legalHoldReference,omitempty"`
	EffectiveAt        string         `json:"effectiveAt,omitempty"`
	ExpiresAt          string         `json:"expiresAt,omitempty"`
}

// RetentionPolicyListResponse is the body for GET /admin/retention/policies
// and GET /admin/retention/policies/{tenantSlug}.
type RetentionPolicyListResponse struct {
	Policies []RetentionPolicy `json:"policies"`
}

func (c *RetentionAdminClient) List(ctx context.Context) (*RetentionPolicyListResponse, error) {
	var out RetentionPolicyListResponse
	err := c.admin.request(ctx, http.MethodGet, "/retention/policies", nil, &out)
	return &out, err
}

func (c *RetentionAdminClient) Upsert(
	ctx context.Context,
	tenantSlug string,
	req RetentionPolicyUpsertRequest,
) (*RetentionPolicy, error) {
	var out RetentionPolicy
	err := c.admin.request(
		ctx,
		http.MethodPut,
		"/retention/policies/"+url.PathEscape(tenantSlug),
		req,
		&out,
	)
	return &out, err
}

// Get returns every retention policy row for one tenant.
func (c *RetentionAdminClient) Get(ctx context.Context, tenantSlug string) (*RetentionPolicyListResponse, error) {
	return adminGetByID[RetentionPolicyListResponse](ctx, c.admin, "/retention/policies/", tenantSlug)
}

func (c *RetentionAdminClient) Delete(ctx context.Context, tenantSlug string) error {
	return c.admin.request(
		ctx,
		http.MethodDelete,
		"/retention/policies/"+url.PathEscape(tenantSlug),
		nil,
		nil,
	)
}

// RetentionStoreBreakdown is one classified store's projection in a retention
// preview or apply. The counts are server-computed; do not re-derive them.
type RetentionStoreBreakdown struct {
	Store             string   `json:"store"`
	RetentionClass    string   `json:"retentionClass"`
	CandidateCount    int64    `json:"candidateCount"`
	EstimatedAffected int64    `json:"estimatedAffected"`
	SampleSubjects    []string `json:"sampleSubjects"`
	LegalHoldCount    int      `json:"legalHoldCount"`
}

// RetentionRunPreview is the body for both preview and apply. LastResult is
// "preview" for a dry run and "applied" for a live run; LastAffectedCounts is
// set only for applies.
type RetentionRunPreview struct {
	PerStore           []RetentionStoreBreakdown `json:"perStore"`
	NextRunAt          string                    `json:"nextRunAt,omitempty"`
	LastResult         string                    `json:"lastResult"`
	LastAffectedCounts map[string]int64          `json:"lastAffectedCounts,omitempty"`
	LastError          string                    `json:"lastError,omitempty"`
	SafeNextAction     string                    `json:"safeNextAction"`
	StartedAt          string                    `json:"startedAt"`
	CompletedAt        string                    `json:"completedAt"`
}

// RetentionRun is one historical retention run.
type RetentionRun struct {
	StartAt            string           `json:"startAt"`
	EndAt              string           `json:"endAt"`
	LastResult         string           `json:"lastResult"`
	LastAffectedCounts map[string]int64 `json:"lastAffectedCounts"`
	LastError          string           `json:"lastError,omitempty"`
	AuditID            *int64           `json:"auditId,omitempty"`
}

// RetentionRunsListResponse is the body for GET
// /admin/retention/policies/{slug}/runs. Follow NextOffset while HasMore.
type RetentionRunsListResponse struct {
	Runs       []RetentionRun `json:"runs"`
	NextOffset int            `json:"nextOffset"`
	HasMore    bool           `json:"hasMore"`
}

// Preview asks the server to compute a deletion estimate without applying it.
// The estimate is server-issued; the SDK must surface it verbatim and never
// round or re-derive the per-store counts.
func (c *RetentionAdminClient) Preview(
	ctx context.Context,
	tenantSlug string,
) (*RetentionRunPreview, error) {
	var out RetentionRunPreview
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/retention/policies/"+url.PathEscape(tenantSlug)+"/preview",
		nil,
		&out,
	)
	return &out, err
}

// Apply submits the destructive retention run. The server is the authority
// for whether deletion actually happens; the SDK must not pre-announce state.
func (c *RetentionAdminClient) Apply(
	ctx context.Context,
	tenantSlug string,
) (*RetentionRunPreview, error) {
	var out RetentionRunPreview
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/retention/policies/"+url.PathEscape(tenantSlug)+"/apply",
		nil,
		&out,
	)
	return &out, err
}

// RetentionRunsPage selects a page of retention runs. Zero values use the
// server defaults.
type RetentionRunsPage struct {
	Limit  int
	Offset int
}

// ListRuns returns the retention runs for a single tenant. Empty runs list
// is the canonical "no runs yet" response, not an error.
func (c *RetentionAdminClient) ListRuns(
	ctx context.Context,
	tenantSlug string,
	page ...RetentionRunsPage,
) (*RetentionRunsListResponse, error) {
	var selected RetentionRunsPage
	if len(page) > 0 {
		selected = page[0]
	}
	path := siteQueryPath("/retention/policies/"+url.PathEscape(tenantSlug)+"/runs", url.Values{
		"limit": {sitePositiveInt(selected.Limit)}, "offset": {sitePositiveInt(selected.Offset)},
	})
	var out RetentionRunsListResponse
	err := c.admin.request(ctx, http.MethodGet, path, nil, &out)
	return &out, err
}

// StorageAlertAdminClient owns tenant-scoped storage alert rules. The list and
// delete surfaces are tenant-safe (effective tenant collapsed to 404).
type StorageAlertAdminClient struct {
	admin *AdminClient
}

// StorageAlertChannel is how a storage alert is delivered.
type StorageAlertChannel string

const (
	StorageAlertChannelEmail   StorageAlertChannel = "email"
	StorageAlertChannelWebhook StorageAlertChannel = "webhook"
)

type StorageAlertRule struct {
	RuleID           string                  `json:"ruleId"`
	TenantSlug       string                  `json:"tenantSlug"`
	Metric           string                  `json:"metric"`
	ThresholdPercent int                     `json:"thresholdPercent"`
	Channel          StorageAlertChannel     `json:"channel"`
	EmailRecipient   string                  `json:"emailRecipient,omitempty"`
	Enabled          bool                    `json:"enabled"`
	Evaluation       *StorageAlertEvaluation `json:"evaluation,omitempty"`
	CreatedAt        string                  `json:"createdAt"`
	UpdatedAt        string                  `json:"updatedAt"`
}

// StorageAlertEvaluation is the latest evaluation of a storage alert rule.
type StorageAlertEvaluation struct {
	Breached      bool   `json:"breached"`
	MeasuredValue int64  `json:"measuredValue"`
	EvaluatedAt   string `json:"evaluatedAt"`
	FiredAt       string `json:"firedAt,omitempty"`
}

// StorageAlertRuleCreateRequest is the body for POST
// /admin/storage/alerts/{tenantSlug}. Channel is email or webhook.
type StorageAlertRuleCreateRequest struct {
	Metric           string              `json:"metric"`
	ThresholdPercent int                 `json:"thresholdPercent"`
	Channel          StorageAlertChannel `json:"channel"`
	EmailRecipient   string              `json:"emailRecipient"`
	Enabled          bool                `json:"enabled"`
}

type StorageAlertRuleListResponse struct {
	Rules []StorageAlertRule `json:"rules"`
}

func (c *StorageAlertAdminClient) ListRules(
	ctx context.Context,
	tenantSlug string,
) (*StorageAlertRuleListResponse, error) {
	var out StorageAlertRuleListResponse
	err := c.admin.request(
		ctx,
		http.MethodGet,
		"/storage/alerts/"+url.PathEscape(tenantSlug),
		nil,
		&out,
	)
	return &out, err
}

func (c *StorageAlertAdminClient) CreateRule(
	ctx context.Context,
	tenantSlug string,
	req StorageAlertRuleCreateRequest,
) (*StorageAlertRule, error) {
	var out StorageAlertRule
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/storage/alerts/"+url.PathEscape(tenantSlug),
		req,
		&out,
	)
	return &out, err
}

func (c *StorageAlertAdminClient) DeleteRule(ctx context.Context, tenantSlug string, ruleID string) error {
	return c.admin.request(
		ctx,
		http.MethodDelete,
		"/storage/alerts/"+url.PathEscape(tenantSlug)+"/"+url.PathEscape(ruleID),
		nil,
		nil,
	)
}

// AuditAdminClient owns the company-scoped audit read surface.
type AuditAdminClient struct {
	admin *AdminClient
}

type AuditScope string

const (
	AuditScopeTenant AuditScope = "tenant"
	AuditScopeGlobal AuditScope = "global"
)

type AuditOutcome string

const (
	AuditOutcomeAttempted AuditOutcome = "attempted"
	AuditOutcomeSuccess   AuditOutcome = "success"
	AuditOutcomeFailure   AuditOutcome = "failure"
	AuditOutcomeDenied    AuditOutcome = "denied"
	AuditOutcomePending   AuditOutcome = "pending"
	AuditOutcomeUnknown   AuditOutcome = "unknown"
)

type AuditDisclosureState string

const (
	AuditDisclosureAvailable   AuditDisclosureState = "available"
	AuditDisclosureRedacted    AuditDisclosureState = "redacted"
	AuditDisclosureNotRecorded AuditDisclosureState = "not_recorded"
)

type AuditChange struct {
	Field  string `json:"field"`
	Before any    `json:"before,omitempty"`
	After  any    `json:"after,omitempty"`
}

type AuditNetworkDisclosure struct {
	IPAddress      string               `json:"ipAddress,omitempty"`
	IPAddressState AuditDisclosureState `json:"ipAddressState"`
	UserAgent      string               `json:"userAgent,omitempty"`
	UserAgentState AuditDisclosureState `json:"userAgentState"`
}

type AuditRetentionDisclosure struct {
	EventMaxAgeSeconds     int64 `json:"eventMaxAgeSeconds"`
	IPAddressMaxAgeSeconds int64 `json:"ipAddressMaxAgeSeconds"`
	UserAgentMaxAgeSeconds int64 `json:"userAgentMaxAgeSeconds"`
}

type AuditEvent struct {
	// EventID is the stable public audit event UUID.
	EventID          string                  `json:"eventId"`
	TenantSlug       string                  `json:"tenantSlug"`
	ActorKind        string                  `json:"actorKind"`
	ActorReference   string                  `json:"actorReference,omitempty"`
	ActorDisplayName string                  `json:"actorDisplayName"`
	ActorRoles       []string                `json:"actorRoles,omitempty"`
	Action           string                  `json:"action"`
	ResourceType     string                  `json:"resourceType"`
	ResourceID       string                  `json:"resourceId,omitempty"`
	CreatedAt        string                  `json:"createdAt"`
	Outcome          AuditOutcome            `json:"outcome"`
	CorrelationID    string                  `json:"correlationId,omitempty"`
	OperationID      string                  `json:"operationId,omitempty"`
	Changes          []AuditChange           `json:"changes,omitempty"`
	Details          map[string]any          `json:"details,omitempty"`
	Network          *AuditNetworkDisclosure `json:"network,omitempty"`
	DomainHistory    *AuditDomainHistory     `json:"domainHistory,omitempty"`
}

// AuditDomainHistory is the domain-history record an audit event links to.
type AuditDomainHistory struct {
	Source        string `json:"source"`
	SourceID      string `json:"sourceId"`
	CreatedAt     string `json:"createdAt"`
	Action        string `json:"action,omitempty"`
	ResourceType  string `json:"resourceType,omitempty"`
	ResourceID    string `json:"resourceId,omitempty"`
	State         string `json:"state,omitempty"`
	PreviousState string `json:"previousState,omitempty"`
	CurrentState  string `json:"currentState,omitempty"`
	StreamVersion int64  `json:"streamVersion,omitempty"`
}

// AuditListCursor is the next page's cursor; Cursor is empty, and absent on the wire, on the last
// page.
type AuditListCursor struct {
	Cursor string `json:"cursor,omitempty"`
}

type AuditListResponse struct {
	Events           []AuditEvent              `json:"events"`
	NextCursor       AuditListCursor           `json:"nextCursor"`
	CoverageBeginsAt string                    `json:"coverageBeginsAt,omitempty"`
	Retention        *AuditRetentionDisclosure `json:"retention,omitempty"`
}

type AuditListOptions struct {
	Scope              AuditScope
	CompanySlug        string
	AffectedTenantSlug string
	Since              string
	Until              string
	ActorKind          string
	ActorReference     string
	Action             string
	ResourceType       string
	ResourceID         string
	Outcome            AuditOutcome
	CorrelationID      string
	// EnvironmentID narrows the events to one environment's realm.
	EnvironmentID string
	Limit         int
	Cursor        string
}

// AuditExportFormat is the file format an audit export is served in.
type AuditExportFormat string

const (
	AuditExportFormatCSV  AuditExportFormat = "csv"
	AuditExportFormatJSON AuditExportFormat = "json"
)

// AuditExport describes an audit export ExportEvents wrote: the media type the server served,
// the bytes written and the answer's request id.
type AuditExport struct {
	ContentType string
	ByteSize    int64
	Metadata    ResponseMetadata
}

type ReportingPackAuditEvent struct {
	Action           string `json:"action"`
	ActorReference   string `json:"actorReference,omitempty"`
	ActorDisplayName string `json:"actorDisplayName"`
	ResourceType     string `json:"resourceType"`
	PackKey          string `json:"packKey"`
	CreatedAt        string `json:"createdAt"`
}

type ReportingPackAuditListResponse struct {
	Events []ReportingPackAuditEvent `json:"events"`
}

func (c *AuditAdminClient) ListEvents(
	ctx context.Context,
	opts AuditListOptions,
) (*AuditListResponse, error) {
	path := "/audit/events"
	if params := auditListParams(opts); params != "" {
		path += "?" + params
	}
	var out AuditListResponse
	err := c.admin.request(ctx, http.MethodGet, path, nil, &out)
	if err == nil {
		for _, event := range out.Events {
			if err := validateAuditEventUUID(event.EventID); err != nil {
				return nil, err
			}
		}
	}
	return &out, err
}

func (c *AuditAdminClient) GetEvent(ctx context.Context, eventID string, options ...AuditListOptions) (*AuditEvent, error) {
	path := "/audit/events/" + url.PathEscape(eventID)
	if len(options) > 0 {
		if params := auditLookupParams(options[0]); params != "" {
			path += "?" + params
		}
	}
	var out AuditEvent
	err := c.admin.request(ctx, http.MethodGet, path, nil, &out)
	if err == nil {
		err = validateAuditEventUUID(out.EventID)
	}
	return &out, err
}

// ExportEvents streams the audit events matching opts into w in the given format, JSON when
// empty. On error w may hold a partial export.
func (c *AuditAdminClient) ExportEvents(
	ctx context.Context, opts AuditListOptions, format AuditExportFormat, w io.Writer,
) (*AuditExport, error) {
	if format == "" {
		format = AuditExportFormatJSON
	}
	if format != AuditExportFormatCSV && format != AuditExportFormatJSON {
		return nil, fmt.Errorf("custd: audit export format must be csv or json")
	}
	params := auditExportParams(opts)
	params.Set("format", string(format))
	path := c.admin.actingCompanyPath("/audit/events/export?" + params.Encode())
	download, err := c.admin.openAdminDownload(ctx, path, "audit export")
	if err != nil {
		return nil, err
	}
	written, copyErr := io.Copy(w, download.body)
	if err := errors.Join(copyErr, download.body.Close()); err != nil {
		return nil, fmt.Errorf("custd: stream audit export: %w", err)
	}
	return &AuditExport{ContentType: download.header.Get("Content-Type"), ByteSize: written, Metadata: download.metadata}, nil
}

func (c *AuditAdminClient) ListReportingPackEvents(ctx context.Context, packKey string) (*ReportingPackAuditListResponse, error) {
	var out ReportingPackAuditListResponse
	err := c.admin.request(ctx, http.MethodGet, "/reporting-packs/audit-events?packKey="+url.QueryEscape(packKey), nil, &out)
	return &out, err
}

// OffboardingAdminClient owns the offboarding schedule and one-off request
// surfaces.
type OffboardingAdminClient struct {
	admin *AdminClient
}

type OffboardingSchedule struct {
	TenantSlug      string `json:"tenantSlug"`
	EffectiveAt     string `json:"effectiveAt"`
	GracePeriodDays int    `json:"gracePeriodDays"`
	Reason          string `json:"reason"`
	Status          string `json:"status"`
	UpdatedAt       string `json:"updatedAt,omitempty"`
}

type OffboardingScheduleRequest struct {
	TenantSlug      string `json:"tenantSlug"`
	EffectiveAt     string `json:"effectiveAt"`
	GracePeriodDays int    `json:"gracePeriodDays"`
	Reason          string `json:"reason"`
	Status          string `json:"status,omitempty"`
}

type OffboardingScheduleListResponse struct {
	Schedules []OffboardingSchedule `json:"schedules"`
}

type OffboardingCancelRequest struct {
	Reason string `json:"reason"`
}

// OffboardingState is where a tenant offboarding request stands.
type OffboardingState string

const (
	OffboardingStatePreview         OffboardingState = "preview"
	OffboardingStateRequested       OffboardingState = "requested"
	OffboardingStateExporting       OffboardingState = "exporting"
	OffboardingStateExported        OffboardingState = "exported"
	OffboardingStateConfirming      OffboardingState = "confirming"
	OffboardingStateConfirmed       OffboardingState = "confirmed"
	OffboardingStateWaiverRequested OffboardingState = "waiver_requested"
	OffboardingStateWaiverApproved  OffboardingState = "waiver_approved"
	OffboardingStateFencing         OffboardingState = "fencing"
	OffboardingStateRevoking        OffboardingState = "revoking"
	OffboardingStateDeleting        OffboardingState = "deleting"
	OffboardingStateComplete        OffboardingState = "complete"
	OffboardingStateFailed          OffboardingState = "failed"
	OffboardingStateCancelled       OffboardingState = "cancelled"
)

// OffboardingRequest is the public request state returned by RequestOffboarding,
// GetRequest, and the acknowledgement endpoint.
type OffboardingRequest struct {
	RequestUUID string           `json:"requestUuid"`
	State       OffboardingState `json:"state"`
	RequestedAt string           `json:"requestedAt"`
}

// OffboardingRequestCreate carries the body for POST /offboarding. Confirmation
// is the human-typed string the server compares against the tenant slug before
// accepting the destructive transition.
type OffboardingRequestCreate struct {
	Confirmation   string `json:"confirmation"`
	IdempotencyKey string `json:"-"`
}

// Schedule writes a delayed offboarding schedule. The server checks TenantSlug
// against the authenticated tenant before persisting the schedule.
func (c *OffboardingAdminClient) Schedule(
	ctx context.Context,
	req OffboardingScheduleRequest,
) (*OffboardingSchedule, error) {
	var out OffboardingSchedule
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/offboarding/schedules",
		req,
		&out,
	)
	return &out, err
}

func (c *OffboardingAdminClient) ListSchedules(ctx context.Context) (*OffboardingScheduleListResponse, error) {
	var out OffboardingScheduleListResponse
	err := c.admin.request(ctx, http.MethodGet, "/offboarding/schedules", nil, &out)
	return &out, err
}

// GetSchedule reads the delayed offboarding schedule for a single tenant. It
// targets the per-tenant route GET /offboarding/schedules/{tenantSlug}, which
// is distinct from the global ListSchedules collection read.
func (c *OffboardingAdminClient) GetSchedule(ctx context.Context, tenantSlug string) (*OffboardingSchedule, error) {
	return adminGetByID[OffboardingSchedule](ctx, c.admin, "/offboarding/schedules/", tenantSlug)
}

func (c *OffboardingAdminClient) CancelSchedule(
	ctx context.Context,
	tenantSlug string,
	req OffboardingCancelRequest,
) (*OffboardingSchedule, error) {
	var out OffboardingSchedule
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/offboarding/schedules/"+url.PathEscape(tenantSlug)+"/cancel",
		req,
		&out,
	)
	return &out, err
}

// RequestOffboarding submits a one-off offboarding request for the effective
// tenant via POST /offboarding. The Confirmation field must match the tenant
// slug the server reads from the auth context; mismatches fail with 400.
func (c *OffboardingAdminClient) RequestOffboarding(
	ctx context.Context,
	req OffboardingRequestCreate,
) (*OffboardingRequest, error) {
	var out OffboardingRequest
	headers := map[string]string{}
	if req.IdempotencyKey != "" {
		headers["Idempotency-Key"] = req.IdempotencyKey
	}
	err := c.admin.requestWithHeaders(ctx, http.MethodPost, "/offboarding", req, &out, headers)
	return &out, err
}

func (c *OffboardingAdminClient) GetRequest(ctx context.Context, requestUUID string) (*OffboardingRequest, error) {
	return adminGetByID[OffboardingRequest](ctx, c.admin, "/offboarding/", requestUUID)
}

func (c *OffboardingAdminClient) CancelRequest(
	ctx context.Context,
	requestUUID string,
	req OffboardingCancelRequest,
) (*OffboardingRequest, error) {
	var out OffboardingRequest
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/offboarding/"+url.PathEscape(requestUUID)+"/cancel",
		req,
		&out,
	)
	return &out, err
}

func (c *OffboardingAdminClient) ConfirmRequest(
	ctx context.Context,
	requestUUID string,
) (*OffboardingRequest, error) {
	var out OffboardingRequest
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/offboarding/"+url.PathEscape(requestUUID)+"/confirm",
		nil,
		&out,
	)
	return &out, err
}

// OffboardingPreviewStore is one row of the server-computed preview
// inventory. EstimatedCount is server-computed; the SDK must not re-derive it.
type OffboardingPreviewStore struct {
	Store           string `json:"store"`
	Kind            string `json:"kind"`
	RetentionClass  string `json:"retention_class"`
	EstimatedCount  int64  `json:"estimated_count"`
	SourceAuthority string `json:"source_authority,omitempty"`
}

// OffboardingPreviewExclusion identifies a known store omitted from the
// preview and explains why it could not be classified.
type OffboardingPreviewExclusion struct {
	Store  string `json:"store"`
	Reason string `json:"reason"`
}

// OffboardingPreviewResponse is the body for POST
// /admin/offboarding/requests/{requestUuid}/preview.
type OffboardingPreviewResponse struct {
	RequestUUID            string                        `json:"requestUuid"`
	GeneratedAt            string                        `json:"generatedAt"`
	ExpiresAt              string                        `json:"expiresAt"`
	Stores                 []OffboardingPreviewStore     `json:"stores"`
	Exclusions             []OffboardingPreviewExclusion `json:"exclusions,omitempty"`
	PreviewInventoryDigest string                        `json:"previewInventoryDigest"`
	Complete               bool                          `json:"complete"`
	Partial                bool                          `json:"partial"`
}

// OffboardingWaiver is the typed waiver echoed in an offboarding receipt.
// Authorization and approval are server-owned; callers cannot submit waiver
// metadata to the execute endpoint.
type OffboardingWaiver struct {
	Role      string `json:"role"`
	Timestamp string `json:"timestamp"`
}

// OffboardingExportResponse is the body for POST
// /admin/offboarding/requests/{requestUuid}/export.
type OffboardingExportResponse struct {
	RequestUUID            string `json:"requestUuid"`
	ChecksumSHA256         string `json:"checksumSha256"`
	ByteSize               int64  `json:"byteSize"`
	RecordCount            int    `json:"recordCount"`
	GeneratedAt            string `json:"generatedAt"`
	ExpiresAt              string `json:"expiresAt"`
	PreviewInventoryDigest string `json:"previewInventoryDigest"`
}

// OffboardingDownload describes the bundle Download wrote and the integrity it verified: the
// declared length and SHA-256 both matched the bytes written.
type OffboardingDownload struct {
	ContentType    string
	ChecksumSHA256 string
	ByteSize       int64
	Metadata       ResponseMetadata
}

// OffboardingAcknowledgeResponse is the body for POST
// /admin/offboarding/requests/{requestUuid}/acknowledge.
type OffboardingAcknowledgeResponse = OffboardingRequest

// OffboardingExecuteResponse is the content-free receipt returned by execute.
type OffboardingExecuteResponse = OffboardingReceiptResponse

// OffboardingRetryResponse is the content-free receipt returned by retry.
type OffboardingRetryResponse = OffboardingReceiptResponse

// OffboardingReceiptPerStore is one row of the receipt's per-store summary.
// DeletedCount is server-issued; RetainedExceptionsCount covers legal holds
// and equivalent exclusions the SDK must not collapse.
type OffboardingReceiptPerStore struct {
	Store                   string `json:"store"`
	RetentionClass          string `json:"retention_class"`
	DeletedCount            int64  `json:"deleted_count"`
	RetainedExceptionsCount int64  `json:"retained_exceptions_count"`
}

// OffboardingReceiptResponse is the body for GET
// /admin/offboarding/requests/{requestUuid}/receipt. FinalState is the
// terminal state of the request; SHA256 is the unkeyed integrity checksum the
// client must store alongside its offboarding record.
type OffboardingReceiptResponse struct {
	CompanyID       int64                        `json:"company_id"`
	RequestedByKind string                       `json:"requested_by_kind"`
	RequestedAt     string                       `json:"requested_at"`
	CompletedAt     string                       `json:"completed_at"`
	FinalState      OffboardingState             `json:"final_state"`
	PerStore        []OffboardingReceiptPerStore `json:"per_store"`
	Waiver          *OffboardingWaiver           `json:"waiver,omitempty"`
	SHA256          string                       `json:"sha256"`
}

// Preview asks the server to compute the per-store inventory estimate for
// the offboarding request. The result is server-issued and must be surfaced
// verbatim; the SDK must not re-derive EstimatedCount.
func (c *OffboardingAdminClient) Preview(
	ctx context.Context,
	requestUUID string,
) (*OffboardingPreviewResponse, error) {
	var out OffboardingPreviewResponse
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/offboarding/requests/"+url.PathEscape(requestUUID)+"/preview",
		nil,
		&out,
	)
	return &out, err
}

// Export triggers the destructive export packaging for a request. The
// response is the per-request artifact metadata; the download URL is
// fetched separately via Download.
func (c *OffboardingAdminClient) Export(
	ctx context.Context,
	requestUUID string,
) (*OffboardingExportResponse, error) {
	var out OffboardingExportResponse
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/offboarding/requests/"+url.PathEscape(requestUUID)+"/export",
		nil,
		&out,
	)
	return &out, err
}

// Download streams the authenticated export bundle into w, hashing it as it copies. It refuses a
// response without valid Content-Length and X-Checksum-SHA256 headers before writing anything,
// and fails when the bytes written do not match them. On error w may hold a partial or
// unverified bundle, which the caller must discard.
func (c *OffboardingAdminClient) Download(
	ctx context.Context, requestUUID string, w io.Writer,
) (*OffboardingDownload, error) {
	path := "/offboarding/requests/" + url.PathEscape(requestUUID) + "/download"
	download, err := c.admin.openAdminDownload(ctx, path, "offboarding download")
	if err != nil {
		return nil, err
	}
	result, err := writeVerifiedOffboardingDownload(download, w)
	if closeErr := download.body.Close(); err == nil && closeErr != nil {
		return nil, fmt.Errorf("custd: close offboarding download: %w", closeErr)
	}
	return result, err
}

func writeVerifiedOffboardingDownload(download *adminDownload, w io.Writer) (*OffboardingDownload, error) {
	declared, err := strconv.ParseInt(strings.TrimSpace(download.header.Get("Content-Length")), 10, 64)
	if err != nil || declared < 0 {
		return nil, errors.New("custd: offboarding download content length is invalid")
	}
	checksum := strings.ToLower(strings.TrimSpace(download.header.Get("X-Checksum-SHA256")))
	if decoded, err := hex.DecodeString(checksum); err != nil || len(decoded) != sha256.Size {
		return nil, errors.New("custd: offboarding download checksum header is invalid")
	}
	hash := sha256.New()
	written, err := io.Copy(io.MultiWriter(w, hash), io.LimitReader(download.body, declared+1))
	if err != nil {
		return nil, fmt.Errorf("custd: stream offboarding download: %w", err)
	}
	if written != declared {
		return nil, errors.New("custd: offboarding download content length mismatch")
	}
	if checksum != hex.EncodeToString(hash.Sum(nil)) {
		return nil, errors.New("custd: offboarding download checksum mismatch")
	}
	return &OffboardingDownload{
		ContentType: download.header.Get("Content-Type"), ChecksumSHA256: checksum,
		ByteSize: written, Metadata: download.metadata,
	}, nil
}

// Acknowledge records that the export was downloaded successfully and its
// inventory was confirmed. It must not be called merely after Preview.
func (c *OffboardingAdminClient) Acknowledge(
	ctx context.Context,
	requestUUID string,
) (*OffboardingAcknowledgeResponse, error) {
	var out OffboardingAcknowledgeResponse
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/offboarding/requests/"+url.PathEscape(requestUUID)+"/acknowledge",
		nil,
		&out,
	)
	return &out, err
}

// Execute triggers the destructive phase. Authorization and approval are
// server-owned; callers cannot submit waiver metadata.
func (c *OffboardingAdminClient) Execute(
	ctx context.Context,
	requestUUID string,
) (*OffboardingExecuteResponse, error) {
	var out OffboardingExecuteResponse
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/offboarding/requests/"+url.PathEscape(requestUUID)+"/execute",
		nil,
		&out,
	)
	return &out, err
}

// Retry re-arms an offboarding request that previously failed. The server
// decides whether the request is retryable; the SDK does not pre-filter.
func (c *OffboardingAdminClient) Retry(
	ctx context.Context,
	requestUUID string,
) (*OffboardingRetryResponse, error) {
	var out OffboardingRetryResponse
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/offboarding/requests/"+url.PathEscape(requestUUID)+"/retry",
		nil,
		&out,
	)
	return &out, err
}

// Receipt returns the terminal offboarding receipt for a request. The
// SHA256 is an unkeyed integrity checksum the client must retain alongside
// its offboarding record; it is not an authenticity signature.
func (c *OffboardingAdminClient) Receipt(
	ctx context.Context,
	requestUUID string,
) (*OffboardingReceiptResponse, error) {
	var out OffboardingReceiptResponse
	err := c.admin.request(
		ctx,
		http.MethodGet,
		"/offboarding/requests/"+url.PathEscape(requestUUID)+"/receipt",
		nil,
		&out,
	)
	return &out, err
}

func auditListParams(opts AuditListOptions) string {
	params := auditFilterParams(opts)
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Cursor != "" {
		params.Set("cursor", opts.Cursor)
	}
	return params.Encode()
}

func auditExportParams(opts AuditListOptions) url.Values {
	return auditFilterParams(opts)
}

func auditFilterParams(opts AuditListOptions) url.Values {
	params := url.Values{}
	if opts.Scope != "" {
		params.Set("scope", string(opts.Scope))
	}
	if opts.CompanySlug != "" {
		params.Set("companySlug", opts.CompanySlug)
	}
	if opts.AffectedTenantSlug != "" {
		params.Set("affectedTenantSlug", opts.AffectedTenantSlug)
	}
	if opts.Since != "" {
		params.Set("since", opts.Since)
	}
	if opts.Until != "" {
		params.Set("until", opts.Until)
	}
	if opts.ActorKind != "" {
		params.Set("actorKind", opts.ActorKind)
	}
	if opts.ActorReference != "" {
		params.Set("actorReference", opts.ActorReference)
	}
	if opts.Action != "" {
		params.Set("action", opts.Action)
	}
	if opts.ResourceType != "" {
		params.Set("resourceType", opts.ResourceType)
	}
	if opts.ResourceID != "" {
		params.Set("resourceId", opts.ResourceID)
	}
	if opts.Outcome != "" {
		params.Set("outcome", string(opts.Outcome))
	}
	if opts.CorrelationID != "" {
		params.Set("correlationId", opts.CorrelationID)
	}
	if opts.EnvironmentID != "" {
		params.Set("environmentId", opts.EnvironmentID)
	}
	return params
}

func auditLookupParams(opts AuditListOptions) string {
	params := url.Values{}
	if opts.Scope != "" {
		params.Set("scope", string(opts.Scope))
	}
	if opts.CompanySlug != "" {
		params.Set("companySlug", opts.CompanySlug)
	}
	return params.Encode()
}

var auditEventUUIDPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func validateAuditEventUUID(value string) error {
	if !auditEventUUIDPattern.MatchString(value) {
		return fmt.Errorf("custd: audit event eventId must be a UUID")
	}
	return nil
}
