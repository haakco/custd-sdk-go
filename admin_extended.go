package custd

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// TenantStorageAdminClient owns tenant-scoped storage location registration.
// Locations are server-prefixed: the SDK submits clientLocation and the
// server returns a serverAssignedPrefix that the SDK must use for raw
// landing writes. Tenant is derived from the auth context; wrong-tenant
// reads collapse to an empty list indistinguishable from "no locations".
type TenantStorageAdminClient struct {
	admin *AdminClient
}

// TenantStorageLocation is the per-tenant storage entry. ClientLocation is
// the client-supplied bucket URI; ServerAssignedPrefix is the writable
// prefix the server mints on the tenant's behalf.
type TenantStorageLocation struct {
	ID                   string `json:"id"`
	TenantSlug           string `json:"tenantSlug"`
	ClientLocation       string `json:"clientLocation"`
	ServerAssignedPrefix string `json:"serverAssignedPrefix"`
	Status               string `json:"status"`
	CreatedAt            string `json:"createdAt,omitempty"`
	ExpiresAt            string `json:"expiresAt,omitempty"`
}

// TenantStorageListResponse is the body for GET /tenant-storage-locations.
type TenantStorageListResponse struct {
	Locations []TenantStorageLocation `json:"locations"`
}

// TenantStorageCreateRequest is the body for POST /tenant-storage-locations.
// Tenant is server-derived; callers must not pre-fill TenantSlug.
type TenantStorageCreateRequest struct {
	TenantSlug     string `json:"tenantSlug,omitempty"`
	ClientLocation string `json:"clientLocation"`
}

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
// download surface returns a short-lived signed URL the SDK must surface
// only to the caller; it must not be logged or echoed into error messages.
type SubjectExportAdminClient struct {
	admin *AdminClient
}

// SubjectExportSubject is the typed selector the server returns alongside
// an export request. The value is a server-side identifier (e.g. a user
// UUID); the SDK must not echo it into logs or error messages.
type SubjectExportSubject struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// SubjectExport is the receipt returned for a subject export request. The
// Checksum and ArtifactSize are present only once the request is in
// terminal ready state.
type SubjectExport struct {
	RequestID    string               `json:"requestId"`
	TenantSlug   string               `json:"tenantSlug"`
	Subject      SubjectExportSubject `json:"subject"`
	Scope        string               `json:"scope"`
	State        string               `json:"state"`
	CreatedAt    string               `json:"createdAt,omitempty"`
	ExpiresAt    string               `json:"expiresAt,omitempty"`
	Checksum     string               `json:"checksum,omitempty"`
	ArtifactSize int64                `json:"artifactSize,omitempty"`
}

// SubjectExportListResponse is the body for GET /admin/subject-exports.
type SubjectExportListResponse struct {
	Exports []SubjectExport `json:"exports"`
}

// SubjectExportCreateRequest is the body for POST /admin/subject-exports.
// IdempotencyKey is required for safe retries.
type SubjectExportCreateRequest struct {
	TenantSlug     string               `json:"tenantSlug"`
	Subject        SubjectExportSubject `json:"subject"`
	Scope          string               `json:"scope"`
	IdempotencyKey string               `json:"idempotencyKey"`
}

// SubjectExportDownloadResponse is the body for GET
// /admin/subject-exports/{requestId}/download. The DownloadURL is a
// short-lived signed URL the SDK must hand back to the caller without
// logging the URL value or the underlying subject identifier.
type SubjectExportDownloadResponse struct {
	RequestID   string `json:"requestId"`
	DownloadURL string `json:"downloadUrl"`
	ExpiresAt   string `json:"expiresAt,omitempty"`
}

func (c *SubjectExportAdminClient) Create(
	ctx context.Context,
	req SubjectExportCreateRequest,
) (*SubjectExport, error) {
	var out SubjectExport
	err := c.admin.request(ctx, http.MethodPost, "/subject-exports", req, &out)
	return &out, err
}

func (c *SubjectExportAdminClient) List(ctx context.Context) (*SubjectExportListResponse, error) {
	var out SubjectExportListResponse
	err := c.admin.request(ctx, http.MethodGet, "/subject-exports", nil, &out)
	return &out, err
}

func (c *SubjectExportAdminClient) Get(ctx context.Context, requestID string) (*SubjectExport, error) {
	return adminGetByID[SubjectExport](ctx, c.admin, "/subject-exports/", requestID)
}

func (c *SubjectExportAdminClient) Cancel(ctx context.Context, requestID string) error {
	return c.admin.request(
		ctx,
		http.MethodPost,
		"/subject-exports/"+url.PathEscape(requestID)+"/cancel",
		nil,
		nil,
	)
}

// Download returns a short-lived signed URL. The DownloadURL field is
// sensitive; callers must not log the URL or echo it into error messages.
func (c *SubjectExportAdminClient) Download(
	ctx context.Context,
	requestID string,
) (*SubjectExportDownloadResponse, error) {
	var out SubjectExportDownloadResponse
	err := c.admin.request(
		ctx,
		http.MethodGet,
		"/subject-exports/"+url.PathEscape(requestID)+"/download",
		nil,
		&out,
	)
	return &out, err
}

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
// a subject. The value is server-side identifier; do not log it.
type PrivacyErasureSelector struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// PrivacyErasureStoreProgress tracks per-store progress of an erasure.
// State==retained is terminal for the legal_hold store and means the row
// must not be deleted; callers must surface this verbatim.
type PrivacyErasureStoreProgress struct {
	Store        string `json:"store"`
	State        string `json:"state"`
	DeletedCount int    `json:"deletedCount,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

// PrivacyErasure is the receipt returned for an erasure request.
type PrivacyErasure struct {
	RequestUUID      string                        `json:"requestUuid"`
	TenantSlug       string                        `json:"tenantSlug"`
	Selector         PrivacyErasureSelector        `json:"selector"`
	State            string                        `json:"state"`
	PerStoreProgress []PrivacyErasureStoreProgress `json:"perStoreProgress,omitempty"`
	CreatedAt        string                        `json:"createdAt,omitempty"`
	CompletedAt      string                        `json:"completedAt,omitempty"`
}

// PrivacyErasureCreateRequest is the body for POST /admin/privacy/erasures.
type PrivacyErasureCreateRequest struct {
	TenantSlug string                 `json:"tenantSlug"`
	Selector   PrivacyErasureSelector `json:"selector"`
	Reason     string                 `json:"reason"`
}

// PrivacyErasureListResponse is the body for GET /admin/privacy/erasures.
type PrivacyErasureListResponse struct {
	Erasures []PrivacyErasure `json:"erasures"`
}

func (c *PrivacyErasureAdminClient) Create(
	ctx context.Context,
	req PrivacyErasureCreateRequest,
) (*PrivacyErasure, error) {
	var out PrivacyErasure
	err := c.admin.request(ctx, http.MethodPost, "/privacy/erasures", req, &out)
	return &out, err
}

func (c *PrivacyErasureAdminClient) List(ctx context.Context) (*PrivacyErasureListResponse, error) {
	var out PrivacyErasureListResponse
	err := c.admin.request(ctx, http.MethodGet, "/privacy/erasures", nil, &out)
	return &out, err
}

func (c *PrivacyErasureAdminClient) Get(ctx context.Context, requestUUID string) (*PrivacyErasure, error) {
	return adminGetByID[PrivacyErasure](ctx, c.admin, "/privacy/erasures/", requestUUID)
}

func (c *PrivacyErasureAdminClient) Force(ctx context.Context, requestUUID string) (*PrivacyErasure, error) {
	var out PrivacyErasure
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

type PrivacyRule struct {
	TenantSlug          string   `json:"tenantSlug,omitempty"`
	Purposes            []string `json:"purposes"`
	HardDeleteAfterDays int      `json:"hardDeleteAfterDays,omitempty"`
}

type PrivacyRuleUpdate struct {
	Purposes            []string `json:"purposes"`
	HardDeleteAfterDays int      `json:"hardDeleteAfterDays,omitempty"`
}

type PrivacyRulesResponse struct {
	TenantSlug          string   `json:"tenantSlug"`
	Purposes            []string `json:"purposes"`
	HardDeleteAfterDays int      `json:"hardDeleteAfterDays"`
}

type PrivacyIdentifierMapRequest struct {
	// ExternalID is the plain identifier the SDK consumer already knows. It is
	// consumed once on the wire; do not log or echo the response payload for
	// this request back to a place where ExternalID could appear.
	ExternalID string `json:"externalId"`
}

type PrivacyIdentifierMapping struct {
	IdentifierID         string `json:"identifierId"`
	InternalIDHash       string `json:"internalIdHash"`
	InternalIDHashPrefix string `json:"internalIdHashPrefix"`
	SaltVersion          int    `json:"saltVersion"`
	CreatedAt            string `json:"createdAt,omitempty"`
}

// GetRules returns the privacy rules attached to the effective tenant.
func (c *PrivacyAdminClient) GetRules(ctx context.Context) (*PrivacyRulesResponse, error) {
	var out PrivacyRulesResponse
	err := c.admin.request(ctx, http.MethodGet, "/privacy/rules", nil, &out)
	return &out, err
}

// SetRules replaces the privacy rules for the effective tenant.
func (c *PrivacyAdminClient) SetRules(ctx context.Context, req PrivacyRuleUpdate) (*PrivacyRulesResponse, error) {
	var out PrivacyRulesResponse
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
) ([]PrivacyIdentifierMapping, error) {
	var out []PrivacyIdentifierMapping
	err := c.admin.request(
		ctx,
		http.MethodGet,
		"/privacy/identifiers/"+url.PathEscape(companySlug),
		nil,
		&out,
	)
	return out, err
}

// RetentionAdminClient owns per-tenant retention policies. Effective-tenant
// authority is enforced server-side; wrong-tenant requests return 404.
type RetentionAdminClient struct {
	admin *AdminClient
}

type RetentionPolicy struct {
	TenantSlug          string   `json:"tenantSlug"`
	MaxAgeDays          int      `json:"maxAgeDays"`
	HardDeleteAfterDays int      `json:"hardDeleteAfterDays"`
	ApplyToEventTypes   []string `json:"applyToEventTypes"`
	ApplyToDataSpaces   []string `json:"applyToDataSpaces"`
}

type RetentionPolicyUpsertRequest struct {
	MaxAgeDays          int      `json:"maxAgeDays"`
	HardDeleteAfterDays int      `json:"hardDeleteAfterDays"`
	ApplyToEventTypes   []string `json:"applyToEventTypes"`
	ApplyToDataSpaces   []string `json:"applyToDataSpaces"`
}

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

func (c *RetentionAdminClient) Get(ctx context.Context, tenantSlug string) (*RetentionPolicy, error) {
	return adminGetByID[RetentionPolicy](ctx, c.admin, "/retention/policies/", tenantSlug)
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

// RetentionRunDeletion is the per-store deletion estimate a preview returns.
// The Count is server-computed; the SDK must not infer it client-side.
type RetentionRunDeletion struct {
	Store string `json:"store"`
	Count int    `json:"count"`
}

// RetentionRunPreview is the body for POST /admin/retention/policies/{slug}/preview.
// The PreviewId is server-issued; the SDK does not mint it.
type RetentionRunPreview struct {
	PreviewID          string                 `json:"previewId"`
	TenantSlug         string                 `json:"tenantSlug"`
	EstimatedDeletions []RetentionRunDeletion `json:"estimatedDeletions"`
	PreviewedAt        string                 `json:"previewedAt,omitempty"`
}

// RetentionRun is the body element for GET /admin/retention/policies/{slug}/runs.
// CompletedAt is empty while the run is in flight.
type RetentionRun struct {
	RunID        string `json:"runId"`
	TenantSlug   string `json:"tenantSlug"`
	State        string `json:"state"`
	StartedAt    string `json:"startedAt,omitempty"`
	CompletedAt  string `json:"completedAt,omitempty"`
	DeletedCount int    `json:"deletedCount,omitempty"`
}

// RetentionRunsListResponse is the body for GET /admin/retention/policies/{slug}/runs.
type RetentionRunsListResponse struct {
	Runs []RetentionRun `json:"runs"`
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
) (*RetentionRun, error) {
	var out RetentionRun
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/retention/policies/"+url.PathEscape(tenantSlug)+"/apply",
		nil,
		&out,
	)
	return &out, err
}

// ListRuns returns the retention runs for a single tenant. Empty runs list
// is the canonical "no runs yet" response, not an error.
func (c *RetentionAdminClient) ListRuns(
	ctx context.Context,
	tenantSlug string,
) (*RetentionRunsListResponse, error) {
	var out RetentionRunsListResponse
	err := c.admin.request(
		ctx,
		http.MethodGet,
		"/retention/policies/"+url.PathEscape(tenantSlug)+"/runs",
		nil,
		&out,
	)
	return &out, err
}

// StorageAlertAdminClient owns tenant-scoped storage alert rules. The list and
// delete surfaces are tenant-safe (effective tenant collapsed to 404).
type StorageAlertAdminClient struct {
	admin *AdminClient
}

type StorageAlertRule struct {
	RuleID           string `json:"ruleId"`
	TenantSlug       string `json:"tenantSlug"`
	Metric           string `json:"metric"`
	ThresholdPercent int    `json:"thresholdPercent"`
	Channel          string `json:"channel"`
	Enabled          bool   `json:"enabled"`
	CreatedAt        string `json:"createdAt,omitempty"`
	UpdatedAt        string `json:"updatedAt,omitempty"`
}

type StorageAlertRuleCreateRequest struct {
	Metric           string `json:"metric"`
	ThresholdPercent int    `json:"thresholdPercent"`
	Channel          string `json:"channel"`
	Enabled          bool   `json:"enabled"`
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

type AuditEvent struct {
	EventID      string `json:"eventId"`
	Action       string `json:"action"`
	ActorID      string `json:"actorId"`
	ActorKind    string `json:"actorKind"`
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId"`
	IPAddress    string `json:"ipAddress"`
	Metadata     string `json:"metadata,omitempty"`
	CreatedAt    string `json:"createdAt"`
}

type AuditListCursor struct {
	Cursor string `json:"cursor"`
}

type AuditListResponse struct {
	Events     []AuditEvent     `json:"events"`
	NextCursor *AuditListCursor `json:"nextCursor"`
}

type AuditListOptions struct {
	ResourceType string
	ResourceID   string
	Limit        int
	Cursor       string
}

type ReportingPackAuditEvent struct {
	Action       string `json:"action"`
	ActorID      string `json:"actorId"`
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId"`
	PackKey      string `json:"packKey"`
	CreatedAt    string `json:"createdAt"`
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
	return &out, err
}

func (c *AuditAdminClient) GetEvent(ctx context.Context, eventID string) (*AuditEvent, error) {
	return adminGetByID[AuditEvent](ctx, c.admin, "/audit/events/", eventID)
}

func (c *AuditAdminClient) ListReportingPackEvents(ctx context.Context) (*ReportingPackAuditListResponse, error) {
	var out ReportingPackAuditListResponse
	err := c.admin.request(ctx, http.MethodGet, "/reporting-packs/audit-events", nil, &out)
	return &out, err
}

// OffboardingAdminClient owns the offboarding schedule and one-off request
// surfaces. Schedule writes the effective tenant server-side; callers must not
// pre-fill TenantSlug on the request body. The tenant is derived from the
// authenticated client context.
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
	EffectiveAt     string `json:"effectiveAt"`
	GracePeriodDays int    `json:"gracePeriodDays"`
	Reason          string `json:"reason"`
	Status          string `json:"status"`
}

type OffboardingScheduleListResponse struct {
	Schedules []OffboardingSchedule `json:"schedules"`
}

type OffboardingCancelRequest struct {
	Reason string `json:"reason"`
}

// OffboardingRequest is the receipt returned for one-off offboarding requests.
// It is the response shape for RequestOffboarding, GetRequest, and the
// single-tenant collection read.
type OffboardingRequest struct {
	RequestUUID string `json:"requestUuid"`
	TenantSlug  string `json:"tenantSlug"`
	Status      string `json:"status"`
	RequestedBy string `json:"requestedBy"`
	RequestedAt string `json:"requestedAt,omitempty"`
}

// OffboardingRequestCreate carries the body for POST /offboarding. Confirmation
// is the human-typed string the server compares against the tenant slug before
// accepting the destructive transition.
type OffboardingRequestCreate struct {
	Confirmation string `json:"confirmation"`
}

// Schedule writes a delayed offboarding schedule for the effective tenant.
// The server pulls the tenant from the auth context; do not include TenantSlug
// in the request body. The collection endpoint is POST /offboarding/schedules.
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
) error {
	return c.admin.request(
		ctx,
		http.MethodPost,
		"/offboarding/schedules/"+url.PathEscape(tenantSlug)+"/cancel",
		req,
		nil,
	)
}

// RequestOffboarding submits a one-off offboarding request for the effective
// tenant via POST /offboarding. The Confirmation field must match the tenant
// slug the server reads from the auth context; mismatches fail with 400.
func (c *OffboardingAdminClient) RequestOffboarding(
	ctx context.Context,
	req OffboardingRequestCreate,
) (*OffboardingRequest, error) {
	var out OffboardingRequest
	err := c.admin.request(ctx, http.MethodPost, "/offboarding", req, &out)
	return &out, err
}

func (c *OffboardingAdminClient) GetRequest(ctx context.Context, requestUUID string) (*OffboardingRequest, error) {
	return adminGetByID[OffboardingRequest](ctx, c.admin, "/offboarding/", requestUUID)
}

func (c *OffboardingAdminClient) CancelRequest(ctx context.Context, requestUUID string) error {
	return c.admin.request(
		ctx,
		http.MethodPost,
		"/offboarding/"+url.PathEscape(requestUUID)+"/cancel",
		nil,
		nil,
	)
}

func (c *OffboardingAdminClient) ConfirmRequest(ctx context.Context, requestUUID string) error {
	return c.admin.request(
		ctx,
		http.MethodPost,
		"/offboarding/"+url.PathEscape(requestUUID)+"/confirm",
		nil,
		nil,
	)
}

// OffboardingPerStore is one row of the per-store inventory the preview
// endpoint returns. EstimatedCount is server-computed; the SDK must not
// re-derive it.
type OffboardingPerStore struct {
	Store          string `json:"store"`
	Kind           string `json:"kind"`
	RetentionClass string `json:"retention_class"`
	EstimatedCount int    `json:"estimated_count"`
}

// OffboardingPreviewResponse is the body for POST
// /admin/offboarding/requests/{requestUuid}/preview.
type OffboardingPreviewResponse struct {
	RequestUUID            string                `json:"requestUuid"`
	PreviewInventoryDigest string                `json:"previewInventoryDigest,omitempty"`
	PerStore               []OffboardingPerStore `json:"perStore"`
}

// OffboardingWaiver is the typed waiver the execute endpoint requires.
// Role identifies the actor (e.g. client_owner); Reason is the human-readable
// rationale. Timestamp is server-stamped on accept.
type OffboardingWaiver struct {
	Role      string `json:"role"`
	Reason    string `json:"reason"`
	Timestamp string `json:"timestamp,omitempty"`
}

// OffboardingExecuteRequest is the body for POST
// /admin/offboarding/requests/{requestUuid}/execute. Waiver is required for
// destructive execution; an empty Role returns a 400 waiver_required error
// the SDK must surface without retry.
type OffboardingExecuteRequest struct {
	Waiver OffboardingWaiver `json:"waiver"`
}

// OffboardingExportResponse is the body for POST
// /admin/offboarding/requests/{requestUuid}/export. Complete=false means
// the server is still gathering inventory; callers must poll.
type OffboardingExportResponse struct {
	RequestUUID      string `json:"requestUuid"`
	ExportArtifactID string `json:"exportArtifactId,omitempty"`
	SchemaVersion    string `json:"schemaVersion,omitempty"`
	GeneratedAt      string `json:"generatedAt,omitempty"`
	ExpiresAt        string `json:"expiresAt,omitempty"`
	Complete         bool   `json:"complete"`
	Checksum         string `json:"checksum,omitempty"`
}

// OffboardingDownloadResponse is the body for GET
// /admin/offboarding/requests/{requestUuid}/download. The DownloadURL is
// short-lived; callers must not log it or echo it into error messages.
type OffboardingDownloadResponse struct {
	RequestUUID string `json:"requestUuid"`
	DownloadURL string `json:"downloadUrl"`
	ExpiresAt   string `json:"expiresAt,omitempty"`
}

// OffboardingAcknowledgeResponse is the body for POST
// /admin/offboarding/requests/{requestUuid}/acknowledge.
type OffboardingAcknowledgeResponse struct {
	RequestUUID    string `json:"requestUuid"`
	State          string `json:"state,omitempty"`
	AcknowledgedAt string `json:"acknowledgedAt,omitempty"`
}

// OffboardingExecuteResponse is the body for POST
// /admin/offboarding/requests/{requestUuid}/execute. The Waiver is echoed
// back with the server-stamped timestamp.
type OffboardingExecuteResponse struct {
	RequestUUID string            `json:"requestUuid"`
	State       string            `json:"state,omitempty"`
	ExecutedAt  string            `json:"executedAt,omitempty"`
	Waiver      OffboardingWaiver `json:"waiver,omitempty"`
}

// OffboardingRetryResponse is the body for POST
// /admin/offboarding/requests/{requestUuid}/retry.
type OffboardingRetryResponse struct {
	RequestUUID string `json:"requestUuid"`
	State       string `json:"state,omitempty"`
	RetriedAt   string `json:"retriedAt,omitempty"`
}

// OffboardingReceiptPerStore is one row of the receipt's per-store summary.
// DeletedCount is server-issued; RetainedExceptionsCount covers legal holds
// and equivalent exclusions the SDK must not collapse.
type OffboardingReceiptPerStore struct {
	Store                   string `json:"store"`
	RetentionClass          string `json:"retention_class"`
	DeletedCount            int    `json:"deleted_count"`
	RetainedExceptionsCount int    `json:"retained_exceptions_count"`
}

// OffboardingReceiptResponse is the body for GET
// /admin/offboarding/requests/{requestUuid}/receipt. FinalState is the
// terminal state of the request; SHA256 is the signed digest the client
// must store alongside its offboarding record.
type OffboardingReceiptResponse struct {
	RequestUUID       string                       `json:"requestUuid"`
	TenantSlug        string                       `json:"tenantSlug"`
	FinalState        string                       `json:"finalState"`
	RequestedByUserID string                       `json:"requestedByUserId,omitempty"`
	RequestedAt       string                       `json:"requestedAt,omitempty"`
	CompletedAt       string                       `json:"completedAt,omitempty"`
	PerStore          []OffboardingReceiptPerStore `json:"perStore"`
	Waiver            *OffboardingWaiver           `json:"waiver,omitempty"`
	SHA256            string                       `json:"sha256,omitempty"`
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

// Download returns a short-lived signed URL for the offboarding export
// artifact. The DownloadURL is sensitive; callers must not log it or echo
// it into error messages.
func (c *OffboardingAdminClient) Download(
	ctx context.Context,
	requestUUID string,
) (*OffboardingDownloadResponse, error) {
	var out OffboardingDownloadResponse
	err := c.admin.request(
		ctx,
		http.MethodGet,
		"/offboarding/requests/"+url.PathEscape(requestUUID)+"/download",
		nil,
		&out,
	)
	return &out, err
}

// Acknowledge records that the operator (or client) has accepted the
// preview. After acknowledgment the server is willing to accept Execute.
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

// Execute triggers the destructive phase. The server requires a non-empty
// Waiver.Role; an empty waiver returns 400 waiver_required, which the
// SDK surfaces without retry.
func (c *OffboardingAdminClient) Execute(
	ctx context.Context,
	requestUUID string,
	req OffboardingExecuteRequest,
) (*OffboardingExecuteResponse, error) {
	var out OffboardingExecuteResponse
	err := c.admin.request(
		ctx,
		http.MethodPost,
		"/offboarding/requests/"+url.PathEscape(requestUUID)+"/execute",
		req,
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
// SHA256 digest is the signed evidence the client must retain alongside
// its offboarding record.
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
	params := url.Values{}
	if opts.ResourceType != "" {
		params.Set("resourceType", opts.ResourceType)
	}
	if opts.ResourceID != "" {
		params.Set("resourceId", opts.ResourceID)
	}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Cursor != "" {
		params.Set("cursor", opts.Cursor)
	}
	return params.Encode()
}
