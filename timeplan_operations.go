package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// timePlanOperationsPath is the operator surface of the durable time-plan
// machinery: recovery health, the derived-event outbox, and the source-owned
// allocator proof. It is tenant-scoped like every other time-plan route.
const timePlanOperationsPath = "/time-plans/operations"

// TimePlanOperationalStatus is the redacted, tenant-scoped operator health
// contract. It carries counts and timestamps only, never run content.
type TimePlanOperationalStatus struct {
	PendingOutbox   int64      `json:"pendingOutbox"`
	FailedOutbox    int64      `json:"failedOutbox"`
	PublishedOutbox int64      `json:"publishedOutbox"`
	ActiveRuns      int64      `json:"activeRuns"`
	StaleRuns       int64      `json:"staleRuns"`
	OldestFailureAt *time.Time `json:"oldestFailureAt,omitempty"`
	CheckedAt       time.Time  `json:"checkedAt"`
}

// TimePlanDurationObservationIdentity identifies one persisted system duration
// observation and its idempotency key.
type TimePlanDurationObservationIdentity struct {
	ObservationUUID string `json:"observationUuid"`
	IdempotencyKey  string `json:"idempotencyKey"`
	RunUUID         string `json:"runUuid"`
	BlockUUID       string `json:"blockUuid"`
	PlanUUID        string `json:"planUuid"`
	SemanticKey     string `json:"semanticKey"`
}

// TimePlanDurationObservationReadResponse is the named envelope for one
// tenant-scoped run's non-deleted system duration observations.
type TimePlanDurationObservationReadResponse struct {
	RunUUID         string                                `json:"runUuid"`
	NonDeletedCount int64                                 `json:"nonDeletedCount"`
	Observations    []TimePlanDurationObservationIdentity `json:"observations"`
}

// TimePlanRunRepairResponse confirms the repaired run projection and the
// server time of the repair.
type TimePlanRunRepairResponse struct {
	Run        TimePlanRun `json:"run"`
	RepairedAt time.Time   `json:"repairedAt"`
}

// TimePlanProjectionLossResponse records one proof-scoped projection loss.
type TimePlanProjectionLossResponse struct {
	RunUUID     string    `json:"runUuid"`
	DeletedRows int64     `json:"deletedRows"`
	LostAt      time.Time `json:"lostAt"`
}

// TimePlanOutboxReplayRequest bounds one operator replay operation.
type TimePlanOutboxReplayRequest struct {
	Limit int `json:"limit,omitempty"`
}

// TimePlanOutboxReplayResponse records one bounded replay operation.
type TimePlanOutboxReplayResponse struct {
	Requeued   int64     `json:"requeued"`
	Requested  int       `json:"requested"`
	ReplayedAt time.Time `json:"replayedAt"`
}

// GetTimePlanOperationalStatus reads the redacted tenant-scoped recovery
// health of the time-plan machinery. The request identifier is returned
// because the operator command quotes it in its output.
func (c *TimePlanAdminClient) GetTimePlanOperationalStatus(
	ctx context.Context, companySlug string,
) (*TimePlanOperationalStatus, ResponseMetadata, error) {
	var out TimePlanOperationalStatus
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodGet, timePlanSDKPath(timePlanOperationsPath+"/status", companySlug, 0), nil, &out,
	)
	return &out, metadata, err
}

// ListTimePlanDurationObservations reads the persisted system duration
// observations of one tenant-scoped run.
func (c *TimePlanAdminClient) ListTimePlanDurationObservations(
	ctx context.Context, companySlug, runUUID string,
) (*TimePlanDurationObservationReadResponse, error) {
	var out TimePlanDurationObservationReadResponse
	err := c.admin.request(
		ctx, http.MethodGet,
		timePlanSDKPath("/time-plans/runs/"+url.PathEscape(runUUID)+"/duration-observations", companySlug, 0),
		nil, &out,
	)
	return &out, err
}

// RepairTimePlanRun rebuilds one tenant-scoped run projection from its
// immutable transition history. The request identifier is returned because
// the operator command quotes it in its output.
func (c *TimePlanAdminClient) RepairTimePlanRun(
	ctx context.Context, companySlug, runUUID string,
) (*TimePlanRunRepairResponse, ResponseMetadata, error) {
	var out TimePlanRunRepairResponse
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodPost,
		timePlanSDKPath("/time-plans/runs/"+url.PathEscape(runUUID)+"/repair", companySlug, 0),
		nil, &out,
	)
	return &out, metadata, err
}

// DeleteTimePlanRunProjection removes the rebuildable duration observations of
// one generated-proof run so the worker repair path can be observed.
func (c *TimePlanAdminClient) DeleteTimePlanRunProjection(
	ctx context.Context, companySlug, runUUID string,
) (*TimePlanProjectionLossResponse, error) {
	var out TimePlanProjectionLossResponse
	err := c.admin.request(
		ctx, http.MethodPost,
		timePlanSDKPath("/time-plans/runs/"+url.PathEscape(runUUID)+"/projection-loss", companySlug, 0),
		nil, &out,
	)
	return &out, err
}

// ReplayTimePlanOutbox requeues a bounded set of failed derived events. The
// request identifier is returned because the operator command quotes it in its
// output.
func (c *TimePlanAdminClient) ReplayTimePlanOutbox(
	ctx context.Context, companySlug string, req TimePlanOutboxReplayRequest,
) (*TimePlanOutboxReplayResponse, ResponseMetadata, error) {
	var out TimePlanOutboxReplayResponse
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodPost, timePlanSDKPath(timePlanOperationsPath+"/outbox/replay", companySlug, 0), req, &out,
	)
	return &out, metadata, err
}
