package custd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const maxReportExportDownloadBytes = 64 << 20

type ReportExportCreateRequest struct {
	DashboardKey   string                 `json:"dashboardKey"`
	Formats        []string               `json:"formats"`
	Parameters     map[string]interface{} `json:"parameters"`
	IdempotencyKey string                 `json:"idempotencyKey,omitempty"`
}

type ReportExportArtifact struct {
	ID                string   `json:"id"`
	ArtifactKey       string   `json:"artifactKey"`
	Format            string   `json:"format"`
	Filename          string   `json:"filename"`
	MediaType         string   `json:"mediaType"`
	ObjectBytes       int64    `json:"objectBytes"`
	ObjectSHA256      string   `json:"objectSha256"`
	Warnings          []string `json:"warnings,omitempty"`
	FallbackForFormat string   `json:"fallbackForFormat,omitempty"`
	RendererVersion   string   `json:"rendererVersion,omitempty"`
}

type ReportExportJob struct {
	ID                 string                 `json:"id"`
	PackKey            string                 `json:"packKey"`
	PackGeneration     int                    `json:"packGeneration"`
	DashboardKey       string                 `json:"dashboardKey"`
	Formats            []string               `json:"formats"`
	ParametersDigest   string                 `json:"parametersDigest"`
	SnapshotSHA256     string                 `json:"snapshotSha256,omitempty"`
	State              string                 `json:"state"`
	ProgressStage      string                 `json:"progressStage"`
	ProgressCounters   map[string]int64       `json:"progressCounters,omitempty"`
	QueuedAt           string                 `json:"queuedAt"`
	StartedAt          *string                `json:"startedAt,omitempty"`
	FinishedAt         *string                `json:"finishedAt,omitempty"`
	ExpiresAt          *string                `json:"expiresAt,omitempty"`
	FailureCategory    string                 `json:"failureCategory,omitempty"`
	FailureMessage     string                 `json:"failureMessage,omitempty"`
	CancellationReason string                 `json:"cancellationReason,omitempty"`
	AttemptCount       int                    `json:"attemptCount"`
	NextAttemptAt      *string                `json:"nextAttemptAt,omitempty"`
	CleanupState       string                 `json:"cleanupState"`
	CleanupAttempts    int                    `json:"cleanupAttempts"`
	NextCleanupAt      *string                `json:"nextCleanupAt,omitempty"`
	Artifacts          []ReportExportArtifact `json:"artifacts,omitempty"`
}

func (r *ReportingClient) CreateExport(ctx context.Context, input ReportExportCreateRequest) (*ReportExportJob, error) {
	if input.DashboardKey == "" || len(input.Formats) == 0 || len(input.Formats) > 8 {
		return nil, fmt.Errorf("custd: report export requires a dashboard key and 1 to 8 formats")
	}
	var out ReportExportJob
	err := r.request(ctx, http.MethodPost, "/exports", input, &out)
	return &out, err
}

func (r *ReportingClient) ListExports(ctx context.Context, limit int) ([]ReportExportJob, error) {
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("custd: report export limit must be between 1 and 100")
	}
	var out []ReportExportJob
	err := r.request(ctx, http.MethodGet, "/exports?limit="+url.QueryEscape(fmt.Sprint(limit)), nil, &out)
	return out, err
}

func (r *ReportingClient) GetExport(ctx context.Context, exportID string) (*ReportExportJob, error) {
	if !preparedDataUUIDPattern.MatchString(exportID) {
		return nil, fmt.Errorf("custd: exportId must be a UUID")
	}
	var out ReportExportJob
	err := r.request(ctx, http.MethodGet, "/exports/"+url.PathEscape(exportID), nil, &out)
	return &out, err
}

func (r *ReportingClient) CancelExport(ctx context.Context, exportID, reason string) (*ReportExportJob, error) {
	if !preparedDataUUIDPattern.MatchString(exportID) {
		return nil, fmt.Errorf("custd: exportId must be a UUID")
	}
	var out ReportExportJob
	err := r.request(ctx, http.MethodPost, "/exports/"+url.PathEscape(exportID)+"/cancel", map[string]string{"reason": reason}, &out)
	return &out, err
}

func (r *ReportingClient) DownloadExport(ctx context.Context, exportID, artifactID string) ([]byte, error) {
	if !preparedDataUUIDPattern.MatchString(exportID) || !preparedDataUUIDPattern.MatchString(artifactID) {
		return nil, fmt.Errorf("custd: exportId and artifactId must be UUIDs")
	}
	path := "/exports/" + url.PathEscape(exportID) + "/artifacts/" + url.PathEscape(artifactID)
	if r.client.config.HTTPClient != nil {
		resp, err := r.client.config.HTTPClient.Do(&HTTPRequest{Method: http.MethodGet, URL: r.endpoint(path), Headers: r.client.headers(false)})
		if err != nil {
			return nil, fmt.Errorf("custd: report export download failed: %w", err)
		}
		if err := r.client.checkStatus(resp.StatusCode, resp.Body); err != nil {
			return nil, err
		}
		return boundedReportExportBytes(resp.Body)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.endpoint(path), nil)
	if err != nil {
		return nil, fmt.Errorf("custd: create report export download: %w", err)
	}
	for key, value := range r.client.headers(false) {
		req.Header.Set(key, value)
	}
	resp, err := r.client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("custd: report export download failed: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxReportExportDownloadBytes+1))
	if err != nil {
		return nil, fmt.Errorf("custd: read report export download: %w", err)
	}
	if err := r.client.checkStatus(resp.StatusCode, body); err != nil {
		return nil, err
	}
	return boundedReportExportBytes(body)
}

func boundedReportExportBytes(body []byte) ([]byte, error) {
	if len(body) > maxReportExportDownloadBytes {
		return nil, fmt.Errorf("custd: report export download exceeds 64 MiB")
	}
	return body, nil
}
