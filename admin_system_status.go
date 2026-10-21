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

// ScheduledTaskListResponse lists the durable tasks.
type ScheduledTaskListResponse struct {
	Tasks []ScheduledTaskSummary `json:"tasks"`
}

// SystemStatusStorage is the fleet-wide storage position.
type SystemStatusStorage struct {
	TenantCount          int       `json:"tenantCount"`
	RetainedLogicalBytes int64     `json:"retainedLogicalBytes"`
	EffectiveLimitBytes  int64     `json:"effectiveLimitBytes"`
	PercentUsed          float64   `json:"percentUsed"`
	MeasuredAt           time.Time `json:"measuredAt"`
}

// SystemStatusTenant is one tenant's storage position, ranked by the server.
type SystemStatusTenant struct {
	CompanySlug          string  `json:"companySlug"`
	CompanyName          string  `json:"companyName"`
	RankReason           string  `json:"rankReason,omitempty"`
	RetainedLogicalBytes int64   `json:"retainedLogicalBytes"`
	EffectiveLimitBytes  int64   `json:"effectiveLimitBytes"`
	PercentUsed          float64 `json:"percentUsed"`
	WarningState         string  `json:"warningState"`
	RejectedEvents       int64   `json:"rejectedEvents,omitempty"`
}

// SystemStatusStage is one pipeline stage's health. The per-stage progress and
// signal detail is not modelled here yet; a caller that needs it reads the
// raw answer until those shapes are added.
type SystemStatusStage struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

// SystemStatusKPIs are the fleet counters. A zero value means "not measured",
// not "measured as zero", for the fields the server marks omitempty.
type SystemStatusKPIs struct {
	EventsPerMinute          float64 `json:"eventsPerMinute,omitempty"`
	BacklogDepth             int64   `json:"backlogDepth,omitempty"`
	InvariantViolations      int64   `json:"invariantViolations,omitempty"`
	ClaimBacklog             int64   `json:"claimBacklog,omitempty"`
	ClaimBacklogOldestAge    int64   `json:"claimBacklogOldestAgeSeconds,omitempty"`
	FailedRawLandingObjects  int64   `json:"failedRawLandingObjects,omitempty"`
	RawLandingUploadFailures int64   `json:"rawLandingUploadFailures,omitempty"`
}

// SystemStatus is the single high-level view: one overall status first, then
// the stages, storage and tenants to drill into.
type SystemStatus struct {
	OverallStatus string               `json:"overallStatus"`
	GeneratedAt   time.Time            `json:"generatedAt"`
	KPIs          SystemStatusKPIs     `json:"kpis"`
	Stages        []SystemStatusStage  `json:"stages"`
	Storage       SystemStatusStorage  `json:"storage"`
	TopTenants    []SystemStatusTenant `json:"topTenants"`
}

const (
	scheduledTasksPath    = "/scheduled-tasks"
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

// GetScheduledTask reads one task.
func (c *SystemAdminClient) GetScheduledTask(
	ctx context.Context,
	taskName string,
) (*ScheduledTaskSummary, ResponseMetadata, error) {
	var out ScheduledTaskSummary
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
	var out ScheduledTaskRun
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodPost, scheduledTasksPath+"/"+url.PathEscape(taskName)+"/run", nil, &out,
	)
	return &out, metadata, err
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

// RetryScheduledTaskRun re-queues one failed run.
func (c *SystemAdminClient) RetryScheduledTaskRun(
	ctx context.Context,
	runUUID string,
) (ResponseMetadata, error) {
	path := scheduledTaskRunsPath + url.PathEscape(runUUID) + "/retry"
	return c.admin.requestWithMetadata(ctx, http.MethodPost, path, nil, nil)
}

func (c *SystemAdminClient) taskAction(
	ctx context.Context,
	taskName string,
	action string,
) (ResponseMetadata, error) {
	path := scheduledTasksPath + "/" + url.PathEscape(taskName) + action
	return c.admin.requestWithMetadata(ctx, http.MethodPost, path, nil, nil)
}
