package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

// Generated-data cleanup: a plan previews what a scoped cleanup would delete,
// and a job deletes it in phases and keeps a proof. The credential names the
// tenant, except that ingest-api reads a system administrator's cleanup-job
// tenant only from the X-Custd-Effective-Company header, so the cleanup-job
// calls send ClientConfig.ActingCompanySlug there. The server then requires an
// active support grant for that company.

// CleanupJetStreamSelection names a stream and the consumers a cleanup drains.
type CleanupJetStreamSelection struct {
	Stream          string   `json:"stream"`
	Consumers       []string `json:"consumers,omitempty"`
	ExecutionStatus string   `json:"executionStatus,omitempty"`
}

// CleanupPlanCreateRequest asks for a generated-data cleanup plan. Times are RFC 3339.
type CleanupPlanCreateRequest struct {
	Mode               string                      `json:"mode"`
	Reason             string                      `json:"reason"`
	DryRun             bool                        `json:"dryRun"`
	Execute            bool                        `json:"execute"`
	ApprovedBy         string                      `json:"approvedBy"`
	ApprovedS3Prefixes []string                    `json:"approvedS3Prefixes"`
	EventTypeSlug      string                      `json:"eventTypeSlug"`
	StartedAt          string                      `json:"startedAt"`
	EndedAt            string                      `json:"endedAt"`
	JetStream          []CleanupJetStreamSelection `json:"jetStream"`
}

// CleanupPlanScope is the frozen scope a plan was computed for.
type CleanupPlanScope struct {
	Mode               string                      `json:"mode"`
	Reason             string                      `json:"reason,omitempty"`
	TenantSlug         string                      `json:"tenantSlug,omitempty"`
	EventTypeSlug      string                      `json:"eventTypeSlug,omitempty"`
	StartedAt          *time.Time                  `json:"startedAt,omitempty"`
	EndedAt            *time.Time                  `json:"endedAt,omitempty"`
	ApprovedS3Prefixes []string                    `json:"approvedS3Prefixes"`
	JetStream          []CleanupJetStreamSelection `json:"jetStream,omitempty"`
	JetStreamExecution string                      `json:"jetStreamExecution,omitempty"`
}

// CleanupPlanItem is what one phase of a plan selects.
type CleanupPlanItem struct {
	Phase           string `json:"phase"`
	DataClass       string `json:"dataClass"`
	StoragePrefix   string `json:"storagePrefix,omitempty"`
	SelectedCount   int64  `json:"selectedCount"`
	SelectedBytes   int64  `json:"selectedBytes"`
	ExecutionStatus string `json:"executionStatus,omitempty"`
}

// CleanupPlan is a previewed generated-data cleanup.
type CleanupPlan struct {
	PlanUUID        string            `json:"planUuid"`
	Mode            string            `json:"mode"`
	Status          string            `json:"status"`
	Reason          string            `json:"reason,omitempty"`
	SelectedRows    int64             `json:"selectedRows"`
	SelectedObjects int64             `json:"selectedObjects"`
	EstimatedBytes  int64             `json:"estimatedBytes"`
	Scope           CleanupPlanScope  `json:"scope"`
	Items           []CleanupPlanItem `json:"items,omitempty"`
	CreatedAt       *time.Time        `json:"createdAt,omitempty"`
}

// CleanupPlanList lists the tenant's recent cleanup plans.
type CleanupPlanList struct {
	Plans []CleanupPlan `json:"plans"`
}

// CleanupApproval is who approved a cleanup job and until when.
type CleanupApproval struct {
	Identity  string    `json:"Identity"`
	ExpiresAt time.Time `json:"ExpiresAt"`
}

// CleanupJobCreateRequest enqueues a scoped cleanup job. An empty TenantSlug
// uses the credential's tenant.
type CleanupJobCreateRequest struct {
	Mode               string                      `json:"mode"`
	Reason             string                      `json:"reason"`
	TenantSlug         string                      `json:"tenantSlug,omitempty"`
	EventTypeSlug      string                      `json:"eventTypeSlug,omitempty"`
	StartedAt          *time.Time                  `json:"startedAt,omitempty"`
	EndedAt            *time.Time                  `json:"endedAt,omitempty"`
	ApprovedS3Prefixes []string                    `json:"approvedS3Prefixes"`
	JetStream          []CleanupJetStreamSelection `json:"jetStream,omitempty"`
	Approval           CleanupApproval             `json:"approval"`
}

// CleanupObjectProgress is how far an object-storage phase has deleted.
type CleanupObjectProgress struct {
	SourceKind                string     `json:"sourceKind"`
	CursorDirection           string     `json:"cursorDirection"`
	CursorID                  int64      `json:"cursorId"`
	CursorReceivedAt          *time.Time `json:"cursorReceivedAt,omitempty"`
	BatchNumber               int64      `json:"batchNumber"`
	LastBatchIdentity         string     `json:"lastBatchIdentity,omitempty"`
	LedgerSelectedCount       int64      `json:"ledgerSelectedCount"`
	ProviderAcknowledgedCount int64      `json:"providerAcknowledgedCount"`
	ProviderRequestCount      int64      `json:"providerRequestCount"`
	ProviderFailureCount      int64      `json:"providerFailureCount"`
	SelectedBytes             int64      `json:"selectedBytes"`
	RetryCount                int64      `json:"retryCount"`
	PendingBatchIdentity      string     `json:"pendingBatchIdentity,omitempty"`
	PendingBatchKeyCount      int        `json:"pendingBatchKeyCount"`
	PendingBatchAttemptCount  int64      `json:"pendingBatchAttemptCount,omitempty"`
	PendingBatchTerminal      bool       `json:"pendingBatchTerminal,omitempty"`
	LastProgressAt            time.Time  `json:"lastProgressAt"`
	Version                   int64      `json:"version"`
}

// CleanupJobItem is one phase of a cleanup job.
type CleanupJobItem struct {
	ID             int64                  `json:"id"`
	JobID          int64                  `json:"jobId"`
	Phase          string                 `json:"phase"`
	DataClass      string                 `json:"dataClass"`
	StoragePrefix  string                 `json:"storagePrefix"`
	SelectedCount  int64                  `json:"selectedCount"`
	SelectedBytes  int64                  `json:"selectedBytes"`
	DeletedCount   int64                  `json:"deletedCount"`
	VerifiedCount  int64                  `json:"verifiedCount"`
	Status         string                 `json:"status"`
	FailureReason  string                 `json:"failureReason,omitempty"`
	ObjectProgress *CleanupObjectProgress `json:"objectProgress,omitempty"`
}

// CleanupJob is a cleanup job and its phases.
type CleanupJob struct {
	JobID         int64            `json:"jobId"`
	JobUUID       string           `json:"jobUuid"`
	Status        string           `json:"status"`
	TenantSlug    string           `json:"tenantSlug"`
	ScopeHash     string           `json:"scopeHash"`
	HoldKind      string           `json:"holdKind"`
	Approval      CleanupApproval  `json:"approval"`
	FailureReason string           `json:"failureReason,omitempty"`
	Items         []CleanupJobItem `json:"items"`
	CreatedAt     time.Time        `json:"createdAt"`
}

// CleanupJobList lists the tenant's cleanup jobs.
type CleanupJobList struct {
	Jobs []CleanupJob `json:"jobs"`
}

// CleanupJobProof is the proof a completed job recorded. Proof is the
// server's proof document, kept exactly as stored.
type CleanupJobProof struct {
	JobUUID string          `json:"jobUuid"`
	Status  string          `json:"status"`
	Proof   json.RawMessage `json:"proof"`
}

// ListCleanupPlans returns the tenant's recent cleanup plans.
func (c *IngestAdminClient) ListCleanupPlans(ctx context.Context) (*CleanupPlanList, error) {
	var out CleanupPlanList
	err := c.admin.requestNonAdmin(ctx, http.MethodGet, "/admin/cleanup/plans", nil, &out)
	return &out, err
}

// CreateCleanupPlan previews a generated-data cleanup.
func (c *IngestAdminClient) CreateCleanupPlan(ctx context.Context, req CleanupPlanCreateRequest) (*CleanupPlan, error) {
	var out CleanupPlan
	err := c.admin.requestNonAdmin(ctx, http.MethodPost, "/admin/cleanup/plans", req, &out)
	return &out, err
}

// ListCleanupJobs returns the tenant's cleanup jobs.
func (c *IngestAdminClient) ListCleanupJobs(ctx context.Context) (*CleanupJobList, error) {
	var out CleanupJobList
	err := c.admin.requestNonAdminWithHeaders(ctx, http.MethodGet, "/cleanup/jobs", nil, &out, c.cleanupJobHeaders())
	return &out, err
}

// CreateCleanupJob enqueues a scoped cleanup job.
func (c *IngestAdminClient) CreateCleanupJob(ctx context.Context, req CleanupJobCreateRequest) (*CleanupJob, error) {
	var out CleanupJob
	err := c.admin.requestNonAdminWithHeaders(ctx, http.MethodPost, "/cleanup/jobs", req, &out, c.cleanupJobHeaders())
	return &out, err
}

// GetCleanupJob returns one cleanup job and its phases.
func (c *IngestAdminClient) GetCleanupJob(ctx context.Context, jobUUID string) (*CleanupJob, error) {
	var out CleanupJob
	err := c.admin.requestNonAdminWithHeaders(ctx, http.MethodGet, cleanupJobPath(jobUUID), nil, &out, c.cleanupJobHeaders())
	return &out, err
}

// CancelCleanupJob cancels a cleanup job and returns its state.
func (c *IngestAdminClient) CancelCleanupJob(ctx context.Context, jobUUID string) (*CleanupJob, error) {
	var out CleanupJob
	err := c.admin.requestNonAdminWithHeaders(ctx, http.MethodPost, cleanupJobPath(jobUUID), nil, &out, c.cleanupJobHeaders())
	return &out, err
}

// GetCleanupJobProof returns a completed job's proof; an unfinished job is a conflict.
func (c *IngestAdminClient) GetCleanupJobProof(ctx context.Context, jobUUID string) (*CleanupJobProof, error) {
	var out CleanupJobProof
	err := c.admin.requestNonAdminWithHeaders(ctx, http.MethodGet, cleanupJobPath(jobUUID)+"/proof", nil, &out, c.cleanupJobHeaders())
	return &out, err
}

// cleanupJobHeaders names the acting company where ingest-api's cleanup-job
// routes read it. An unset ActingCompanySlug sends nothing, so the credential's
// own company applies.
func (c *IngestAdminClient) cleanupJobHeaders() map[string]string {
	company := c.admin.client.config.ActingCompanySlug
	if company == "" {
		return nil
	}
	return map[string]string{effectiveCompanyHeader: company}
}

func cleanupJobPath(jobUUID string) string {
	return "/cleanup/jobs/" + url.PathEscape(jobUUID)
}
