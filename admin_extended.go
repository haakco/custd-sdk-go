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
	"strconv"
	"strings"
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

// OffboardingRequest is the public request state returned by RequestOffboarding,
// GetRequest, and the acknowledgement endpoint.
type OffboardingRequest struct {
	RequestUUID string `json:"requestUuid"`
	State       string `json:"state"`
	RequestedAt string `json:"requestedAt"`
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
	Reason    string `json:"reason"`
	Timestamp string `json:"timestamp,omitempty"`
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

const maxOffboardingDownloadBytes int64 = 64 << 20

// OffboardingDownloadResponse contains authenticated export bytes and the
// integrity metadata verified against the response headers.
type OffboardingDownloadResponse struct {
	Bytes          []byte
	ChecksumSHA256 string
	ByteSize       int64
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
	CompanyID         int64                        `json:"company_id"`
	RequestedByUserID *int64                       `json:"requested_by_user_id,omitempty"`
	RequestedByActor  string                       `json:"requested_by_actor"`
	RequestedAt       string                       `json:"requested_at"`
	CompletedAt       string                       `json:"completed_at"`
	FinalState        string                       `json:"final_state"`
	PerStore          []OffboardingReceiptPerStore `json:"per_store"`
	Waiver            *OffboardingWaiver           `json:"waiver,omitempty"`
	SHA256            string                       `json:"sha256"`
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

// Download returns the authenticated export bytes. It rejects missing or
// inconsistent integrity headers and responses larger than 64 MiB.
func (c *OffboardingAdminClient) Download(
	ctx context.Context,
	requestUUID string,
) (*OffboardingDownloadResponse, error) {
	path := "/offboarding/requests/" + url.PathEscape(requestUUID) + "/download"
	if c.admin.client.config.HTTPClient != nil {
		return c.downloadViaDoer(path)
	}
	return c.downloadViaHTTP(ctx, path)
}

func (c *OffboardingAdminClient) downloadViaDoer(path string) (*OffboardingDownloadResponse, error) {
	resp, err := c.admin.client.config.HTTPClient.Do(&HTTPRequest{
		Method: http.MethodGet, URL: c.admin.endpoint(path), Headers: c.admin.client.headers(false),
	})
	if err != nil {
		return nil, fmt.Errorf("custd: offboarding download failed: %w", err)
	}
	if err := c.admin.client.checkStatus(resp.StatusCode, resp.Body); err != nil {
		return nil, err
	}
	return verifiedOffboardingDownload(resp.Body, responseHeader(resp.Headers, "Content-Length"), responseHeader(resp.Headers, "X-Checksum-SHA256"))
}

func (c *OffboardingAdminClient) downloadViaHTTP(ctx context.Context, path string) (*OffboardingDownloadResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.admin.endpoint(path), nil)
	if err != nil {
		return nil, fmt.Errorf("custd: create offboarding download: %w", err)
	}
	for key, value := range c.admin.client.headers(false) {
		req.Header.Set(key, value)
	}
	resp, err := c.admin.client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("custd: offboarding download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.ContentLength > maxOffboardingDownloadBytes {
		return nil, errors.New("custd: offboarding download exceeds 64 MiB")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxOffboardingDownloadBytes+1))
	if err != nil {
		return nil, fmt.Errorf("custd: read offboarding download: %w", err)
	}
	if err := c.admin.client.checkStatus(resp.StatusCode, body); err != nil {
		return nil, err
	}
	return verifiedOffboardingDownload(body, resp.Header.Get("Content-Length"), resp.Header.Get("X-Checksum-SHA256"))
}

func verifiedOffboardingDownload(body []byte, lengthHeader, checksumHeader string) (*OffboardingDownloadResponse, error) {
	if int64(len(body)) > maxOffboardingDownloadBytes {
		return nil, errors.New("custd: offboarding download exceeds 64 MiB")
	}
	declared, err := strconv.ParseInt(strings.TrimSpace(lengthHeader), 10, 64)
	if err != nil || declared < 0 {
		return nil, errors.New("custd: offboarding download content length is invalid")
	}
	if declared > maxOffboardingDownloadBytes {
		return nil, errors.New("custd: offboarding download exceeds 64 MiB")
	}
	checksum := strings.ToLower(strings.TrimSpace(checksumHeader))
	if decoded, err := hex.DecodeString(checksum); err != nil || len(decoded) != sha256.Size {
		return nil, errors.New("custd: offboarding download checksum header is invalid")
	}
	actual := sha256.Sum256(body)
	if checksum != hex.EncodeToString(actual[:]) {
		return nil, errors.New("custd: offboarding download checksum mismatch")
	}
	if declared != int64(len(body)) {
		return nil, errors.New("custd: offboarding download content length mismatch")
	}
	return &OffboardingDownloadResponse{Bytes: body, ChecksumSHA256: checksum, ByteSize: declared}, nil
}

func responseHeader(headers map[string]string, name string) string {
	for key, value := range headers {
		if strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
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
