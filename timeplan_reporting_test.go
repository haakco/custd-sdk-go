package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestTimePlanVersionAndReportReadsUseTenantScopedRoutes(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	tests := []timePlanRequestCase{
		{
			name: "published version",
			call: func() error {
				_, err := client.Admin.TimePlans.GetPublishedTimePlanVersion(context.Background(), "acme", "plan-1", "version-1")
				return err
			},
			wantMethod:    http.MethodGet,
			wantURL:       "http://localhost:8080/api/v1/admin/time-plans/plans/plan-1/versions/version-1?companySlug=acme",
			wantEmptyBody: true,
		},
		{
			name: "completed duration report",
			call: func() error {
				_, err := client.Admin.TimePlans.GetCompletedDurationReport(context.Background(), "acme", "run-1")
				return err
			},
			wantMethod:    http.MethodGet,
			wantURL:       "http://localhost:8080/api/v1/admin/time-plans/runs/run-1/report?companySlug=acme",
			wantEmptyBody: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertTimePlanRequestCase(t, doer, test)
		})
	}
}

func TestTimePlanCompletedRunExportUsesTheReportingRoute(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	test := timePlanRequestCase{
		name: "completed run export",
		call: func() error {
			_, err := client.Admin.TimePlans.CreateCompletedRunExport(context.Background(), CompletedRunExportRequest{
				Source:  CompletedRunExportSource{ContractVersion: "completed-run-export.v1"},
				Formats: []ExportFormat{FormatCSV},
			})
			return err
		},
		wantMethod:   http.MethodPost,
		wantURL:      "http://localhost:8080/api/v1/reporting/exports/completed-run",
		wantJSONKeys: []string{"source", "formats"},
	}

	assertTimePlanRequestCase(t, doer, test)
}

func TestTimePlanPublishedVersionDecodesServerAllocation(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK,
		`{"uuid":"version-1","planUuid":"plan-1","versionNumber":2,"definitionHash":"hash-1","schemaVersion":"time-plan.definition.v1",`+
			`"allocatorVersion":"largest-remainder.v1","publishedAt":"2026-09-02T12:00:00Z","horizonMs":60000,"blockCount":1,`+
			`"allocations":[{"blockId":"block-1","sequence":0,"durationMs":60000}]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	version, err := client.Admin.TimePlans.GetPublishedTimePlanVersion(context.Background(), "acme", "plan-1", "version-1")
	if err != nil {
		t.Fatalf("GetPublishedTimePlanVersion error: %v", err)
	}
	if version.UUID != "version-1" || version.VersionNumber != 2 || version.HorizonMS != 60000 {
		t.Fatalf("version = %+v", version)
	}
	if len(version.Allocations) != 1 || version.Allocations[0].BlockID != "block-1" {
		t.Fatalf("allocations = %+v", version.Allocations)
	}
	if version.PublishedAt.IsZero() {
		t.Fatalf("publishedAt was not decoded: %+v", version)
	}
}

func TestTimePlanCompletedDurationReportDecodesMeasurementContract(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK,
		`{"tenant_id":"acme","summary":{"run_id":"run-1","plan_id":"plan-1","version_id":"version-1","run_status":"completed",`+
			`"stream_version":4,"completed_at":"2026-09-02T12:00:00Z","baseline_horizon_ms":60000,"lost_ms":0,"unused_ms":0,`+
			`"overrun_ms":0,"planned_total_ms":60000,"current_total_ms":60000,"actual_total_ms":59000,"final_variance_ms":1000,`+
			`"actual_complete":true,"block_count":1,"actual_block_count":1,"completed_block_count":1,"skipped_block_count":0,`+
			`"cancelled_block_count":0,"not_run_block_count":0,"censored_block_count":0},`+
			`"blocks":[{"block_uuid":"block-1","sequence":0,"semantic_key":"focus","planned_ms":60000,"current_ms":60000,`+
			`"actual_ms":59000,"status":"completed","outcome":"completed"}],"adjustments":[],`+
			`"annotations":[{"uuid":"annotation-1","type":"note","text":"hello","recorded_at":"2026-09-02T12:00:00Z","actor_kind":"human"}],`+
			`"prediction_comparisons":[],"provenance":{"report_version":"completed-duration-report.v1","input_hash":"hash-1",`+
			`"generated_at":"2026-09-02T12:00:00Z","corrected":false,"adjustment_count":0,"annotation_count":1,"prediction_count":0}}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	report, err := client.Admin.TimePlans.GetCompletedDurationReport(context.Background(), "acme", "run-1")
	if err != nil {
		t.Fatalf("GetCompletedDurationReport error: %v", err)
	}
	if report.TenantID != "acme" || report.Summary.RunID != "run-1" || report.Summary.RunStatus != CompletedDurationRunStatusCompleted {
		t.Fatalf("report = %+v", report)
	}
	if report.Summary.ActualTotalMS != 59000 || report.Summary.FinalVarianceMS != 1000 || !report.Summary.ActualComplete {
		t.Fatalf("summary = %+v", report.Summary)
	}
	if len(report.Blocks) != 1 || report.Blocks[0].BlockUUID != "block-1" || report.Blocks[0].Outcome != CompletedDurationOutcomeCompleted {
		t.Fatalf("blocks = %+v", report.Blocks)
	}
	if report.Blocks[0].ActualMS == nil || *report.Blocks[0].ActualMS != 59000 {
		t.Fatalf("block actual = %+v", report.Blocks[0])
	}
	if len(report.Annotations) != 1 || report.Annotations[0].Text != "hello" {
		t.Fatalf("annotations = %+v", report.Annotations)
	}
	if report.Provenance.ReportVersion != "completed-duration-report.v1" || report.Provenance.AnnotationCount != 1 {
		t.Fatalf("provenance = %+v", report.Provenance)
	}
}

func TestTimePlanCompletedRunExportDecodesQueuedJob(t *testing.T) {
	doer := newCaptureDoer(http.StatusAccepted,
		`{"id":"export-1","packKey":"completed_run","packGeneration":3,"dashboardKey":"completed-run","formats":["csv"],`+
			`"parametersDigest":"digest-1","state":"queued","progressStage":"queued","queuedAt":"2026-09-02T12:00:00Z",`+
			`"progressCounters":[{"key":"blocks","count":2}],"attemptCount":0,"cleanupState":"pending","cleanupAttempts":0}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	job, err := client.Admin.TimePlans.CreateCompletedRunExport(context.Background(), CompletedRunExportRequest{
		Source:         CompletedRunExportSource{ContractVersion: "completed-run-export.v1"},
		Formats:        []ExportFormat{FormatCSV},
		IdempotencyKey: "export-1",
	})
	if err != nil {
		t.Fatalf("CreateCompletedRunExport error: %v", err)
	}
	if job.ID != "export-1" || job.State != "queued" || job.PackGeneration != 3 {
		t.Fatalf("job = %+v", job)
	}
	if len(job.ProgressCounters) != 1 || job.ProgressCounters[0].Key != "blocks" || job.ProgressCounters[0].Count != 2 {
		t.Fatalf("progress counters = %+v", job.ProgressCounters)
	}
	var payload map[string]any
	if err := json.Unmarshal(doer.requests[0].Body, &payload); err != nil {
		t.Fatalf("decode request body: %v", err)
	}
	if payload["idempotencyKey"] != "export-1" {
		t.Fatalf("idempotencyKey = %+v", payload["idempotencyKey"])
	}
}
