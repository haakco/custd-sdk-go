package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestMeasurementProjectsCreateUsesAdminAPI(t *testing.T) {
	doer := newCaptureDoer(http.StatusCreated, `{"projectUuid":"project-123","projectSlug":"checkout-runway","name":"Checkout Runway","kind":"deadline_forecast","status":"active"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080/")
	var req MeasurementProjectCreate
	if err := json.Unmarshal(readContractFixture(t, "measurement-project-create.json"), &req); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	project, err := client.Admin.Measurement.Projects.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if project.ProjectUUID != "project-123" || project.ProjectSlug != "checkout-runway" {
		t.Fatalf("project = %+v", project)
	}
	assertSiteRequests(t, doer.requests, []string{
		"POST http://localhost:8080/api/v1/admin/measurement/projects",
	})
}

func TestMeasurementObservationBulkValidatesRowResults(t *testing.T) {
	doer := newCaptureDoer(http.StatusAccepted, string(readContractFixture(t, "measurement-observation-bulk-response.json")))
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	var req MeasurementObservationBulkRequest
	if err := json.Unmarshal(readContractFixture(t, "measurement-observation-bulk-request.json"), &req); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	response, err := client.Admin.Measurement.Projects.SubmitObservations(context.Background(), "checkout-runway", req)
	if err != nil {
		t.Fatalf("SubmitObservations returned error: %v", err)
	}

	if len(response.Results) != 2 || response.Results[0].ObservationUUID == "" {
		t.Fatalf("response = %+v", response)
	}
	assertSiteRequests(t, doer.requests, []string{
		"POST http://localhost:8080/api/v1/admin/measurement/projects/checkout-runway/observations:bulk",
	})
}

func TestMeasurementObservationBulkFailsOnResultCountMismatch(t *testing.T) {
	doer := newCaptureDoer(http.StatusAccepted, `{"results":[]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	_, err := client.Admin.Measurement.Projects.SubmitObservations(context.Background(), "checkout-runway", measurementBulkRequest(1))
	if err == nil {
		t.Fatal("expected result-count mismatch error, got nil")
	}
	if !strings.Contains(err.Error(), "measurement result count 0 does not match submitted row count 1") {
		t.Fatalf("error = %v", err)
	}
}

func TestMeasurementObservationBulkFailsOnMissingSuccessfulObservationUUID(t *testing.T) {
	doer := newCaptureDoer(http.StatusAccepted, `{"results":[{"rowIndex":1,"success":true,"status":202}]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	_, err := client.Admin.Measurement.Projects.SubmitObservations(context.Background(), "checkout-runway", measurementBulkRequest(1))
	if err == nil {
		t.Fatal("expected missing observation UUID error, got nil")
	}
	if !strings.Contains(err.Error(), "measurement result 0 missing observationUuid") {
		t.Fatalf("error = %v", err)
	}
}

func measurementBulkRequest(count int) MeasurementObservationBulkRequest {
	rows := make([]MeasurementObservationInput, count)
	for i := range rows {
		rows[i] = MeasurementObservationInput{
			SeriesSlug: "checkout-completions",
			ObservedAt: "2026-07-01T00:00:00Z",
			Value:      42,
		}
	}
	return MeasurementObservationBulkRequest{Rows: rows}
}
