package custd

import "time"

// The completed-run export source is the typed handoff from the canonical
// completed report to the reporting export worker. It keeps the report's
// typed sections instead of a renderer-specific payload.

// CompletedRunExportSource is the typed handoff from the canonical completed
// report to the reporting export owner.
type CompletedRunExportSource struct {
	ContractVersion string                      `json:"contractVersion"`
	Identity        CompletedRunExportIdentity  `json:"identity"`
	Sections        CompletedRunExportSections  `json:"sections"`
	Redaction       CompletedRunExportRedaction `json:"redaction"`
	Metadata        CompletedRunExportMetadata  `json:"metadata"`
	Integrity       CompletedRunExportIntegrity `json:"integrity"`
}

// CompletedRunExportIdentity binds a completed report to its tenant and the
// published reporting-pack generation selected for the export.
type CompletedRunExportIdentity struct {
	TenantID       string `json:"tenantId"`
	RunID          string `json:"runId"`
	PlanID         string `json:"planId"`
	VersionID      string `json:"versionId"`
	PackKey        string `json:"packKey"`
	PackGeneration int    `json:"packGeneration"`
}

// CompletedRunExportSectionKind is one section of the completed-run source.
type CompletedRunExportSectionKind string

const (
	CompletedRunExportSectionSummary              CompletedRunExportSectionKind = "summary"
	CompletedRunExportSectionBlocks               CompletedRunExportSectionKind = "blocks"
	CompletedRunExportSectionAdjustments          CompletedRunExportSectionKind = "adjustments"
	CompletedRunExportSectionAnnotationsActions   CompletedRunExportSectionKind = "annotations_actions"
	CompletedRunExportSectionPredictionComparison CompletedRunExportSectionKind = "prediction_comparison"
	CompletedRunExportSectionProvenance           CompletedRunExportSectionKind = "provenance"
)

// CompletedRunExportSectionOrder names the ordered, typed source sections.
type CompletedRunExportSectionOrder []CompletedRunExportSectionKind

// CompletedRunExportBlockSection preserves the canonical block order.
type CompletedRunExportBlockSection struct {
	Items []CompletedDurationReportBlock `json:"items"`
}

// CompletedRunExportAdjustmentSection preserves the canonical adjustment
// order.
type CompletedRunExportAdjustmentSection struct {
	Items []CompletedDurationReportAdjustment `json:"items"`
}

// CompletedRunExportAnnotationsActionsSection preserves the canonical
// annotation order.
type CompletedRunExportAnnotationsActionsSection struct {
	Items []CompletedDurationReportAnnotation `json:"items"`
}

// CompletedRunExportPredictionComparisonSection preserves the canonical
// prediction comparisons.
type CompletedRunExportPredictionComparisonSection struct {
	Items []CompletedDurationReportPredictionComparison `json:"items"`
}

// CompletedRunExportSections is the closed, ordered set of typed report
// sections.
type CompletedRunExportSections struct {
	Order                CompletedRunExportSectionOrder                `json:"order"`
	Summary              CompletedDurationReportSummary                `json:"summary"`
	Blocks               CompletedRunExportBlockSection                `json:"blocks"`
	Adjustments          CompletedRunExportAdjustmentSection           `json:"adjustments"`
	AnnotationsActions   CompletedRunExportAnnotationsActionsSection   `json:"annotationsActions"`
	PredictionComparison CompletedRunExportPredictionComparisonSection `json:"predictionComparison"`
	Provenance           CompletedDurationReportProvenance             `json:"provenance"`
}

// CompletedRunExportRedaction records the result of the canonical report's
// fail-closed annotation redaction.
type CompletedRunExportRedaction struct {
	Result                  string `json:"result"`
	RedactedAnnotationCount int    `json:"redactedAnnotationCount"`
}

// CompletedRunExportMetadata carries the generation and correction facts of
// the canonical measurement report.
type CompletedRunExportMetadata struct {
	ReportVersion string    `json:"reportVersion"`
	GeneratedAt   time.Time `json:"generatedAt"`
	Corrected     bool      `json:"corrected"`
}

// CompletedRunExportIntegrity binds the source report and the resulting typed
// envelope without storing an opaque serialized payload.
type CompletedRunExportIntegrity struct {
	SourceInputSHA256  string `json:"sourceInputSha256"`
	SourceReportSHA256 string `json:"sourceReportSha256"`
	ExportSHA256       string `json:"exportSha256"`
}

// CompletedRunExportRequest is the typed handoff from a canonical completed
// time-plan report to the tenant-scoped report-export worker.
type CompletedRunExportRequest struct {
	Source         CompletedRunExportSource `json:"source"`
	Formats        []ExportFormat           `json:"formats"`
	IdempotencyKey string                   `json:"idempotencyKey,omitempty"`
}

// CompletedRunExportCounter is one bounded lifecycle counter of an export job.
type CompletedRunExportCounter struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

// CompletedRunExportJob is the report-export job the completed-run export
// queues.
//
// It is declared next to the completed-run export instead of reusing
// ReportExportJob because the reporting export contract returns
// progressCounters as a typed list of counters, while ReportExportJob models
// them as a map.
type CompletedRunExportJob struct {
	ID                 string                      `json:"id"`
	PackKey            string                      `json:"packKey"`
	PackGeneration     int                         `json:"packGeneration"`
	DashboardKey       string                      `json:"dashboardKey"`
	Formats            []ExportFormat              `json:"formats"`
	ParametersDigest   string                      `json:"parametersDigest"`
	SnapshotSHA256     string                      `json:"snapshotSha256,omitempty"`
	State              string                      `json:"state"`
	ProgressStage      string                      `json:"progressStage"`
	ProgressCounters   []CompletedRunExportCounter `json:"progressCounters,omitempty"`
	QueuedAt           time.Time                   `json:"queuedAt"`
	StartedAt          *time.Time                  `json:"startedAt,omitempty"`
	FinishedAt         *time.Time                  `json:"finishedAt,omitempty"`
	ExpiresAt          *time.Time                  `json:"expiresAt,omitempty"`
	FailureCategory    string                      `json:"failureCategory,omitempty"`
	FailureMessage     string                      `json:"failureMessage,omitempty"`
	CancellationReason string                      `json:"cancellationReason,omitempty"`
	AttemptCount       int                         `json:"attemptCount"`
	NextAttemptAt      *time.Time                  `json:"nextAttemptAt,omitempty"`
	CleanupState       string                      `json:"cleanupState"`
	CleanupAttempts    int                         `json:"cleanupAttempts"`
	NextCleanupAt      *time.Time                  `json:"nextCleanupAt,omitempty"`
	Artifacts          []ReportExportArtifact      `json:"artifacts,omitempty"`
}
