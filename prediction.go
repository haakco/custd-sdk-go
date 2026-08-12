package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// PredictionAdminClient owns the tenant-scoped configurable signal
// prediction lifecycle. Every method requires the company slug explicitly so
// callers cannot accidentally rely on an ambient tenant when configuring a
// definition or source.
type PredictionAdminClient struct {
	admin *AdminClient
}

type PredictionDefinition struct {
	UUID                  string    `json:"uuid"`
	DefinitionKey         string    `json:"definition_key"`
	DisplayName           string    `json:"display_name"`
	Description           string    `json:"description,omitempty"`
	Status                string    `json:"status"`
	ScheduleKind          string    `json:"schedule_kind"`
	DefaultHorizonSeconds int       `json:"default_horizon_seconds,omitempty"`
	IsPaused              bool      `json:"is_paused"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type PredictionDefinitionCreateRequest struct {
	DefinitionKey string `json:"definition_key"`
	DisplayName   string `json:"display_name"`
	Description   string `json:"description,omitempty"`
}

type PredictionDefinitionUpdateRequest struct {
	DefinitionKey string `json:"definition_key"`
	DisplayName   string `json:"display_name,omitempty"`
	Description   string `json:"description,omitempty"`
	Revision      int64  `json:"revision"`
}

type PredictionDefinitionListResponse struct {
	Items         []PredictionDefinition `json:"items"`
	NextPageToken string                 `json:"next_page_token,omitempty"`
}

type PredictionVersion struct {
	UUID           string          `json:"uuid"`
	VersionUUID    string          `json:"version_uuid"`
	VersionNumber  int             `json:"version_number"`
	VersionStatus  string          `json:"version_status"`
	DefinitionHash string          `json:"definition_hash"`
	Definition     json.RawMessage `json:"definition"`
	FeatureCount   int             `json:"feature_count"`
	SourceCount    int             `json:"source_count"`
	CreatedBy      string          `json:"created_by"`
	CreatedAt      time.Time       `json:"created_at"`
}

type PredictionVersionPublishRequest struct {
	Definition json.RawMessage `json:"definition"`
	CreatedBy  string          `json:"created_by"`
}

type PredictionActivateRequest struct {
	VersionUUID string `json:"version_uuid"`
}

type PredictionRollbackRequest struct {
	VersionUUID string `json:"version_uuid"`
	Reason      string `json:"reason,omitempty"`
}

type PredictionPauseRequest struct {
	Reason string `json:"reason,omitempty"`
}

type PredictionRunNowRequest struct {
	WorkerID string `json:"worker_id,omitempty"`
}

type PredictionSignalSource struct {
	UUID                   string     `json:"uuid"`
	SourceKey              string     `json:"source_key"`
	SourceMode             string     `json:"source_mode"`
	DisplayName            string     `json:"display_name"`
	Description            string     `json:"description,omitempty"`
	PollIntervalSeconds    int        `json:"poll_interval_seconds,omitempty"`
	SourceStatus           string     `json:"source_status"`
	IsPaused               bool       `json:"is_paused"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
	LastSucceededAt        *time.Time `json:"last_succeeded_at,omitempty"`
	ConsecutiveFailedCount int        `json:"consecutive_failed_count"`
}

type PredictionSignalSourceCreateRequest struct {
	SourceKey           string          `json:"source_key"`
	SourceMode          string          `json:"source_mode"`
	DisplayName         string          `json:"display_name"`
	Description         string          `json:"description,omitempty"`
	Configuration       json.RawMessage `json:"configuration"`
	PollIntervalSeconds int             `json:"poll_interval_seconds,omitempty"`
}

type PredictionRunSummary struct {
	RunUUID         string    `json:"run_uuid"`
	AsOf            time.Time `json:"as_of_at"`
	HorizonEnd      time.Time `json:"horizon_end_at"`
	Output          float64   `json:"output"`
	Baseline        float64   `json:"baseline"`
	OverrideApplied bool      `json:"override_applied"`
	InputHash       string    `json:"input_hash"`
	EngineVersion   string    `json:"engine_version"`
	WarningCount    int       `json:"warning_count"`
	GeneratedAt     time.Time `json:"generated_at"`
	DurationMillis  int       `json:"duration_milliseconds,omitempty"`
}

type PredictionOutcomeSummary struct {
	RunUUID        string     `json:"run_uuid"`
	Resolution     string     `json:"resolution"`
	OutcomeKind    string     `json:"outcome_kind"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	ObservedValue  int        `json:"observed_value"`
	IsLateEvidence bool       `json:"is_late_evidence"`
}

type PredictionEvaluationSummary struct {
	WindowStartAt     time.Time `json:"window_start_at"`
	WindowEndAt       time.Time `json:"window_end_at"`
	ResolvedRunCount  int       `json:"resolved_run_count"`
	PositiveCount     int       `json:"positive_count"`
	NegativeCount     int       `json:"negative_count"`
	BrierScore        *float64  `json:"brier_score,omitempty"`
	LogLoss           *float64  `json:"log_loss,omitempty"`
	SparseBucketCount int       `json:"sparse_bucket_count"`
}

type PredictionThresholdEvent struct {
	VersionUUID     string    `json:"version_uuid"`
	HysteresisState string    `json:"hysteresis_state"`
	Direction       string    `json:"direction"`
	ObservedValue   float64   `json:"observed_value"`
	Threshold       float64   `json:"threshold"`
	TriggerRunUUID  string    `json:"trigger_run_uuid"`
	DedupKey        string    `json:"dedup_key"`
	EmittedAt       time.Time `json:"emitted_at"`
}

func (c *PredictionAdminClient) ListDefinitions(
	ctx context.Context, companySlug string, pageSize int, pageToken string,
) (*PredictionDefinitionListResponse, error) {
	var out PredictionDefinitionListResponse
	err := c.admin.request(ctx, http.MethodGet, predictionPath(companySlug, "/definitions", pageSize, pageToken), nil, &out)
	return &out, err
}

func (c *PredictionAdminClient) GetDefinition(
	ctx context.Context, companySlug, definitionUUID string,
) (*PredictionDefinition, error) {
	var out PredictionDefinition
	err := c.admin.request(ctx, http.MethodGet, predictionResourcePath(companySlug, "/definitions/"+url.PathEscape(definitionUUID)), nil, &out)
	return &out, err
}

func (c *PredictionAdminClient) CreateDefinition(
	ctx context.Context, companySlug string, req PredictionDefinitionCreateRequest,
) (*PredictionDefinition, error) {
	var out PredictionDefinition
	err := c.admin.request(ctx, http.MethodPost, predictionPath(companySlug, "/definitions", 0, ""), req, &out)
	return &out, err
}

func (c *PredictionAdminClient) UpdateDefinition(
	ctx context.Context, companySlug, definitionUUID string, req PredictionDefinitionUpdateRequest,
) (*PredictionDefinition, error) {
	var out PredictionDefinition
	err := c.admin.request(ctx, http.MethodPatch, predictionResourcePath(companySlug, "/definitions/"+url.PathEscape(definitionUUID)), req, &out)
	return &out, err
}

func (c *PredictionAdminClient) GetVersion(
	ctx context.Context, companySlug, definitionUUID, versionUUID string,
) (*PredictionVersion, error) {
	var out PredictionVersion
	path := "/definitions/" + url.PathEscape(definitionUUID) + "/versions/" + url.PathEscape(versionUUID)
	err := c.admin.request(ctx, http.MethodGet, predictionResourcePath(companySlug, path), nil, &out)
	return &out, err
}

func (c *PredictionAdminClient) PublishVersion(
	ctx context.Context, companySlug, definitionUUID string, req PredictionVersionPublishRequest,
) (*PredictionVersion, error) {
	var out PredictionVersion
	err := c.admin.request(ctx, http.MethodPost, predictionResourcePath(companySlug, "/definitions/"+url.PathEscape(definitionUUID)+"/publish"), req, &out)
	return &out, err
}

func (c *PredictionAdminClient) ActivateVersion(
	ctx context.Context, companySlug, definitionUUID string, req PredictionActivateRequest,
) (*PredictionVersion, error) {
	var out PredictionVersion
	err := c.admin.request(ctx, http.MethodPost, predictionResourcePath(companySlug, "/definitions/"+url.PathEscape(definitionUUID)+"/activate"), req, &out)
	return &out, err
}

func (c *PredictionAdminClient) RollbackVersion(
	ctx context.Context, companySlug, definitionUUID string, req PredictionRollbackRequest,
) (*PredictionVersion, error) {
	var out PredictionVersion
	err := c.admin.request(ctx, http.MethodPost, predictionResourcePath(companySlug, "/definitions/"+url.PathEscape(definitionUUID)+"/rollback"), req, &out)
	return &out, err
}

func (c *PredictionAdminClient) PauseDefinition(
	ctx context.Context, companySlug, definitionUUID string, req PredictionPauseRequest,
) error {
	return c.admin.request(ctx, http.MethodPost, predictionResourcePath(companySlug, "/definitions/"+url.PathEscape(definitionUUID)+"/pause"), req, nil)
}

func (c *PredictionAdminClient) ResumeDefinition(ctx context.Context, companySlug, definitionUUID string) error {
	return c.admin.request(ctx, http.MethodPost, predictionResourcePath(companySlug, "/definitions/"+url.PathEscape(definitionUUID)+"/resume"), nil, nil)
}

func (c *PredictionAdminClient) ArchiveDefinition(ctx context.Context, companySlug, definitionUUID string) error {
	return c.admin.request(ctx, http.MethodPost, predictionResourcePath(companySlug, "/definitions/"+url.PathEscape(definitionUUID)+"/archive"), nil, nil)
}

func (c *PredictionAdminClient) RunNow(
	ctx context.Context, companySlug, definitionUUID string, req PredictionRunNowRequest,
) error {
	return c.admin.request(ctx, http.MethodPost, predictionResourcePath(companySlug, "/definitions/"+url.PathEscape(definitionUUID)+"/run-now"), req, nil)
}

func (c *PredictionAdminClient) ListRuns(
	ctx context.Context, companySlug, definitionUUID string, pageSize int,
) ([]PredictionRunSummary, error) {
	var out []PredictionRunSummary
	err := c.admin.request(ctx, http.MethodGet, predictionPath(companySlug, "/definitions/"+url.PathEscape(definitionUUID)+"/runs", pageSize, ""), nil, &out)
	return out, err
}

func (c *PredictionAdminClient) ListOutcomes(
	ctx context.Context, companySlug, definitionUUID string, pageSize int,
) ([]PredictionOutcomeSummary, error) {
	var out []PredictionOutcomeSummary
	err := c.admin.request(ctx, http.MethodGet, predictionPath(companySlug, "/definitions/"+url.PathEscape(definitionUUID)+"/outcomes", pageSize, ""), nil, &out)
	return out, err
}

func (c *PredictionAdminClient) GetEvaluation(
	ctx context.Context, companySlug, definitionUUID string,
) (*PredictionEvaluationSummary, error) {
	var out PredictionEvaluationSummary
	err := c.admin.request(ctx, http.MethodGet, predictionResourcePath(companySlug, "/definitions/"+url.PathEscape(definitionUUID)+"/evaluations"), nil, &out)
	return &out, err
}

func (c *PredictionAdminClient) ListThresholdEvents(
	ctx context.Context, companySlug, definitionUUID string, pageSize int,
) ([]PredictionThresholdEvent, error) {
	var out []PredictionThresholdEvent
	err := c.admin.request(ctx, http.MethodGet, predictionPath(companySlug, "/definitions/"+url.PathEscape(definitionUUID)+"/threshold-events", pageSize, ""), nil, &out)
	return out, err
}

func (c *PredictionAdminClient) ListSignalSources(
	ctx context.Context, companySlug string, pageSize int, pageToken string,
) ([]PredictionSignalSource, error) {
	var out []PredictionSignalSource
	err := c.admin.request(ctx, http.MethodGet, predictionPath(companySlug, "/sources", pageSize, pageToken), nil, &out)
	return out, err
}

func (c *PredictionAdminClient) GetSignalSource(
	ctx context.Context, companySlug, sourceUUID string,
) (*PredictionSignalSource, error) {
	var out PredictionSignalSource
	err := c.admin.request(ctx, http.MethodGet, predictionResourcePath(companySlug, "/sources/"+url.PathEscape(sourceUUID)), nil, &out)
	return &out, err
}

func (c *PredictionAdminClient) CreateSignalSource(
	ctx context.Context, companySlug string, req PredictionSignalSourceCreateRequest,
) (*PredictionSignalSource, error) {
	var out PredictionSignalSource
	err := c.admin.request(ctx, http.MethodPost, predictionPath(companySlug, "/sources", 0, ""), req, &out)
	return &out, err
}

func (c *PredictionAdminClient) ActivateSignalSource(ctx context.Context, companySlug, sourceUUID string) error {
	return c.admin.request(ctx, http.MethodPost, predictionResourcePath(companySlug, "/sources/"+url.PathEscape(sourceUUID)+"/activate"), nil, nil)
}

func (c *PredictionAdminClient) ArchiveSignalSource(ctx context.Context, companySlug, sourceUUID string) error {
	return c.admin.request(ctx, http.MethodPost, predictionResourcePath(companySlug, "/sources/"+url.PathEscape(sourceUUID)+"/archive"), nil, nil)
}

func predictionResourcePath(companySlug, path string) string {
	return predictionPath(companySlug, path, 0, "")
}

func predictionPath(companySlug, path string, pageSize int, pageToken string) string {
	values := url.Values{}
	values.Set("companySlug", companySlug)
	if pageSize > 0 {
		values.Set("pageSize", strconv.Itoa(pageSize))
	}
	if pageToken != "" {
		values.Set("pageToken", pageToken)
	}
	return "/measurement/predictions" + path + "?" + values.Encode()
}
