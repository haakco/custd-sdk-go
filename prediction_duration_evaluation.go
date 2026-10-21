package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// DurationEvaluationRequest selects one tenant-scoped duration history and an
// immutable prediction version for rolling-origin evaluation. The window and
// the limit narrow the history the evaluation reads; the version fixes what
// was compared.
type DurationEvaluationRequest struct {
	PlanUUID              string    `json:"planUuid"`
	SemanticKey           string    `json:"semanticKey"`
	PredictionVersionUUID string    `json:"predictionVersionUuid,omitempty"`
	WindowStart           time.Time `json:"windowStart,omitempty"`
	AsOf                  time.Time `json:"asOf,omitempty"`
	Limit                 int       `json:"limit,omitempty"`
}

// DurationEvaluationResponse contains the selected definition, chronological
// folds, outcome counts, and version lifecycle evidence.
type DurationEvaluationResponse struct {
	DefinitionUUID  string                             `json:"definitionUuid"`
	PlanUUID        string                             `json:"planUuid"`
	SemanticKey     string                             `json:"semanticKey"`
	Artifact        DurationForecastEvaluationArtifact `json:"artifact"`
	Outcomes        DurationEvaluationOutcomeCounts    `json:"outcomes"`
	VersionEvidence DurationEvaluationVersionEvidence  `json:"versionEvidence"`
}

// DurationEvaluationOutcomeCounts keeps eligible and excluded outcomes
// separate at the typed API boundary.
type DurationEvaluationOutcomeCounts struct {
	Eligible  int `json:"eligible"`
	Skipped   int `json:"skipped"`
	Cancelled int `json:"cancelled"`
	NotRun    int `json:"notRun"`
	Censored  int `json:"censored"`
}

// DurationEvaluationVersion identifies the selected immutable version.
type DurationEvaluationVersion struct {
	UUID          string    `json:"uuid"`
	VersionNumber int       `json:"versionNumber"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
}

// DurationEvaluationVersionHistoryEntry records one audited lifecycle change.
type DurationEvaluationVersionHistoryEntry struct {
	VersionUUID    string    `json:"versionUuid"`
	TransitionKind string    `json:"transitionKind"`
	FromStatus     string    `json:"fromStatus"`
	ToStatus       string    `json:"toStatus"`
	Actor          string    `json:"actor"`
	Reason         string    `json:"reason"`
	TransitionedAt time.Time `json:"transitionedAt"`
}

// DurationEvaluationVersionHistory is the named lifecycle history envelope.
type DurationEvaluationVersionHistory struct {
	Items []DurationEvaluationVersionHistoryEntry `json:"items"`
}

// DurationEvaluationVersionEvidence binds the result to the selected and
// active versions plus the forward-only audited history.
type DurationEvaluationVersionEvidence struct {
	Selected          DurationEvaluationVersion        `json:"selected"`
	ActiveVersionUUID string                           `json:"activeVersionUuid,omitempty"`
	History           DurationEvaluationVersionHistory `json:"history"`
}

// DurationForecastEvaluationArtifact is the digest-bound, typed output of a
// rolling-origin duration comparison. It has no persistence or transport side
// effects; a caller chooses where to store or publish it.
type DurationForecastEvaluationArtifact struct {
	SchemaVersion         string                     `json:"schema_version"`
	Source                string                     `json:"source"`
	PredictionVersionUUID string                     `json:"prediction_version_uuid"`
	Evaluation            DurationForecastEvaluation `json:"evaluation"`
	ContentSHA256         string                     `json:"content_sha256"`
}

// DurationForecastEvaluation is a deterministic rolling-origin result. A fold
// is emitted only when the preceding history reaches the published minimum, so
// fewer observations produce a valid empty evaluation.
type DurationForecastEvaluation struct {
	TenantID               string                            `json:"tenant_id"`
	ProjectID              string                            `json:"project_id"`
	SeriesID               string                            `json:"series_id"`
	Definition             DurationForecastDefinition        `json:"definition"`
	MinimumTrainingHistory int                               `json:"minimum_training_history"`
	FoldCount              int                               `json:"fold_count"`
	PrimaryMetric          string                            `json:"primary_metric"`
	Metrics                DurationForecastEvaluationMetrics `json:"metrics"`
	Folds                  []DurationForecastEvaluationFold  `json:"folds"`
	InputHash              string                            `json:"input_hash"`
	AsOf                   time.Time                         `json:"as_of"`
}

// DurationForecastDefinition is the immutable, typed configuration for one
// duration prediction version.
type DurationForecastDefinition struct {
	SchemaVersion  string `json:"schema_version"`
	Method         string `json:"method"`
	MethodVersion  string `json:"method_version"`
	PriorWeight    int64  `json:"prior_weight"`
	MinimumHistory int    `json:"minimum_history"`
}

// DurationForecastEvaluationMetrics keeps the nominal comparator beside the
// shrinkage candidate without exposing a method-keyed map at the boundary.
type DurationForecastEvaluationMetrics struct {
	Nominal   DurationForecastEvaluationMetric `json:"nominal"`
	Shrinkage DurationForecastEvaluationMetric `json:"shrinkage"`
}

// DurationForecastEvaluationMetric is one method's primary absolute-error
// result and its interval coverage summary.
type DurationForecastEvaluationMetric struct {
	Method              string                             `json:"method"`
	EvaluatedFoldCount  int                                `json:"evaluated_fold_count"`
	MeanAbsoluteErrorMS float64                            `json:"mean_absolute_error_ms"`
	Interval            DurationForecastIntervalEvaluation `json:"interval"`
}

// DurationForecastIntervalEvaluation contains coverage and width for the
// available empirical intervals. Unavailable sparse intervals are omitted.
type DurationForecastIntervalEvaluation struct {
	AvailableFoldCount int     `json:"available_fold_count"`
	CoveredFoldCount   int     `json:"covered_fold_count"`
	Coverage           float64 `json:"coverage"`
	MeanWidthMS        float64 `json:"mean_width_ms"`
}

// DurationForecastEvaluationFold records one chronological holdout and both
// predictions used for the baseline comparison.
type DurationForecastEvaluationFold struct {
	FoldNumber               int                      `json:"fold_number"`
	HeldOutRunID             string                   `json:"held_out_run_id"`
	ObservedAt               time.Time                `json:"observed_at"`
	TrainingSampleCount      int                      `json:"training_sample_count"`
	ActualMS                 int64                    `json:"actual_ms"`
	NominalPointMS           int64                    `json:"nominal_point_ms"`
	ShrinkagePointMS         int64                    `json:"shrinkage_point_ms"`
	NominalAbsoluteErrorMS   int64                    `json:"nominal_absolute_error_ms"`
	ShrinkageAbsoluteErrorMS int64                    `json:"shrinkage_absolute_error_ms"`
	ShrinkageInterval        DurationForecastInterval `json:"shrinkage_interval"`
}

// DurationForecastInterval is the empirical interval of one forecast. A bound
// the history could not establish is absent rather than zero.
type DurationForecastInterval struct {
	P10MS *int64 `json:"p10_ms,omitempty"`
	P90MS *int64 `json:"p90_ms,omitempty"`
}

// EvaluateDurationHistory posts one tenant-scoped rolling-origin duration
// evaluation request for the selected prediction definition.
//
// The answer carries the server's request identifier, because a proof that
// cannot quote the request it answered is a proof nobody can follow up.
func (c *PredictionAdminClient) EvaluateDurationHistory(
	ctx context.Context,
	companySlug string,
	definitionUUID string,
	req DurationEvaluationRequest,
) (*DurationEvaluationResponse, ResponseMetadata, error) {
	var out DurationEvaluationResponse
	path := predictionResourcePath(
		companySlug, "/definitions/"+url.PathEscape(definitionUUID)+"/duration-evaluations",
	)
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodPost, path, req, &out)
	if err != nil {
		return nil, metadata, err
	}
	return &out, metadata, nil
}
