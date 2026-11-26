package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// PredictionClient reads the tenant's prediction definitions and their runs,
// outcomes and evaluations from the analytics API. The credential names the
// tenant; changes to a definition go through Admin.Predictions.
type PredictionClient struct {
	measurement *MeasurementClient
}

// PredictionDefinitionRead is one active prediction definition.
type PredictionDefinitionRead struct {
	UUID          string    `json:"uuid"`
	DefinitionKey string    `json:"definitionKey"`
	DisplayName   string    `json:"displayName"`
	Status        string    `json:"status"`
	IsPaused      bool      `json:"isPaused"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// PredictionDefinitionReadList lists the tenant's active prediction definitions.
type PredictionDefinitionReadList struct {
	Items []PredictionDefinitionRead `json:"items"`
}

// PredictionRunRead is one run of a prediction definition.
type PredictionRunRead struct {
	RunUUID           string    `json:"runUuid"`
	AsOf              time.Time `json:"asOf"`
	HorizonEnd        time.Time `json:"horizonEnd"`
	Output            float64   `json:"output"`
	Baseline          float64   `json:"baseline"`
	OverrideApplied   bool      `json:"overrideApplied"`
	InputHash         string    `json:"inputHash"`
	ContributionCount int       `json:"contributionCount"`
	GeneratedAt       time.Time `json:"generatedAt"`
}

// PredictionRunReadList lists a definition's runs, newest first.
type PredictionRunReadList struct {
	Items []PredictionRunRead `json:"items"`
}

// PredictionOutcomeRead is how one run resolved.
type PredictionOutcomeRead struct {
	RunUUID        string     `json:"runUuid"`
	Resolution     string     `json:"resolution"`
	OutcomeKind    string     `json:"outcomeKind"`
	ResolvedAt     *time.Time `json:"resolvedAt,omitempty"`
	ObservedValue  int        `json:"observedValue"`
	IsLateEvidence bool       `json:"isLateEvidence"`
}

// PredictionOutcomeReadList lists a definition's outcomes.
type PredictionOutcomeReadList struct {
	Items []PredictionOutcomeRead `json:"items"`
}

// PredictionEvaluationRead is the latest scored window of a definition's
// resolved runs.
type PredictionEvaluationRead struct {
	WindowStartAt     time.Time `json:"windowStartAt"`
	WindowEndAt       time.Time `json:"windowEndAt"`
	ResolvedRunCount  int       `json:"resolvedRunCount"`
	PositiveCount     int       `json:"positiveCount"`
	NegativeCount     int       `json:"negativeCount"`
	BrierScore        *float64  `json:"brierScore,omitempty"`
	LogLoss           *float64  `json:"logLoss,omitempty"`
	SparseBucketCount int       `json:"sparseBucketCount"`
}

const predictionReadDefinitionsPath = "/predictions/definitions"

// ListDefinitions returns the tenant's active prediction definitions.
func (c *PredictionClient) ListDefinitions(ctx context.Context) (*PredictionDefinitionReadList, error) {
	var out PredictionDefinitionReadList
	err := c.measurement.request(ctx, http.MethodGet, predictionReadDefinitionsPath, nil, &out)
	return &out, err
}

// GetDefinition returns one prediction definition.
func (c *PredictionClient) GetDefinition(ctx context.Context, definitionUUID string) (*PredictionDefinitionRead, error) {
	var out PredictionDefinitionRead
	err := c.measurement.request(ctx, http.MethodGet, predictionReadDefinitionPath(definitionUUID, ""), nil, &out)
	return &out, err
}

// ListRuns returns a definition's runs; a non-zero since keeps only runs as of
// that time or later.
func (c *PredictionClient) ListRuns(ctx context.Context, definitionUUID string, since time.Time) (*PredictionRunReadList, error) {
	values := url.Values{}
	if !since.IsZero() {
		values.Set("since", since.UTC().Format(time.RFC3339))
	}
	var out PredictionRunReadList
	path := siteQueryPath(predictionReadDefinitionPath(definitionUUID, "/runs"), values)
	err := c.measurement.request(ctx, http.MethodGet, path, nil, &out)
	return &out, err
}

// ListOutcomes returns how a definition's runs resolved.
func (c *PredictionClient) ListOutcomes(ctx context.Context, definitionUUID string) (*PredictionOutcomeReadList, error) {
	var out PredictionOutcomeReadList
	err := c.measurement.request(ctx, http.MethodGet, predictionReadDefinitionPath(definitionUUID, "/outcomes"), nil, &out)
	return &out, err
}

// GetLatestEvaluation returns the latest evaluation of a definition's resolved runs.
func (c *PredictionClient) GetLatestEvaluation(ctx context.Context, definitionUUID string) (*PredictionEvaluationRead, error) {
	var out PredictionEvaluationRead
	err := c.measurement.request(ctx, http.MethodGet, predictionReadDefinitionPath(definitionUUID, "/evaluations"), nil, &out)
	return &out, err
}

func predictionReadDefinitionPath(definitionUUID, resource string) string {
	return predictionReadDefinitionsPath + "/" + url.PathEscape(definitionUUID) + resource
}
