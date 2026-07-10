package custd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const measurementEndpoint = "/api/v1/measurement"

type MeasurementAdminClient struct {
	Projects *MeasurementProjectAdminClient
}

type MeasurementProjectAdminClient struct {
	admin *AdminClient
}

type MeasurementClient struct {
	Projects *MeasurementProjectClient
	client   *CustdClient
}

type MeasurementProjectClient struct {
	measurement *MeasurementClient
}

type MeasurementProjectCreate struct {
	ProjectCode string                    `json:"projectCode"`
	Name        string                    `json:"name"`
	Kind        string                    `json:"kind"`
	Description string                    `json:"description,omitempty"`
	Series      []MeasurementSeriesCreate `json:"series"`
	Target      MeasurementTargetCreate   `json:"target"`
}

type MeasurementProject struct {
	ProjectUUID string `json:"projectUuid"`
	ProjectCode string `json:"projectCode"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
}

type MeasurementProjectList struct {
	Projects []MeasurementProject `json:"projects"`
}

type MeasurementSeriesCreate struct {
	SeriesCode          string `json:"seriesCode"`
	Name                string `json:"name"`
	UnitSlug            string `json:"unitSlug"`
	CompletionDirection string `json:"completionDirection"`
	Source              string `json:"source"`
}

type MeasurementTargetCreate struct {
	TargetCode  string  `json:"targetCode"`
	Name        string  `json:"name"`
	TargetValue float64 `json:"targetValue"`
	TargetDate  string  `json:"targetDate,omitempty"`
	State       string  `json:"state"`
}

type MeasurementObservationInput struct {
	SeriesUUID     string            `json:"seriesUuid"`
	ObservedAt     string            `json:"observedAt"`
	Value          float64           `json:"value"`
	IdempotencyKey string            `json:"idempotencyKey,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type MeasurementObservationBulkRequest struct {
	Rows []MeasurementObservationInput `json:"rows"`
}

type MeasurementObservationResult struct {
	RowIndex        int    `json:"rowIndex"`
	Success         bool   `json:"success"`
	Status          int    `json:"status,omitempty"`
	ObservationUUID string `json:"observationUuid,omitempty"`
	Type            string `json:"type,omitempty"`
	Title           string `json:"title,omitempty"`
	Detail          string `json:"detail,omitempty"`
}

type MeasurementObservationBulkResponse struct {
	ImportID string                         `json:"importId"`
	Accepted int                            `json:"accepted"`
	Rejected int                            `json:"rejected"`
	Results  []MeasurementObservationResult `json:"results"`
}

type MeasurementCSVImportResponse struct {
	ImportID string                         `json:"importId"`
	Accepted int                            `json:"accepted"`
	Rejected int                            `json:"rejected"`
	Results  []MeasurementObservationResult `json:"results"`
}

type MeasurementImportRun struct {
	ImportID    string                         `json:"importId"`
	ProjectUUID string                         `json:"projectUuid"`
	Source      string                         `json:"source"`
	RowCount    int                            `json:"rowCount"`
	Accepted    int                            `json:"accepted"`
	Rejected    int                            `json:"rejected"`
	CreatedAt   string                         `json:"createdAt"`
	CompletedAt string                         `json:"completedAt,omitempty"`
	Results     []MeasurementObservationResult `json:"results"`
}

type MeasurementCalibrationExerciseCreate struct {
	Question    string         `json:"question"`
	Confidence  float64        `json:"confidence"`
	LowerBound  float64        `json:"lowerBound"`
	UpperBound  float64        `json:"upperBound"`
	ActualValue *float64       `json:"actualValue,omitempty"`
	Contained   *bool          `json:"contained,omitempty"`
	ResolvedAt  string         `json:"resolvedAt,omitempty"`
	PriorUUID   string         `json:"priorUuid,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type MeasurementCalibrationExercise struct {
	ExerciseUUID string  `json:"exerciseUuid"`
	ProjectUUID  string  `json:"projectUuid"`
	Status       string  `json:"status"`
	Question     string  `json:"question"`
	Confidence   float64 `json:"confidence"`
	LowerBound   float64 `json:"lowerBound"`
	UpperBound   float64 `json:"upperBound"`
	ActualValue  float64 `json:"actualValue,omitempty"`
	Contained    *bool   `json:"contained,omitempty"`
	ResolvedAt   string  `json:"resolvedAt,omitempty"`
	CreatedAt    string  `json:"createdAt,omitempty"`
	UpdatedAt    string  `json:"updatedAt,omitempty"`
}

type MeasurementDecisionModelCreate struct {
	DecisionCode string                        `json:"decisionCode"`
	Name         string                        `json:"name"`
	Description  string                        `json:"description,omitempty"`
	Status       string                        `json:"status"`
	Objective    string                        `json:"objective,omitempty"`
	Choices      []MeasurementDecisionChoice   `json:"choices"`
	Variables    []MeasurementDecisionVariable `json:"variables"`
	Metadata     map[string]any                `json:"metadata,omitempty"`
}

type MeasurementDecisionChoice struct {
	ChoiceCode  string                         `json:"choiceCode"`
	Name        string                         `json:"name"`
	Description string                         `json:"description,omitempty"`
	Expression  *MeasurementDecisionExpression `json:"expression"`
}

type MeasurementDecisionVariable struct {
	VariableCode     string         `json:"variableCode"`
	Name             string         `json:"name"`
	UnitSlug         string         `json:"unit,omitempty"`
	DistributionSlug string         `json:"distributionSlug"`
	Parameters       map[string]any `json:"parameters,omitempty"`
	MeasurementCost  float64        `json:"measurementCost,omitempty"`
}

type MeasurementDecisionExpression struct {
	Kind         string                          `json:"kind"`
	Value        *float64                        `json:"value,omitempty"`
	VariableCode string                          `json:"variableCode,omitempty"`
	Children     []MeasurementDecisionExpression `json:"children,omitempty"`
}

type MeasurementDecisionModel struct {
	DecisionUUID string                                `json:"decisionUuid"`
	DecisionCode string                                `json:"decisionCode"`
	ProjectUUID  string                                `json:"projectUuid"`
	Status       string                                `json:"status"`
	Name         string                                `json:"name"`
	Description  string                                `json:"description,omitempty"`
	Objective    string                                `json:"objective,omitempty"`
	Choices      []MeasurementDecisionChoiceResponse   `json:"choices,omitempty"`
	Variables    []MeasurementDecisionVariableResponse `json:"variables,omitempty"`
	CreatedAt    string                                `json:"createdAt,omitempty"`
	UpdatedAt    string                                `json:"updatedAt,omitempty"`
}

type MeasurementDecisionChoiceResponse struct {
	ChoiceUUID  string                         `json:"choiceUuid,omitempty"`
	ChoiceCode  string                         `json:"choiceCode"`
	Name        string                         `json:"name"`
	Description string                         `json:"description,omitempty"`
	Expression  *MeasurementDecisionExpression `json:"expression,omitempty"`
}

type MeasurementDecisionVariableResponse struct {
	VariableUUID     string         `json:"variableUuid,omitempty"`
	VariableCode     string         `json:"variableCode"`
	Name             string         `json:"name"`
	UnitSlug         string         `json:"unit,omitempty"`
	DistributionSlug string         `json:"distributionSlug"`
	Parameters       map[string]any `json:"parameters,omitempty"`
	MeasurementCost  float64        `json:"measurementCost,omitempty"`
}

type MeasurementForecastRunRequest struct {
	Seed        int64 `json:"seed,omitempty"`
	SampleCount int   `json:"sampleCount,omitempty"`
}

type MeasurementForecastRun struct {
	ForecastRunUUID string                       `json:"forecastRunUuid"`
	ProjectUUID     string                       `json:"projectUuid"`
	Status          string                       `json:"status"`
	Method          string                       `json:"method"`
	GeneratedAt     string                       `json:"generatedAt"`
	InputCount      int                          `json:"inputCount"`
	SampleCount     int                          `json:"sampleCount"`
	Seed            int64                        `json:"seed"`
	Percentiles     map[string]string            `json:"percentiles"`
	Warnings        []MeasurementForecastWarning `json:"warnings,omitempty"`
	Provenance      map[string]string            `json:"provenance"`
}

type MeasurementForecastWarning struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type MeasurementSimulationRunRequest struct {
	Seed        int64 `json:"seed,omitempty"`
	SampleCount int   `json:"sampleCount,omitempty"`
}

type MeasurementSimulationRun struct {
	SimulationRunUUID string                           `json:"simulationRunUuid"`
	ProjectUUID       string                           `json:"projectUuid"`
	DecisionUUID      string                           `json:"decisionUuid"`
	Status            string                           `json:"status"`
	Method            string                           `json:"method"`
	GeneratedAt       string                           `json:"generatedAt"`
	CompletedAt       string                           `json:"completedAt"`
	SampleCount       int                              `json:"sampleCount"`
	Seed              int64                            `json:"seed"`
	Variables         []MeasurementVOIVariableResponse `json:"variables"`
	Warnings          []MeasurementSimulationWarning   `json:"warnings,omitempty"`
	Provenance        map[string]string                `json:"provenance"`
}

type MeasurementVOIVariableResponse struct {
	VariableCode    string  `json:"variableCode"`
	Name            string  `json:"name"`
	Rank            int     `json:"rank"`
	EVPI            float64 `json:"evpi"`
	EVPPI           float64 `json:"evppi,omitempty"`
	MeasurementCost float64 `json:"measurementCost"`
	WorthMeasuring  bool    `json:"worthMeasuring"`
}

type MeasurementSimulationWarning struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

func newMeasurementAdminClient(admin *AdminClient) *MeasurementAdminClient {
	return &MeasurementAdminClient{Projects: &MeasurementProjectAdminClient{admin: admin}}
}

func newMeasurementClient(client *CustdClient) *MeasurementClient {
	measurement := &MeasurementClient{client: client}
	measurement.Projects = &MeasurementProjectClient{measurement: measurement}
	return measurement
}

func (c *MeasurementProjectAdminClient) Create(
	ctx context.Context,
	req MeasurementProjectCreate,
) (*MeasurementProject, error) {
	var project MeasurementProject
	err := c.admin.request(ctx, http.MethodPost, "/measurement/projects", req, &project)
	return &project, err
}

func (c *MeasurementProjectAdminClient) List(ctx context.Context) (*MeasurementProjectList, error) {
	var projects MeasurementProjectList
	err := c.admin.request(ctx, http.MethodGet, "/measurement/projects", nil, &projects)
	return &projects, err
}

func (c *MeasurementProjectAdminClient) Get(ctx context.Context, projectUuid string) (*MeasurementProject, error) {
	return adminGetByID[MeasurementProject](ctx, c.admin, "/measurement/projects/", projectUuid)
}

func (c *MeasurementProjectAdminClient) SubmitObservation(
	ctx context.Context,
	projectUuid string,
	observation MeasurementObservationInput,
) (*MeasurementObservationBulkResponse, error) {
	return c.SubmitObservations(ctx, projectUuid, MeasurementObservationBulkRequest{
		Rows: []MeasurementObservationInput{observation},
	})
}

func (c *MeasurementProjectAdminClient) SubmitObservations(
	ctx context.Context,
	projectUuid string,
	req MeasurementObservationBulkRequest,
) (*MeasurementObservationBulkResponse, error) {
	var response MeasurementObservationBulkResponse
	path := "/measurement/projects/" + url.PathEscape(projectUuid) + "/observations:bulk"
	if err := c.admin.request(ctx, http.MethodPost, path, req, &response); err != nil {
		return nil, err
	}
	if err := validateMeasurementResults(response.Results, len(req.Rows)); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *MeasurementProjectAdminClient) ImportCSVString(
	ctx context.Context,
	projectUuid string,
	csv string,
	expectedRows int,
) (*MeasurementCSVImportResponse, error) {
	var response MeasurementCSVImportResponse
	path := "/measurement/projects/" + url.PathEscape(projectUuid) + "/observations:csv"
	if err := c.admin.request(ctx, http.MethodPost, path, map[string]string{"csv": csv}, &response); err != nil {
		return nil, err
	}
	if err := validateMeasurementResults(response.Results, expectedRows); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *MeasurementProjectAdminClient) GetImportRun(
	ctx context.Context,
	projectUuid string,
	importID string,
) (*MeasurementImportRun, error) {
	var response MeasurementImportRun
	path := "/measurement/projects/" + url.PathEscape(projectUuid) + "/imports/" + url.PathEscape(importID)
	err := c.admin.request(ctx, http.MethodGet, path, nil, &response)
	return &response, err
}

func (c *MeasurementProjectAdminClient) CreateCalibrationExercise(
	ctx context.Context,
	projectUuid string,
	req MeasurementCalibrationExerciseCreate,
) (*MeasurementCalibrationExercise, error) {
	var response MeasurementCalibrationExercise
	path := "/measurement/projects/" + url.PathEscape(projectUuid) + "/calibration-exercises"
	err := c.admin.request(ctx, http.MethodPost, path, req, &response)
	return &response, err
}

func (c *MeasurementProjectAdminClient) CreateDecisionModel(
	ctx context.Context,
	projectUuid string,
	req MeasurementDecisionModelCreate,
) (*MeasurementDecisionModel, error) {
	var response MeasurementDecisionModel
	path := "/measurement/projects/" + url.PathEscape(projectUuid) + "/decisions"
	err := c.admin.request(ctx, http.MethodPost, path, req, &response)
	return &response, err
}

func (c *MeasurementProjectClient) RunForecast(
	ctx context.Context,
	projectUuid string,
	req MeasurementForecastRunRequest,
) (*MeasurementForecastRun, error) {
	var response MeasurementForecastRun
	path := "/projects/" + url.PathEscape(projectUuid) + "/forecast-runs"
	err := c.measurement.request(ctx, http.MethodPost, path, req, &response)
	return &response, err
}

func (c *MeasurementProjectClient) GetForecastRun(
	ctx context.Context,
	projectUuid string,
	forecastRunUUID string,
) (*MeasurementForecastRun, error) {
	var response MeasurementForecastRun
	path := "/projects/" + url.PathEscape(projectUuid) + "/forecast-runs/" + url.PathEscape(forecastRunUUID)
	err := c.measurement.request(ctx, http.MethodGet, path, nil, &response)
	return &response, err
}

func (c *MeasurementProjectClient) RunDecisionSimulation(
	ctx context.Context,
	projectUuid string,
	decisionUuid string,
	req MeasurementSimulationRunRequest,
) (*MeasurementSimulationRun, error) {
	var response MeasurementSimulationRun
	path := measurementDecisionSimulationPath(projectUuid, decisionUuid)
	err := c.measurement.request(ctx, http.MethodPost, path, req, &response)
	return &response, err
}

func (c *MeasurementProjectClient) GetDecisionSimulationRun(
	ctx context.Context,
	projectUuid string,
	decisionUuid string,
	simulationRunUUID string,
) (*MeasurementSimulationRun, error) {
	var response MeasurementSimulationRun
	path := measurementDecisionSimulationPath(projectUuid, decisionUuid) + "/" + url.PathEscape(simulationRunUUID)
	err := c.measurement.request(ctx, http.MethodGet, path, nil, &response)
	return &response, err
}

func validateMeasurementResults(results []MeasurementObservationResult, submittedRows int) error {
	if len(results) != submittedRows {
		return fmt.Errorf(
			"custd: measurement result count %d does not match submitted row count %d",
			len(results), submittedRows,
		)
	}
	for i, result := range results {
		if result.Success && result.ObservationUUID == "" {
			return fmt.Errorf("custd: measurement result %d missing observationUuid", i)
		}
	}
	return nil
}

func measurementDecisionSimulationPath(projectUuid string, decisionUuid string) string {
	return "/projects/" + url.PathEscape(projectUuid) +
		"/decisions/" + url.PathEscape(decisionUuid) + "/simulation-runs"
}

func (c *MeasurementClient) request(ctx context.Context, method string, path string, payload any, out any) error {
	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("custd: marshal measurement request: %w", err)
		}
	}
	if c.client.config.HTTPClient != nil {
		return c.requestViaDoer(method, path, body, out)
	}
	return c.requestViaHTTP(ctx, method, path, body, out)
}

func (c *MeasurementClient) requestViaDoer(method string, path string, body []byte, out any) error {
	resp, err := c.client.config.HTTPClient.Do(&HTTPRequest{
		Method:  method,
		URL:     c.endpoint(path),
		Headers: c.client.headers(false),
		Body:    body,
	})
	if err != nil {
		return fmt.Errorf("custd: measurement request failed: %w", err)
	}
	if err := c.client.checkStatus(resp.StatusCode, resp.Body); err != nil {
		return err
	}
	return decodeAdminResponse(resp.Body, out)
}

func (c *MeasurementClient) requestViaHTTP(ctx context.Context, method string, path string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.endpoint(path), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("custd: create measurement request: %w", err)
	}
	for k, v := range c.client.headers(false) {
		req.Header.Set(k, v)
	}
	resp, err := c.client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("custd: measurement request failed: %w", err)
	}
	// nolint:errcheck // response body fully read below; a close error cannot affect the already-read measurement response
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	if err := c.client.checkStatus(resp.StatusCode, respBody); err != nil {
		return err
	}
	return decodeAdminResponse(respBody, out)
}

func (c *MeasurementClient) endpoint(path string) string {
	return strings.TrimRight(c.client.config.BaseURL, "/") + measurementEndpoint + path
}
