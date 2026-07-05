package custd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type MeasurementAdminClient struct {
	Projects *MeasurementProjectAdminClient
}

type MeasurementProjectAdminClient struct {
	admin *AdminClient
}

type MeasurementProjectCreate struct {
	ProjectSlug string                    `json:"projectSlug"`
	Name        string                    `json:"name"`
	Kind        string                    `json:"kind"`
	Description string                    `json:"description,omitempty"`
	Series      []MeasurementSeriesCreate `json:"series"`
	Target      MeasurementTargetCreate   `json:"target"`
}

type MeasurementProject struct {
	ProjectUUID string `json:"projectUuid"`
	ProjectSlug string `json:"projectSlug"`
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
}

type MeasurementProjectList struct {
	Projects []MeasurementProject `json:"projects"`
}

type MeasurementSeriesCreate struct {
	SeriesSlug          string `json:"seriesSlug"`
	Name                string `json:"name"`
	Unit                string `json:"unit"`
	CompletionDirection string `json:"completionDirection"`
	Source              string `json:"source"`
}

type MeasurementTargetCreate struct {
	TargetSlug  string  `json:"targetSlug"`
	Name        string  `json:"name"`
	TargetValue float64 `json:"targetValue"`
	TargetDate  string  `json:"targetDate,omitempty"`
	State       string  `json:"state"`
}

type MeasurementObservationInput struct {
	SeriesSlug     string            `json:"seriesSlug"`
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

func newMeasurementAdminClient(admin *AdminClient) *MeasurementAdminClient {
	return &MeasurementAdminClient{Projects: &MeasurementProjectAdminClient{admin: admin}}
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

func (c *MeasurementProjectAdminClient) Get(ctx context.Context, projectSlug string) (*MeasurementProject, error) {
	return adminGetByID[MeasurementProject](ctx, c.admin, "/measurement/projects/", projectSlug)
}

func (c *MeasurementProjectAdminClient) SubmitObservation(
	ctx context.Context,
	projectSlug string,
	observation MeasurementObservationInput,
) (*MeasurementObservationBulkResponse, error) {
	return c.SubmitObservations(ctx, projectSlug, MeasurementObservationBulkRequest{
		Rows: []MeasurementObservationInput{observation},
	})
}

func (c *MeasurementProjectAdminClient) SubmitObservations(
	ctx context.Context,
	projectSlug string,
	req MeasurementObservationBulkRequest,
) (*MeasurementObservationBulkResponse, error) {
	var response MeasurementObservationBulkResponse
	path := "/measurement/projects/" + url.PathEscape(projectSlug) + "/observations:bulk"
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
	projectSlug string,
	csv string,
	expectedRows int,
) (*MeasurementCSVImportResponse, error) {
	var response MeasurementCSVImportResponse
	path := "/measurement/projects/" + url.PathEscape(projectSlug) + "/observations:csv"
	if err := c.admin.request(ctx, http.MethodPost, path, map[string]string{"csv": csv}, &response); err != nil {
		return nil, err
	}
	if err := validateMeasurementResults(response.Results, expectedRows); err != nil {
		return nil, err
	}
	return &response, nil
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
