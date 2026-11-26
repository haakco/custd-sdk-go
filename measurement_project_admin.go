package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// The project lifecycle, delivery estimates and decision models a tenant
// administrator manages on a measurement project. An empty companySlug uses the
// credential's own tenant.

// MeasurementProjectDelivery is what done means for a project and when it is due.
type MeasurementProjectDelivery struct {
	CompletedAt      *time.Time `json:"completedAt,omitempty"`
	DefinitionOfDone string     `json:"definitionOfDone"`
	TargetDate       *time.Time `json:"targetDate,omitempty"`
}

// MeasurementProjectRun links a project to the series of similar runs it belongs to.
type MeasurementProjectRun struct {
	ForecastMethod string `json:"forecastMethod,omitempty"`
	SeriesRootUUID string `json:"seriesRootUuid,omitempty"`
	Similar        bool   `json:"similar"`
}

type MeasurementProjectUpdate struct {
	// DefinitionOfDone and Description are left unchanged when nil and cleared by "".
	DefinitionOfDone   *string                   `json:"definitionOfDone,omitempty"`
	DeliveryTargetDate *time.Time                `json:"deliveryTargetDate,omitempty"`
	Description        *string                   `json:"description,omitempty"`
	ForecastMethod     MeasurementForecastMethod `json:"forecastMethod,omitempty"`
	Name               string                    `json:"name,omitempty"`
	RunSimilar         *bool                     `json:"runSimilar,omitempty"`
	Status             string                    `json:"status,omitempty"`
}

type MeasurementProjectComplete struct {
	CompletedAt time.Time `json:"completedAt"`
}

type MeasurementEstimate struct {
	EstimateUUID string    `json:"estimateUuid"`
	EstimatedAt  time.Time `json:"estimatedAt"`
	HighDate     time.Time `json:"highDate"`
	LowDate      time.Time `json:"lowDate"`
	Note         string    `json:"note"`
}

type MeasurementEstimateStatus struct {
	AgeDays float64 `json:"ageDays"`
	// ChanceByDeliveryTarget is nil when the project has no delivery target.
	ChanceByDeliveryTarget *float64            `json:"chanceByDeliveryTarget,omitempty"`
	Estimate               MeasurementEstimate `json:"estimate"`
	// MovementDays is nil for the first estimate, which has nothing to move from.
	MovementDays *float64 `json:"movementDays,omitempty"`
}

type MeasurementEstimateList struct {
	DeliveryTargetDate *time.Time                 `json:"deliveryTargetDate,omitempty"`
	Estimates          []MeasurementEstimate      `json:"estimates"`
	Latest             *MeasurementEstimateStatus `json:"latest,omitempty"`
}

type MeasurementEstimateCreate struct {
	EstimatedAt *time.Time `json:"estimatedAt,omitempty"`
	HighDate    time.Time  `json:"highDate"`
	LowDate     time.Time  `json:"lowDate"`
	Note        string     `json:"note,omitempty"`
}

type MeasurementRunCreate struct {
	// Defaults to the source project's name.
	Name        string `json:"name,omitempty"`
	ProjectCode string `json:"projectCode"`
	// Deadline for the new run's target.
	TargetDate *time.Time `json:"targetDate,omitempty"`
}

type MeasurementDecisionModelList struct {
	Decisions []MeasurementDecisionModel `json:"decisions"`
}

type MeasurementDecisionModelUpdateRequest struct {
	Description string `json:"description,omitempty"`
	Name        string `json:"name,omitempty"`
	Objective   string `json:"objective,omitempty"`
	Status      string `json:"status,omitempty"`
}

// Update updates a measurement project.
func (c *MeasurementProjectAdminClient) Update(ctx context.Context, companySlug string, projectUUID string, req MeasurementProjectUpdate) (*MeasurementProject, error) {
	var out MeasurementProject
	err := c.admin.request(ctx, http.MethodPatch, siteQueryPath("/measurement/projects/"+url.PathEscape(projectUUID), url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// Cancel cancels a measurement project.
func (c *MeasurementProjectAdminClient) Cancel(ctx context.Context, companySlug string, projectUUID string) (*MeasurementProject, error) {
	var out MeasurementProject
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/measurement/projects/"+url.PathEscape(projectUUID)+"/cancel", url.Values{companySlugQueryParameter: {companySlug}}), nil, &out)
	return &out, err
}

// Complete completes a measurement project with its actual date.
func (c *MeasurementProjectAdminClient) Complete(ctx context.Context, companySlug string, projectUUID string, req MeasurementProjectComplete) (*MeasurementProject, error) {
	var out MeasurementProject
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/measurement/projects/"+url.PathEscape(projectUUID)+"/complete", url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// ListEstimates lists a measurement project's delivery estimates.
func (c *MeasurementProjectAdminClient) ListEstimates(ctx context.Context, companySlug string, projectUUID string) (*MeasurementEstimateList, error) {
	var out MeasurementEstimateList
	err := c.admin.request(ctx, http.MethodGet, siteQueryPath("/measurement/projects/"+url.PathEscape(projectUUID)+"/estimates", url.Values{companySlugQueryParameter: {companySlug}}), nil, &out)
	return &out, err
}

// AddEstimate records a delivery estimate.
func (c *MeasurementProjectAdminClient) AddEstimate(ctx context.Context, companySlug string, projectUUID string, req MeasurementEstimateCreate) (*MeasurementEstimate, error) {
	var out MeasurementEstimate
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/measurement/projects/"+url.PathEscape(projectUUID)+"/estimates", url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// AddRun starts a new run of a measurement project.
func (c *MeasurementProjectAdminClient) AddRun(ctx context.Context, companySlug string, projectUUID string, req MeasurementRunCreate) (*MeasurementProject, error) {
	var out MeasurementProject
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/measurement/projects/"+url.PathEscape(projectUUID)+"/runs", url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// AddSeries creates measurement series.
func (c *MeasurementProjectAdminClient) AddSeries(ctx context.Context, companySlug string, projectUUID string, req MeasurementSeriesCreate) (*MeasurementSeries, error) {
	var out MeasurementSeries
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/measurement/projects/"+url.PathEscape(projectUUID)+"/series", url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// AddTarget creates a measurement target.
func (c *MeasurementProjectAdminClient) AddTarget(ctx context.Context, companySlug string, projectUUID string, req MeasurementTargetCreate) (*MeasurementTarget, error) {
	var out MeasurementTarget
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/measurement/projects/"+url.PathEscape(projectUUID)+"/targets", url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// ReopenTarget reopens a met measurement target.
func (c *MeasurementProjectAdminClient) ReopenTarget(ctx context.Context, companySlug string, projectUUID string, targetUUID string) (*MeasurementTarget, error) {
	var out MeasurementTarget
	err := c.admin.request(ctx, http.MethodPost, siteQueryPath("/measurement/projects/"+url.PathEscape(projectUUID)+"/targets/"+url.PathEscape(targetUUID)+"/reopen", url.Values{companySlugQueryParameter: {companySlug}}), nil, &out)
	return &out, err
}

// ListDecisions lists measurement decision models.
func (c *MeasurementProjectAdminClient) ListDecisions(ctx context.Context, companySlug string, projectUUID string) (*MeasurementDecisionModelList, error) {
	var out MeasurementDecisionModelList
	err := c.admin.request(ctx, http.MethodGet, siteQueryPath("/measurement/projects/"+url.PathEscape(projectUUID)+"/decisions", url.Values{companySlugQueryParameter: {companySlug}}), nil, &out)
	return &out, err
}

// GetDecision gets a measurement decision model.
func (c *MeasurementProjectAdminClient) GetDecision(ctx context.Context, companySlug string, projectUUID string, decisionUUID string) (*MeasurementDecisionModel, error) {
	var out MeasurementDecisionModel
	err := c.admin.request(ctx, http.MethodGet, siteQueryPath("/measurement/projects/"+url.PathEscape(projectUUID)+"/decisions/"+url.PathEscape(decisionUUID), url.Values{companySlugQueryParameter: {companySlug}}), nil, &out)
	return &out, err
}

// UpdateDecision updates a measurement decision model.
func (c *MeasurementProjectAdminClient) UpdateDecision(ctx context.Context, companySlug string, projectUUID string, decisionUUID string, req MeasurementDecisionModelUpdateRequest) (*MeasurementDecisionModel, error) {
	var out MeasurementDecisionModel
	err := c.admin.request(ctx, http.MethodPatch, siteQueryPath("/measurement/projects/"+url.PathEscape(projectUUID)+"/decisions/"+url.PathEscape(decisionUUID), url.Values{companySlugQueryParameter: {companySlug}}), req, &out)
	return &out, err
}

// DeleteDecision archives a measurement decision model.
func (c *MeasurementProjectAdminClient) DeleteDecision(ctx context.Context, companySlug string, projectUUID string, decisionUUID string) error {
	return c.admin.request(ctx, http.MethodDelete, siteQueryPath("/measurement/projects/"+url.PathEscape(projectUUID)+"/decisions/"+url.PathEscape(decisionUUID), url.Values{companySlugQueryParameter: {companySlug}}), nil, nil)
}
