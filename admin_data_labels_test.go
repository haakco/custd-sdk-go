package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestDataLabelsAdminClientRoutes(t *testing.T) {
	doer := &dataLabelDoer{}
	client := NewClient(&ClientConfig{BaseURL: "http://localhost:8080", APIKey: "token", HTTPClient: doer})
	doer.responses = dataLabelAdminResponses()
	labels := client.Admin.DataLabels
	ctx := context.Background()
	create := DataLabelDefinitionCreateRequest{Key: "app.plan", DisplayName: "Plan", Description: "Plan", AllowedScopes: []string{"event"}, Sensitivity: "internal", IntendedUse: "Reporting", PropagationPolicy: "none"}
	update := DataLabelDefinitionUpdateRequest{DisplayName: "Plan", Description: "Updated", AllowedScopes: []string{"event"}, Sensitivity: "internal", IntendedUse: "Reporting", PropagationPolicy: "none"}
	_, _ = labels.List(ctx, true)
	_, _ = labels.Catalogue(ctx, true)
	_, _ = labels.Get(ctx, "definition/1", true)
	_, _ = labels.Create(ctx, create)
	_, _ = labels.Update(ctx, "definition/1", update)
	_ = labels.Disable(ctx, "definition/1")
	_, _ = labels.CreateValue(ctx, "definition/1", DataLabelValueCreateRequest{Value: "paid", DisplayName: "Paid", Description: "Paid"})
	_, _ = labels.UpdateValue(ctx, "value/1", DataLabelValueUpdateRequest{DisplayName: "Paid", Description: "Updated"})
	_ = labels.DisableValue(ctx, "value/1")
	_, _ = labels.ListUsage(ctx)
	_, _ = labels.ListAssignments(ctx)
	_ = labels.SetEventTypeDefault(ctx, "page/view", "definition/1", EventTypeDataLabelDefaultRequest{ValueUUID: "value-1"})
	_ = labels.RemoveEventTypeDefault(ctx, "page/view", "definition/1")
	_ = labels.SetSchemaFieldAssignment(ctx, "schema/1", SchemaFieldDataLabelAssignmentRequest{FieldPath: "/user/id", DefinitionUUID: "definition-1", ValueUUID: "value-1"})
	_ = labels.RemoveSchemaFieldAssignment(ctx, "assignment/1")
	want := []string{
		"GET /api/v1/admin/data-labels?includeDisabled=true", "GET /api/v1/admin/data-labels/catalogue?includeDisabled=true",
		"GET /api/v1/admin/data-labels/definition%2F1?includeDisabled=true",
		"POST /api/v1/admin/data-labels", "PATCH /api/v1/admin/data-labels/definition%2F1", "POST /api/v1/admin/data-labels/definition%2F1/disable",
		"POST /api/v1/admin/data-labels/definition%2F1/values", "PATCH /api/v1/admin/data-label-values/value%2F1", "POST /api/v1/admin/data-label-values/value%2F1/disable",
		"GET /api/v1/admin/data-labels/usage", "GET /api/v1/admin/data-label-assignments",
		"PUT /api/v1/admin/event-types/page%2Fview/data-label-defaults/definition%2F1", "DELETE /api/v1/admin/event-types/page%2Fview/data-label-defaults/definition%2F1",
		"PUT /api/v1/admin/event-schemas/schema%2F1/field-data-labels", "DELETE /api/v1/admin/data-label-assignments/schema-fields/assignment%2F1",
	}
	if got := requestMethodsAndPaths(doer.requests); !equalStrings(got, want) {
		t.Fatalf("routes = %#v, want %#v", got, want)
	}
}

type dataLabelDoer struct {
	requests  []*HTTPRequest
	responses []*HTTPResponse
}

func (d *dataLabelDoer) Do(req *HTTPRequest) (*HTTPResponse, error) {
	d.requests = append(d.requests, req)
	response := d.responses[0]
	d.responses = d.responses[1:]
	return response, nil
}

func dataLabelAdminResponses() []*HTTPResponse {
	bodies := []any{map[string]any{"definitions": []any{}}, map[string]any{"catalogue": map[string]any{"labels": []any{}}, "assignments": []any{}, "reportingPacks": []any{}, "fingerprint": "sha256"}, map[string]any{"uuid": "definition-1", "values": []any{}}, map[string]any{"uuid": "definition-1"}, map[string]any{"uuid": "definition-1"}, nil, map[string]any{"uuid": "value-1"}, map[string]any{"uuid": "value-1"}, nil, map[string]any{"usage": []any{}}, map[string]any{"eventTypeDefaults": []any{}, "schemaFieldAssignments": []any{}}, nil, nil, nil, nil}
	out := make([]*HTTPResponse, 0, len(bodies))
	for _, body := range bodies {
		data, _ := json.Marshal(body)
		status := http.StatusOK
		if body == nil {
			status = http.StatusNoContent
			data = nil
		}
		out = append(out, &HTTPResponse{StatusCode: status, Body: data})
	}
	return out
}

func requestMethodsAndPaths(requests []*HTTPRequest) []string {
	out := make([]string, len(requests))
	for i, req := range requests {
		out[i] = req.Method + " " + req.URL[len("http://localhost:8080"):]
	}
	return out
}
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
