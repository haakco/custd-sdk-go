package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// The templates of AdminSiteReport.Template and a saved report.
const (
	AdminSiteReportTemplateExecutive = "executive"
	AdminSiteReportTemplateFull      = "full"
)

// AdminSiteMaxSavedReports is how many saved reports a domain keeps.
const AdminSiteMaxSavedReports = 20

// SavedSiteReportNameTakenCode is the problem code a 409 carries when the domain already has a
// saved report with that name.
const SavedSiteReportNameTakenCode = "saved_site_report_name_taken"

// SavedSiteReportLimitCode is the problem code a 409 carries once the domain keeps
// AdminSiteMaxSavedReports saved reports.
const SavedSiteReportLimitCode = "saved_site_report_limit"

// AdminSiteReportOptions chooses a report's template (AdminSiteReportTemplate*) and days (1..90)
// of search numbers; zero values use the server's full template and 28 days.
type AdminSiteReportOptions struct {
	Template string
	Days     int
}

// AdminSiteSavedReportRequest saves, or replaces, a named report built now. Template and Days are
// omitted when unset, so the server uses full and 28.
type AdminSiteSavedReportRequest struct {
	Name     string `json:"name"`
	Template string `json:"template,omitempty"`
	Days     int    `json:"days,omitempty"`
}

// AdminSiteSavedReportSummary is one saved report without its content. GeneratedAt is when its
// content was built; LiveShares counts unexpired share links that open it.
type AdminSiteSavedReportSummary struct {
	ReportUUID  string    `json:"reportUuid"`
	Name        string    `json:"name"`
	Template    string    `json:"template"`
	Days        int       `json:"days"`
	GeneratedAt time.Time `json:"generatedAt"`
	CreatedAt   time.Time `json:"createdAt"`
	LiveShares  int       `json:"liveShares"`
}

// AdminSiteSavedReport is a saved report's summary and its stored content.
type AdminSiteSavedReport struct {
	AdminSiteSavedReportSummary
	Report AdminSiteReport `json:"report"`
}

// AdminSiteSavedReportList lists a domain's saved reports and how many it may keep.
type AdminSiteSavedReportList struct {
	Reports    []AdminSiteSavedReportSummary `json:"reports"`
	MaxReports int                           `json:"maxReports"`
}

// ListSavedReports returns the domain's saved reports without their content.
func (c *SiteDomainAdminClient) ListSavedReports(ctx context.Context, domainUUID string) (*AdminSiteSavedReportList, error) {
	var reports AdminSiteSavedReportList
	err := c.request(ctx, http.MethodGet, siteSavedReportsPath(domainUUID), nil, &reports)
	return &reports, err
}

// SaveReport builds a report now and keeps it under a name unique to the domain (201).
func (c *SiteDomainAdminClient) SaveReport(
	ctx context.Context, domainUUID string, req AdminSiteSavedReportRequest,
) (*AdminSiteSavedReport, error) {
	return c.savedReport(ctx, http.MethodPost, siteSavedReportsPath(domainUUID), req)
}

// GetSavedReport returns one saved report and its stored content.
func (c *SiteDomainAdminClient) GetSavedReport(
	ctx context.Context, domainUUID, reportUUID string,
) (*AdminSiteSavedReport, error) {
	return c.savedReport(ctx, http.MethodGet, siteSavedReportPath(domainUUID, reportUUID), nil)
}

// ReplaceSavedReport renames a saved report and builds its content again.
func (c *SiteDomainAdminClient) ReplaceSavedReport(
	ctx context.Context, domainUUID, reportUUID string, req AdminSiteSavedReportRequest,
) (*AdminSiteSavedReport, error) {
	return c.savedReport(ctx, http.MethodPut, siteSavedReportPath(domainUUID, reportUUID), req)
}

// DeleteSavedReport deletes a saved report.
func (c *SiteDomainAdminClient) DeleteSavedReport(ctx context.Context, domainUUID, reportUUID string) error {
	return c.request(ctx, http.MethodDelete, siteSavedReportPath(domainUUID, reportUUID), nil, nil)
}

func (c *SiteDomainAdminClient) savedReport(
	ctx context.Context, method, path string, payload any,
) (*AdminSiteSavedReport, error) {
	var report AdminSiteSavedReport
	err := c.request(ctx, method, path, payload, &report)
	return &report, err
}

func siteSavedReportsPath(domainUUID string) string {
	return siteDomainPath(domainUUID) + "/saved-reports"
}

func siteSavedReportPath(domainUUID, reportUUID string) string {
	return siteSavedReportsPath(domainUUID) + "/" + url.PathEscape(reportUUID)
}
