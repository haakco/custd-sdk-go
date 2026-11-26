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

type UncertaintyModelList struct {
	Models []UncertaintyModel `json:"models"`
}

type UncertaintyInputSource struct {
	Field       string `json:"field"`
	Kind        string `json:"kind"`
	ProjectUUID string `json:"projectUuid"`
}

type UncertaintyInput struct {
	Distribution string                  `json:"distribution,omitempty"`
	High         *float64                `json:"high,omitempty"`
	Low          *float64                `json:"low,omitempty"`
	Samples      []float64               `json:"samples,omitempty"`
	Shape        string                  `json:"shape"`
	Source       *UncertaintyInputSource `json:"source,omitempty"`
	Successes    *int64                  `json:"successes,omitempty"`
	Trials       *int64                  `json:"trials,omitempty"`
	Value        *float64                `json:"value,omitempty"`
}

type UncertaintyQuantity struct {
	Description string            `json:"description,omitempty"`
	Formula     string            `json:"formula,omitempty"`
	Input       *UncertaintyInput `json:"input,omitempty"`
	Kind        string            `json:"kind"`
	Name        string            `json:"name"`
	Unit        string            `json:"unit,omitempty"`
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

type UncertaintyRunSummary struct {
	CreatedAt             time.Time `json:"createdAt"`
	ModelUUID             string    `json:"modelUuid"`
	RunUUID               string    `json:"runUuid"`
	SampleCount           int64     `json:"sampleCount"`
	Seed                  int64     `json:"seed"`
	SourceForecastRunUUID string    `json:"sourceForecastRunUuid,omitempty"`
	Trigger               string    `json:"trigger"`
	VersionNumber         int64     `json:"versionNumber"`
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

type UncertaintyBoundInput struct {
	Detail              string                 `json:"detail,omitempty"`
	ForecastGeneratedAt *time.Time             `json:"forecastGeneratedAt,omitempty"`
	ForecastRunUUID     string                 `json:"forecastRunUuid,omitempty"`
	Quantity            string                 `json:"quantity"`
	Source              UncertaintyInputSource `json:"source"`
	State               string                 `json:"state"`
	Value               float64                `json:"value,omitempty"`
}

type UncertaintyInputInformation struct {
	ExpectedValueOfPerfectInformation float64 `json:"expectedValueOfPerfectInformation"`
	Name                              string  `json:"name"`
}

type UncertaintyInformation struct {
	ActProbability                    float64                       `json:"actProbability"`
	ExpectedValueOfPerfectInformation float64                       `json:"expectedValueOfPerfectInformation"`
	Inputs                            []UncertaintyInputInformation `json:"inputs"`
	Recommended                       string                        `json:"recommended"`
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
	Trigger               string                       `json:"trigger"`
	VersionNumber         int64                        `json:"versionNumber"`
}

// List lists the tenant's uncertainty models.
func (c *UncertaintyModelAdminClient) List(ctx context.Context, companySlug string, source string) (*UncertaintyModelList, error) {
	var out UncertaintyModelList
	err := c.admin.request(ctx, http.MethodGet, siteQueryPath("/uncertainty-models", url.Values{companySlugQueryParameter: {companySlug}, "source": {source}}), nil, &out)
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
