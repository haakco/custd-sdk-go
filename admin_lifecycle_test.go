package custd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

// These tests cover lifecycle behaviour the route conformance rows do not:
// request bodies and query parameters the SDK sends, typed server problems,
// tenant selection, and offboarding download integrity. Response bodies are
// the server's generated examples or are built from the server's own types.

const lifecycleBaseURL = "http://localhost:8080/api/v1/admin"

// lifecycleExample returns one side ("minimal" or "maximal") of a server example.
func lifecycleExample(t *testing.T, operationID, side string) string {
	t.Helper()
	var example map[string]json.RawMessage
	if err := json.Unmarshal(readContractFixture(t, "api/examples/"+operationID+".json"), &example); err != nil {
		t.Fatalf("parse example %s: %v", operationID, err)
	}
	return string(example[side])
}

// notFoundProblem is the body httpx.WriteError emits for apierr.NotFound(detail).
func notFoundProblem(detail, instance string) string {
	return `{"type":"not_found","title":"Not Found","status":404,"detail":"` + detail +
		`","instance":"` + instance + `","traceId":"req-1","retryability":"none","nextAction":{"action":"none"}}`
}

func requireProblem(t *testing.T, err error, status int) *Problem {
	t.Helper()
	var requestErr *RequestError
	if !errors.As(err, &requestErr) || requestErr.Problem == nil {
		t.Fatalf("error = %T %v, want a typed problem", err, err)
	}
	if requestErr.StatusCode != status || requestErr.Retryable {
		t.Fatalf("error status = %d retryable = %v, want %d and not retryable", requestErr.StatusCode, requestErr.Retryable, status)
	}
	return requestErr.Problem
}

func TestSubjectExport_SendsBodyLimitAndScope(t *testing.T) {
	doer := newCaptureDoer(http.StatusAccepted, lifecycleExample(t, "createSubjectExport", "maximal"))
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	ctx := context.Background()

	if _, err := client.Admin.SubjectExports.Create(ctx, SubjectExportCreateRequest{
		OpaqueSubjectID: "subject-1", Scope: SubjectExportScopeBoth, IdempotencyKey: "export-1",
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got := string(doer.requests[0].Body); got != `{"opaqueSubjectId":"subject-1","scope":"both","idempotencyKey":"export-1"}` {
		t.Fatalf("Create body = %s", got)
	}
	doer.status, doer.body = http.StatusOK, lifecycleExample(t, "listSubjectExports", "minimal")
	for _, limit := range []int{25, 0} {
		if _, err := client.Admin.SubjectExports.List(ctx, limit); err != nil {
			t.Fatalf("List(%d): %v", limit, err)
		}
	}
	doer.body = lifecycleExample(t, "downloadSubjectExport", "maximal")
	if _, err := client.Admin.SubjectExports.Download(ctx, "r-1", SubjectExportScopePortability); err != nil {
		t.Fatalf("Download: %v", err)
	}
	assertSiteRequests(t, doer.requests, []string{
		"POST " + lifecycleBaseURL + "/subject-exports",
		"GET " + lifecycleBaseURL + "/subject-exports?limit=25",
		"GET " + lifecycleBaseURL + "/subject-exports",
		"GET " + lifecycleBaseURL + "/subject-exports/r-1/download?scope=portability",
	})
}

// A download whose artifact is not available is the server's typed 404. The
// SDK surfaces it once, without retry, and keeps the request path, which names
// the tenant's export, out of the error text.
func TestSubjectExport_UnavailableArtifactSurfacesTheProblem(t *testing.T) {
	const instance = "/api/v1/admin/subject-exports/0195f0a0-export/download"
	doer := newCaptureDoer(http.StatusNotFound, notFoundProblem("subject export artifact not found", instance))
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	_, err := client.Admin.SubjectExports.Download(context.Background(), "0195f0a0-export", SubjectExportScopeAccess)
	problem := requireProblem(t, err, http.StatusNotFound)
	if problem.Detail != "subject export artifact not found" || len(doer.requests) != 1 {
		t.Fatalf("problem = %+v after %d requests", problem, len(doer.requests))
	}
	if strings.Contains(err.Error(), "0195f0a0-export") {
		t.Fatalf("error leaked the export path: %s", err)
	}
}

// TestPrivacyErasure_SelectsTheTenant proves create sends the tenant and selector
// in the body, reads name a tenant only when asked, and a wrong-tenant read is
// the server's plain 404 rather than a distinguishable denial.
func TestPrivacyErasure_SelectsTheTenant(t *testing.T) {
	doer := newCaptureDoer(http.StatusAccepted, lifecycleExample(t, "createPrivacyErasure", "minimal"))
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	ctx := context.Background()

	if _, err := client.Admin.Erasures.Create(ctx, PrivacyErasureCreateRequest{
		CompanySlug: "acme", Selector: PrivacyErasureSelector{Type: "user_id", Values: []string{"u-1"}},
		Reason: "gdpr_erasure_request",
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if body := string(doer.requests[0].Body); !strings.Contains(body, `"companySlug":"acme"`) ||
		!strings.Contains(body, `"selector":{"type":"user_id","values":["u-1"]}`) {
		t.Fatalf("Create body = %s", body)
	}
	doer.status, doer.body = http.StatusOK, lifecycleExample(t, "listPrivacyErasures", "minimal")
	if _, err := client.Admin.Erasures.List(ctx, ""); err != nil {
		t.Fatalf("List: %v", err)
	}
	if _, err := client.Admin.Erasures.List(ctx, "beta"); err != nil {
		t.Fatalf("List beta: %v", err)
	}
	doer.status, doer.body = http.StatusNotFound, notFoundProblem("privacy: erasure job not found", "/api/v1/admin/privacy/erasures/e-1")
	_, err := client.Admin.Erasures.Get(ctx, "e-1", "beta")
	if problem := requireProblem(t, err, http.StatusNotFound); problem.Type != "not_found" {
		t.Fatalf("problem = %+v", problem)
	}
	assertSiteRequests(t, doer.requests, []string{
		"POST " + lifecycleBaseURL + "/privacy/erasures",
		"GET " + lifecycleBaseURL + "/privacy/erasures",
		"GET " + lifecycleBaseURL + "/privacy/erasures?companySlug=beta",
		"GET " + lifecycleBaseURL + "/privacy/erasures/e-1?companySlug=beta",
	})
}

// TestPrivacyErasure_ForceDecodesTheWrapper uses the server's
// privacyErasureForceResponse shape, which has no generated example.
func TestPrivacyErasure_ForceDecodesTheWrapper(t *testing.T) {
	request := lifecycleExample(t, "getPrivacyErasure", "minimal")
	doer := newCaptureDoer(http.StatusAccepted, `{"request":`+request+`,"safe_next_action":"","safe_next_action_code":""}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	forced, err := client.Admin.Erasures.Force(context.Background(), "e-1")
	if err != nil {
		t.Fatalf("Force: %v", err)
	}
	if forced.Request == nil || forced.Request.RequestUUID != "example" {
		t.Fatalf("forced = %+v", forced)
	}
}

// TestRetention_PreviewApplyRuns covers the run routes, whose server types
// (retention.RetentionRunPreview and RetentionRunsListResponse) have no
// generated schema or example. Legal-hold rows are counted per store.
func TestRetention_PreviewApplyRuns(t *testing.T) {
	const run = `{"perStore":[{"store":"raw","retentionClass":"raw","candidateCount":142,"estimatedAffected":140,` +
		`"sampleSubjects":["s-1"],"legalHoldCount":2}],"lastResult":"%s","safeNextAction":"none",` +
		`"startedAt":"2026-07-31T10:00:00Z","completedAt":"2026-07-31T10:00:01Z"}`
	doer := newCaptureDoer(http.StatusOK, strings.Replace(run, "%s", "preview", 1))
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	ctx := context.Background()

	preview, err := client.Admin.Retention.Preview(ctx, "acme")
	if err != nil {
		t.Fatalf("Preview: %v", err)
	}
	if preview.LastResult != "preview" || len(preview.PerStore) != 1 ||
		preview.PerStore[0].CandidateCount != 142 || preview.PerStore[0].LegalHoldCount != 2 {
		t.Fatalf("preview = %+v", preview)
	}
	doer.body = strings.Replace(run, "%s", "applied", 1)
	if applied, err := client.Admin.Retention.Apply(ctx, "acme"); err != nil || applied.LastResult != "applied" {
		t.Fatalf("Apply = %+v, %v", applied, err)
	}
	doer.body = `{"runs":[{"startAt":"2026-07-30T10:00:00Z","endAt":"2026-07-31T10:00:00Z","lastResult":"applied",` +
		`"lastAffectedCounts":{"raw":172},"auditId":7}],"nextOffset":1,"hasMore":true}`
	runs, err := client.Admin.Retention.ListRuns(ctx, "acme")
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(runs.Runs) != 1 || runs.Runs[0].LastAffectedCounts["raw"] != 172 || !runs.HasMore || runs.NextOffset != 1 {
		t.Fatalf("runs = %+v", runs)
	}
}

// TestOffboarding_DownloadAndExecute proves the archive bytes arrive exactly,
// with the server's integrity headers and the caller's credential, and that the
// destructive execute step sends no body.
func TestOffboarding_DownloadAndExecute(t *testing.T) {
	archive := "tenant archive bytes"
	sum := sha256.Sum256([]byte(archive))
	doer := newCaptureDoer(http.StatusOK, archive)
	doer.headers = map[string]string{
		"Content-Length":    strconv.Itoa(len(archive)),
		"X-Checksum-SHA256": hex.EncodeToString(sum[:]),
	}
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	ctx := context.Background()

	download, err := client.Admin.Offboarding.Download(ctx, "ob-1")
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if string(download.Bytes) != archive || download.ChecksumSHA256 != hex.EncodeToString(sum[:]) ||
		download.ByteSize != int64(len(archive)) {
		t.Fatalf("download = %+v", download)
	}
	if got := doer.requests[0].Headers["Authorization"]; got != "Bearer admin-token" {
		t.Fatalf("download authorization = %q", got)
	}

	doer.headers, doer.body = nil, lifecycleExample(t, "executeOffboarding", "maximal")
	if _, err := client.Admin.Offboarding.Execute(ctx, "ob-1"); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if got := doer.requests[1].Body; len(got) != 0 {
		t.Fatalf("execute body = %q, want empty", got)
	}
}

func TestOffboarding_DownloadFailsClosedOnInvalidIntegrityMetadata(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
	}{
		{"missing checksum", map[string]string{"Content-Length": "3"}},
		{"wrong checksum", map[string]string{"Content-Length": "3", "X-Checksum-SHA256": strings.Repeat("0", 64)}},
		{"length mismatch", map[string]string{"Content-Length": "2", "X-Checksum-SHA256": strings.Repeat("0", 64)}},
		{"declared oversize", map[string]string{"Content-Length": strconv.FormatInt(maxOffboardingDownloadBytes+1, 10), "X-Checksum-SHA256": strings.Repeat("0", 64)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, "abc")
			doer.headers = tt.headers
			client := newAdminTestClient(t, doer, "http://localhost:8080")
			if _, err := client.Admin.Offboarding.Download(context.Background(), "request-1"); err == nil {
				t.Fatal("Download succeeded with invalid integrity metadata")
			}
		})
	}
}

// TestOffboarding_ErasureIncompleteBlocksConfirm covers the cross-pipeline
// safety rule: a destructive confirm is refused while a subject erasure is not
// terminal. The body is the server's OffboardingWorkflowConflictResponse.
func TestOffboarding_ErasureIncompleteBlocksConfirm(t *testing.T) {
	doer := newCaptureDoer(http.StatusConflict, `{"error":"subject erasure has not reached a terminal state",`+
		`"code":"erasure_incomplete","safe_next_action":"retry_erasure"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	_, err := client.Admin.Offboarding.ConfirmRequest(context.Background(), "ob-1")
	problem := requireProblem(t, err, http.StatusConflict)
	if problem.Code != "erasure_incomplete" || problem.SafeNextAction != "retry_erasure" {
		t.Fatalf("problem = %+v, want code and recovery guidance", problem)
	}
	if !strings.Contains(err.Error(), "retry_erasure") {
		t.Fatalf("error did not surface safeNextAction retry_erasure: %s", err)
	}
}
