package custd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Tenant analytics reads beside the event query: the reporting dashboard
// catalogue, dashboard data generations and draft pack previews, the realtime
// active-visitor count, the dashboard summary and the per-day event count. The
// credential names the tenant.

// ReportingDraftPreviewRequest renders one query against a pack definition that
// has not been published. TenantSlug must be the credential's tenant unless the
// caller is a system administrator.
type ReportingDraftPreviewRequest struct {
	TenantSlug string                `json:"tenantSlug"`
	Definition Pack                  `json:"definition"`
	Query      ReportingQueryRequest `json:"query"`
}

// PreviewDraft renders a query against an unpublished pack definition.
func (r *ReportingClient) PreviewDraft(ctx context.Context, req ReportingDraftPreviewRequest) (*RenderedWidgetData, error) {
	var out RenderedWidgetData
	err := r.request(ctx, http.MethodPost, "/draft-preview", req, &out)
	return &out, err
}

// AnalyticsActiveNow is how many sessions were active on a site in the window.
type AnalyticsActiveNow struct {
	ActiveNow     int `json:"activeNow"`
	WindowSeconds int `json:"windowSeconds"`
}

// ActiveNow counts the sessions active on one site now. A 503 means realtime
// counts are unavailable and the caller should hide them.
func (a *AnalyticsEventClient) ActiveNow(ctx context.Context, siteUUID string) (*AnalyticsActiveNow, error) {
	if strings.TrimSpace(siteUUID) == "" {
		return nil, fmt.Errorf("custd: active-now requires a site")
	}
	var out AnalyticsActiveNow
	err := a.request(ctx, http.MethodGet, "/api/v1/analytics/active-now?"+url.Values{"site": {siteUUID}}.Encode(), nil, &out)
	return &out, err
}

// AnalyticsDashboardTenant names the tenant a summary describes.
type AnalyticsDashboardTenant struct {
	CompanySlug string `json:"companySlug"`
}

// AnalyticsDashboardEvents totals the tenant's stored events.
type AnalyticsDashboardEvents struct {
	Total            int64      `json:"total"`
	OldestEventAt    *time.Time `json:"oldestEventAt"`
	LatestEventAt    *time.Time `json:"latestEventAt"`
	LatestReceivedAt *time.Time `json:"latestReceivedAt"`
	LoadedDayCount   int        `json:"loadedDayCount"`
}

// AnalyticsDashboardCoverage is the span of days with stored events.
type AnalyticsDashboardCoverage struct {
	FirstDay       string `json:"firstDay"`
	LastDay        string `json:"lastDay"`
	LoadedDayCount int    `json:"loadedDayCount"`
}

// AnalyticsDashboardLoadedDay is one stored day and its archive state.
type AnalyticsDashboardLoadedDay struct {
	Date              string     `json:"date"`
	EventCount        int64      `json:"eventCount"`
	ArchiveStatus     string     `json:"archiveStatus"`
	ArchiveVerifiedAt *time.Time `json:"archiveVerifiedAt"`
	QueryHref         string     `json:"queryHref"`
}

// AnalyticsDashboardDogfoodSource is one internal source's delivery health.
type AnalyticsDashboardDogfoodSource struct {
	SourceSystem    string     `json:"sourceSystem"`
	SourceCompany   string     `json:"sourceCompany"`
	Environment     string     `json:"environment"`
	EventCount      int64      `json:"eventCount"`
	LatestEventAt   *time.Time `json:"latestEventAt"`
	LatestEventType string     `json:"latestEventType"`
	FailureCount    int64      `json:"failureCount"`
	RejectionCount  int64      `json:"rejectionCount"`
	Status          string     `json:"status"`
}

// AnalyticsDashboardDogfoodMetric is one internal operational metric.
type AnalyticsDashboardDogfoodMetric struct {
	SourceSystem string     `json:"sourceSystem"`
	Service      string     `json:"service"`
	MetricName   string     `json:"metricName"`
	MetricKind   string     `json:"metricKind"`
	Value        float64    `json:"value"`
	Unit         string     `json:"unit"`
	WindowStart  *time.Time `json:"windowStart"`
	WindowEnd    *time.Time `json:"windowEnd"`
	Route        string     `json:"route,omitempty"`
	Status       string     `json:"status,omitempty"`
}

// AnalyticsDashboardDogfood is the internal dogfood view; empty for client tenants.
type AnalyticsDashboardDogfood struct {
	Sources []AnalyticsDashboardDogfoodSource `json:"sources"`
	Metrics []AnalyticsDashboardDogfoodMetric `json:"metrics"`
}

// AnalyticsDashboardSummary is the tenant's stored-event overview.
type AnalyticsDashboardSummary struct {
	Tenant      AnalyticsDashboardTenant      `json:"tenant"`
	Events      AnalyticsDashboardEvents      `json:"events"`
	Coverage    AnalyticsDashboardCoverage    `json:"coverage"`
	LoadedDays  []AnalyticsDashboardLoadedDay `json:"loadedDays"`
	Dogfood     AnalyticsDashboardDogfood     `json:"dogfood"`
	GeneratedAt time.Time                     `json:"generatedAt"`
}

// DashboardSummary returns the tenant's stored-event overview.
func (a *AnalyticsEventClient) DashboardSummary(ctx context.Context) (*AnalyticsDashboardSummary, error) {
	var out AnalyticsDashboardSummary
	err := a.request(ctx, http.MethodGet, "/api/v1/analytics/dashboard-summary", nil, &out)
	return &out, err
}

// AnalyticsTelemetryEventCount is how many events the tenant stored on a day.
type AnalyticsTelemetryEventCount struct {
	Count int64 `json:"count"`
}

// TelemetryEventCount counts the tenant's events on one UTC date
// (YYYY-MM-DD), optionally of one event type.
func (a *AnalyticsEventClient) TelemetryEventCount(ctx context.Context, date, eventType string) (*AnalyticsTelemetryEventCount, error) {
	if strings.TrimSpace(date) == "" {
		return nil, fmt.Errorf("custd: event count requires a date")
	}
	var out AnalyticsTelemetryEventCount
	path := siteQueryPath("/api/v1/telemetry/event-count", url.Values{"date": {date}, "eventType": {eventType}})
	err := a.request(ctx, http.MethodGet, path, nil, &out)
	return &out, err
}

type ReportingDashboardCatalogueTrust struct {
	ServerOwned  bool `json:"serverOwned"`
	TenantScoped bool `json:"tenantScoped"`
}

type ReportingDashboardCatalogueEntry struct {
	DefaultRange   string                           `json:"defaultRange"`
	Description    string                           `json:"description"`
	ExportFormats  []string                         `json:"exportFormats"`
	Group          string                           `json:"group"`
	Interactions   []string                         `json:"interactions"`
	Key            string                           `json:"key"`
	RefreshSeconds int64                            `json:"refreshSeconds"`
	Title          string                           `json:"title"`
	Trust          ReportingDashboardCatalogueTrust `json:"trust"`
}

type ReportingDashboardCatalogueResponse struct {
	Dashboards []ReportingDashboardCatalogueEntry `json:"dashboards"`
}

type ReportingDashboardGenerationRequest struct {
	DashboardKey string                 `json:"dashboardKey"`
	Parameters   ReportExportParameters `json:"parameters"`
}

type ReportingDatasetRowCount struct {
	Count int64  `json:"count"`
	Key   string `json:"key"`
}

type ReportingSnapshotColumn struct {
	Description string `json:"description,omitempty"`
	Key         string `json:"key"`
	Label       string `json:"label"`
	Nullable    bool   `json:"nullable"`
	Type        string `json:"type"`
	Unit        string `json:"unit,omitempty"`
}

type ReportingSnapshotDatasetSchema struct {
	Columns []ReportingSnapshotColumn `json:"columns"`
	Key     string                    `json:"key"`
	Label   string                    `json:"label"`
}

type ReportingDashboardGenerationResponse struct {
	ByteLength       int64                            `json:"byteLength,omitempty"`
	Compression      string                           `json:"compression,omitempty"`
	ContractVersion  int64                            `json:"contractVersion,omitempty"`
	DatasetRowCounts []ReportingDatasetRowCount       `json:"datasetRowCounts,omitempty"`
	DatasetSchemas   []ReportingSnapshotDatasetSchema `json:"datasetSchemas,omitempty"`
	DownloadURL      string                           `json:"downloadUrl,omitempty"`
	GenerationID     string                           `json:"generationId"`
	MediaType        string                           `json:"mediaType,omitempty"`
	PollAfterSeconds int64                            `json:"pollAfterSeconds,omitempty"`
	SnapshotSha256   string                           `json:"snapshotSha256,omitempty"`
	State            string                           `json:"state"`
	URLExpiresAt     *time.Time                       `json:"urlExpiresAt,omitempty"`
	ValidUntil       *time.Time                       `json:"validUntil,omitempty"`
}

// REUSED existing:

// ListDashboards lists reporting dashboards.
func (r *ReportingClient) ListDashboards(ctx context.Context) (*ReportingDashboardCatalogueResponse, error) {
	var out ReportingDashboardCatalogueResponse
	err := r.request(ctx, http.MethodGet, "/dashboards", nil, &out)
	return &out, err
}

// GenerateDashboardData creates dashboard data generation.
func (r *ReportingClient) GenerateDashboardData(ctx context.Context, req ReportingDashboardGenerationRequest) (*ReportingDashboardGenerationResponse, error) {
	var out ReportingDashboardGenerationResponse
	err := r.request(ctx, http.MethodPost, "/dashboard-data-generations", req, &out)
	return &out, err
}
