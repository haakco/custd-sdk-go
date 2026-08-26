package custd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

// These tests cover the per-namespace lifecycle surfaces wired to
// contract-fixtures/lifecycle/. Every namespace loads its own happy-path
// (and where present, isolation / legal-hold / error) fixtures and asserts
// the SDK decodes them into the documented typed shape. The test doer
// mirrors the request capture pattern from admin_extended_test.go so URL
// and method are validated alongside the body decode.

// lifecycleDoer returns the same body for every request and records each
// call. Lifecycle tests rarely need to vary responses, but a couple need
// status shifts to surface a 4xx error envelope.
type lifecycleDoer struct {
	requests []*HTTPRequest
	status   int
	body     string
	headers  map[string]string
}

func (d *lifecycleDoer) Do(req *HTTPRequest) (*HTTPResponse, error) {
	d.requests = append(d.requests, req)
	return &HTTPResponse{StatusCode: d.status, Body: []byte(d.body), Headers: d.headers}, nil
}

func newLifecycleTestClient(t *testing.T, doer *lifecycleDoer, baseURL string) *CustdClient {
	t.Helper()
	client := NewClient(&ClientConfig{
		BaseURL:       baseURL,
		APIKey:        "admin-token",
		HTTPClient:    doer,
		FlushInterval: time.Hour,
	})
	t.Cleanup(func() { _ = client.Close(context.Background()) })
	return client
}

// TestTenantStorage_ListParses covers the happy path for the storage
// location list. It asserts the SDK decodes the per-location typed fields
// (id, tenant slug, server-prefix, status) and surfaces a non-empty list.
func TestTenantStorage_ListParses(t *testing.T) {
	body := readLifecycleFixture(t, "tenant-storage", "valid-list-response.json")
	doer := &lifecycleDoer{status: http.StatusOK, body: string(body)}
	client := newLifecycleTestClient(t, doer, "http://localhost:8080/")

	list, err := client.Admin.TenantStorage.List(context.Background())
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(list.Locations) != 2 {
		t.Fatalf("expected 2 locations, got %d (%+v)", len(list.Locations), list.Locations)
	}
	first := list.Locations[0]
	if first.ID != "loc_acme_warehouse" {
		t.Fatalf("first.ID = %q", first.ID)
	}
	if first.TenantSlug != "acme" {
		t.Fatalf("first.TenantSlug = %q", first.TenantSlug)
	}
	if first.ServerAssignedPrefix != "raw/acme/2026-07-31/" {
		t.Fatalf("first.ServerAssignedPrefix = %q", first.ServerAssignedPrefix)
	}
	if first.Status != "active" {
		t.Fatalf("first.Status = %q", first.Status)
	}
	if doer.requests[0].Method != http.MethodGet ||
		doer.requests[0].URL != "http://localhost:8080/api/v1/tenant-storage-locations" {
		t.Fatalf("List request = %+v", doer.requests[0])
	}
}

// TestTenantStorage_Isolation covers the cross-tenant isolation rule: when
// the SDK is authenticated as a different tenant, the list must collapse to
// an empty array indistinguishable from "no locations".
func TestTenantStorage_Isolation(t *testing.T) {
	body := readLifecycleFixture(t, "tenant-storage", "isolation-other-tenant-response.json")
	doer := &lifecycleDoer{status: http.StatusOK, body: string(body)}
	client := newLifecycleTestClient(t, doer, "http://localhost:8080/")

	list, err := client.Admin.TenantStorage.List(context.Background())
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(list.Locations) != 0 {
		t.Fatalf("isolation list must be empty, got %+v", list.Locations)
	}
}

// TestSubjectExport_CreateAndForce covers the happy path: a Create
// returns a typed receipt, and a Force is recorded on the same request.
func TestSubjectExport_CreateAndForce(t *testing.T) {
	doer := &lifecycleDoer{
		status: http.StatusCreated,
		body:   string(readLifecycleFixture(t, "subject-exports", "valid-create-response.json")),
	}
	client := newLifecycleTestClient(t, doer, "http://localhost:8080/")

	created, err := client.Admin.SubjectExports.Create(context.Background(), SubjectExportCreateRequest{
		TenantSlug: "acme",
		Subject: SubjectExportSubject{
			Type:  "userUuid",
			Value: "01J5K7N4Y8X9Z2B6V3D1M0Q7RJ",
		},
		Scope:          "portability",
		IdempotencyKey: "export-acme-01J5K7N4Y8X9Z2B6V3D1M0Q7RJ-2026-07-31",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if created.RequestID != "se_01J5K7N4Y8X9Z2B6V3D1M0Q7RJ" {
		t.Fatalf("created.RequestID = %q", created.RequestID)
	}
	if created.State != "queued" {
		t.Fatalf("created.State = %q", created.State)
	}
	if created.Subject.Type != "userUuid" {
		t.Fatalf("created.Subject.Type = %q", created.Subject.Type)
	}
	if doer.requests[0].Method != http.MethodPost ||
		doer.requests[0].URL != "http://localhost:8080/api/v1/admin/subject-exports" {
		t.Fatalf("Create request = %+v", doer.requests[0])
	}

	// Switch the doer to the force fixture and verify Force decodes a
	// ready-state response.
	doer.status = http.StatusOK
	doer.body = string(readLifecycleFixture(t, "subject-exports", "valid-force-response.json"))
	forced, err := client.Admin.SubjectExports.Force(context.Background(), "se_01J5K7N4Y8X9Z2B6V3D1M0Q7RJ")
	if err != nil {
		t.Fatalf("Force error: %v", err)
	}
	if forced.State != "ready" {
		t.Fatalf("forced.State = %q", forced.State)
	}
	if doer.requests[1].Method != http.MethodPost ||
		doer.requests[1].URL != "http://localhost:8080/api/v1/admin/subject-exports/se_01J5K7N4Y8X9Z2B6V3D1M0Q7RJ/force" {
		t.Fatalf("Force request = %+v", doer.requests[1])
	}
}

// TestSubjectExport_Expiry covers the terminal "download URL expired" path.
// The server returns 4xx with a stable error code the SDK must surface
// without retry. We assert the error envelope is decoded into Problem.
func TestSubjectExport_Expiry(t *testing.T) {
	body := readLifecycleFixture(t, "subject-exports", "expired-download-response.json")
	doer := &lifecycleDoer{status: http.StatusGone, body: string(body)}
	client := newLifecycleTestClient(t, doer, "http://localhost:8080/")

	_, err := client.Admin.SubjectExports.Download(context.Background(), "se_01J5K7N4Y8X9Z2B6V3D1M0Q7RJ")
	if err == nil {
		t.Fatalf("expected terminal error for expired download")
	}
	problem := parseProblem(body)
	if problem == nil {
		t.Fatalf("expected Problem envelope in body, got nil")
	}
	if problem.Code != "download_expired" {
		t.Fatalf("problem.Code = %q", problem.Code)
	}
	// The error string must not echo the signed URL value.
	if got := err.Error(); contains(got, "signed.example.invalid") {
		t.Fatalf("error leaked signed URL: %s", got)
	}
}

// TestPrivacyErasure_HappyPath covers Create + Get with per-store
// progress. Asserts every store is decoded, including the deletedCount.
func TestPrivacyErasure_HappyPath(t *testing.T) {
	createBody := readLifecycleFixture(t, "privacy-erasures", "valid-create-response.json")
	doer := &lifecycleDoer{status: http.StatusCreated, body: string(createBody)}
	client := newLifecycleTestClient(t, doer, "http://localhost:8080/")

	created, err := client.Admin.Erasures.Create(context.Background(), PrivacyErasureCreateRequest{
		TenantSlug: "acme",
		Selector: PrivacyErasureSelector{
			Type:  "userUuid",
			Value: "01J5K7N4Y8X9Z2B6V3D1M0Q7RJ",
		},
		Reason: "gdpr_erasure_request",
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if created.RequestUUID != "pe_01J5K7N4Y8X9Z2B6V3D1M0Q7RJ" {
		t.Fatalf("created.RequestUUID = %q", created.RequestUUID)
	}

	getBody := readLifecycleFixture(t, "privacy-erasures", "valid-get-response.json")
	doer.status = http.StatusOK
	doer.body = string(getBody)
	got, err := client.Admin.Erasures.Get(context.Background(), "pe_01J5K7N4Y8X9Z2B6V3D1M0Q7RJ")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if got.State != "complete" {
		t.Fatalf("got.State = %q", got.State)
	}
	if len(got.PerStoreProgress) != 5 {
		t.Fatalf("expected 5 store rows, got %d", len(got.PerStoreProgress))
	}
	for _, row := range got.PerStoreProgress {
		if row.Store == "" || row.State == "" {
			t.Fatalf("empty store or state in %+v", row)
		}
	}
}

// TestPrivacyErasure_LegalHold covers the legal-hold retention rule. The
// legal_hold store must surface as state=retained with deletedCount=0; the
// SDK must not collapse it or treat it as terminal-complete.
func TestPrivacyErasure_LegalHold(t *testing.T) {
	body := readLifecycleFixture(t, "privacy-erasures", "legal-hold-retained.json")
	doer := &lifecycleDoer{status: http.StatusOK, body: string(body)}
	client := newLifecycleTestClient(t, doer, "http://localhost:8080/")

	got, err := client.Admin.Erasures.Get(context.Background(), "pe_01J5K7N4Y8X9Z2B6V3D1M0Q7RJ")
	if err != nil {
		t.Fatalf("Get error: %v", err)
	}
	if got.State != "partial" {
		t.Fatalf("got.State = %q, want partial", got.State)
	}
	var legalHold *PrivacyErasureStoreProgress
	for i, row := range got.PerStoreProgress {
		if row.Store == "legal_hold" {
			legalHold = &got.PerStoreProgress[i]
			break
		}
	}
	if legalHold == nil {
		t.Fatalf("legal_hold store missing from perStoreProgress: %+v", got.PerStoreProgress)
	}
	if legalHold.State != "retained" {
		t.Fatalf("legalHold.State = %q, want retained", legalHold.State)
	}
	if legalHold.DeletedCount != 0 {
		t.Fatalf("legalHold.DeletedCount = %d, want 0", legalHold.DeletedCount)
	}
	if legalHold.Reason != "legal_hold" {
		t.Fatalf("legalHold.Reason = %q, want legal_hold", legalHold.Reason)
	}
}

// TestRetention_PreviewApplyRuns covers the full retention run lifecycle:
// preview returns server-computed counts, apply mints a run, and runs
// surfaces the run history including a complete and an in-flight run.
func TestRetention_PreviewApplyRuns(t *testing.T) {
	previewBody := readLifecycleFixture(t, "retention", "valid-preview-response.json")
	applyBody := readLifecycleFixture(t, "retention", "valid-apply-response.json")
	runsBody := readLifecycleFixture(t, "retention", "valid-runs-response.json")

	doer := &lifecycleDoer{status: http.StatusOK, body: string(previewBody)}
	client := newLifecycleTestClient(t, doer, "http://localhost:8080/")

	preview, err := client.Admin.Retention.Preview(context.Background(), "acme")
	if err != nil {
		t.Fatalf("Preview error: %v", err)
	}
	if preview.PreviewID == "" {
		t.Fatalf("preview.PreviewID empty")
	}
	if len(preview.EstimatedDeletions) != 3 {
		t.Fatalf("expected 3 estimatedDeletions, got %d", len(preview.EstimatedDeletions))
	}
	if preview.EstimatedDeletions[0].Store != "raw_landing" ||
		preview.EstimatedDeletions[0].Count != 142 {
		t.Fatalf("first deletion = %+v", preview.EstimatedDeletions[0])
	}
	if doer.requests[0].URL != "http://localhost:8080/api/v1/admin/retention/policies/acme/preview" {
		t.Fatalf("Preview URL = %s", doer.requests[0].URL)
	}
	if doer.requests[0].Method != http.MethodPost {
		t.Fatalf("Preview method = %s", doer.requests[0].Method)
	}

	doer.body = string(applyBody)
	applied, err := client.Admin.Retention.Apply(context.Background(), "acme")
	if err != nil {
		t.Fatalf("Apply error: %v", err)
	}
	if applied.RunID == "" {
		t.Fatalf("applied.RunID empty")
	}
	if applied.State != "running" {
		t.Fatalf("applied.State = %q", applied.State)
	}

	doer.body = string(runsBody)
	runs, err := client.Admin.Retention.ListRuns(context.Background(), "acme")
	if err != nil {
		t.Fatalf("ListRuns error: %v", err)
	}
	if len(runs.Runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs.Runs))
	}
	if runs.Runs[0].State != "complete" || runs.Runs[0].DeletedCount != 172 {
		t.Fatalf("first run = %+v", runs.Runs[0])
	}
	if runs.Runs[1].State != "running" {
		t.Fatalf("second run.State = %q", runs.Runs[1].State)
	}
}

// TestRetention_SelectorlessScope covers the error envelope the server
// returns when a policy is upserted with a scope that requires an explicit
// selector the SDK does not yet support. The SDK must surface the error
// without retry.
func TestRetention_SelectorlessScope(t *testing.T) {
	body := readLifecycleFixture(t, "retention", "invalid-selectorless-scope.json")
	doer := &lifecycleDoer{status: http.StatusBadRequest, body: string(body)}
	client := newLifecycleTestClient(t, doer, "http://localhost:8080/")

	_, err := client.Admin.Retention.Upsert(context.Background(), "acme", RetentionPolicyUpsertRequest{
		MaxAgeDays:          30,
		HardDeleteAfterDays: 60,
	})
	if err == nil {
		t.Fatalf("expected selector_required error")
	}
	problem := parseProblem(body)
	if problem == nil {
		t.Fatalf("expected Problem envelope in body")
	}
	// The fixture uses a flat {error, message} shape; verify the SDK
	// surfaces a non-empty message either via Problem or status.
	if err.Error() == "" {
		t.Fatalf("empty error string")
	}
}

// TestOffboarding_FullLifecycle covers the end-to-end offboarding
// lifecycle: request create, preview, export, download, acknowledge,
// execute, and receipt. The receipt must include the per-store deletion
// summary and its unkeyed SHA256 integrity checksum.
func TestOffboarding_FullLifecycle(t *testing.T) {
	client := newLifecycleTestClient(t,
		&lifecycleDoer{
			status: http.StatusCreated,
			body:   string(readLifecycleFixture(t, "offboarding", "valid-request-create-response.json")),
		},
		"http://localhost:8080/",
	)

	created, err := client.Admin.Offboarding.RequestOffboarding(context.Background(), OffboardingRequestCreate{
		Confirmation: "acme",
	})
	if err != nil {
		t.Fatalf("RequestOffboarding error: %v", err)
	}
	if created.RequestUUID != "ob_01J5K7N4Y8X9Z2B6V3D1M0Q7RJ" {
		t.Fatalf("created.RequestUUID = %q", created.RequestUUID)
	}
	if created.State != "preview" || created.RequestedAt == "" {
		t.Fatalf("created = %+v", created)
	}

	doer, ok := client.config.HTTPClient.(*lifecycleDoer)
	if !ok {
		t.Fatalf("expected lifecycleDoer transport")
	}
	doer.status = http.StatusOK
	doer.body = string(readLifecycleFixture(t, "offboarding", "valid-preview-response.json"))
	preview, err := client.Admin.Offboarding.Preview(context.Background(), "ob_01J5K7N4Y8X9Z2B6V3D1M0Q7RJ")
	if err != nil {
		t.Fatalf("Preview error: %v", err)
	}
	if preview.PreviewInventoryDigest == "" || preview.GeneratedAt == "" || preview.ExpiresAt == "" {
		t.Fatalf("PreviewInventoryDigest empty")
	}
	if !preview.Complete || preview.Partial {
		t.Fatalf("preview completeness = complete:%v partial:%v", preview.Complete, preview.Partial)
	}
	if len(preview.Stores) != 3 {
		t.Fatalf("expected 3 stores, got %d", len(preview.Stores))
	}
	if preview.Stores[0].RetentionClass != "operational" {
		t.Fatalf("first preview store = %+v", preview.Stores[0])
	}

	doer.body = string(readLifecycleFixture(t, "offboarding", "valid-export-response.json"))
	export, err := client.Admin.Offboarding.Export(context.Background(), "ob_01J5K7N4Y8X9Z2B6V3D1M0Q7RJ")
	if err != nil {
		t.Fatalf("Export error: %v", err)
	}
	if export.ChecksumSHA256 == "" || export.ByteSize != 4096 || export.RecordCount != 1357 {
		t.Fatalf("export metadata = %+v", export)
	}
	if export.PreviewInventoryDigest == "" || export.GeneratedAt == "" || export.ExpiresAt == "" {
		t.Fatalf("export timestamps/digest = %+v", export)
	}

	var downloadFixture struct {
		BodyBase64     string `json:"bodyBase64"`
		ChecksumSHA256 string `json:"checksumSha256"`
		ByteSize       int64  `json:"byteSize"`
	}
	if err := json.Unmarshal(readLifecycleFixture(t, "offboarding", "valid-download-binary.json"), &downloadFixture); err != nil {
		t.Fatalf("decode download fixture: %v", err)
	}
	downloadBytes, err := base64.StdEncoding.DecodeString(downloadFixture.BodyBase64)
	if err != nil {
		t.Fatalf("decode download bytes: %v", err)
	}
	doer.body = string(downloadBytes)
	doer.headers = map[string]string{
		"Content-Length":    strconv.FormatInt(downloadFixture.ByteSize, 10),
		"X-Checksum-SHA256": downloadFixture.ChecksumSHA256,
	}
	download, err := client.Admin.Offboarding.Download(context.Background(), "ob_01J5K7N4Y8X9Z2B6V3D1M0Q7RJ")
	if err != nil {
		t.Fatalf("Download error: %v", err)
	}
	if string(download.Bytes) != string(downloadBytes) ||
		download.ChecksumSHA256 != downloadFixture.ChecksumSHA256 ||
		download.ByteSize != downloadFixture.ByteSize {
		t.Fatalf("download = %+v, want exact fixture bytes and metadata", download)
	}
	if got := doer.requests[len(doer.requests)-1].Headers["Authorization"]; got != "Bearer admin-token" {
		t.Fatalf("download authorization = %q", got)
	}

	doer.body = string(readLifecycleFixture(t, "offboarding", "valid-acknowledge-response.json"))
	ack, err := client.Admin.Offboarding.Acknowledge(context.Background(), "ob_01J5K7N4Y8X9Z2B6V3D1M0Q7RJ")
	if err != nil {
		t.Fatalf("Acknowledge error: %v", err)
	}
	if ack.State != "requested" || ack.RequestedAt == "" {
		t.Fatalf("ack.State = %q", ack.State)
	}

	doer.body = string(readLifecycleFixture(t, "offboarding", "valid-execute-response.json"))
	doer.headers = nil
	exec, err := client.Admin.Offboarding.Execute(context.Background(), "ob_01J5K7N4Y8X9Z2B6V3D1M0Q7RJ")
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if exec.CompanyID != 42 || exec.FinalState != "complete" {
		t.Fatalf("exec = %+v", exec)
	}
	if exec.Waiver == nil || exec.Waiver.Role != "client_owner" {
		t.Fatalf("exec.Waiver = %+v", exec.Waiver)
	}
	if got := doer.requests[len(doer.requests)-1].Body; len(got) != 0 {
		t.Fatalf("execute body = %q, want empty", got)
	}

	doer.body = string(readLifecycleFixture(t, "offboarding", "valid-receipt-response.json"))
	receipt, err := client.Admin.Offboarding.Receipt(context.Background(), "ob_01J5K7N4Y8X9Z2B6V3D1M0Q7RJ")
	if err != nil {
		t.Fatalf("Receipt error: %v", err)
	}
	if receipt.CompanyID != 42 || receipt.FinalState != "complete" {
		t.Fatalf("receipt.FinalState = %q", receipt.FinalState)
	}
	if receipt.RequestedByActor != "user:u_01J5K7N4Y8X9Z2B6V3D1M0Q7RJ" {
		t.Fatalf("receipt.RequestedByActor = %q", receipt.RequestedByActor)
	}
	if receipt.RequestedByUserID == nil || *receipt.RequestedByUserID != 7 {
		t.Fatalf("receipt.RequestedByUserID = %v", receipt.RequestedByUserID)
	}
	if receipt.SHA256 == "" {
		t.Fatalf("receipt.SHA256 empty")
	}
	if len(receipt.PerStore) != 3 {
		t.Fatalf("expected 3 perStore rows in receipt, got %d", len(receipt.PerStore))
	}
	for _, row := range receipt.PerStore {
		if row.Store == "" || row.RetentionClass == "" {
			t.Fatalf("incomplete perStore row: %+v", row)
		}
	}

	machineReceipt := map[string]any{}
	if err := json.Unmarshal([]byte(doer.body), &machineReceipt); err != nil {
		t.Fatalf("decode receipt fixture: %v", err)
	}
	machineReceipt["requested_by_actor"] = "client:tiao-lifecycle"
	machineReceipt["requested_by_user_id"] = nil
	machineBody, err := json.Marshal(machineReceipt)
	if err != nil {
		t.Fatalf("encode machine receipt fixture: %v", err)
	}
	doer.body = string(machineBody)
	machine, err := client.Admin.Offboarding.Receipt(context.Background(), "ob_01J5K7N4Y8X9Z2B6V3D1M0Q7RJ")
	if err != nil {
		t.Fatalf("machine Receipt error: %v", err)
	}
	if machine.RequestedByActor != "client:tiao-lifecycle" {
		t.Fatalf("machine.RequestedByActor = %q", machine.RequestedByActor)
	}
	if machine.RequestedByUserID != nil {
		t.Fatalf("machine.RequestedByUserID = %d, want nil", *machine.RequestedByUserID)
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
			doer := &lifecycleDoer{status: http.StatusOK, body: "abc", headers: tt.headers}
			client := newLifecycleTestClient(t, doer, "http://localhost:8080/")
			if _, err := client.Admin.Offboarding.Download(context.Background(), "request-1"); err == nil {
				t.Fatal("Download succeeded with invalid integrity metadata")
			}
		})
	}
}

// TestOffboarding_ErasureIncompleteBlocksConfirm covers the cross-pipeline
// safety rule: a destructive offboarding confirm must be rejected when a
// related erasure is not yet terminal-complete. The server returns a
// safeNextAction the SDK must surface verbatim.
func TestOffboarding_ErasureIncompleteBlocksConfirm(t *testing.T) {
	body := readLifecycleFixture(t, "offboarding", "incomplete-erasure-blocks-confirm.json")
	doer := &lifecycleDoer{status: http.StatusConflict, body: string(body)}
	client := newLifecycleTestClient(t, doer, "http://localhost:8080/")

	_, err := client.Admin.Offboarding.ConfirmRequest(context.Background(), "ob_01J5K7N4Y8X9Z2B6V3D1M0Q7RJ")
	if err == nil {
		t.Fatalf("expected erasure_incomplete error")
	}
	// Verify the SDK's error contains the safeNextAction guidance.
	if !contains(err.Error(), "retry_erasure") {
		t.Fatalf("error did not surface safeNextAction retry_erasure: %s", err.Error())
	}
	var sendErr *sendError
	if !errors.As(err, &sendErr) || sendErr.Problem == nil {
		t.Fatalf("error type = %T, want typed problem", err)
	}
	if sendErr.Problem.Code != "erasure_incomplete" || sendErr.Problem.SafeNextAction != "retry_erasure" {
		t.Fatalf("problem = %+v, want code and recovery guidance", sendErr.Problem)
	}
}

// contains is a tiny substring helper kept local so the test file does
// not pull strings/bytes just for two probes.
func contains(haystack, needle string) bool {
	if needle == "" {
		return true
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
