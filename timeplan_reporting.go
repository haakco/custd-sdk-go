package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// completedRunExportPath is the report-export route that queues a completed
// time-plan run. It lives outside the admin namespace because the reporting
// export worker owns it.
const completedRunExportPath = "/reporting/exports/completed-run"

// TimePlanPublishedVersionDetails is the immutable, server-computed view of a
// sealed definition version. The Admin API allocates from persisted version
// blocks, so this DTO is the proof client's source-owned definition boundary.
type TimePlanPublishedVersionDetails struct {
	UUID             string               `json:"uuid"`
	PlanUUID         string               `json:"planUuid"`
	VersionNumber    int                  `json:"versionNumber"`
	DefinitionHash   string               `json:"definitionHash"`
	SchemaVersion    string               `json:"schemaVersion"`
	AllocatorVersion string               `json:"allocatorVersion"`
	PublishedAt      time.Time            `json:"publishedAt"`
	HorizonMS        int64                `json:"horizonMs"`
	BlockCount       int                  `json:"blockCount"`
	Allocations      []TimePlanAllocation `json:"allocations"`
}

// GetPublishedTimePlanVersion reads one immutable, tenant-scoped definition
// version and its server-computed allocation. The version UUID is required so
// a proof cannot accidentally observe a mutable draft or a later revision.
func (c *TimePlanAdminClient) GetPublishedTimePlanVersion(
	ctx context.Context, companySlug, planUUID, versionUUID string,
) (*TimePlanPublishedVersionDetails, error) {
	var out TimePlanPublishedVersionDetails
	path := timePlanSDKPath(
		"/time-plans/plans/"+url.PathEscape(planUUID)+"/versions/"+url.PathEscape(versionUUID),
		companySlug, 0,
	)
	err := c.admin.request(ctx, http.MethodGet, path, nil, &out)
	return &out, err
}

// GetCompletedDurationReport reads the canonical terminal report of one
// tenant-scoped run. The measurement contract owns its validated DTO shape.
func (c *TimePlanAdminClient) GetCompletedDurationReport(
	ctx context.Context, companySlug, runUUID string,
) (*CompletedDurationReport, error) {
	var out CompletedDurationReport
	path := timePlanSDKPath("/time-plans/runs/"+url.PathEscape(runUUID)+"/report", companySlug, 0)
	err := c.admin.request(ctx, http.MethodGet, path, nil, &out)
	return &out, err
}

// CreateCompletedRunExport queues one validated completed-run source through
// the analytics report-export owner.
//
// The export route sits outside the admin namespace, so it is issued against
// the tenant-scoped /api/v1 surface. The answer is the queued export job.
func (c *TimePlanAdminClient) CreateCompletedRunExport(
	ctx context.Context, req CompletedRunExportRequest,
) (*CompletedRunExportJob, error) {
	var out CompletedRunExportJob
	err := c.admin.requestNonAdmin(ctx, http.MethodPost, completedRunExportPath, req, &out)
	return &out, err
}
