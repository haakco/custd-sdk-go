package custd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Platform and tenant operations: realm domain images, app lifecycle,
// tenant storage limits, consent records, storage and RLS state, event type
// and category settings, accepted raw-landing claims, error issues and managed
// audit provisioning. An empty companySlug uses the credential's own tenant.

// EventTypeAdminClient reads and changes event type and event category settings.
type EventTypeAdminClient struct {
	admin *AdminClient
}

// IngestClaimAdminClient inspects and repairs accepted raw-landing claims.
type IngestClaimAdminClient struct {
	admin *AdminClient
}

// RLSAdminClient reports the row-level security catalog and reconciles it
// through reviewed plans.
type RLSAdminClient struct {
	admin *AdminClient
}

// ErrorIssueAdminClient reads a tenant's grouped error issues and changes their state.
type ErrorIssueAdminClient struct {
	admin *AdminClient
}

// ManagedAuditAdminClient provisions the tenant's managed audit producer.
type ManagedAuditAdminClient struct {
	admin *AdminClient
}

// idempotencyKeyHeader states a retry identity so a repeated call is recognised
// rather than repeating the mutation.
func idempotencyKeyHeader(key string) map[string]string {
	return map[string]string{"Idempotency-Key": strings.TrimSpace(key)}
}

// HistoricalCleanup runs the deletion cleanup again for an already deleted
// tenant and returns its content-free receipt. companyID must name the same
// tenant as slug; the idempotency key makes a retry finish the same cleanup.
func (c *TenantAdminClient) HistoricalCleanup(
	ctx context.Context, slug string, companyID int64, idempotencyKey string,
) (*OffboardingReceiptResponse, error) {
	if companyID <= 0 {
		return nil, fmt.Errorf("custd: company id must be a positive integer")
	}
	if strings.TrimSpace(idempotencyKey) == "" {
		return nil, fmt.Errorf("custd: historical cleanup requires an idempotency key")
	}
	var out OffboardingReceiptResponse
	path := "/tenants/" + url.PathEscape(slug) + "/historical-cleanup?companyId=" + strconv.FormatInt(companyID, 10)
	err := c.admin.requestWithHeaders(ctx, http.MethodPost, path, nil, &out, idempotencyKeyHeader(idempotencyKey))
	return &out, err
}

// ErrorIssueStateFilter selects error issues by triage state; all lists every state.
type ErrorIssueStateFilter string

const (
	ErrorIssueStateFilterOpen     ErrorIssueStateFilter = "open"
	ErrorIssueStateFilterResolved ErrorIssueStateFilter = "resolved"
	ErrorIssueStateFilterIgnored  ErrorIssueStateFilter = "ignored"
	ErrorIssueStateFilterAll      ErrorIssueStateFilter = "all"
)

// ErrorIssueSort orders an error issue list.
type ErrorIssueSort string

const (
	ErrorIssueSortRecent   ErrorIssueSort = "recent"
	ErrorIssueSortFrequent ErrorIssueSort = "frequent"
)

// ErrorIssueListOptions filters and pages a tenant's error issues. Empty
// fields and a zero Limit use the server defaults.
type ErrorIssueListOptions struct {
	CompanySlug string
	SiteUUID    string
	State       ErrorIssueStateFilter
	Sort        ErrorIssueSort
	Limit       int64
	Cursor      string
}

type AuthDomainImageOverrideSetRequest struct {
	Component         string    `json:"component"`
	ExpiresAt         time.Time `json:"expiresAt"`
	ImageDigest       string    `json:"imageDigest"`
	IncidentReference string    `json:"incidentReference"`
}

type AuthDomainImageOverrideStatus struct {
	ActorKind         string     `json:"actorKind"`
	ActorSubject      string     `json:"actorSubject,omitempty"`
	Component         string     `json:"component"`
	EndedAt           *time.Time `json:"endedAt,omitempty"`
	EndedBy           string     `json:"endedBy,omitempty"`
	ExpiresAt         time.Time  `json:"expiresAt"`
	ImageDigest       string     `json:"imageDigest"`
	IncidentReference string     `json:"incidentReference"`
	ReceiptID         string     `json:"receiptId"`
	SetAt             time.Time  `json:"setAt"`
	State             string     `json:"state"`
}

type MintAuthDomainRequest struct {
	SiteDomainID string `json:"siteDomainId"`
}

type AppItemEnvironmentLink struct {
	CreatedAt time.Time `json:"createdAt"`
	// When the item expires unless it changes; its links stop working then.
	ItemExpiresAt time.Time   `json:"itemExpiresAt"`
	ItemUUID      string      `json:"itemUuid"`
	Kind          AppItemKind `json:"kind"`
	LinkUUID      string      `json:"linkUuid"`
	Role          AppItemRole `json:"role"`
}

type AppItemEnvironmentLinkList struct {
	Links []AppItemEnvironmentLink `json:"links"`
}

type AppDeployCredentialSummary struct {
	ClientID  string    `json:"clientId"`
	CreatedAt time.Time `json:"createdAt"`
}

type AppDeployCredentialList struct {
	Credentials []AppDeployCredentialSummary `json:"credentials"`
}

// AppSignInMode is who may sign in to an app.
type AppSignInMode string

const (
	AppSignInModeOff        AppSignInMode = "off"
	AppSignInModeInviteOnly AppSignInMode = "invite_only"
	AppSignInModeOpen       AppSignInMode = "open"
)

type AppSignInRequest struct {
	ExpectedRevision int64         `json:"expectedRevision"`
	Mode             AppSignInMode `json:"mode"`
}

type StorageLimitUpdateRequest struct {
	HardLimitEnabled      bool  `json:"hardLimitEnabled"`
	IncludedStorageBytes  int64 `json:"includedStorageBytes"`
	PurchasedStorageBytes int64 `json:"purchasedStorageBytes"`
	WarnPercent           int64 `json:"warnPercent"`
}

type StorageLimit struct {
	HardLimitEnabled      bool  `json:"hardLimitEnabled"`
	IncludedStorageBytes  int64 `json:"includedStorageBytes"`
	PurchasedStorageBytes int64 `json:"purchasedStorageBytes"`
	WarnPercent           int64 `json:"warnPercent"`
}

// PrivacyConsentPurpose is the purpose a consent record covers.
type PrivacyConsentPurpose string

const (
	PrivacyConsentPurposeAnalytics          PrivacyConsentPurpose = "analytics"
	PrivacyConsentPurposeProductImprovement PrivacyConsentPurpose = "product_improvement"
	PrivacyConsentPurposeMarketing          PrivacyConsentPurpose = "marketing"
)

type PrivacyConsentRecord struct {
	ExpiresAt  *time.Time            `json:"expiresAt,omitempty"`
	Granted    bool                  `json:"granted"`
	GrantedAt  *time.Time            `json:"grantedAt,omitempty"`
	Purpose    PrivacyConsentPurpose `json:"purpose"`
	RecordedAt time.Time             `json:"recordedAt"`
	RevokedAt  *time.Time            `json:"revokedAt,omitempty"`
	Source     string                `json:"source"`
	SubjectID  string                `json:"subjectId"`
}

type PrivacyConsentListResponse struct {
	CompanySlug string                 `json:"companySlug"`
	Records     []PrivacyConsentRecord `json:"records"`
	SubjectID   string                 `json:"subjectId"`
}

type PrivacyConsentRecordRequest struct {
	CompanySlug string                `json:"companySlug"`
	ExpiresAt   *time.Time            `json:"expiresAt,omitempty"`
	Granted     bool                  `json:"granted"`
	Purpose     PrivacyConsentPurpose `json:"purpose"`
	Source      string                `json:"source"`
	SubjectID   string                `json:"subjectId"`
}

type StorageAlertDelivery struct {
	Error     string     `json:"error,omitempty"`
	State     string     `json:"state"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

type StorageAlertMeasurement struct {
	LimitBytes       int64  `json:"limitBytes"`
	MeasuredValue    int64  `json:"measuredValue"`
	ThresholdPercent int64  `json:"thresholdPercent"`
	WarningState     string `json:"warningState"`
}

type StorageAlertEvent struct {
	Channel        StorageAlertChannel     `json:"channel"`
	Delivery       StorageAlertDelivery    `json:"delivery"`
	EmailRecipient string                  `json:"emailRecipient,omitempty"`
	EventID        string                  `json:"eventId"`
	Measurement    StorageAlertMeasurement `json:"measurement"`
	Metric         string                  `json:"metric"`
	RuleID         string                  `json:"ruleId,omitempty"`
	TenantSlug     string                  `json:"tenantSlug"`
	TriggeredAt    time.Time               `json:"triggeredAt"`
}

type StorageAlertEventListResponse struct {
	Events []StorageAlertEvent `json:"events"`
}

type TimePlanPredictionSnapshot struct {
	BlockID               int64     `json:"blockId"`
	BlockUUID             string    `json:"blockUuid"`
	InputHash             string    `json:"inputHash"`
	Method                string    `json:"method"`
	MethodVersion         string    `json:"methodVersion"`
	P10Ms                 int64     `json:"p10Ms,omitempty"`
	P90Ms                 int64     `json:"p90Ms,omitempty"`
	PointMs               int64     `json:"pointMs"`
	PredictedAt           time.Time `json:"predictedAt"`
	PredictionVersionUUID string    `json:"predictionVersionUuid,omitempty"`
	RunUUID               string    `json:"runUuid"`
	SampleCount           int64     `json:"sampleCount"`
	SnapshotUUID          string    `json:"snapshotUuid"`
	Warnings              []string  `json:"warnings"`
}

type TimePlanPredictionSnapshotList struct {
	Snapshots []TimePlanPredictionSnapshot `json:"snapshots"`
}

type StorageUsageBreakdown struct {
	DbCoreLogicalBytes          int64  `json:"dbCoreLogicalBytes"`
	DbRawInlineLogicalBytes     int64  `json:"dbRawInlineLogicalBytes"`
	ParquetArchiveAttribution   string `json:"parquetArchiveAttribution"`
	S3ParquetArchiveBytes       int64  `json:"s3ParquetArchiveBytes"`
	S3ParquetArchiveObjectCount int64  `json:"s3ParquetArchiveObjectCount"`
	S3RawBytes                  int64  `json:"s3RawBytes"`
	S3RawObjectCount            int64  `json:"s3RawObjectCount"`
}

type StorageMeasurementState struct {
	CalculationVersion int64  `json:"calculationVersion"`
	CompletenessState  string `json:"completenessState"`
	SourceWatermark    int64  `json:"sourceWatermark"`
}

type TenantStorageUsage struct {
	AttributedStorageBytes int64                   `json:"attributedStorageBytes"`
	Breakdown              *StorageUsageBreakdown  `json:"breakdown,omitempty"`
	CompanyName            string                  `json:"companyName"`
	CompanySlug            string                  `json:"companySlug"`
	EffectiveLimitBytes    int64                   `json:"effectiveLimitBytes"`
	Limit                  StorageLimit            `json:"limit"`
	MeasuredAt             time.Time               `json:"measuredAt"`
	MeasurementState       StorageMeasurementState `json:"measurementState"`
	PercentUsed            float64                 `json:"percentUsed"`
	RemainingBytes         int64                   `json:"remainingBytes"`
	RetainedEventCount     int64                   `json:"retainedEventCount"`
	RetainedLogicalBytes   int64                   `json:"retainedLogicalBytes"`
	Warning                bool                    `json:"warning"`
	WarningState           string                  `json:"warningState"`
}

type StorageUsageSummary struct {
	AttributedStorageBytes int64                   `json:"attributedStorageBytes"`
	Breakdown              *StorageUsageBreakdown  `json:"breakdown,omitempty"`
	EffectiveLimitBytes    int64                   `json:"effectiveLimitBytes"`
	MeasuredAt             time.Time               `json:"measuredAt"`
	MeasurementState       StorageMeasurementState `json:"measurementState"`
	PercentUsed            float64                 `json:"percentUsed"`
	RemainingBytes         int64                   `json:"remainingBytes"`
	RetainedEventCount     int64                   `json:"retainedEventCount"`
	RetainedLogicalBytes   int64                   `json:"retainedLogicalBytes"`
	TenantCount            int64                   `json:"tenantCount"`
	Tenants                []TenantStorageUsage    `json:"tenants"`
	Warning                bool                    `json:"warning"`
	WarningState           string                  `json:"warningState"`
}

type DataSpaceUsageRow struct {
	CompanyName          string                  `json:"companyName"`
	Enabled              bool                    `json:"enabled"`
	MeasuredAt           time.Time               `json:"measuredAt"`
	MeasurementState     StorageMeasurementState `json:"measurementState"`
	RetainedEventCount   int64                   `json:"retainedEventCount"`
	RetainedLogicalBytes int64                   `json:"retainedLogicalBytes"`
	RevokedAt            *time.Time              `json:"revokedAt,omitempty"`
	Slug                 string                  `json:"slug"`
}

type DataSpaceUsageSummary struct {
	DataSpaceCount       int64                   `json:"dataSpaceCount"`
	DataSpaces           []DataSpaceUsageRow     `json:"dataSpaces"`
	MeasuredAt           time.Time               `json:"measuredAt"`
	MeasurementState     StorageMeasurementState `json:"measurementState"`
	ParentCompanySlug    string                  `json:"parentCompanySlug"`
	RetainedEventCount   int64                   `json:"retainedEventCount"`
	RetainedLogicalBytes int64                   `json:"retainedLogicalBytes"`
}

type EventCategoryConfig struct {
	IsAudit          bool            `json:"isAudit"`
	Name             string          `json:"name"`
	RetentionDays    *int64          `json:"retentionDays"`
	RollupDimensions json.RawMessage `json:"rollupDimensions,omitempty"`
	Slug             string          `json:"slug"`
}

type EventCategoryConfigListResponse struct {
	EventCategories []EventCategoryConfig `json:"eventCategories"`
}

type EventRollupDimension struct {
	Lookups []string `json:"lookups"`
	Slot    string   `json:"slot"`
}

type RollupDimensionProfile struct {
	Dimensions []EventRollupDimension `json:"dimensions,omitempty"`
}

type EventCategoryConfigUpdateRequest struct {
	IsAudit          bool                    `json:"isAudit"`
	RetentionDays    *int64                  `json:"retentionDays"`
	RollupDimensions *RollupDimensionProfile `json:"rollupDimensions,omitempty"`
}

type EventTypeConfig struct {
	CategoryIsAudit       bool            `json:"categoryIsAudit"`
	CategoryRetentionDays *int64          `json:"categoryRetentionDays"`
	CategorySlug          string          `json:"categorySlug"`
	Name                  string          `json:"name"`
	RetentionDays         *int64          `json:"retentionDays"`
	RollupDimensions      json.RawMessage `json:"rollupDimensions,omitempty"`
	Slug                  string          `json:"slug"`
}

type EventTypeConfigListResponse struct {
	EventTypes []EventTypeConfig `json:"eventTypes"`
}

type EventTypeConfigUpdateRequest struct {
	RetentionDays    *int64                  `json:"retentionDays"`
	RollupDimensions *RollupDimensionProfile `json:"rollupDimensions,omitempty"`
}

type IngestClaim struct {
	ByteCount       int64      `json:"byteCount"`
	ClaimAttempts   int64      `json:"claimAttempts"`
	ClaimExpiresAt  *time.Time `json:"claimExpiresAt,omitempty"`
	EventCount      int64      `json:"eventCount"`
	FailedAt        *time.Time `json:"failedAt,omitempty"`
	FailureClass    string     `json:"failureClass,omitempty"`
	ObjectID        int64      `json:"objectId"`
	ObjectUUID      string     `json:"objectUuid"`
	ProcessedAt     *time.Time `json:"processedAt,omitempty"`
	ProcessingState string     `json:"processingState"`
	WindowStart     time.Time  `json:"windowStart"`
}

type IngestClaimList struct {
	Claims    []IngestClaim `json:"claims"`
	Truncated bool          `json:"truncated"`
}

type IngestClaimTerminalRequest struct {
	FailureClass string `json:"failureClass"`
}

type RLSPolicySpec struct {
	Enabled    bool   `json:"enabled"`
	Force      bool   `json:"force"`
	PolicyName string `json:"policy_name"`
}

type RLSTableDiff struct {
	Classification string         `json:"classification"`
	Desired        *RLSPolicySpec `json:"desired,omitempty"`
	Drift          bool           `json:"drift"`
	Live           *RLSPolicySpec `json:"live,omitempty"`
	Reason         string         `json:"reason,omitempty"`
	Table          string         `json:"table"`
}

type RLSStatus struct {
	CatalogDigest  string         `json:"catalog_digest"`
	CatalogVersion string         `json:"catalog_version"`
	Tables         []RLSTableDiff `json:"tables"`
}

type RLSPlanRequest struct {
	ExpiresInSeconds *int64 `json:"expires_in_seconds,omitempty"`
	GitCommit        string `json:"git_commit"`
	Label            string `json:"label"`
}

type RLSPlan struct {
	ActorID             string          `json:"actor_id"`
	CatalogDigest       string          `json:"catalog_digest"`
	CatalogVersion      string          `json:"catalog_version"`
	CreatedAt           time.Time       `json:"created_at"`
	ExpectedState       json.RawMessage `json:"expected_state"`
	ExpiresAt           time.Time       `json:"expires_at"`
	GitCommit           string          `json:"git_commit"`
	IdempotencyKey      string          `json:"idempotency_key"`
	Label               string          `json:"label"`
	PlanUUID            string          `json:"plan_uuid"`
	ReversibleSQLDigest string          `json:"reversible_sql_digest"`
	TargetTables        []string        `json:"target_tables"`
}

type RLSExecuteRequest struct {
	Force *bool `json:"force,omitempty"`
}

type RLSPolicyDiff struct {
	Converged bool          `json:"converged"`
	Desired   RLSPolicySpec `json:"desired"`
	Drift     []string      `json:"drift,omitempty"`
	Live      RLSPolicySpec `json:"live"`
	Table     string        `json:"table"`
}

// RLSExecuteOutcome is how an RLS reconciliation execution ended.
type RLSExecuteOutcome string

const (
	RLSExecuteOutcomeAttempted RLSExecuteOutcome = "attempted"
	RLSExecuteOutcomeSuccess   RLSExecuteOutcome = "success"
	RLSExecuteOutcomeFailure   RLSExecuteOutcome = "failure"
)

type RLSExecuteResponse struct {
	Diff        []RLSPolicyDiff   `json:"diff"`
	FinishedAt  time.Time         `json:"finished_at"`
	Outcome     RLSExecuteOutcome `json:"outcome"`
	PlanUUID    string            `json:"plan_uuid"`
	RollbackSQL string            `json:"rollback_sql"`
	StartedAt   time.Time         `json:"started_at"`
}

type RLSProofResponse struct {
	After       map[string]RLSPolicySpec `json:"after"`
	AttemptedAt time.Time                `json:"attempted_at"`
	Before      map[string]RLSPolicySpec `json:"before"`
	CompletedAt time.Time                `json:"completed_at"`
	Diff        []RLSPolicyDiff          `json:"diff"`
	PlanUUID    string                   `json:"plan_uuid"`
	RollbackSQL string                   `json:"rollback_sql"`
}

// ErrorIssueState is the triage state of an error issue.
type ErrorIssueState string

const (
	ErrorIssueStateOpen     ErrorIssueState = "open"
	ErrorIssueStateResolved ErrorIssueState = "resolved"
	ErrorIssueStateIgnored  ErrorIssueState = "ignored"
)

type ErrorIssue struct {
	Culprit         string          `json:"culprit"`
	Environments    []string        `json:"environments"`
	EventCount      int64           `json:"eventCount"`
	ExceptionType   string          `json:"exceptionType"`
	Fingerprint     string          `json:"fingerprint"`
	FirstRelease    string          `json:"firstRelease"`
	FirstSeenAt     time.Time       `json:"firstSeenAt"`
	LastRelease     string          `json:"lastRelease"`
	LastSeenAt      time.Time       `json:"lastSeenAt"`
	Platform        string          `json:"platform"`
	RegressedAt     *time.Time      `json:"regressedAt,omitempty"`
	ResolvedAt      *time.Time      `json:"resolvedAt,omitempty"`
	ResolvedRelease string          `json:"resolvedRelease,omitempty"`
	SiteName        string          `json:"siteName"`
	SiteUUID        string          `json:"siteUuid"`
	State           ErrorIssueState `json:"state"`
	Title           string          `json:"title"`
	UUID            string          `json:"uuid"`
}

type ErrorIssueListResponse struct {
	Issues     []ErrorIssue `json:"issues"`
	NextCursor string       `json:"nextCursor,omitempty"`
	OpenCount  int64        `json:"openCount"`
}

type ErrorIssueFrame struct {
	Column      int64    `json:"column,omitempty"`
	ContextLine string   `json:"contextLine,omitempty"`
	Filename    string   `json:"filename,omitempty"`
	Function    string   `json:"function,omitempty"`
	InApp       bool     `json:"inApp,omitempty"`
	Line        int64    `json:"line,omitempty"`
	Module      string   `json:"module,omitempty"`
	PostContext []string `json:"postContext,omitempty"`
	PreContext  []string `json:"preContext,omitempty"`
}

type ErrorIssueException struct {
	Frames        []ErrorIssueFrame `json:"frames,omitempty"`
	FramesOmitted int64             `json:"framesOmitted,omitempty"`
	// Handled is nil when the reporting SDK did not say; an explicit false is kept.
	Handled *bool  `json:"handled,omitempty"`
	Module  string `json:"module,omitempty"`
	Type    string `json:"type"`
	Value   string `json:"value,omitempty"`
}

type ErrorIssueEvent struct {
	CorrelationID  string                `json:"correlationId,omitempty"`
	Environment    string                `json:"environment,omitempty"`
	Exceptions     []ErrorIssueException `json:"exceptions"`
	Fingerprint    string                `json:"fingerprint"`
	Level          string                `json:"level,omitempty"`
	Platform       string                `json:"platform,omitempty"`
	Release        string                `json:"release,omitempty"`
	RuntimeName    string                `json:"runtimeName,omitempty"`
	RuntimeVersion string                `json:"runtimeVersion,omitempty"`
	SDKName        string                `json:"sdkName,omitempty"`
	SDKVersion     string                `json:"sdkVersion,omitempty"`
	SiteUUID       string                `json:"siteUuid,omitempty"`
	Truncated      bool                  `json:"truncated,omitempty"`
}

type ErrorIssueDetail struct {
	CompanyName     string          `json:"companyName"`
	Issue           ErrorIssue      `json:"issue"`
	LatestEvent     ErrorIssueEvent `json:"latestEvent"`
	LatestEventUUID string          `json:"latestEventUuid"`
}

type ErrorIssueStateRequest struct {
	State ErrorIssueState `json:"state"`
}

type PublicSiteConfig struct {
	AllowedOrigins        []string            `json:"allowedOrigins"`
	Environment           string              `json:"environment"`
	IdentityMode          BrowserIdentityMode `json:"identityMode"`
	PageViewSchemaVersion string              `json:"pageViewSchemaVersion,omitempty"`
	SiteUUID              string              `json:"siteUuid"`
}

// REUSED existing: AppSummary ProducerProvisionPublicClient

// SetDomainImageOverride pins one realm component to an exact incident image identity.
func (c *AuthProjectAdminClient) SetDomainImageOverride(ctx context.Context, projectID string, environmentID string, req AuthDomainImageOverrideSetRequest, options AuthProjectRequestOptions) (*AuthDomainBinding, error) {
	var out AuthDomainBinding
	err := c.admin.requestWithHeaders(ctx, http.MethodPut, "/auth-projects/"+url.PathEscape(projectID)+"/environments/"+url.PathEscape(environmentID)+"/domain/image-override", req, &out, authProjectHeaders(options))
	return &out, err
}

// ClearDomainImageOverride ends the realm's incident image override.
func (c *AuthProjectAdminClient) ClearDomainImageOverride(ctx context.Context, projectID string, environmentID string, options AuthProjectRequestOptions) (*AuthDomainBinding, error) {
	var out AuthDomainBinding
	err := c.admin.requestWithHeaders(ctx, http.MethodDelete, "/auth-projects/"+url.PathEscape(projectID)+"/environments/"+url.PathEscape(environmentID)+"/domain/image-override", nil, &out, authProjectHeaders(options))
	return &out, err
}

// MintDomain reserves a platform-generated issuer hostname for an installed environment.
func (c *AuthProjectAdminClient) MintDomain(ctx context.Context, projectID string, environmentID string, req MintAuthDomainRequest, options AuthProjectRequestOptions) (*AuthDomainBinding, error) {
	var out AuthDomainBinding
	err := c.admin.requestWithHeaders(ctx, http.MethodPost, "/auth-projects/"+url.PathEscape(projectID)+"/environments/"+url.PathEscape(environmentID)+"/domain/mint", req, &out, authProjectHeaders(options))
	return &out, err
}

// ListAppLinks lists the environment's app item links.
func (c *AuthProjectAdminClient) ListAppLinks(ctx context.Context, projectID string, environmentID string, options AuthProjectRequestOptions) (*AppItemEnvironmentLinkList, error) {
	var out AppItemEnvironmentLinkList
	err := c.admin.requestWithHeaders(ctx, http.MethodGet, "/auth-projects/"+url.PathEscape(projectID)+"/environments/"+url.PathEscape(environmentID)+"/app-links", nil, &out, authProjectHeaders(options))
	return &out, err
}

// ListAppDeployCredentials lists an app's CI deploy credentials.
func (c *AuthProjectAdminClient) ListAppDeployCredentials(ctx context.Context, slug string, options AuthProjectRequestOptions) (*AppDeployCredentialList, error) {
	var out AppDeployCredentialList
	err := c.admin.requestWithHeaders(ctx, http.MethodGet, "/apps/"+url.PathEscape(slug)+"/deploy-credentials", nil, &out, authProjectHeaders(options))
	return &out, err
}

// PauseApp pauses an app.
func (c *AuthProjectAdminClient) PauseApp(ctx context.Context, slug string, options AuthProjectRequestOptions) (*AppSummary, error) {
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out AppSummary
	err := c.admin.requestWithHeaders(ctx, http.MethodPost, "/apps/"+url.PathEscape(slug)+"/pause", nil, &out, authProjectHeaders(options))
	return &out, err
}

// ResumeApp resumes an app.
func (c *AuthProjectAdminClient) ResumeApp(ctx context.Context, slug string, options AuthProjectRequestOptions) (*AppSummary, error) {
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out AppSummary
	err := c.admin.requestWithHeaders(ctx, http.MethodPost, "/apps/"+url.PathEscape(slug)+"/resume", nil, &out, authProjectHeaders(options))
	return &out, err
}

// RepairApp repairs an app's setup.
func (c *AuthProjectAdminClient) RepairApp(ctx context.Context, slug string, options AuthProjectRequestOptions) (*AppSummary, error) {
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out AppSummary
	err := c.admin.requestWithHeaders(ctx, http.MethodPost, "/apps/"+url.PathEscape(slug)+"/repair", nil, &out, authProjectHeaders(options))
	return &out, err
}

// SetAppSignIn sets an app's sign-in mode.
func (c *AuthProjectAdminClient) SetAppSignIn(ctx context.Context, slug string, req AppSignInRequest, options AuthProjectRequestOptions) (*AppSummary, error) {
	if err := requireAuthProjectIdempotencyKey(options); err != nil {
		return nil, err
	}
	var out AppSummary
	err := c.admin.requestWithHeaders(ctx, http.MethodPost, "/apps/"+url.PathEscape(slug)+"/sign-in", req, &out, authProjectHeaders(options))
	return &out, err
}

// SetStorageLimit updates tenant storage limit.
func (c *TenantAdminClient) SetStorageLimit(ctx context.Context, slug string, req StorageLimitUpdateRequest) (*StorageLimit, error) {
	var out StorageLimit
	err := c.admin.request(ctx, http.MethodPut, "/tenants/"+url.PathEscape(slug)+"/storage-limit", req, &out)
	return &out, err
}

// ReplaceUnknownSecret replaces an OAuth client secret after unknown delivery.
func (c *OAuthClientAdminClient) ReplaceUnknownSecret(ctx context.Context, clientID string, options OAuthClientRequestOptions) (*AdminOAuthClientSecretResponse, error) {
	if err := options.requireIdempotencyKey(); err != nil {
		return nil, err
	}
	var out AdminOAuthClientSecretResponse
	err := c.admin.requestWithHeaders(ctx, http.MethodPost, "/oauth-clients/"+url.PathEscape(clientID)+"/replace-unknown-secret", nil, &out, options.headers())
	return &out, err
}

// ListConsent lists a subject's consent records.
func (c *PrivacyAdminClient) ListConsent(ctx context.Context, companySlug string, subjectID string) (*PrivacyConsentListResponse, error) {
	var out PrivacyConsentListResponse
	err := c.admin.request(ctx, http.MethodGet, siteQueryPath("/privacy/consent", url.Values{companySlugQueryParameter: {companySlug}, "subjectId": {subjectID}}), nil, &out)
	return &out, err
}

// RecordConsent records a consent grant or revocation.
func (c *PrivacyAdminClient) RecordConsent(ctx context.Context, req PrivacyConsentRecordRequest) (*PrivacyConsentRecord, error) {
	var out PrivacyConsentRecord
	err := c.admin.request(ctx, http.MethodPost, "/privacy/consent", req, &out)
	return &out, err
}

// DeleteConsent revokes a subject's consent for one purpose.
func (c *PrivacyAdminClient) DeleteConsent(ctx context.Context, companySlug string, subjectID string, purpose string, source string) (*PrivacyConsentRecord, error) {
	var out PrivacyConsentRecord
	err := c.admin.request(ctx, http.MethodDelete, siteQueryPath("/privacy/consent/"+url.PathEscape(subjectID), url.Values{companySlugQueryParameter: {companySlug}, "purpose": {purpose}, "source": {source}}), nil, &out)
	return &out, err
}

// ListEvents lists recent storage alert events for a tenant.
func (c *StorageAlertAdminClient) ListEvents(ctx context.Context, tenantSlug string) (*StorageAlertEventListResponse, error) {
	var out StorageAlertEventListResponse
	err := c.admin.request(ctx, http.MethodGet, "/storage/alerts/"+url.PathEscape(tenantSlug)+"/events", nil, &out)
	return &out, err
}

// ListRunPredictions lists immutable time-plan prediction snapshots.
func (c *TimePlanAdminClient) ListRunPredictions(ctx context.Context, companySlug string, runUUID string, limit int64) (*TimePlanPredictionSnapshotList, error) {
	var out TimePlanPredictionSnapshotList
	err := c.admin.request(ctx, http.MethodGet, siteQueryPath("/time-plans/runs/"+url.PathEscape(runUUID)+"/predictions", url.Values{companySlugQueryParameter: {companySlug}, "limit": {positiveIntQuery(limit)}}), nil, &out)
	return &out, err
}

// ListStorageUsage lists tenant storage usage.
func (c *UsageAdminClient) ListStorageUsage(ctx context.Context) (*StorageUsageSummary, error) {
	var out StorageUsageSummary
	err := c.admin.request(ctx, http.MethodGet, "/storage-usage", nil, &out)
	return &out, err
}

// Approve marks a domain verified on a HaakCo system administrator's confirmation.
func (c *SiteDomainAdminClient) Approve(ctx context.Context, companySlug string, domainUUID string) (*AdminSiteDomain, error) {
	var out AdminSiteDomain
	err := c.request(ctx, http.MethodPost, siteQueryPath("/site-domains/"+url.PathEscape(domainUUID)+"/approval", url.Values{companySlugQueryParameter: {companySlug}}), nil, &out)
	return &out, err
}

// RevokeApproval withdraws a HaakCo system administrator's approval of a domain.
func (c *SiteDomainAdminClient) RevokeApproval(ctx context.Context, companySlug string, domainUUID string) (*AdminSiteDomain, error) {
	var out AdminSiteDomain
	err := c.request(ctx, http.MethodDelete, siteQueryPath("/site-domains/"+url.PathEscape(domainUUID)+"/approval", url.Values{companySlugQueryParameter: {companySlug}}), nil, &out)
	return &out, err
}

// Get gets a producer.
func (c *ProducerProvisioningClient) Get(ctx context.Context, clientID string) (*ProducerProvisionPublicClient, error) {
	var out ProducerProvisionPublicClient
	err := c.provisioning.request(ctx, http.MethodGet, "/producer-provisioning/"+url.PathEscape(clientID), nil, &out)
	return &out, err
}

// Usage gets data space usage.
func (c *DataSpaceProvisioningClient) Usage(ctx context.Context, parentCompanySlug string) (*DataSpaceUsageSummary, error) {
	var out DataSpaceUsageSummary
	err := c.provisioning.request(ctx, http.MethodGet, dataSpacePath("/data-spaces/usage", parentCompanySlug), nil, &out)
	return &out, err
}

// ListCategories lists event categories.
func (c *EventTypeAdminClient) ListCategories(ctx context.Context) (*EventCategoryConfigListResponse, error) {
	var out EventCategoryConfigListResponse
	err := c.admin.request(ctx, http.MethodGet, "/event-categories", nil, &out)
	return &out, err
}

// UpdateCategory updates event category.
func (c *EventTypeAdminClient) UpdateCategory(ctx context.Context, slug string, req EventCategoryConfigUpdateRequest) error {
	return c.admin.request(ctx, http.MethodPatch, "/event-categories/"+url.PathEscape(slug), req, nil)
}

// List lists event types.
func (c *EventTypeAdminClient) List(ctx context.Context) (*EventTypeConfigListResponse, error) {
	var out EventTypeConfigListResponse
	err := c.admin.request(ctx, http.MethodGet, "/event-types", nil, &out)
	return &out, err
}

// Update updates event type.
func (c *EventTypeAdminClient) Update(ctx context.Context, slug string, req EventTypeConfigUpdateRequest) error {
	return c.admin.request(ctx, http.MethodPatch, "/event-types/"+url.PathEscape(slug), req, nil)
}

// List lists accepted raw-landing claims.
func (c *IngestClaimAdminClient) List(ctx context.Context, tenantSlug string, limit int64) (*IngestClaimList, error) {
	var out IngestClaimList
	err := c.admin.request(ctx, http.MethodGet, siteQueryPath("/ingest/claims/"+url.PathEscape(tenantSlug), url.Values{"limit": {positiveIntQuery(limit)}}), nil, &out)
	return &out, err
}

// Retry retries an accepted raw-landing claim.
func (c *IngestClaimAdminClient) Retry(ctx context.Context, tenantSlug string, objectUUID string) (*IngestClaim, error) {
	var out IngestClaim
	err := c.admin.request(ctx, http.MethodPost, "/ingest/claims/"+url.PathEscape(tenantSlug)+"/"+url.PathEscape(objectUUID)+"/retry", nil, &out)
	return &out, err
}

// MarkTerminal terminallies classify an accepted raw-landing claim.
func (c *IngestClaimAdminClient) MarkTerminal(ctx context.Context, tenantSlug string, objectUUID string, req IngestClaimTerminalRequest) (*IngestClaim, error) {
	var out IngestClaim
	err := c.admin.request(ctx, http.MethodPost, "/ingest/claims/"+url.PathEscape(tenantSlug)+"/"+url.PathEscape(objectUUID)+"/terminal", req, &out)
	return &out, err
}

// Status gets RLS catalog status.
func (c *RLSAdminClient) Status(ctx context.Context) (*RLSStatus, error) {
	var out RLSStatus
	err := c.admin.request(ctx, http.MethodGet, "/rls/status", nil, &out)
	return &out, err
}

// CreatePlan creates an RLS reconciliation plan.
func (c *RLSAdminClient) CreatePlan(ctx context.Context, req RLSPlanRequest) (*RLSPlan, error) {
	var out RLSPlan
	err := c.admin.request(ctx, http.MethodPost, "/rls/reconciliation-plans", req, &out)
	return &out, err
}

// ExecutePlan executes an RLS reconciliation plan.
func (c *RLSAdminClient) ExecutePlan(ctx context.Context, planUUID string, req RLSExecuteRequest, idempotencyKey string) (*RLSExecuteResponse, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return nil, fmt.Errorf("custd: executePlan requires an idempotency key")
	}
	var out RLSExecuteResponse
	err := c.admin.requestWithHeaders(ctx, http.MethodPost, "/rls/reconciliation-plans/"+url.PathEscape(planUUID)+"/execute", req, &out, idempotencyKeyHeader(idempotencyKey))
	return &out, err
}

// GetPlanProof gets an RLS reconciliation plan proof.
func (c *RLSAdminClient) GetPlanProof(ctx context.Context, planUUID string) (*RLSProofResponse, error) {
	var out RLSProofResponse
	err := c.admin.request(ctx, http.MethodGet, "/rls/reconciliation-plans/"+url.PathEscape(planUUID)+"/proof", nil, &out)
	return &out, err
}

// List lists a tenant's error issues.
func (c *ErrorIssueAdminClient) List(ctx context.Context, options ErrorIssueListOptions) (*ErrorIssueListResponse, error) {
	var out ErrorIssueListResponse
	path := siteQueryPath("/error-issues", url.Values{
		companySlugQueryParameter: {options.CompanySlug}, "siteUuid": {options.SiteUUID}, "state": {string(options.State)},
		"sort": {string(options.Sort)}, "limit": {positiveIntQuery(options.Limit)}, "cursor": {options.Cursor},
	})
	err := c.admin.requestNonAdmin(ctx, http.MethodGet, c.admin.actingCompanyPath(path), nil, &out)
	return &out, err
}

// Get gets one error issue with its latest error.
func (c *ErrorIssueAdminClient) Get(ctx context.Context, companySlug string, issueUUID string) (*ErrorIssueDetail, error) {
	var out ErrorIssueDetail
	err := c.admin.requestNonAdmin(ctx, http.MethodGet, c.admin.actingCompanyPath(siteQueryPath("/error-issues/"+url.PathEscape(issueUUID), url.Values{companySlugQueryParameter: {companySlug}})), nil, &out)
	return &out, err
}

// SetState resolve,s ignore or reopen an error issue.
func (c *ErrorIssueAdminClient) SetState(ctx context.Context, companySlug string, issueUUID string, req ErrorIssueStateRequest) (*ErrorIssue, error) {
	var out ErrorIssue
	err := c.admin.requestNonAdmin(ctx, http.MethodPost, c.admin.actingCompanyPath(siteQueryPath("/error-issues/"+url.PathEscape(issueUUID)+"/state", url.Values{companySlugQueryParameter: {companySlug}})), req, &out)
	return &out, err
}

// GetPublicConfig gets public site config.
func (c *SiteAdminClient) GetPublicConfig(ctx context.Context, siteUUID string) (*PublicSiteConfig, error) {
	var out PublicSiteConfig
	err := c.admin.requestNonAdmin(ctx, http.MethodGet, "/sites/"+url.PathEscape(siteUUID)+"/config", nil, &out)
	return &out, err
}

// ProvisionProducer provisions a managed audit producer. The server validates the
// request as it does producer provisioning and answers with the one-time credential.
func (c *ManagedAuditAdminClient) ProvisionProducer(
	ctx context.Context,
	req ProducerProvisionRequest,
) (*ProducerProvisionResponse, error) {
	var out ProducerProvisionResponse
	err := c.admin.requestNonAdmin(ctx, http.MethodPost, "/managed-audit/provision", req, &out)
	return &out, err
}
