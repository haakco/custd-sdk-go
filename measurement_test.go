package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestMeasurementProjectsCreateUsesAdminAPI(t *testing.T) {
	doer := newCaptureDoer(http.StatusCreated, `{"projectUuid":"project-123","projectCode":"checkout-runway","name":"Checkout Runway","kind":"deadline_forecast","status":"active","series":[{"seriesUuid":"series-123","seriesCode":"checkout-completions","name":"Checkout completions","unitSlug":"count","completionDirection":"increase","source":"manual"}],"target":{"targetUuid":"target-123","targetCode":"release","name":"Release","targetValue":100,"state":"active"}}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080/")
	var req MeasurementProjectCreate
	if err := json.Unmarshal(readContractFixture(t, "measurement-project-create.json"), &req); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	project, err := client.Admin.Measurement.Projects.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if project.ProjectUUID != "project-123" || project.ProjectCode != "checkout-runway" {
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

func TestMeasurementCSVImportValidatesRowResults(t *testing.T) {
	doer := newCaptureDoer(http.StatusAccepted, string(readContractFixture(t, "measurement-csv-import-response.json")))
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	response, err := client.Admin.Measurement.Projects.ImportCSVString(
		context.Background(),
		"checkout-runway",
		"seriesUuid,observedAt,value\ncheckout-completions,2026-07-01T00:00:00Z,42.5\n",
		2,
	)
	if err != nil {
		t.Fatalf("ImportCSVString returned error: %v", err)
	}

	if response.Accepted != 1 || response.Rejected != 1 {
		t.Fatalf("response = %+v", response)
	}
	assertSiteRequests(t, doer.requests, []string{
		"POST http://localhost:8080/api/v1/admin/measurement/projects/checkout-runway/observations:csv",
	})
}

func TestMeasurementImportRunGetUsesAdminAPI(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"importId":"import-123","projectCode":"checkout-runway","source":"csv","rowCount":2,"accepted":1,"rejected":1,"createdAt":"2026-07-06T00:00:00Z","results":[{"rowIndex":1,"success":true,"observationUuid":"obs-1"},{"rowIndex":2,"success":false,"status":422,"title":"Invalid measurement observation","detail":"observedAt must be an RFC3339 timestamp"}]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	run, err := client.Admin.Measurement.Projects.GetImportRun(context.Background(), "checkout-runway", "import-123")
	if err != nil {
		t.Fatalf("GetImportRun returned error: %v", err)
	}
	if run.ImportID != "import-123" || run.RowCount != 2 || len(run.Results) != 2 {
		t.Fatalf("run = %+v", run)
	}
	assertSiteRequests(t, doer.requests, []string{
		"GET http://localhost:8080/api/v1/admin/measurement/projects/checkout-runway/imports/import-123",
	})
}

func TestMeasurementCalibrationExerciseCreateUsesAdminAPI(t *testing.T) {
	doer := newCaptureDoer(http.StatusCreated, `{"exerciseUuid":"exercise-123","projectCode":"checkout-runway","status":"scored","question":"Will checkout launch by target date?","confidence":80,"lowerBound":70,"upperBound":110,"actualValue":100,"contained":true,"resolvedAt":"2026-08-31T00:00:00Z"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	actualValue := 100.0
	contained := true

	exercise, err := client.Admin.Measurement.Projects.CreateCalibrationExercise(
		context.Background(),
		"checkout-runway",
		MeasurementCalibrationExerciseCreate{
			Question:    "Will checkout launch by target date?",
			Confidence:  80,
			LowerBound:  70,
			UpperBound:  110,
			ActualValue: &actualValue,
			Contained:   &contained,
			ResolvedAt:  "2026-08-31T00:00:00Z",
		},
	)
	if err != nil {
		t.Fatalf("CreateCalibrationExercise returned error: %v", err)
	}
	if exercise.ExerciseUUID != "exercise-123" || exercise.Status != "scored" {
		t.Fatalf("exercise = %+v", exercise)
	}
	assertSiteRequests(t, doer.requests, []string{
		"POST http://localhost:8080/api/v1/admin/measurement/projects/checkout-runway/calibration-exercises",
	})
}

func TestMeasurementDecisionModelCreateUsesAdminAPI(t *testing.T) {
	doer := newCaptureDoer(http.StatusCreated, `{"decisionUuid":"decision-123","decisionCode":"launch-choice","projectCode":"checkout-runway","status":"active","name":"Launch choice","choices":[{"choiceCode":"ship","name":"Ship","expression":{"kind":"variable","variableCode":"conversion-rate"}}],"variables":[{"variableUuid":"variable-123","variableCode":"conversion-rate","name":"Conversion rate","unitSlug":"percent","distributionSlug":"normal","parameters":{"mean":10,"stddev":2},"measurementCost":1}]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	decision, err := client.Admin.Measurement.Projects.CreateDecisionModel(
		context.Background(),
		"checkout-runway",
		MeasurementDecisionModelCreate{
			DecisionCode: "launch-choice",
			Name:         "Launch choice",
			Status:       "active",
			Choices: []MeasurementDecisionChoice{{
				ChoiceCode: "ship",
				Name:       "Ship",
				Expression: &MeasurementDecisionExpression{Kind: "variable", VariableCode: "conversion-rate"},
			}},
			Variables: []MeasurementDecisionVariable{{
				VariableCode:     "conversion-rate",
				Name:             "Conversion rate",
				UnitSlug:         "percent",
				DistributionSlug: "normal",
				Parameters:       map[string]any{"mean": 10, "stddev": 2},
				MeasurementCost:  1,
			}},
		},
	)
	if err != nil {
		t.Fatalf("CreateDecisionModel returned error: %v", err)
	}
	if decision.DecisionUUID != "decision-123" || decision.DecisionCode != "launch-choice" ||
		decision.ProjectCode != "checkout-runway" || len(decision.Variables) != 1 ||
		decision.Variables[0].VariableUUID != "variable-123" || decision.Variables[0].UnitSlug != "percent" {
		t.Fatalf("decision = %+v", decision)
	}
	// The server reads the variable unit as unitSlug; any other key is dropped.
	if !strings.Contains(string(doer.requests[0].Body), `"unitSlug":"percent"`) {
		t.Fatalf("create body = %s", doer.requests[0].Body)
	}
	assertSiteRequests(t, doer.requests, []string{
		"POST http://localhost:8080/api/v1/admin/measurement/projects/checkout-runway/decisions",
	})
}

func TestMeasurementForecastRunUsesAnalyticsAPI(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"forecastRunUuid":"forecast-123","projectCode":"checkout-runway","status":"completed","method":"bootstrap_quantile","generatedAt":"2026-07-05T12:00:00Z","inputCount":42,"sampleCount":2000,"seed":770501,"percentiles":{"p10":"2026-08-12T00:00:00Z","p50":"2026-08-27T00:00:00Z","p90":"2026-09-18T00:00:00Z"},"warnings":[{"code":"sparse","severity":"warning","message":"limited input"}],"provenance":{"engineVersion":"measurement-go-v1","inputHash":"sha256:abc"}}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	forecast, err := client.Measurement.Projects.RunForecast(
		context.Background(),
		"checkout-runway",
		MeasurementForecastRunRequest{Seed: 770501, SampleCount: 2000},
	)
	if err != nil {
		t.Fatalf("RunForecast returned error: %v", err)
	}
	if forecast.ForecastRunUUID != "forecast-123" || forecast.Percentiles["p50"] == "" {
		t.Fatalf("forecast = %+v", forecast)
	}
	assertSiteRequests(t, doer.requests, []string{
		"POST http://localhost:8080/api/v1/measurement/projects/checkout-runway/forecast-runs",
	})
}

func TestMeasurementDecisionSimulationUsesAnalyticsAPI(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"simulationRunUuid":"simulation-123","projectCode":"checkout-runway","decisionCode":"launch-choice","status":"completed","method":"evpi_rank","generatedAt":"2026-07-05T12:00:00Z","completedAt":"2026-07-05T12:00:01Z","sampleCount":2000,"seed":770501,"variables":[{"variableUuid":"variable-123","name":"Conversion rate","rank":1,"evpi":12.5,"measurementCost":1,"worthMeasuring":true}],"warnings":[{"code":"decision_weight_defaulted","severity":"warning","message":"defaulted"}],"provenance":{"engineVersion":"measurement-go-v1","inputHash":"sha256:def"}}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	simulation, err := client.Measurement.Projects.RunDecisionSimulation(
		context.Background(),
		"checkout-runway",
		"launch-choice",
		MeasurementSimulationRunRequest{Seed: 770501, SampleCount: 2000},
	)
	if err != nil {
		t.Fatalf("RunDecisionSimulation returned error: %v", err)
	}
	if simulation.SimulationRunUUID != "simulation-123" || len(simulation.Variables) != 1 ||
		simulation.Variables[0].VariableUUID != "variable-123" {
		t.Fatalf("simulation = %+v", simulation)
	}
	assertSiteRequests(t, doer.requests, []string{
		"POST http://localhost:8080/api/v1/measurement/projects/checkout-runway/decisions/launch-choice/simulation-runs",
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
			SeriesUUID: "checkout-completions",
			ObservedAt: "2026-07-01T00:00:00Z",
			Value:      42,
		}
	}
	return MeasurementObservationBulkRequest{Rows: rows}
}

// TestMeasurementProjectUpdateClearsWithAnEmptyString proves "" reaches the server, which clears
// the field, while a nil field is left out and kept; an unknown chance and movement stay nil.
func TestMeasurementProjectUpdateClearsWithAnEmptyString(t *testing.T) {
	body, err := json.Marshal(MeasurementProjectUpdate{Description: new(""), Name: "Runway"})
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{"description":"","name":"Runway"}` {
		t.Fatalf("update body = %s", body)
	}
	var status MeasurementEstimateStatus
	if err := json.Unmarshal([]byte(`{"ageDays":1,"chanceByDeliveryTarget":null,"movementDays":null,"estimate":{}}`), &status); err != nil {
		t.Fatal(err)
	}
	if status.ChanceByDeliveryTarget != nil || status.MovementDays != nil {
		t.Fatalf("unknown chance or movement read as a value: %+v", status)
	}
	var first MeasurementForecastStoryEstimate
	if err := json.Unmarshal([]byte(`{"estimatedAt":"2026-11-01T00:00:00Z","high":"2026-12-01T00:00:00Z","low":"2026-11-20T00:00:00Z","widthDays":11}`), &first); err != nil {
		t.Fatal(err)
	}
	if first.MovementDays != nil {
		t.Fatalf("first story estimate movement = %v, want nil", *first.MovementDays)
	}
}
