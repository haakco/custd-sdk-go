package custd

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

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
// pre-fill TenantSlug on the request.
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

type OffboardingRequest struct {
	RequestUUID string `json:"requestUuid"`
	TenantSlug  string `json:"tenantSlug"`
	Status      string `json:"status"`
	RequestedBy string `json:"requestedBy"`
	RequestedAt string `json:"requestedAt,omitempty"`
}

func (c *OffboardingAdminClient) Schedule(
	ctx context.Context,
	tenantSlug string,
	req OffboardingScheduleRequest,
) (*OffboardingSchedule, error) {
	var out OffboardingSchedule
	err := c.admin.request(
		ctx,
		http.MethodPut,
		"/offboarding/schedules/"+url.PathEscape(tenantSlug),
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
