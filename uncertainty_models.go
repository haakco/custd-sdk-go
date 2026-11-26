package custd

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Uncertainty models are a tenant's saved quantity graphs: the admin surface
// stores immutable versions, and the measurement surface runs them.

// UncertaintyModelAdminClient manages a tenant's uncertainty models and their
// versions. A system administrator names the tenant with companySlug; an empty
// companySlug uses the credential's own tenant.
type UncertaintyModelAdminClient struct {
	admin *AdminClient
}

// UncertaintyModelClient runs uncertainty models and reads their runs.
type UncertaintyModelClient struct {
	measurement *MeasurementClient
}

type UncertaintyModel struct {
	CreatedAt     time.Time `json:"createdAt"`
	Description   string    `json:"description"`
	LatestVersion int64     `json:"latestVersion"`
	ModelUUID     string    `json:"modelUuid"`
	Name          string    `json:"name"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// UncertaintyModelSource selects uncertainty models by where they were made; all lists both.
type UncertaintyModelSource string

const (
	UncertaintyModelSourceAll   UncertaintyModelSource = "all"
	UncertaintyModelSourceAdmin UncertaintyModelSource = "admin"
	UncertaintyModelSourceApp   UncertaintyModelSource = "app"
)

type UncertaintyModelList struct {
	Models []UncertaintyModel `json:"models"`
}

// UncertaintyInputSourceKind is the kind of source an uncertainty input reads.
type UncertaintyInputSourceKind string

const (
	UncertaintyInputSourceKindCounterForecast UncertaintyInputSourceKind = "counter_forecast"
)

// UncertaintyInputField is the forecast field an uncertainty input reads.
type UncertaintyInputField string

const (
	UncertaintyInputFieldCompletionDays      UncertaintyInputField = "completion_days"
	UncertaintyInputFieldDeadlineProbability UncertaintyInputField = "deadline_probability"
	UncertaintyInputFieldRatePerDay          UncertaintyInputField = "rate_per_day"
)

type UncertaintyInputSource struct {
	Field       UncertaintyInputField      `json:"field"`
	Kind        UncertaintyInputSourceKind `json:"kind"`
	ProjectUUID string                     `json:"projectUuid"`
}

// UncertaintyInputShape is how an uncertainty input is stated.
type UncertaintyInputShape string

const (
	UncertaintyInputShapePoint      UncertaintyInputShape = "point"
	UncertaintyInputShapeRange      UncertaintyInputShape = "range"
	UncertaintyInputShapeProportion UncertaintyInputShape = "proportion"
	UncertaintyInputShapeSamples    UncertaintyInputShape = "samples"
	UncertaintyInputShapeBound      UncertaintyInputShape = "bound"
)

// UncertaintyDistribution is the distribution an uncertainty input samples from.
type UncertaintyDistribution string

const (
	UncertaintyDistributionNormal    UncertaintyDistribution = "normal"
	UncertaintyDistributionLognormal UncertaintyDistribution = "lognormal"
	UncertaintyDistributionUniform   UncertaintyDistribution = "uniform"
)

type UncertaintyInput struct {
	Distribution UncertaintyDistribution `json:"distribution,omitempty"`
	High         *float64                `json:"high,omitempty"`
	Low          *float64                `json:"low,omitempty"`
	Samples      []float64               `json:"samples,omitempty"`
	Shape        UncertaintyInputShape   `json:"shape"`
	Source       *UncertaintyInputSource `json:"source,omitempty"`
	Successes    *int64                  `json:"successes,omitempty"`
	Trials       *int64                  `json:"trials,omitempty"`
	Value        *float64                `json:"value,omitempty"`
}

// UncertaintyQuantityKind is whether a quantity is an input or a formula.
type UncertaintyQuantityKind string

const (
	UncertaintyQuantityKindInput   UncertaintyQuantityKind = "input"
	UncertaintyQuantityKindFormula UncertaintyQuantityKind = "formula"
)

type UncertaintyQuantity struct {
	Description string                  `json:"description,omitempty"`
	Formula     string                  `json:"formula,omitempty"`
	Input       *UncertaintyInput       `json:"input,omitempty"`
	Kind        UncertaintyQuantityKind `json:"kind"`
	Name        string                  `json:"name"`
	Unit        string                  `json:"unit,omitempty"`
}

type UncertaintyCreateModelRequest struct {
	Description string                `json:"description,omitempty"`
	Name        string                `json:"name"`
	Note        string                `json:"note,omitempty"`
	Quantities  []UncertaintyQuantity `json:"quantities,omitempty"`
}

type UncertaintyQueuedRunFailure struct {
	Attempts int64     `json:"attempts"`
	Error    string    `json:"error"`
	FailedAt time.Time `json:"failedAt"`
}

type UncertaintyVersion struct {
	CreatedAt     time.Time             `json:"createdAt"`
	Note          string                `json:"note"`
	Quantities    []UncertaintyQuantity `json:"quantities"`
	VersionNumber int64                 `json:"versionNumber"`
	VersionUUID   string                `json:"versionUuid"`
}

type UncertaintyModelDetail struct {
	CreatedAt        time.Time                    `json:"createdAt"`
	Description      string                       `json:"description"`
	LatestVersion    int64                        `json:"latestVersion"`
	ModelUUID        string                       `json:"modelUuid"`
	Name             string                       `json:"name"`
	QueuedRunFailure *UncertaintyQueuedRunFailure `json:"queuedRunFailure,omitempty"`
	UpdatedAt        time.Time                    `json:"updatedAt"`
	Version          *UncertaintyVersion          `json:"version,omitempty"`
}

type UncertaintyCopyModelRequest struct {
	Name          string `json:"name,omitempty"`
	VersionNumber *int64 `json:"versionNumber,omitempty"`
}

type UncertaintyVersionSummary struct {
	CreatedAt     time.Time `json:"createdAt"`
	Note          string    `json:"note"`
	QuantityCount int64     `json:"quantityCount"`
	VersionNumber int64     `json:"versionNumber"`
	VersionUUID   string    `json:"versionUuid"`
}

type UncertaintyVersionList struct {
	Versions []UncertaintyVersionSummary `json:"versions"`
}

type UncertaintyCreateVersionRequest struct {
	Note       string                `json:"note,omitempty"`
	Quantities []UncertaintyQuantity `json:"quantities"`
}

// UncertaintyRunTrigger is what started an uncertainty model run.
type UncertaintyRunTrigger string

const (
	UncertaintyRunTriggerRequest       UncertaintyRunTrigger = "request"
	UncertaintyRunTriggerSourceChanged UncertaintyRunTrigger = "source_changed"
)

type UncertaintyRunSummary struct {
	CreatedAt             time.Time             `json:"createdAt"`
	ModelUUID             string                `json:"modelUuid"`
	RunUUID               string                `json:"runUuid"`
	SampleCount           int64                 `json:"sampleCount"`
	Seed                  int64                 `json:"seed"`
	SourceForecastRunUUID string                `json:"sourceForecastRunUuid,omitempty"`
	Trigger               UncertaintyRunTrigger `json:"trigger"`
	VersionNumber         int64                 `json:"versionNumber"`
}

type UncertaintyRunList struct {
	Runs []UncertaintyRunSummary `json:"runs"`
}

type UncertaintyDecision struct {
	MinimumWorthwhile *float64 `json:"minimumWorthwhile,omitempty"`
	Output            string   `json:"output"`
	Threshold         float64  `json:"threshold"`
	ValuePerUnit      float64  `json:"valuePerUnit"`
}

type UncertaintyRunRequest struct {
	Decision          *UncertaintyDecision `json:"decision,omitempty"`
	SampleCount       *int64               `json:"sampleCount,omitempty"`
	Seed              int64                `json:"seed"`
	SensitivityOutput string               `json:"sensitivityOutput,omitempty"`
	VersionNumber     *int64               `json:"versionNumber,omitempty"`
}

// UncertaintyBoundInputState is whether a bound input resolved to a value.
type UncertaintyBoundInputState string

const (
	UncertaintyBoundInputStateResolved      UncertaintyBoundInputState = "resolved"
	UncertaintyBoundInputStateSourceMissing UncertaintyBoundInputState = "source_missing"
	UncertaintyBoundInputStateValueMissing  UncertaintyBoundInputState = "value_missing"
)

// UncertaintyBoundInput is one bound input's resolution. Value is set only when State is resolved,
// so a resolved zero is distinguishable from a missing value.
type UncertaintyBoundInput struct {
	Detail              string                     `json:"detail,omitempty"`
	ForecastGeneratedAt *time.Time                 `json:"forecastGeneratedAt,omitempty"`
	ForecastRunUUID     string                     `json:"forecastRunUuid,omitempty"`
	Quantity            string                     `json:"quantity"`
	Source              UncertaintyInputSource     `json:"source"`
	State               UncertaintyBoundInputState `json:"state"`
	Value               *float64                   `json:"value,omitempty"`
}

type UncertaintyInputInformation struct {
	ExpectedValueOfPerfectInformation float64 `json:"expectedValueOfPerfectInformation"`
	Name                              string  `json:"name"`
}

// UncertaintyRecommendation is what the value of information recommends.
type UncertaintyRecommendation string

const (
	UncertaintyRecommendationAct  UncertaintyRecommendation = "act"
	UncertaintyRecommendationHold UncertaintyRecommendation = "hold"
)

type UncertaintyInformation struct {
	ActProbability                    float64                       `json:"actProbability"`
	ExpectedValueOfPerfectInformation float64                       `json:"expectedValueOfPerfectInformation"`
	Inputs                            []UncertaintyInputInformation `json:"inputs"`
	Recommended                       UncertaintyRecommendation     `json:"recommended"`
	SuggestedNextMeasurement          string                        `json:"suggestedNextMeasurement,omitempty"`
}

type UncertaintyQuantityError struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

type UncertaintyHistogram struct {
	Counts []int64 `json:"counts"`
	Max    float64 `json:"max"`
	Min    float64 `json:"min"`
}

type UncertaintyPercentiles struct {
	P25 float64 `json:"p25"`
	P5  float64 `json:"p5"`
	P50 float64 `json:"p50"`
	P75 float64 `json:"p75"`
	P95 float64 `json:"p95"`
}

type UncertaintyQuantitySummary struct {
	Histogram   UncertaintyHistogram   `json:"histogram"`
	Mean        float64                `json:"mean"`
	Percentiles UncertaintyPercentiles `json:"percentiles"`
	StdDev      float64                `json:"stdDev"`
}

type UncertaintyQuantityResult struct {
	Error   *UncertaintyQuantityError   `json:"error,omitempty"`
	Name    string                      `json:"name"`
	Summary *UncertaintyQuantitySummary `json:"summary,omitempty"`
}

type UncertaintyScatterPoint struct {
	Input  float64 `json:"input"`
	Output float64 `json:"output"`
}

type UncertaintyInputSensitivity struct {
	Name            string                    `json:"name"`
	RankCorrelation float64                   `json:"rankCorrelation"`
	Scatter         []UncertaintyScatterPoint `json:"scatter,omitempty"`
	Share           float64                   `json:"share"`
}

type UncertaintySensitivity struct {
	Inputs []UncertaintyInputSensitivity `json:"inputs"`
	Output string                        `json:"output"`
}

type UncertaintyRunDetail struct {
	BoundInputs           []UncertaintyBoundInput      `json:"boundInputs"`
	CreatedAt             time.Time                    `json:"createdAt"`
	Decision              *UncertaintyDecision         `json:"decision,omitempty"`
	Information           *UncertaintyInformation      `json:"information,omitempty"`
	ModelUUID             string                       `json:"modelUuid"`
	Quantities            []UncertaintyQuantityResult  `json:"quantities"`
	QueuedRunFailure      *UncertaintyQueuedRunFailure `json:"queuedRunFailure,omitempty"`
	RunUUID               string                       `json:"runUuid"`
	SampleCount           int64                        `json:"sampleCount"`
	Seed                  int64                        `json:"seed"`
	Sensitivity           *UncertaintySensitivity      `json:"sensitivity,omitempty"`
	SensitivityOutput     string                       `json:"sensitivityOutput,omitempty"`
	SourceForecastRunUUID string                       `json:"sourceForecastRunUuid,omitempty"`
	Trigger               UncertaintyRunTrigger        `json:"trigger"`
	VersionNumber         int64                        `json:"versionNumber"`
}

// List lists the tenant's uncertainty models.
func (c *UncertaintyModelAdminClient) List(ctx context.Context, companySlug string, source UncertaintyModelSource) (*UncertaintyModelList, error) {
	var out UncertaintyModelList
	err := c.admin.request(ctx, http.MethodGet, siteQueryPath("/uncertainty-models", url.Values{companySlugQueryParameter: {companySlug}, "source": {string(source)}}), nil, &out)
	return &out, err
}

// Create creates an uncertainty model, with version 1 when quantities are given.
func (c *UncertaintyModelAdminClient) Create(ctx context.Context, companySlug string, req UncertaintyCreateModelRequest) (*UncertaintyModelDetail, error) {
	var out UncertaintyModelDetail
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/uncertainty-models", url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// Get gets an uncertainty model with its latest version.
func (c *UncertaintyModelAdminClient) Get(ctx context.Context, companySlug string, modelUUID string) (*UncertaintyModelDetail, error) {
	var out UncertaintyModelDetail
	err := c.admin.request(ctx, http.MethodGet, siteQueryPath("/uncertainty-models/"+url.PathEscape(modelUUID), url.Values{companySlugQueryParameter: {companySlug}}), nil, &out)
	return &out, err
}

// Delete deletes an uncertainty model with its versions and runs; an app item's model is deleted with its item.
func (c *UncertaintyModelAdminClient) Delete(ctx context.Context, companySlug string, modelUUID string) error {
	return c.admin.request(ctx, http.MethodDelete, siteQueryPath("/uncertainty-models/"+url.PathEscape(modelUUID), url.Values{companySlugQueryParameter: {companySlug}}), nil, nil)
}

// Copy copies one version of an uncertainty model into a new model.
func (c *UncertaintyModelAdminClient) Copy(ctx context.Context, companySlug string, modelUUID string, req UncertaintyCopyModelRequest) (*UncertaintyModelDetail, error) {
	var out UncertaintyModelDetail
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/uncertainty-models/"+url.PathEscape(modelUUID)+"/copy", url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// ListVersions lists an uncertainty model's versions, newest first.
func (c *UncertaintyModelAdminClient) ListVersions(ctx context.Context, companySlug string, modelUUID string) (*UncertaintyVersionList, error) {
	var out UncertaintyVersionList
	err := c.admin.request(ctx, http.MethodGet, siteQueryPath("/uncertainty-models/"+url.PathEscape(modelUUID)+"/versions", url.Values{companySlugQueryParameter: {companySlug}}), nil, &out)
	return &out, err
}

// CreateVersion saves a new immutable version of an uncertainty model.
func (c *UncertaintyModelAdminClient) CreateVersion(ctx context.Context, companySlug string, modelUUID string, req UncertaintyCreateVersionRequest) (*UncertaintyVersion, error) {
	var out UncertaintyVersion
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/uncertainty-models/"+url.PathEscape(modelUUID)+"/versions", url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// GetVersion gets one version of an uncertainty model.
func (c *UncertaintyModelAdminClient) GetVersion(ctx context.Context, companySlug string, modelUUID string, versionNumber int64) (*UncertaintyVersion, error) {
	var out UncertaintyVersion
	err := c.admin.request(ctx, http.MethodGet, siteQueryPath("/uncertainty-models/"+url.PathEscape(modelUUID)+"/versions/"+strconv.FormatInt(versionNumber, 10), url.Values{companySlugQueryParameter: {companySlug}}), nil, &out)
	return &out, err
}

// ListRuns lists an uncertainty model's runs.
func (c *UncertaintyModelClient) ListRuns(ctx context.Context, modelUUID string) (*UncertaintyRunList, error) {
	var out UncertaintyRunList
	err := c.measurement.request(ctx, http.MethodGet, "/uncertainty-models/"+url.PathEscape(modelUUID)+"/runs", nil, &out)
	return &out, err
}

// CreateRun runs an uncertainty model.
func (c *UncertaintyModelClient) CreateRun(ctx context.Context, modelUUID string, req UncertaintyRunRequest) (*UncertaintyRunDetail, error) {
	var out UncertaintyRunDetail
	err := c.measurement.request(ctx, http.MethodPost, "/uncertainty-models/"+url.PathEscape(modelUUID)+"/runs", req, &out)
	return &out, err
}

// GetRun gets an uncertainty model run.
func (c *UncertaintyModelClient) GetRun(ctx context.Context, modelUUID string, runUUID string) (*UncertaintyRunDetail, error) {
	var out UncertaintyRunDetail
	err := c.measurement.request(ctx, http.MethodGet, "/uncertainty-models/"+url.PathEscape(modelUUID)+"/runs/"+url.PathEscape(runUUID), nil, &out)
	return &out, err
}
