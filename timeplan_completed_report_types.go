package custd

import "time"

// The completed-duration report is owned by the measurement contract. Its
// field names are snake_case and are reproduced here verbatim: renaming one
// would silently drop the value on the wire.

// CompletedDurationRunStatus is the terminal status of a reported run.
type CompletedDurationRunStatus string

const (
	CompletedDurationRunStatusCompleted CompletedDurationRunStatus = "completed"
	CompletedDurationRunStatusCancelled CompletedDurationRunStatus = "cancelled"
)

// CompletedDurationBlockStatus is the status of one reported block.
type CompletedDurationBlockStatus string

const (
	CompletedDurationBlockStatusCompleted CompletedDurationBlockStatus = "completed"
	CompletedDurationBlockStatusSkipped   CompletedDurationBlockStatus = "skipped"
	CompletedDurationBlockStatusNotRun    CompletedDurationBlockStatus = "not_run"
	CompletedDurationBlockStatusPending   CompletedDurationBlockStatus = "pending"
	CompletedDurationBlockStatusActive    CompletedDurationBlockStatus = "active"
)

// CompletedDurationOutcome is the terminal outcome of one reported block.
type CompletedDurationOutcome string

const (
	CompletedDurationOutcomeCompleted CompletedDurationOutcome = "completed"
	CompletedDurationOutcomeSkipped   CompletedDurationOutcome = "skipped"
	CompletedDurationOutcomeCancelled CompletedDurationOutcome = "cancelled"
	CompletedDurationOutcomeNotRun    CompletedDurationOutcome = "not_run"
	CompletedDurationOutcomeCensored  CompletedDurationOutcome = "censored"
)

// CompletedDurationReport is the canonical terminal report of one completed
// time-plan run.
type CompletedDurationReport struct {
	TenantID              string                                        `json:"tenant_id"`
	Summary               CompletedDurationReportSummary                `json:"summary"`
	Blocks                []CompletedDurationReportBlock                `json:"blocks"`
	Adjustments           []CompletedDurationReportAdjustment           `json:"adjustments"`
	Annotations           []CompletedDurationReportAnnotation           `json:"annotations"`
	PredictionComparisons []CompletedDurationReportPredictionComparison `json:"prediction_comparisons"`
	Provenance            CompletedDurationReportProvenance             `json:"provenance"`
}

// CompletedDurationReportSummary identifies the reported run revision.
type CompletedDurationReportSummaryIdentity struct {
	RunID         string                     `json:"run_id"`
	PlanID        string                     `json:"plan_id"`
	VersionID     string                     `json:"version_id"`
	RunStatus     CompletedDurationRunStatus `json:"run_status"`
	StreamVersion int64                      `json:"stream_version"`
}

// CompletedDurationReportSummarySchedule contains the scheduled and effective
// times of the reported run.
type CompletedDurationReportSummarySchedule struct {
	ScheduledStartsAt *time.Time `json:"scheduled_starts_at,omitempty"`
	ScheduledEndsAt   *time.Time `json:"scheduled_ends_at,omitempty"`
	EffectiveStartsAt *time.Time `json:"effective_starts_at,omitempty"`
	EffectiveEndsAt   *time.Time `json:"effective_ends_at,omitempty"`
}

// CompletedDurationReportSummaryCompletion contains the terminal horizon facts.
type CompletedDurationReportSummaryCompletion struct {
	CompletedAt         time.Time `json:"completed_at"`
	BaselineHorizonMS   int64     `json:"baseline_horizon_ms"`
	ExecutableHorizonMS *int64    `json:"executable_horizon_ms,omitempty"`
	LostMS              int64     `json:"lost_ms"`
	UnusedMS            int64     `json:"unused_ms"`
	OverrunMS           int64     `json:"overrun_ms"`
}

// CompletedDurationReportSummaryTotals contains the planned and actual totals.
type CompletedDurationReportSummaryTotals struct {
	PlannedTotalMS  int64 `json:"planned_total_ms"`
	CurrentTotalMS  int64 `json:"current_total_ms"`
	ActualTotalMS   int64 `json:"actual_total_ms"`
	FinalVarianceMS int64 `json:"final_variance_ms"`
}

// CompletedDurationReportSummaryCounts contains the block outcome counts.
type CompletedDurationReportSummaryCounts struct {
	ActualComplete      bool `json:"actual_complete"`
	BlockCount          int  `json:"block_count"`
	ActualBlockCount    int  `json:"actual_block_count"`
	CompletedBlockCount int  `json:"completed_block_count"`
	SkippedBlockCount   int  `json:"skipped_block_count"`
	CancelledBlockCount int  `json:"cancelled_block_count"`
	NotRunBlockCount    int  `json:"not_run_block_count"`
	CensoredBlockCount  int  `json:"censored_block_count"`
}

// CompletedDurationReportSummary is the flat summary of one reported run.
type CompletedDurationReportSummary struct {
	CompletedDurationReportSummaryIdentity
	CompletedDurationReportSummarySchedule
	CompletedDurationReportSummaryCompletion
	CompletedDurationReportSummaryTotals
	CompletedDurationReportSummaryCounts
}

// CompletedDurationReportBlockIdentity contains the ordered block identity.
type CompletedDurationReportBlockIdentity struct {
	BlockUUID   string   `json:"block_uuid"`
	Sequence    int      `json:"sequence"`
	SemanticKey string   `json:"semantic_key"`
	Category    string   `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// CompletedDurationReportBlockTiming contains the planned and observed
// durations of one reported block.
type CompletedDurationReportBlockTiming struct {
	PlannedMS          int64      `json:"planned_ms"`
	CurrentMS          int64      `json:"current_ms"`
	AllocatedAtStartMS *int64     `json:"allocated_at_start_ms,omitempty"`
	ActualMS           *int64     `json:"actual_ms,omitempty"`
	WallStartedAt      *time.Time `json:"wall_started_at,omitempty"`
	WallEndedAt        *time.Time `json:"wall_ended_at,omitempty"`
}

// CompletedDurationReportBlockOutcome contains the terminal status and outcome
// of one reported block.
type CompletedDurationReportBlockOutcome struct {
	Status  CompletedDurationBlockStatus `json:"status"`
	Outcome CompletedDurationOutcome     `json:"outcome"`
}

// CompletedDurationReportBlock carries the planned, current, actual, and
// outcome facts of one reported block.
type CompletedDurationReportBlock struct {
	CompletedDurationReportBlockIdentity
	CompletedDurationReportBlockTiming
	CompletedDurationReportBlockOutcome
}

// CompletedDurationReportAllocation is one block's exact integer allocation.
type CompletedDurationReportAllocation struct {
	BlockUUID  string `json:"block_uuid"`
	Sequence   int    `json:"sequence"`
	DurationMS int64  `json:"duration_ms"`
}

// CompletedDurationReportAllocationChange is one allocation correction.
type CompletedDurationReportAllocationChange struct {
	BlockUUID string `json:"block_uuid"`
	FromMS    int64  `json:"from_ms"`
	ToMS      int64  `json:"to_ms"`
}

// CompletedDurationReportCalculationReceipt is the allocator receipt of one
// transition.
type CompletedDurationReportCalculationReceipt struct {
	AllocatorVersion string                                    `json:"allocator_version,omitempty"`
	Reason           string                                    `json:"reason,omitempty"`
	Summary          string                                    `json:"summary,omitempty"`
	Source           []CompletedDurationReportAllocation       `json:"source,omitempty"`
	Result           []CompletedDurationReportAllocation       `json:"result,omitempty"`
	Changes          []CompletedDurationReportAllocationChange `json:"changes,omitempty"`
}

// CompletedDurationReportAdjustmentIdentity contains the immutable transition
// identity of one adjustment.
type CompletedDurationReportAdjustmentIdentity struct {
	TransitionUUID string `json:"transition_uuid"`
	StreamVersion  int64  `json:"stream_version"`
	CommandID      string `json:"command_id"`
	Type           string `json:"type"`
}

// CompletedDurationReportAdjustmentActor contains the transition actor and
// timing facts of one adjustment.
type CompletedDurationReportAdjustmentActor struct {
	ActorKind        string     `json:"actor_kind,omitempty"`
	ActorRef         string     `json:"actor_ref,omitempty"`
	OccurredAt       time.Time  `json:"occurred_at"`
	ClientOccurredAt *time.Time `json:"client_occurred_at,omitempty"`
}

// CompletedDurationReportAdjustmentMetadata contains the transition reason and
// lineage of one adjustment.
type CompletedDurationReportAdjustmentMetadata struct {
	Reason                 string `json:"reason,omitempty"`
	PreviousStatus         string `json:"previous_status,omitempty"`
	CurrentStatus          string `json:"current_status,omitempty"`
	AllocatorVersion       string `json:"allocator_version,omitempty"`
	SchemaVersion          string `json:"schema_version,omitempty"`
	SupersedesTransitionID string `json:"supersedes_transition_id,omitempty"`
}

// CompletedDurationReportAdjustment is one immutable transition-ledger item.
type CompletedDurationReportAdjustment struct {
	CompletedDurationReportAdjustmentIdentity
	CompletedDurationReportAdjustmentActor
	CompletedDurationReportAdjustmentMetadata
	Receipt CompletedDurationReportCalculationReceipt `json:"receipt"`
}

// CompletedDurationReportAnnotationIdentity contains the identity of one
// reported annotation.
type CompletedDurationReportAnnotationIdentity struct {
	UUID      string `json:"uuid"`
	BlockUUID string `json:"block_uuid,omitempty"`
	Type      string `json:"type"`
}

// CompletedDurationReportAnnotationContent contains the bounded note and
// action content of one reported annotation.
type CompletedDurationReportAnnotationContent struct {
	Text           string     `json:"text,omitempty"`
	MarkerLabel    string     `json:"marker_label,omitempty"`
	DecisionStatus string     `json:"decision_status,omitempty"`
	AssigneeRef    string     `json:"assignee_ref,omitempty"`
	DueDate        *time.Time `json:"due_date,omitempty"`
	ActionStatus   string     `json:"action_status,omitempty"`
	SupersedesUUID string     `json:"supersedes_uuid,omitempty"`
}

// CompletedDurationReportAnnotationAudit contains the recording and redaction
// facts of one reported annotation.
type CompletedDurationReportAnnotationAudit struct {
	RecordedAt time.Time  `json:"recorded_at"`
	ActorKind  string     `json:"actor_kind"`
	ActorRef   string     `json:"actor_ref,omitempty"`
	RedactedAt *time.Time `json:"redacted_at,omitempty"`
}

// CompletedDurationReportAnnotation is one reported run annotation.
type CompletedDurationReportAnnotation struct {
	CompletedDurationReportAnnotationIdentity
	CompletedDurationReportAnnotationContent
	CompletedDurationReportAnnotationAudit
}

// CompletedDurationReportPrediction is one predicted block duration.
type CompletedDurationReportPrediction struct {
	BlockUUID     string    `json:"block_uuid"`
	Method        string    `json:"method"`
	MethodVersion string    `json:"method_version"`
	SampleCount   int       `json:"sample_count"`
	PointMS       int64     `json:"point_ms"`
	P10MS         *int64    `json:"p10_ms,omitempty"`
	P90MS         *int64    `json:"p90_ms,omitempty"`
	InputHash     string    `json:"input_hash"`
	Warnings      []string  `json:"warnings,omitempty"`
	PredictedAt   time.Time `json:"predicted_at"`
}

// CompletedDurationReportPredictionComparison compares one prediction with
// the observed duration.
type CompletedDurationReportPredictionComparison struct {
	CompletedDurationReportPrediction
	ActualMS        *int64 `json:"actual_ms,omitempty"`
	AbsoluteErrorMS *int64 `json:"absolute_error_ms,omitempty"`
	IntervalCovered *bool  `json:"interval_covered,omitempty"`
}

// CompletedDurationReportProvenance records how the report was generated.
type CompletedDurationReportProvenance struct {
	ReportVersion                string    `json:"report_version"`
	InputHash                    string    `json:"input_hash"`
	GeneratedAt                  time.Time `json:"generated_at"`
	Corrected                    bool      `json:"corrected"`
	MeasurementProjectionVersion string    `json:"measurement_projection_version,omitempty"`
	MeasurementInputHash         string    `json:"measurement_input_hash,omitempty"`
	ForecastEvaluationHash       string    `json:"forecast_evaluation_hash,omitempty"`
	AdjustmentCount              int       `json:"adjustment_count"`
	AnnotationCount              int       `json:"annotation_count"`
	PredictionCount              int       `json:"prediction_count"`
}
