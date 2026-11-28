package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// SystemAdminClient owns operator diagnostics: health, the durable scheduler,
// and the aggregated system status.
//
// These are read-mostly operator surfaces. The scheduler actions are control
// actions, so each returns the request identifier even when it has no body:
// a pause that was refused still needs to be quotable.
type SystemAdminClient struct {
	admin *AdminClient
}

// Health is the process liveness answer.
type Health struct {
	Status string `json:"status"`
}

// ScheduledTaskRun is one execution of a scheduled task.
type ScheduledTaskRun struct {
	RunUUID        string     `json:"runUuid"`
	TaskName       string     `json:"taskName"`
	Trigger        string     `json:"trigger"`
	Status         string     `json:"status"`
	ScheduledFor   time.Time  `json:"scheduledFor"`
	StartedAt      *time.Time `json:"startedAt,omitempty"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	DurationMS     int64      `json:"durationMs"`
	WorkerID       string     `json:"workerId,omitempty"`
	ItemsTotal     int64      `json:"itemsTotal"`
	ItemsCompleted int64      `json:"itemsCompleted"`
	ItemsFailed    int64      `json:"itemsFailed"`
	FailureReason  string     `json:"failureReason,omitempty"`
}

// scheduledTaskRunResponse is the envelope the server wraps a triggered or
// retried run in.
type scheduledTaskRunResponse struct {
	Run ScheduledTaskRun `json:"run"`
}

// ScheduledTaskSummary is one durable task's schedule and current pressure.
type ScheduledTaskSummary struct {
	TaskName           string            `json:"taskName"`
	DisplayName        string            `json:"displayName"`
	Description        string            `json:"description"`
	OwnerService       string            `json:"ownerService"`
	Enabled            bool              `json:"enabled"`
	ScheduleKind       string            `json:"scheduleKind"`
	ScheduleExpression string            `json:"scheduleExpression"`
	Timezone           string            `json:"timezone"`
	NextRunAt          *time.Time        `json:"nextRunAt,omitempty"`
	OverlapPolicy      string            `json:"overlapPolicy"`
	LockBackend        string            `json:"lockBackend"`
	TimeoutSeconds     int               `json:"timeoutSeconds"`
	ManualRunsAllowed  bool              `json:"manualRunsAllowed"`
	ForceRunsAllowed   bool              `json:"forceRunsAllowed"`
	QueuedRuns         int64             `json:"queuedRuns"`
	RunningRuns        int64             `json:"runningRuns"`
	FailedRuns         int64             `json:"failedRuns"`
	LatestRun          *ScheduledTaskRun `json:"latestRun,omitempty"`
}

// ScheduledTaskRunProof is the numeric proof a run records. Zero means the
// task does not report that count.
type ScheduledTaskRunProof struct {
	JobsCompacted         int64 `json:"jobsCompacted,omitempty"`
	RowsCompacted         int64 `json:"rowsCompacted,omitempty"`
	BytesCompacted        int64 `json:"bytesCompacted,omitempty"`
	SourcesRetired        int64 `json:"sourcesRetired,omitempty"`
	BytesRetired          int64 `json:"bytesRetired,omitempty"`
	RegionalTrendReads    int64 `json:"regionalTrendReads,omitempty"`
	RegionalTrendItems    int64 `json:"regionalTrendItems,omitempty"`
	RegionalTrendFailures int64 `json:"regionalTrendFailures,omitempty"`
	HackerNewsReads       int64 `json:"hackerNewsReads,omitempty"`
	HackerNewsItems       int64 `json:"hackerNewsItems,omitempty"`
	HackerNewsFailures    int64 `json:"hackerNewsFailures,omitempty"`
}

// ScheduledTaskRunDetail is one run in a task's history, with its proof.
type ScheduledTaskRunDetail struct {
	ScheduledTaskRun
	ScheduledTaskRunProof
}

// ScheduledTaskDetailSummary is the task a detail read returns; its latest run
// carries proof, unlike the list view's.
type ScheduledTaskDetailSummary struct {
	ScheduledTaskSummary
	LatestRun *ScheduledTaskRunDetail `json:"latestRun,omitempty"`
}

// ScheduledTaskDetail is one task and its recent run history.
type ScheduledTaskDetail struct {
	Task ScheduledTaskDetailSummary `json:"task"`
	Runs []ScheduledTaskRunDetail   `json:"runs"`
}

// ScheduledTaskListResponse lists the durable tasks.
type ScheduledTaskListResponse struct {
	Tasks []ScheduledTaskSummary `json:"tasks"`
}

// SystemStatusStorage is the fleet-wide storage position.
type SystemStatusStorage struct {
	TenantCount            int       `json:"tenantCount"`
	RetainedLogicalBytes   int64     `json:"retainedLogicalBytes"`
	AttributedStorageBytes int64     `json:"attributedStorageBytes"`
	EffectiveLimitBytes    int64     `json:"effectiveLimitBytes"`
	PercentUsed            float64   `json:"percentUsed"`
	MeasuredAt             time.Time `json:"measuredAt"`
}

// SystemStatusTenant is one tenant's storage position, ranked by the server.
type SystemStatusTenant struct {
	CompanySlug            string  `json:"companySlug"`
	CompanyName            string  `json:"companyName"`
	RankReason             string  `json:"rankReason,omitempty"`
	RetainedLogicalBytes   int64   `json:"retainedLogicalBytes"`
	AttributedStorageBytes int64   `json:"attributedStorageBytes"`
	EffectiveLimitBytes    int64   `json:"effectiveLimitBytes"`
	PercentUsed            float64 `json:"percentUsed"`
	WarningState           string  `json:"warningState"`
	RejectedEvents         int64   `json:"rejectedEvents,omitempty"`
	ClaimQueued            int64   `json:"claimQueued,omitempty"`
}

// SystemStageProgress is how a pipeline stage's backlog is trending.
type SystemStageProgress struct {
	Classification         string  `json:"classification,omitempty"`
	Depth                  int64   `json:"depth,omitempty"`
	DrainRatePerSecond     float64 `json:"drainRatePerSecond,omitempty"`
	EtaSeconds             int64   `json:"etaSeconds,omitempty"`
	RetentionExpirySeconds int64   `json:"retentionExpirySeconds,omitempty"`
	Note                   string  `json:"note,omitempty"`
}

// SystemStageSignals are the per-stage diagnostic readings behind a stage's
// status. A nil probe means the probe did not run.
type SystemStageSignals struct {
	HealthProbeOK                    *bool   `json:"healthProbeOk,omitempty"`
	MetricsProbeOK                   *bool   `json:"metricsProbeOk,omitempty"`
	ResourceMetricsOK                *bool   `json:"resourceMetricsOk,omitempty"`
	Pending                          int64   `json:"pending,omitempty"`
	AckPending                       int64   `json:"ackPending,omitempty"`
	Redelivered                      int64   `json:"redelivered,omitempty"`
	ConsumerActive                   *bool   `json:"consumerActive,omitempty"`
	RetentionSecondsLeft             int64   `json:"retentionSecondsLeft,omitempty"`
	DeliveredStreamSequence          int64   `json:"deliveredStreamSequence,omitempty"`
	AckFloorStreamSequence           int64   `json:"ackFloorStreamSequence,omitempty"`
	AckPendingUtilization            float64 `json:"ackPendingUtilization,omitempty"`
	SlowConsumer                     bool    `json:"slowConsumer,omitempty"`
	AcceptedRawLandingPending        int64   `json:"acceptedRawLandingPending,omitempty"`
	AcceptedRawLandingFailed         int64   `json:"acceptedRawLandingFailed,omitempty"`
	AcceptedRawLandingUploadFailures int64   `json:"acceptedRawLandingUploadFailures,omitempty"`
	SegmentPending                   int64   `json:"segmentPending,omitempty"`
	CompactionBacklog                int64   `json:"compactionBacklog,omitempty"`
	CompactionBacklogAgeSeconds      int64   `json:"compactionBacklogAgeSeconds,omitempty"`
	RollupBacklog                    int64   `json:"rollupBacklog,omitempty"`
	MemoryUsageBytes                 int64   `json:"memoryUsageBytes,omitempty"`
	MemoryLimitBytes                 int64   `json:"memoryLimitBytes,omitempty"`
	CPUMilli                         int64   `json:"cpuMilli,omitempty"`
	CPULimitMilli                    int64   `json:"cpuLimitMilli,omitempty"`
	RestartCount                     int64   `json:"restartCount,omitempty"`
	OOMKillCount                     int64   `json:"oomKillCount,omitempty"`
	LivenessFlapCount                int64   `json:"livenessFlapCount,omitempty"`
}

// SystemHealthStatus is the health of the system or one of its stages.
type SystemHealthStatus string

const (
	SystemHealthStatusHealthy  SystemHealthStatus = "healthy"
	SystemHealthStatusDegraded SystemHealthStatus = "degraded"
	SystemHealthStatusFailing  SystemHealthStatus = "failing"
	SystemHealthStatusUnknown  SystemHealthStatus = "unknown"
)

// SystemStatusStage is one pipeline stage's health, with its backlog trend and
// the signals behind the status when the server measured them.
type SystemStatusStage struct {
	Name     string               `json:"name"`
	Status   SystemHealthStatus   `json:"status"`
	Reason   string               `json:"reason,omitempty"`
	Progress *SystemStageProgress `json:"progress,omitempty"`
	Signals  *SystemStageSignals  `json:"signals,omitempty"`
}

// DesiredStateDiagnostics counts the desired-state reconciliation queue.
type DesiredStateDiagnostics struct {
	Backlog                 int64 `json:"backlog,omitempty"`
	OldestPendingAgeSeconds int64 `json:"oldestPendingAgeSeconds,omitempty"`
	TerminalTotal           int64 `json:"terminalTotal,omitempty"`
	ExhaustedTotal          int64 `json:"exhaustedTotal,omitempty"`
}

// SystemStatusKPIs are the fleet counters. A zero value means "not measured",
// not "measured as zero", for the fields the server marks omitempty.
type SystemStatusKPIs struct {
	EventsPerMinute          float64                 `json:"eventsPerMinute,omitempty"`
	BacklogDepth             int64                   `json:"backlogDepth,omitempty"`
	InvariantViolations      int64                   `json:"invariantViolations,omitempty"`
	ClaimBacklog             int64                   `json:"claimBacklog,omitempty"`
	ClaimBacklogOldestAge    int64                   `json:"claimBacklogOldestAgeSeconds,omitempty"`
	FailedRawLandingObjects  int64                   `json:"failedRawLandingObjects,omitempty"`
	RawLandingUploadFailures int64                   `json:"rawLandingUploadFailures,omitempty"`
	DesiredState             DesiredStateDiagnostics `json:"desiredState"`
}

// SystemRealmCapacity is the deployment's auth realm slot pool. When Observed
// is false the pool could not be read: the counts are zero and Reason says why.
type SystemRealmCapacity struct {
	Observed    bool   `json:"observed"`
	Declared    int    `json:"declared"`
	Occupied    int    `json:"occupied"`
	Quarantined int    `json:"quarantined"`
	Unavailable int    `json:"unavailable"`
	Exhausted   bool   `json:"exhausted"`
	Reason      string `json:"reason,omitempty"`
}

// OperationalStageSummary is one step of the received, durable, queryable flow.
type OperationalStageSummary struct {
	Status SystemHealthStatus `json:"status"`
	Reason string             `json:"reason,omitempty"`
}

// OperationalBacklogSummary is the pipeline backlog and how it is trending.
type OperationalBacklogSummary struct {
	Depth          int64  `json:"depth"`
	Classification string `json:"classification"`
}

// OperationalSummary is the one-line answer to "is data flowing": whether
// events are received, durable and queryable, and the backlog between them.
type OperationalSummary struct {
	Status    SystemHealthStatus        `json:"status"`
	Received  OperationalStageSummary   `json:"received"`
	Durable   OperationalStageSummary   `json:"durable"`
	Queryable OperationalStageSummary   `json:"queryable"`
	Backlog   OperationalBacklogSummary `json:"backlog"`
}

// SystemStatus is the single high-level view: one overall status first, then
// the stages, storage and tenants to drill into.
type SystemStatus struct {
	OverallStatus string               `json:"overallStatus"`
	GeneratedAt   time.Time            `json:"generatedAt"`
	Summary       OperationalSummary   `json:"summary"`
	RealmCapacity SystemRealmCapacity  `json:"realmCapacity"`
	KPIs          SystemStatusKPIs     `json:"kpis"`
	Stages        []SystemStatusStage  `json:"stages"`
	Storage       SystemStatusStorage  `json:"storage"`
	TopTenants    []SystemStatusTenant `json:"topTenants"`
}

// The Site Insights scheduler tasks, in the order ListSiteInsightsTasks shows them. Page rechecks
// run inside site-audits; URL inspections, IndexNow, rankings and Bing links inside
// site-search-imports.
const (
	SiteInsightsTaskChecks        = "site-checks"
	SiteInsightsTaskAudits        = "site-audits"
	SiteInsightsTaskSearchImports = "site-search-imports"
	SiteInsightsTaskWebData       = "site-web-data"
	SiteInsightsTaskWebGraph      = "site-web-graph"
	SiteInsightsTaskAccessLogs    = "site-access-logs"
	SiteInsightsTaskSearchPool    = "site-search-pool"
)

// SiteInsightsBacklog is one count of work waiting for a task, summed over every tenant.
type SiteInsightsBacklog struct {
	Name    string `json:"name"`
	Count   int64  `json:"count"`
	Meaning string `json:"meaning"`
}

// SiteInsightsTaskStatus is one Site Insights task: whether this deployment's scheduler registered
// and enabled it, its latest run, its latest failed run among the recent runs and its backlog.
type SiteInsightsTaskStatus struct {
	TaskName      string                `json:"taskName"`
	Registered    bool                  `json:"registered"`
	Enabled       bool                  `json:"enabled"`
	NextRunAt     *time.Time            `json:"nextRunAt,omitempty"`
	LastRunAt     *time.Time            `json:"lastRunAt,omitempty"`
	LastRunStatus string                `json:"lastRunStatus,omitempty"`
	LastFailureAt *time.Time            `json:"lastFailureAt,omitempty"`
	LastFailure   string                `json:"lastFailure,omitempty"`
	Backlog       []SiteInsightsBacklog `json:"backlog"`
}

// SiteInsightsTaskList is every Site Insights scheduler task, counted across all tenants; it names
// no tenant.
type SiteInsightsTaskList struct {
	GeneratedAt time.Time                `json:"generatedAt"`
	Tasks       []SiteInsightsTaskStatus `json:"tasks"`
}

const (
	scheduledTasksPath    = "/scheduled-tasks"
	siteInsightsTasksPath = "/site-insights/tasks"
	scheduledTaskRunsPath = "/scheduled-task-runs/"
	systemStatusPath      = "/system/status"
)

// Status returns the aggregated system view.
func (c *SystemAdminClient) Status(
	ctx context.Context,
) (*SystemStatus, ResponseMetadata, error) {
	var out SystemStatus
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodGet, systemStatusPath, nil, &out)
	return &out, metadata, err
}

// ListScheduledTasks returns the durable tasks with their current pressure.
func (c *SystemAdminClient) ListScheduledTasks(
	ctx context.Context,
) (*ScheduledTaskListResponse, ResponseMetadata, error) {
	var out ScheduledTaskListResponse
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodGet, scheduledTasksPath, nil, &out)
	return &out, metadata, err
}

// ListSiteInsightsTasks returns the Site Insights scheduler tasks with their run state and
// backlog; system administrators only.
func (c *SystemAdminClient) ListSiteInsightsTasks(
	ctx context.Context,
) (*SiteInsightsTaskList, ResponseMetadata, error) {
	var out SiteInsightsTaskList
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodGet, siteInsightsTasksPath, nil, &out)
	return &out, metadata, err
}

// GetScheduledTask reads one task and its recent runs.
func (c *SystemAdminClient) GetScheduledTask(
	ctx context.Context,
	taskName string,
) (*ScheduledTaskDetail, ResponseMetadata, error) {
	var out ScheduledTaskDetail
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodGet, scheduledTasksPath+"/"+url.PathEscape(taskName), nil, &out,
	)
	return &out, metadata, err
}

// RunScheduledTask triggers a run now. The server decides whether the task's
// overlap and force policy permits it.
func (c *SystemAdminClient) RunScheduledTask(
	ctx context.Context,
	taskName string,
) (*ScheduledTaskRun, ResponseMetadata, error) {
	return c.runScheduledTask(ctx, taskName, nil)
}

// ScheduledTaskRunRequest carries the optional operator force decision.
type ScheduledTaskRunRequest struct {
	Force bool `json:"force"`
}

// RunScheduledTaskWithOptions requests a run under the server's force and overlap policy.
func (c *SystemAdminClient) RunScheduledTaskWithOptions(
	ctx context.Context, taskName string, request *ScheduledTaskRunRequest,
) (*ScheduledTaskRun, ResponseMetadata, error) {
	if request == nil {
		return c.RunScheduledTask(ctx, taskName)
	}
	return c.runScheduledTask(ctx, taskName, request)
}

func (c *SystemAdminClient) runScheduledTask(
	ctx context.Context, taskName string, request any,
) (*ScheduledTaskRun, ResponseMetadata, error) {
	var out scheduledTaskRunResponse
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodPost, scheduledTasksPath+"/"+url.PathEscape(taskName)+"/run", request, &out,
	)
	return &out.Run, metadata, err
}

// PauseScheduledTask stops future runs. There is no body, so the metadata is
// the whole result.
func (c *SystemAdminClient) PauseScheduledTask(
	ctx context.Context,
	taskName string,
) (ResponseMetadata, error) {
	return c.taskAction(ctx, taskName, "/pause")
}

// ResumeScheduledTask re-enables future runs.
func (c *SystemAdminClient) ResumeScheduledTask(
	ctx context.Context,
	taskName string,
) (ResponseMetadata, error) {
	return c.taskAction(ctx, taskName, "/resume")
}

// RetryScheduledTaskRun re-queues one failed run and returns the queued run.
func (c *SystemAdminClient) RetryScheduledTaskRun(
	ctx context.Context,
	runUUID string,
) (*ScheduledTaskRun, ResponseMetadata, error) {
	var out scheduledTaskRunResponse
	path := scheduledTaskRunsPath + url.PathEscape(runUUID) + "/retry"
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodPost, path, nil, &out)
	return &out.Run, metadata, err
}

func (c *SystemAdminClient) taskAction(
	ctx context.Context,
	taskName string,
	action string,
) (ResponseMetadata, error) {
	path := scheduledTasksPath + "/" + url.PathEscape(taskName) + action
	return c.admin.requestWithMetadata(ctx, http.MethodPost, path, nil, nil)
}
