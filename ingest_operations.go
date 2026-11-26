package custd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// IngestAdminClient calls the ingest service's tenant operations: replaying
// rejected events, managed audit redaction, monitoring, diagnostics, retention
// cleanup and the raw-landing integrity check. The credential names the tenant.
type IngestAdminClient struct {
	admin *AdminClient
}

// IngestDiagnosticsSection names one route of the ingest diagnostics snapshot.
// Every section answers the same whole snapshot.
type IngestDiagnosticsSection string

const (
	IngestDiagnosticsIngest    IngestDiagnosticsSection = "ingest"
	IngestDiagnosticsBuffers   IngestDiagnosticsSection = "buffers"
	IngestDiagnosticsDatabase  IngestDiagnosticsSection = "database"
	IngestDiagnosticsSegments  IngestDiagnosticsSection = "segments"
	IngestDiagnosticsReporting IngestDiagnosticsSection = "reporting"
)

var ingestDiagnosticsSections = map[IngestDiagnosticsSection]bool{
	IngestDiagnosticsIngest: true, IngestDiagnosticsBuffers: true, IngestDiagnosticsDatabase: true,
	IngestDiagnosticsSegments: true, IngestDiagnosticsReporting: true,
}

// IngestReprocessOutcome is the replay result of one rejected event.
type IngestReprocessOutcome struct {
	UUID    string `json:"uuid"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// IngestReprocessResult counts a replay of the tenant's rejected events.
type IngestReprocessResult struct {
	Processed int                      `json:"processed"`
	Succeeded int                      `json:"succeeded"`
	Failed    int                      `json:"failed"`
	Outcomes  []IngestReprocessOutcome `json:"outcomes"`
}

// ManagedAuditRedactionEvent selects one managed audit event and the fields to redact.
type ManagedAuditRedactionEvent struct {
	Selectors map[string]string `json:"selectors"`
	Fields    []string          `json:"fields"`
}

// ManagedAuditRedactionRequest redacts fields of managed audit events under a
// source system's redaction rule.
type ManagedAuditRedactionRequest struct {
	SourceSystem string                       `json:"sourceSystem"`
	RedactionID  string                       `json:"redactionId"`
	Reason       string                       `json:"reason"`
	Events       []ManagedAuditRedactionEvent `json:"events"`
}

// ManagedAuditRedactionResult is the outcome for one selected event.
type ManagedAuditRedactionResult struct {
	Selectors map[string]string `json:"selectors"`
	Status    string            `json:"status"`
}

// ManagedAuditRedactionResponse reports a redaction request's outcome.
type ManagedAuditRedactionResponse struct {
	Success     bool                          `json:"success"`
	RedactionID string                        `json:"redactionId"`
	ResultShape string                        `json:"resultShape"`
	Results     []ManagedAuditRedactionResult `json:"results"`
}

// IngestMonitoringSummary is the tenant's ingest and storage counters.
type IngestMonitoringSummary struct {
	IngestRequestsTotal             int64 `json:"ingest_requests_total"`
	IngestErrorsTotal               int64 `json:"ingest_errors_total"`
	EventSegmentsVerifiedRowsTotal  int64 `json:"event_segments_verified_rows_total"`
	RawEventsLandedTotal            int64 `json:"raw_events_landed_total"`
	RawStorageS3Total               int64 `json:"raw_storage_s3_total"`
	RawLandingS3UploadFailuresTotal int64 `json:"raw_landing_s3_upload_failures_total"`
}

// IngestMonitoringAlert is one active ingest alert for the tenant.
type IngestMonitoringAlert struct {
	ID       string `json:"id"`
	Severity string `json:"type"`
	Title    string `json:"title"`
	Message  string `json:"message"`
}

// IngestMonitoringAlertList lists the tenant's active ingest alerts.
type IngestMonitoringAlertList struct {
	Alerts []IngestMonitoringAlert `json:"alerts"`
}

// SystemMonitoringTotals is the platform-wide ingest and storage counters.
type SystemMonitoringTotals struct {
	IngestRequestsTotal                       int64 `json:"ingest_requests_total"`
	IngestErrorsTotal                         int64 `json:"ingest_errors_total"`
	RawEventsLandedTotal                      int64 `json:"raw_events_landed_total"`
	RawLandingS3UploadFailuresTotal           int64 `json:"raw_landing_s3_upload_failures_total"`
	DataSafetyInvariantViolationsTotal        int64 `json:"data_safety_invariant_violations_total"`
	StorageUsageCollectorLastSuccessTimestamp int64 `json:"storage_usage_collector_last_success_timestamp"`
	ClaimBacklogCount                         int64 `json:"claim_backlog_count"`
	ClaimBacklogOldestAgeSeconds              int64 `json:"claim_backlog_oldest_age_seconds"`
}

// SystemMonitoringTenant is one of the tenants ranked as needing attention.
type SystemMonitoringTenant struct {
	CompanySlug            string  `json:"companySlug"`
	CompanyName            string  `json:"companyName"`
	RankReason             string  `json:"rankReason"`
	AttributedStorageBytes int64   `json:"attributedStorageBytes"`
	EffectiveLimitBytes    int64   `json:"effectiveLimitBytes"`
	PercentUsed            float64 `json:"percentUsed"`
	RejectedEvents         int64   `json:"rejectedEvents,omitempty"`
}

// SystemMonitoringSummary is the platform-wide monitoring view.
type SystemMonitoringSummary struct {
	GeneratedAt time.Time                `json:"generatedAt"`
	Totals      SystemMonitoringTotals   `json:"totals"`
	TopTenants  []SystemMonitoringTenant `json:"topTenants"`
}

// IngestStatusCounts counts ingest answers by status class.
type IngestStatusCounts struct {
	TwoXX  int64 `json:"2xx"`
	FourXX int64 `json:"4xx"`
	FiveXX int64 `json:"5xx"`
}

// IngestDiagnosticsIngestState is the request side of the ingest service.
type IngestDiagnosticsIngestState struct {
	Available            bool               `json:"available"`
	RequestRatePerSecond float64            `json:"requestRatePerSecond"`
	StatusCounts         IngestStatusCounts `json:"statusCounts"`
	P95LatencyMs         float64            `json:"p95LatencyMs"`
	ActiveRequests       int64              `json:"activeRequests"`
}

// IngestJetStreamCluster is the event stream's replication state.
type IngestJetStreamCluster struct {
	ReplicaCount       int64  `json:"replicaCount"`
	Leader             string `json:"leader,omitempty"`
	Replicated         bool   `json:"replicated"`
	AllReplicasCurrent bool   `json:"allReplicasCurrent"`
}

// IngestJetStreamStreamState is the event stream's size.
type IngestJetStreamStreamState struct {
	Messages int64                  `json:"messages"`
	Bytes    int64                  `json:"bytes"`
	Subjects int64                  `json:"subjects"`
	Cluster  IngestJetStreamCluster `json:"cluster"`
}

// IngestJetStreamConsumerState is the segment writer consumer's progress.
type IngestJetStreamConsumerState struct {
	NumWaiting                int64   `json:"numWaiting"`
	MaxAckPending             int64   `json:"maxAckPending"`
	AckPendingUtilization     float64 `json:"ackPendingUtilization"`
	SlowConsumer              bool    `json:"slowConsumer"`
	DeliveredStreamSequence   int64   `json:"deliveredStreamSequence"`
	DeliveredConsumerSequence int64   `json:"deliveredConsumerSequence"`
	AckFloorStreamSequence    int64   `json:"ackFloorStreamSequence"`
	AckFloorConsumerSequence  int64   `json:"ackFloorConsumerSequence"`
	SequenceGap               int64   `json:"sequenceGap"`
}

// IngestJetStreamState is the event stream and its consumer.
type IngestJetStreamState struct {
	Available            bool                         `json:"available"`
	Stream               string                       `json:"stream,omitempty"`
	Consumer             string                       `json:"consumer,omitempty"`
	Pending              int64                        `json:"pending"`
	AckPending           int64                        `json:"ackPending"`
	Redelivered          int64                        `json:"redelivered"`
	ConsumerActive       bool                         `json:"consumerActive"`
	RetentionSecondsLeft int64                        `json:"retentionSecondsLeft"`
	StreamState          IngestJetStreamStreamState   `json:"streamState"`
	ConsumerState        IngestJetStreamConsumerState `json:"consumerState"`
}

// IngestSegmentState counts verified, pending and failed Parquet segments.
type IngestSegmentState struct {
	Available                          bool  `json:"available"`
	Pending                            int64 `json:"pending"`
	Verified                           int64 `json:"verified"`
	Failed                             int64 `json:"failed"`
	OldestPendingAgeSeconds            int64 `json:"oldestPendingAgeSeconds"`
	CompactedVerified                  int64 `json:"compactedVerified"`
	CompactedFailed                    int64 `json:"compactedFailed"`
	RollupVerified                     int64 `json:"rollupVerified"`
	RollupFailed                       int64 `json:"rollupFailed"`
	CompactionEligibleSources          int64 `json:"compactionEligibleSources"`
	CompactionEstimatedObjectReduction int64 `json:"compactionEstimatedObjectReduction"`
}

// IngestRollupSchedulerState is the rollup scheduler's queue.
type IngestRollupSchedulerState struct {
	IntervalSeconds int64     `json:"intervalSeconds"`
	NextRunAt       time.Time `json:"nextRunAt"`
	LeaderWorkerID  string    `json:"leaderWorkerId"`
	QueuedRuns      int64     `json:"queuedRuns"`
	RunningRuns     int64     `json:"runningRuns"`
	FailedRuns      int64     `json:"failedRuns"`
	QueuedJobs      int64     `json:"queuedJobs"`
	ClaimedJobs     int64     `json:"claimedJobs"`
	FailedJobs      int64     `json:"failedJobs"`
	LatestRunStatus string    `json:"latestRunStatus"`
}

// IngestWorkerState is one background worker's backlog and health.
type IngestWorkerState struct {
	Available         bool                        `json:"available"`
	Enabled           bool                        `json:"enabled"`
	Backlog           int64                       `json:"backlog"`
	LagSeconds        int64                       `json:"lagSeconds"`
	Failed            int64                       `json:"failed"`
	LastSuccessUnix   int64                       `json:"lastSuccessUnix"`
	LastFailureUnix   int64                       `json:"lastFailureUnix"`
	LastFailureReason string                      `json:"lastFailureReason"`
	Scheduler         *IngestRollupSchedulerState `json:"scheduler,omitempty"`
}

// IngestObjectStorageState is the accepted-event raw landing in object storage.
type IngestObjectStorageState struct {
	Available                                               bool    `json:"available"`
	AcceptedRawLandingPending                               int64   `json:"acceptedRawLandingPending"`
	AcceptedRawLandingOldestPendingAgeSeconds               int64   `json:"acceptedRawLandingOldestPendingAgeSeconds"`
	AcceptedRawLandingWritingLeases                         int64   `json:"acceptedRawLandingWritingLeases"`
	AcceptedRawLandingFailedObjects                         int64   `json:"acceptedRawLandingFailedObjects"`
	AcceptedRawLandingConsolidationEligibleWindows          int64   `json:"acceptedRawLandingConsolidationEligibleWindows"`
	AcceptedRawLandingConsolidationSourceObjects            int64   `json:"acceptedRawLandingConsolidationSourceObjects"`
	AcceptedRawLandingConsolidationEstimatedObjectReduction int64   `json:"acceptedRawLandingConsolidationEstimatedObjectReduction"`
	AcceptedRawLandingS3UploadFailuresTotal                 int64   `json:"acceptedRawLandingS3UploadFailuresTotal"`
	AcceptedRawLandingObjectGroupsPerSecond                 float64 `json:"acceptedRawLandingObjectGroupsPerSecond"`
	AcceptedRawLandingUploadDurationMs                      float64 `json:"acceptedRawLandingUploadDurationMs"`
	AcceptedRawLandingVerificationDurationMs                float64 `json:"acceptedRawLandingVerificationDurationMs"`
	AcceptedRawLandingDBMetadataDurationMs                  float64 `json:"acceptedRawLandingDBMetadataDurationMs"`
	AcceptedRawLandingObjectGroupSizeAverage                float64 `json:"acceptedRawLandingObjectGroupSizeAverage"`
	AcceptedRawLandingObjectGroupSizeMax                    float64 `json:"acceptedRawLandingObjectGroupSizeMax"`
	AcceptedRawLandingObjectLimiterSaturation               float64 `json:"acceptedRawLandingObjectLimiterSaturation"`
}

// IngestBufferState is every buffer between ingest and stored segments.
type IngestBufferState struct {
	JetStream  IngestJetStreamState     `json:"jetstream"`
	Segments   IngestSegmentState       `json:"segments"`
	Compaction IngestWorkerState        `json:"compaction"`
	Rollups    IngestWorkerState        `json:"rollups"`
	S3         IngestObjectStorageState `json:"s3"`
}

// IngestTopStatement is one of a database's most expensive statements.
type IngestTopStatement struct {
	Query        string  `json:"query"`
	Calls        int64   `json:"calls"`
	TotalExecMs  float64 `json:"totalExecMs"`
	MeanExecMs   float64 `json:"meanExecMs"`
	RowsReturned int64   `json:"rowsReturned"`
	WALBytes     int64   `json:"walBytes"`
}

// IngestWaitEvent counts sessions waiting on one event.
type IngestWaitEvent struct {
	Type  string `json:"type"`
	Event string `json:"event"`
	Count int64  `json:"count"`
}

// IngestTableSize is one table's stored size.
type IngestTableSize struct {
	TableName   string `json:"tableName"`
	RowEstimate int64  `json:"rowEstimate"`
	TotalBytes  int64  `json:"totalBytes"`
	TableBytes  int64  `json:"tableBytes"`
	IndexBytes  int64  `json:"indexBytes"`
}

// IngestDatabasePoolState is one database's connections and load.
type IngestDatabasePoolState struct {
	Available       bool                 `json:"available"`
	Active          int64                `json:"active"`
	Waiting         int64                `json:"waiting"`
	WALBytes        int64                `json:"walBytes"`
	TopStatements   []IngestTopStatement `json:"topStatements,omitempty"`
	WriteStatements []IngestTopStatement `json:"writeStatements"`
	WaitEvents      []IngestWaitEvent    `json:"waitEvents,omitempty"`
	TableSizes      []IngestTableSize    `json:"tableSizes,omitempty"`
}

// IngestDatabaseState is the core and raw databases.
type IngestDatabaseState struct {
	Core          IngestDatabasePoolState `json:"core"`
	Raw           IngestDatabasePoolState `json:"raw"`
	BytesPerEvent int64                   `json:"bytesPerEvent"`
}

// IngestReportingState says where reporting reads fresh and archived data.
type IngestReportingState struct {
	Available      bool   `json:"available"`
	FreshSource    string `json:"freshSource"`
	ArchiveSource  string `json:"archiveSource"`
	MissingWindows int64  `json:"missingWindows"`
}

// IngestSilentDropState is the rate events are lost without an answer.
type IngestSilentDropState struct {
	DropRatePerSecond float64 `json:"dropRatePerSecond"`
}

// IngestStageProgress is how a pipeline stage's backlog is trending.
type IngestStageProgress struct {
	Classification         string  `json:"classification"`
	Depth                  int64   `json:"depth"`
	DrainRatePerSecond     float64 `json:"drainRatePerSecond,omitempty"`
	EtaSeconds             int64   `json:"etaSeconds,omitempty"`
	RetentionExpirySeconds int64   `json:"retentionExpirySeconds,omitempty"`
	Note                   string  `json:"note,omitempty"`
}

// IngestPipelineStage is one node of the data path, in flow order.
type IngestPipelineStage struct {
	Name      string               `json:"name"`
	Status    string               `json:"status"`
	Headline  string               `json:"headline"`
	DetailKey string               `json:"detailKey"`
	Progress  *IngestStageProgress `json:"progress,omitempty"`
}

// IngestDiagnosticsSnapshot is the whole data-path health view: one status,
// the ordered pipeline, then each section's detail.
type IngestDiagnosticsSnapshot struct {
	Status      string                       `json:"status"`
	GeneratedAt time.Time                    `json:"generatedAt"`
	Ingest      IngestDiagnosticsIngestState `json:"ingest"`
	Buffers     IngestBufferState            `json:"buffers"`
	Database    IngestDatabaseState          `json:"database"`
	Segments    IngestSegmentState           `json:"segments"`
	Reporting   IngestReportingState         `json:"reporting"`
	SilentDrops IngestSilentDropState        `json:"silentDrops"`
	Pipeline    []IngestPipelineStage        `json:"pipeline"`
}

// IngestCleanupResult counts the rows a retention cleanup deleted.
type IngestCleanupResult struct {
	DeletedRaw      int64 `json:"deletedRaw"`
	DeletedEvents   int64 `json:"deletedEvents"`
	DeletedRejected int64 `json:"deletedRejected"`
}

// IngestIntegrityResult is a raw-landing integrity check: how many raw events
// were checked and which had no accepted landing.
type IngestIntegrityResult struct {
	CheckedCount  int      `json:"checkedCount"`
	OrphanedUUIDs []string `json:"orphanedUUIDs"`
}

// Reprocess replays the tenant's rejected events.
func (c *IngestAdminClient) Reprocess(ctx context.Context) (*IngestReprocessResult, error) {
	var out IngestReprocessResult
	err := c.admin.requestNonAdmin(ctx, http.MethodPost, "/reprocess", nil, &out)
	return &out, err
}

// RedactManagedAudit redacts fields of the tenant's managed audit events.
func (c *IngestAdminClient) RedactManagedAudit(
	ctx context.Context, req ManagedAuditRedactionRequest,
) (*ManagedAuditRedactionResponse, error) {
	var out ManagedAuditRedactionResponse
	err := c.admin.requestNonAdmin(ctx, http.MethodPost, "/managed-audit/redactions", req, &out)
	return &out, err
}

// MonitoringSummary returns the tenant's ingest and storage counters.
func (c *IngestAdminClient) MonitoringSummary(ctx context.Context) (*IngestMonitoringSummary, error) {
	var out IngestMonitoringSummary
	err := c.admin.requestNonAdmin(ctx, http.MethodGet, "/monitoring/summary", nil, &out)
	return &out, err
}

// MonitoringAlerts returns the tenant's active ingest alerts.
func (c *IngestAdminClient) MonitoringAlerts(ctx context.Context) (*IngestMonitoringAlertList, error) {
	var alerts []IngestMonitoringAlert
	if err := c.admin.requestNonAdmin(ctx, http.MethodGet, "/monitoring/alerts", nil, &alerts); err != nil {
		return nil, err
	}
	return &IngestMonitoringAlertList{Alerts: emptyIfNil(alerts)}, nil
}

// SystemMonitoringSummary returns the platform-wide monitoring view.
func (c *IngestAdminClient) SystemMonitoringSummary(ctx context.Context) (*SystemMonitoringSummary, error) {
	var out SystemMonitoringSummary
	err := c.admin.requestNonAdmin(ctx, http.MethodGet, "/system/monitoring/summary", nil, &out)
	return &out, err
}

// Diagnostics returns the data-path diagnostics snapshot from one section's route.
func (c *IngestAdminClient) Diagnostics(
	ctx context.Context, section IngestDiagnosticsSection,
) (*IngestDiagnosticsSnapshot, error) {
	if !ingestDiagnosticsSections[section] {
		return nil, fmt.Errorf("custd: unknown ingest diagnostics section %q", section)
	}
	var out IngestDiagnosticsSnapshot
	err := c.admin.requestNonAdmin(ctx, http.MethodGet, "/diagnostics/"+string(section), nil, &out)
	return &out, err
}

// Cleanup deletes the tenant's raw, event and rejected rows past their retention.
func (c *IngestAdminClient) Cleanup(ctx context.Context) (*IngestCleanupResult, error) {
	var out IngestCleanupResult
	err := c.admin.requestNonAdmin(ctx, http.MethodPost, "/admin/cleanup", nil, &out)
	return &out, err
}

// CheckIntegrity checks recent raw events against their accepted landings and
// removes raw events that have none.
func (c *IngestAdminClient) CheckIntegrity(ctx context.Context) (*IngestIntegrityResult, error) {
	var out IngestIntegrityResult
	err := c.admin.requestNonAdmin(ctx, http.MethodPost, "/admin/integrity", nil, &out)
	return &out, err
}

// EventIngestReceipt acknowledges one accepted event.
type EventIngestReceipt struct {
	Success   bool   `json:"success"`
	EventUUID string `json:"eventUuid"`
}

// SendEvent sends one event and waits for its acknowledgement, without the
// local queue or automatic retry. An answer that names another event is refused.
func (c *CustdClient) SendEvent(ctx context.Context, event EventEnvelope) (*EventIngestReceipt, error) {
	events, err := c.prepareConfirmedEvents([]EventEnvelope{event})
	if err != nil {
		return nil, err
	}
	var out EventIngestReceipt
	if err := c.Admin.requestNonAdmin(ctx, http.MethodPost, "/events", events[0], &out); err != nil {
		return nil, err
	}
	if !out.Success || out.EventUUID != events[0].EventUUID {
		return nil, validationErrorf("custd: event response does not acknowledge the sent event")
	}
	return &out, nil
}

// SiteCollectorConfig is a site's public browser write key and the allowed
// origin the collector checks it against.
type SiteCollectorConfig struct {
	WriteKey string
	Origin   string
}

// CollectEvent sends one browser event with a site's public write key. It
// carries none of the client's own credentials.
func (c *CustdClient) CollectEvent(
	ctx context.Context, site SiteCollectorConfig, event EventEnvelope,
) (*EventIngestReceipt, error) {
	events, err := c.prepareConfirmedEvents([]EventEnvelope{event})
	if err != nil {
		return nil, err
	}
	var out EventIngestReceipt
	if err := c.collect(ctx, site, "/api/v1/collect/events", events[0], &out); err != nil {
		return nil, err
	}
	if !out.Success || out.EventUUID != events[0].EventUUID {
		return nil, validationErrorf("custd: collector response does not acknowledge the sent event")
	}
	return &out, nil
}

// CollectEvents sends a batch of browser events with a site's public write key
// and returns each event's outcome. Partial rejection is a valid answer.
func (c *CustdClient) CollectEvents(
	ctx context.Context, site SiteCollectorConfig, request SendEventsRequest,
) (*SendEventsResponse, error) {
	events, err := c.prepareConfirmedEvents(request.Events)
	if err != nil {
		return nil, err
	}
	var out SendEventsResponse
	if err := c.collect(ctx, site, "/api/v1/collect/events/batch", eventBatchRequest{Events: events}, &out); err != nil {
		return nil, err
	}
	if err := validateConfirmedEventResults(events, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *CustdClient) collect(ctx context.Context, site SiteCollectorConfig, path string, payload, out any) error {
	if !strings.HasPrefix(site.WriteKey, "site_pk_") || strings.TrimSpace(site.Origin) == "" {
		return fmt.Errorf("custd: collector requires a site_pk_ write key and an allowed origin")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("custd: encode collector request: %w", err)
	}
	response, err := c.requestApplication(ctx, &HTTPRequest{
		Method: http.MethodPost, URL: strings.TrimRight(c.config.BaseURL, "/") + path, Body: body,
		Headers: map[string]string{
			"Authorization": "Bearer " + site.WriteKey, "Origin": site.Origin,
			"Content-Type": "application/json", "X-Custd-Sdk": sdkIdentity,
		},
	})
	if err != nil {
		return err
	}
	if err := c.checkStatus(response.StatusCode, response.Body); err != nil {
		return err
	}
	if err := unmarshalResponse(response.Body, out); err != nil {
		return validationWrap("custd: decode collector response", err)
	}
	return nil
}
