package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// Tenant email surfaces: campaign email permissions, revisions, runs and the
// welcome funnel; the tenant's email relay; SDK usage and deprecations;
// accounting evidence; and per-event notification destinations. An empty
// companySlug uses the credential's own tenant.

// CampaignEmailAdminClient manages campaign email for a tenant.
type CampaignEmailAdminClient struct {
	admin *AdminClient
}

// EmailRelayAdminClient manages the SMTP relay a tenant's email is sent through.
type EmailRelayAdminClient struct {
	admin *AdminClient
}

// SDKUsageAdminClient reports which SDK versions call Custd and marks versions deprecated.
type SDKUsageAdminClient struct {
	admin *AdminClient
}

// AccountingEvidenceAdminClient appends and lists a tenant's accounting evidence.
type AccountingEvidenceAdminClient struct {
	admin *AdminClient
}

// NotificationDestinationAdminClient sets where one platform event is sent for a tenant.
type NotificationDestinationAdminClient struct {
	admin *AdminClient
}

type CampaignEmailRecipient struct {
	Email string `json:"email"`
}

type CampaignEmailPermission struct {
	ConfirmedAt *time.Time `json:"confirmedAt,omitempty"`
	Eligible    bool       `json:"eligible"`
	WithdrawnAt *time.Time `json:"withdrawnAt,omitempty"`
}

type CampaignEmailConfirmation struct {
	ConfirmedAt time.Time `json:"confirmedAt"`
	Email       string    `json:"email"`
	EvidenceRef string    `json:"evidenceRef"`
	Source      string    `json:"source"`
}

type PublishCampaignEmailRevisionRequest struct {
	ArtifactJSON string `json:"artifactJson"`
	Revision     string `json:"revision"`
}

type CampaignEmailRevision struct {
	ArtifactJSON string    `json:"artifactJson"`
	PublishedAt  time.Time `json:"publishedAt"`
	Revision     string    `json:"revision"`
}

type CampaignRenderRequest struct {
	Slots map[string]string `json:"slots"`
}

type CampaignRenderResult struct {
	CompanyUUID string `json:"companyUuid"`
	HTML        string `json:"html"`
	Subject     string `json:"subject"`
	Text        string `json:"text"`
}

type CampaignTestSelection struct {
	Revision string            `json:"revision"`
	Values   map[string]string `json:"values"`
}

type CampaignRunScheduleRequest struct {
	AudienceDigest string                `json:"audienceDigest,omitempty"`
	Campaign       CampaignTestSelection `json:"campaign"`
	ID             string                `json:"id"`
	Recipients     []string              `json:"recipients"`
	ScheduledAt    time.Time             `json:"scheduledAt"`
}

type CampaignRunRecipientStatus struct {
	DeliveryUUID string `json:"deliveryUuid,omitempty"`
	Position     int64  `json:"position"`
	Reason       string `json:"reason,omitempty"`
	State        string `json:"state"`
}

type CampaignRunStatus struct {
	Cancelled   int64                        `json:"cancelled"`
	CancelledAt *time.Time                   `json:"cancelledAt,omitempty"`
	Deliveries  []CampaignRunRecipientStatus `json:"deliveries"`
	Excluded    int64                        `json:"excluded"`
	Failed      int64                        `json:"failed"`
	ID          string                       `json:"id"`
	Pending     int64                        `json:"pending"`
	Processing  int64                        `json:"processing"`
	Recipients  int64                        `json:"recipients"`
	Revision    string                       `json:"revision"`
	ScheduledAt time.Time                    `json:"scheduledAt"`
	Sent        int64                        `json:"sent"`
}

type CampaignWelcomeDefinition struct {
	Reminder CampaignTestSelection `json:"reminder"`
	Version  string                `json:"version"`
	Welcome  CampaignTestSelection `json:"welcome"`
}

type CampaignWelcomeState struct {
	Paused  bool   `json:"paused"`
	Version string `json:"version"`
}

type CampaignWelcomePause struct {
	Paused *bool `json:"paused"`
}

type CampaignWelcomeActivation struct {
	ActivatedAt time.Time `json:"activatedAt"`
	ContactID   string    `json:"contactId"`
}

type CampaignWelcomeEnrollment struct {
	ActivatedAt        *time.Time `json:"activatedAt,omitempty"`
	ContactID          string     `json:"contactId"`
	EarliestReminderAt *time.Time `json:"earliestReminderAt,omitempty"`
	ReminderRunID      string     `json:"reminderRunId,omitempty"`
	Version            string     `json:"version,omitempty"`
	WelcomeRunID       string     `json:"welcomeRunId,omitempty"`
}

type CampaignWelcomeSignup struct {
	ConfirmedAt    time.Time `json:"confirmedAt"`
	ContactID      string    `json:"contactId"`
	Email          string    `json:"email"`
	EvidenceRef    string    `json:"evidenceRef"`
	EvidenceSource string    `json:"evidenceSource"`
}

type EmailRelayEffectiveSender struct {
	FromAddress     string `json:"fromAddress"`
	FromDisplayName string `json:"fromDisplayName"`
	HasCredential   bool   `json:"hasCredential"`
	Relay           string `json:"relay"`
}

type EmailRelay struct {
	CreatedAt       time.Time `json:"createdAt"`
	Enabled         bool      `json:"enabled"`
	FromAddress     string    `json:"fromAddress"`
	FromDisplayName string    `json:"fromDisplayName"`
	HasCredential   bool      `json:"hasCredential"`
	Host            string    `json:"host"`
	Port            int64     `json:"port"`
	TLSMode         string    `json:"tlsMode"`
	UpdatedAt       time.Time `json:"updatedAt"`
	Username        string    `json:"username"`
}

// EmailRelayView is a tenant's own relay and the sender its next message carries. Relay is nil
// when the tenant has set no relay of its own and sends through Custd's.
type EmailRelayView struct {
	Effective EmailRelayEffectiveSender `json:"effective"`
	Relay     *EmailRelay               `json:"relay"`
}

type EmailRelayWriteRequest struct {
	Enabled         bool   `json:"enabled"`
	FromAddress     string `json:"fromAddress"`
	FromDisplayName string `json:"fromDisplayName"`
	Host            string `json:"host"`
	Password        string `json:"password"`
	Port            int64  `json:"port"`
	TLSMode         string `json:"tlsMode"`
	Username        string `json:"username"`
}

type EmailRelayTestSendRequest struct {
	Campaign        *CampaignTestSelection `json:"campaign,omitempty"`
	ReceiptRevision string                 `json:"receiptRevision,omitempty"`
	Recipient       string                 `json:"recipient"`
}

type EmailRelayTestSendResult struct {
	Accepted      bool                      `json:"accepted"`
	DeliveryUuids []string                  `json:"deliveryUuids"`
	MessageUUID   string                    `json:"messageUuid"`
	Sender        EmailRelayEffectiveSender `json:"sender"`
}

type SDKVersionDeprecation struct {
	DeprecatedAt       time.Time `json:"deprecatedAt"`
	MarkedBy           string    `json:"markedBy"`
	Note               string    `json:"note"`
	RecommendedVersion string    `json:"recommendedVersion"`
	SDK                string    `json:"sdk"`
	VersionScope       string    `json:"versionScope"`
}

type SDKUsageItem struct {
	Deprecation *SDKVersionDeprecation `json:"deprecation,omitempty"`
	FirstSeenAt time.Time              `json:"firstSeenAt"`
	LastSeenAt  time.Time              `json:"lastSeenAt"`
	Requests    int64                  `json:"requests"`
	SDK         string                 `json:"sdk"`
	State       string                 `json:"state"`
	Version     string                 `json:"version"`
}

type SDKUsageListResponse struct {
	Items []SDKUsageItem `json:"items"`
}

type SDKVersionDeprecationRequest struct {
	Note               string `json:"note"`
	RecommendedVersion string `json:"recommendedVersion"`
}

// AccountingEvidenceKind is what one accounting evidence record states, or what it targets.
type AccountingEvidenceKind string

const (
	AccountingEvidenceKindPurchase   AccountingEvidenceKind = "purchase"
	AccountingEvidenceKindBalance    AccountingEvidenceKind = "balance"
	AccountingEvidenceKindCorrection AccountingEvidenceKind = "correction"
)

type AccountingEvidence struct {
	Actor              string                 `json:"actor"`
	AuditCorrelationID string                 `json:"auditCorrelationId"`
	Balance            float64                `json:"balance,omitempty"`
	Delta              float64                `json:"delta,omitempty"`
	EffectiveAt        time.Time              `json:"effectiveAt"`
	ID                 string                 `json:"id"`
	Kind               AccountingEvidenceKind `json:"kind"`
	ObservedAt         time.Time              `json:"observedAt"`
	Quantity           float64                `json:"quantity,omitempty"`
	Reason             string                 `json:"reason"`
	SupersedesID       string                 `json:"supersedesId,omitempty"`
	TargetID           string                 `json:"targetId,omitempty"`
	TargetKind         AccountingEvidenceKind `json:"targetKind,omitempty"`
	TenantKey          string                 `json:"tenantKey"`
	Unit               Unit                   `json:"unit"`
}

type AccountingEvidenceListResponse struct {
	Evidence []AccountingEvidence `json:"evidence"`
}

type AccountingEvidenceCreateRequest struct {
	AuditCorrelationID string                 `json:"auditCorrelationId"`
	Balance            *float64               `json:"balance,omitempty"`
	Delta              *float64               `json:"delta,omitempty"`
	EffectiveAt        time.Time              `json:"effectiveAt"`
	ID                 string                 `json:"id"`
	Kind               AccountingEvidenceKind `json:"kind"`
	ObservedAt         time.Time              `json:"observedAt"`
	Quantity           *float64               `json:"quantity,omitempty"`
	Reason             string                 `json:"reason"`
	SupersedesID       string                 `json:"supersedesId,omitempty"`
	TargetID           string                 `json:"targetId,omitempty"`
	TargetKind         AccountingEvidenceKind `json:"targetKind,omitempty"`
	Unit               Unit                   `json:"unit"`
}

type AccountingEvidenceResponse struct {
	Evidence AccountingEvidence `json:"evidence"`
}

// NotificationEventSlug is an event a notification destination can carry.
type NotificationEventSlug string

const (
	NotificationEventSlugEmailRelayTestSend    NotificationEventSlug = "email_relay_test_send"
	NotificationEventSlugIdentityVerifyEmail   NotificationEventSlug = "identity_verify_email"
	NotificationEventSlugIdentityResetPassword NotificationEventSlug = "identity_reset_password"
	NotificationEventSlugUserWelcome           NotificationEventSlug = "user_welcome"
	NotificationEventSlugApplicationInvitation NotificationEventSlug = "application_invitation"
	NotificationEventSlugErrorIssueAlert       NotificationEventSlug = "error_issue_alert"
	NotificationEventSlugSiteFindingOpened     NotificationEventSlug = "site_finding_opened"
	NotificationEventSlugSiteIncident          NotificationEventSlug = "site_incident"
	NotificationEventSlugSiteAIAllowance       NotificationEventSlug = "site_ai_allowance"
	NotificationEventSlugSiteStatusPageConfirm NotificationEventSlug = "site_status_page_confirm"
	NotificationEventSlugSiteStatusPageUpdate  NotificationEventSlug = "site_status_page_update"
)

// NotificationDestinationEvent is an event whose destination a tenant can read and set.
type NotificationDestinationEvent string

const (
	NotificationDestinationEventErrorIssueAlert   NotificationDestinationEvent = "error_issue_alert"
	NotificationDestinationEventSiteFindingOpened NotificationDestinationEvent = "site_finding_opened"
	NotificationDestinationEventSiteIncident      NotificationDestinationEvent = "site_incident"
	NotificationDestinationEventSiteAIAllowance   NotificationDestinationEvent = "site_ai_allowance"
)

// NotificationChannel is how a notification destination is delivered.
type NotificationChannel string

const (
	NotificationChannelEmail   NotificationChannel = "email"
	NotificationChannelWebhook NotificationChannel = "webhook"
	NotificationChannelSlack   NotificationChannel = "slack"
)

type NotificationDestination struct {
	Channel        NotificationChannel   `json:"channel"`
	Configured     bool                  `json:"configured"`
	EmailRecipient string                `json:"emailRecipient,omitempty"`
	Enabled        bool                  `json:"enabled"`
	EventSlug      NotificationEventSlug `json:"eventSlug"`
}

type NotificationDestinationRequest struct {
	Channel        NotificationChannel `json:"channel"`
	EmailRecipient string              `json:"emailRecipient,omitempty"`
	Enabled        bool                `json:"enabled"`
}

// REUSED existing:

// CheckPermission checks current campaign email permission.
func (c *CampaignEmailAdminClient) CheckPermission(ctx context.Context, companySlug string, req CampaignEmailRecipient) (*CampaignEmailPermission, error) {
	var out CampaignEmailPermission
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/campaign-email-permissions/check", url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// ConfirmPermission records an existing verified campaign opt-in.
func (c *CampaignEmailAdminClient) ConfirmPermission(ctx context.Context, companySlug string, req CampaignEmailConfirmation) (*CampaignEmailPermission, error) {
	var out CampaignEmailPermission
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/campaign-email-permissions/confirm", url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// WithdrawPermission withdraws permission to send campaign email.
func (c *CampaignEmailAdminClient) WithdrawPermission(ctx context.Context, companySlug string, req CampaignEmailRecipient) (*CampaignEmailPermission, error) {
	var out CampaignEmailPermission
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/campaign-email-permissions/withdraw", url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// CreateRevision publishes a trusted-build email campaign artifact.
func (c *CampaignEmailAdminClient) CreateRevision(ctx context.Context, companySlug string, req PublishCampaignEmailRevisionRequest) (*CampaignEmailRevision, error) {
	var out CampaignEmailRevision
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/campaign-email-revisions", url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// GetRevision reads an exact email campaign artifact revision.
func (c *CampaignEmailAdminClient) GetRevision(ctx context.Context, companySlug string, revision string) (*CampaignEmailRevision, error) {
	var out CampaignEmailRevision
	err := c.admin.request(ctx, http.MethodGet, siteQueryPath("/campaign-email-revisions/"+url.PathEscape(revision), url.Values{companySlugQueryParameter: {companySlug}}), nil, &out)
	return &out, err
}

// RenderRevision renders a published email campaign revision.
func (c *CampaignEmailAdminClient) RenderRevision(ctx context.Context, revision string, req CampaignRenderRequest) (*CampaignRenderResult, error) {
	var out CampaignRenderResult
	err := c.admin.request(ctx, http.MethodPost, "/campaign-email-revisions/"+url.PathEscape(revision)+"/render", req, &out)
	return &out, err
}

// PreviewRevision previews without sending a published email campaign revision.
func (c *CampaignEmailAdminClient) PreviewRevision(ctx context.Context, companySlug, revision string, req CampaignRenderRequest) (*CampaignRenderResult, error) {
	var out CampaignRenderResult
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/campaign-email-revisions/"+url.PathEscape(revision)+"/preview", url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// StartRun schedules a pinned campaign run.
func (c *CampaignEmailAdminClient) StartRun(ctx context.Context, companySlug string, req CampaignRunScheduleRequest) (*CampaignRunStatus, error) {
	var out CampaignRunStatus
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/campaign-email-runs", url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// GetRun inspects campaign run progress.
func (c *CampaignEmailAdminClient) GetRun(ctx context.Context, companySlug string, runID string) (*CampaignRunStatus, error) {
	var out CampaignRunStatus
	err := c.admin.request(ctx, http.MethodGet, siteQueryPath("/campaign-email-runs/"+url.PathEscape(runID), url.Values{companySlugQueryParameter: {companySlug}}), nil, &out)
	return &out, err
}

// CancelRun cancels pending campaign delivery.
func (c *CampaignEmailAdminClient) CancelRun(ctx context.Context, companySlug string, runID string) (*CampaignRunStatus, error) {
	var out CampaignRunStatus
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/campaign-email-runs/"+url.PathEscape(runID)+"/cancel", url.Values{companySlugQueryParameter: {companySlug}}), nil, &out)
	return &out, err
}

// PutWelcomeDefinition configures a pinned welcome funnel.
func (c *CampaignEmailAdminClient) PutWelcomeDefinition(ctx context.Context, companySlug string, req CampaignWelcomeDefinition) (*CampaignWelcomeState, error) {
	var out CampaignWelcomeState
	err := c.admin.request(ctx, http.MethodPut, siteQueryPath("/campaign-email-welcome/definition", url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// SetWelcomePaused pauses or resume welcome funnel submissions.
func (c *CampaignEmailAdminClient) SetWelcomePaused(ctx context.Context, companySlug string, req CampaignWelcomePause) (*CampaignWelcomeState, error) {
	var out CampaignWelcomeState
	err := c.admin.request(ctx, http.MethodPut, siteQueryPath("/campaign-email-welcome/pause", url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// RecordWelcomeActivation records first-party contact activation and exit the reminder.
func (c *CampaignEmailAdminClient) RecordWelcomeActivation(ctx context.Context, companySlug string, req CampaignWelcomeActivation) (*CampaignWelcomeEnrollment, error) {
	var out CampaignWelcomeEnrollment
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/campaign-email-welcome/activation", url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// RecordWelcomeConfirmedSignup enters a contact with recorded opt-in evidence.
func (c *CampaignEmailAdminClient) RecordWelcomeConfirmedSignup(ctx context.Context, companySlug string, req CampaignWelcomeSignup) (*CampaignWelcomeEnrollment, error) {
	var out CampaignWelcomeEnrollment
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/campaign-email-welcome/confirmed-signup", url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// GetWelcomeContact inspects a contact's welcome enrollment and delivery runs.
func (c *CampaignEmailAdminClient) GetWelcomeContact(ctx context.Context, companySlug string, contactID string) (*CampaignWelcomeEnrollment, error) {
	var out CampaignWelcomeEnrollment
	err := c.admin.request(ctx, http.MethodGet, siteQueryPath("/campaign-email-welcome/contacts/"+url.PathEscape(contactID), url.Values{companySlugQueryParameter: {companySlug}}), nil, &out)
	return &out, err
}

// Get gets a tenant's email relay and effective sender.
func (c *EmailRelayAdminClient) Get(ctx context.Context, tenantSlug string) (*EmailRelayView, error) {
	var out EmailRelayView
	err := c.admin.request(ctx, http.MethodGet, "/email-relay/"+url.PathEscape(tenantSlug), nil, &out)
	return &out, err
}

// Put creates or replace a tenant's email relay.
func (c *EmailRelayAdminClient) Put(ctx context.Context, tenantSlug string, req EmailRelayWriteRequest) (*EmailRelay, error) {
	var out EmailRelay
	err := c.admin.request(ctx, http.MethodPut, "/email-relay/"+url.PathEscape(tenantSlug), req, &out)
	return &out, err
}

// Delete clears a tenant's email relay.
func (c *EmailRelayAdminClient) Delete(ctx context.Context, tenantSlug string) error {
	return c.admin.request(ctx, http.MethodDelete, "/email-relay/"+url.PathEscape(tenantSlug), nil, nil)
}

// TestSend sends a test email on a tenant's effective relay.
func (c *EmailRelayAdminClient) TestSend(ctx context.Context, tenantSlug string, req EmailRelayTestSendRequest) (*EmailRelayTestSendResult, error) {
	var out EmailRelayTestSendResult
	err := c.admin.request(ctx, http.MethodPost, "/email-relay/"+url.PathEscape(tenantSlug)+"/test-send", req, &out)
	return &out, err
}

// List lists SDK usage.
func (c *SDKUsageAdminClient) List(ctx context.Context) (*SDKUsageListResponse, error) {
	var out SDKUsageListResponse
	err := c.admin.request(ctx, http.MethodGet, "/sdk-usage", nil, &out)
	return &out, err
}

// PutDeprecation marks an SDK version deprecated.
func (c *SDKUsageAdminClient) PutDeprecation(ctx context.Context, sDK string, versionScope string, req SDKVersionDeprecationRequest) (*SDKVersionDeprecation, error) {
	var out SDKVersionDeprecation
	err := c.admin.request(ctx, http.MethodPut, "/sdk-usage/deprecations/"+url.PathEscape(sDK)+"/"+url.PathEscape(versionScope), req, &out)
	return &out, err
}

// DeleteDeprecation clears an SDK version deprecation.
func (c *SDKUsageAdminClient) DeleteDeprecation(ctx context.Context, sDK string, versionScope string) error {
	return c.admin.request(ctx, http.MethodDelete, "/sdk-usage/deprecations/"+url.PathEscape(sDK)+"/"+url.PathEscape(versionScope), nil, nil)
}

// List lists tenant accounting evidence.
func (c *AccountingEvidenceAdminClient) List(ctx context.Context, asOf string, limit int64) (*AccountingEvidenceListResponse, error) {
	var out AccountingEvidenceListResponse
	err := c.admin.request(ctx, http.MethodGet, siteQueryPath("/accounting/evidence", url.Values{"asOf": {asOf}, "limit": {positiveIntQuery(limit)}}), nil, &out)
	return &out, err
}

// Create appends tenant accounting evidence.
func (c *AccountingEvidenceAdminClient) Create(ctx context.Context, req AccountingEvidenceCreateRequest) (*AccountingEvidenceResponse, error) {
	var out AccountingEvidenceResponse
	err := c.admin.request(ctx, http.MethodPost, "/accounting/evidence", req, &out)
	return &out, err
}

// Get gets where one event is sent for the tenant.
func (c *NotificationDestinationAdminClient) Get(ctx context.Context, companySlug string, eventSlug NotificationDestinationEvent) (*NotificationDestination, error) {
	var out NotificationDestination
	err := c.admin.requestNonAdmin(ctx, http.MethodGet, c.admin.actingCompanyPath(siteQueryPath("/notification-destinations/"+url.PathEscape(string(eventSlug)), url.Values{companySlugQueryParameter: {companySlug}})), nil, &out)
	return &out, err
}

// Put sets where one event is sent for the tenant.
func (c *NotificationDestinationAdminClient) Put(ctx context.Context, companySlug string, eventSlug NotificationDestinationEvent, req NotificationDestinationRequest) (*NotificationDestination, error) {
	var out NotificationDestination
	err := c.admin.requestNonAdmin(ctx, http.MethodPut, c.admin.actingCompanyPath(siteQueryPath("/notification-destinations/"+url.PathEscape(string(eventSlug)), url.Values{companySlugQueryParameter: {companySlug}})), req, &out)
	return &out, err
}
