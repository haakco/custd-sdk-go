package custd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// realmWireServer is a stateful stand-in for the admin API's project-auth revision
// rules, so a provision test exercises the real ordering contract instead of a
// permissive fixed response. It enforces what the service enforces:
//
//   - Apply with expected 0 inserts environment revision 1; a later write must name
//     the current revision or it conflicts.
//   - A directory transition compares its expected revision against the environment
//     revision and advances it.
//   - A client reconcile requires the current revision and records its observation
//     without advancing the revision.
//   - Status is reconciled only when the recorded observation describes the current
//     revision.
//   - An Idempotency-Key must be 8 to 128 characters, and a repeated key returns the
//     recorded response instead of writing again.
type realmWireServer struct {
	projectID       string
	environmentID   string
	revision        int64
	observedAt      int64
	operations      map[string]string
	requests        []*HTTPRequest
	statusFailures  int
	skipObservation bool
}

func newRealmWireServer(projectID, environmentID string) *realmWireServer {
	return &realmWireServer{projectID: projectID, environmentID: environmentID, operations: map[string]string{}}
}

func (s *realmWireServer) Do(req *HTTPRequest) (*HTTPResponse, error) {
	s.requests = append(s.requests, req)
	key := req.Headers["Idempotency-Key"]
	if key != "" && (len(key) < 8 || len(key) > 128) {
		return jsonWireResponse(http.StatusBadRequest, `{"title":"invalid_idempotency_key"}`), nil
	}
	switch {
	case req.Method == http.MethodPost && strings.HasSuffix(req.URL, "/apply"):
		return s.apply(req, key), nil
	case req.Method == http.MethodPost && strings.HasSuffix(req.URL, "/directories/enable"):
		return s.enable(req, key), nil
	case req.Method == http.MethodPost && strings.HasSuffix(req.URL, "/clients/reconcile"):
		return s.reconcile(req, key), nil
	case req.Method == http.MethodGet && strings.HasSuffix(req.URL, "/status"):
		return s.status(), nil
	}
	return jsonWireResponse(http.StatusNotFound, `{"title":"not_found"}`), nil
}

func (s *realmWireServer) apply(req *HTTPRequest, key string) *HTTPResponse {
	if recorded, ok := s.operations[key]; ok {
		return jsonWireResponse(http.StatusOK, recorded)
	}
	var body struct {
		ExpectedRevision int64 `json:"expectedRevision"`
	}
	_ = json.Unmarshal(req.Body, &body)
	switch {
	case s.revision == 0 && body.ExpectedRevision == 0:
		s.revision = 1
	case s.revision != 0 && body.ExpectedRevision == s.revision:
		s.revision++
	default:
		return jsonWireResponse(http.StatusConflict, `{"title":"revision_conflict"}`)
	}
	operation := s.operationJSON("project_auth_config_update", key)
	s.operations[key] = operation
	return jsonWireResponse(http.StatusOK, operation)
}

func (s *realmWireServer) enable(req *HTTPRequest, key string) *HTTPResponse {
	if recorded, ok := s.operations[key]; ok {
		return jsonWireResponse(http.StatusOK, recorded)
	}
	var body struct {
		ExpectedRevision int64 `json:"expectedRevision"`
	}
	_ = json.Unmarshal(req.Body, &body)
	if body.ExpectedRevision != s.revision {
		return jsonWireResponse(http.StatusConflict, `{"title":"revision_conflict"}`)
	}
	s.revision++
	operation := s.operationJSON("directory_enable", key)
	s.operations[key] = operation
	return jsonWireResponse(http.StatusOK, operation)
}

func (s *realmWireServer) reconcile(req *HTTPRequest, key string) *HTTPResponse {
	if recorded, ok := s.operations[key]; ok {
		return jsonWireResponse(http.StatusOK, recorded)
	}
	var body struct {
		ExpectedRevision int64 `json:"expectedRevision"`
	}
	_ = json.Unmarshal(req.Body, &body)
	if body.ExpectedRevision != s.revision {
		return jsonWireResponse(http.StatusConflict, `{"title":"client_sync_stale"}`)
	}
	if !s.skipObservation {
		s.observedAt = s.revision
	}
	report := fmt.Sprintf(
		`{"environmentId":%q,"revision":%d,"checkedAt":"2026-10-16T00:00:00Z","superseded":false,"clients":[]}`,
		s.environmentID, s.revision,
	)
	s.operations[key] = report
	return jsonWireResponse(http.StatusOK, report)
}

func (s *realmWireServer) status() *HTTPResponse {
	if s.statusFailures > 0 {
		s.statusFailures--
		return jsonWireResponse(http.StatusInternalServerError, `{"title":"status_unavailable"}`)
	}
	reconciled := s.observedAt != 0 && s.observedAt == s.revision
	clientSync := "null"
	if s.observedAt != 0 {
		clientSync = fmt.Sprintf(
			`{"clientIds":[],"revision":%d,"checkedAt":"2026-10-16T00:00:00Z","current":%t,"registrations":[]}`,
			s.observedAt, s.observedAt == s.revision,
		)
	}
	body := fmt.Sprintf(
		`{"projectId":%q,"environmentId":%q,"revision":%d,`+
			`"desired":{"identityMode":"isolated","registrationPolicy":"invite_only","loginPaused":false,"audiences":[],"profileFields":[]},`+
			`"reconciled":%t,"reconcileNote":"","loginReady":false,"loginNote":"","clientSync":%s}`,
		s.projectID, s.environmentID, s.revision, reconciled, clientSync,
	)
	return jsonWireResponse(http.StatusOK, body)
}

func (s *realmWireServer) operationJSON(kind, key string) string {
	return fmt.Sprintf(
		`{"id":"op-%d","kind":%q,"status":"applied","revision":%d,"idempotencyKey":%q,`+
			`"appliedAt":"2026-10-16T00:00:00Z","replayed":false}`,
		s.revision, kind, s.revision, key,
	)
}

func jsonWireResponse(status int, body string) *HTTPResponse {
	return &HTTPResponse{StatusCode: status, Body: []byte(body)}
}

func newAdminDoerClient(t *testing.T, doer HTTPDoer) *CustdClient {
	t.Helper()
	client := NewClient(&ClientConfig{BaseURL: "http://localhost:8080", APIKey: "admin-token", HTTPClient: doer})
	t.Cleanup(func() { _ = client.Close(context.Background()) })
	return client
}

func realmProvisionRequest() AuthProjectRealmProvisionRequest {
	return AuthProjectRealmProvisionRequest{
		DesiredStateRequest: AuthProjectDesiredStateRequest{
			DesiredState: AuthProjectDesiredState{
				IdentityMode:       AuthProjectIdentityIsolated,
				RegistrationPolicy: "invite_only",
			},
			ExpectedRevision: 0,
		},
		DirectoryPoolID: "pool-1",
	}
}

// TestAuthProjectProvisionRealmEnforcesServerRevisionSequence is the stateful wire
// regression: the fake enforces the real revision rules, so a helper that enables
// the directory at the caller's expected revision (0) gets a conflict after the
// apply advanced the environment to revision 1. The corrected helper derives the
// directory revision from the apply and re-reconciles at the directory's revision.
func TestAuthProjectProvisionRealmEnforcesServerRevisionSequence(t *testing.T) {
	server := newRealmWireServer("project-1", "environment-1")
	client := newAdminDoerClient(t, server)

	receipt, err := client.Admin.AuthProjects.ProvisionRealm(
		context.Background(), "project-1", "environment-1", realmProvisionRequest(),
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: "realm-provision-1"},
	)
	if err != nil {
		t.Fatalf("ProvisionRealm returned error: %v", err)
	}

	if len(server.requests) != 4 {
		t.Fatalf("provision made %d requests, want 4: %v", len(server.requests), realmRequestPaths(server))
	}
	wantPaths := []string{"/apply", "/directories/enable", "/clients/reconcile", "/status"}
	for index, suffix := range wantPaths {
		if !strings.HasSuffix(server.requests[index].URL, suffix) {
			t.Fatalf("request %d url = %s, want suffix %s", index, server.requests[index].URL, suffix)
		}
	}
	wantKeys := []string{
		"realm-provision-1:apply",
		"realm-provision-1:directory-enable",
		"realm-provision-1:client-reconcile",
		"",
	}
	for index, wantKey := range wantKeys {
		if got := server.requests[index].Headers["Idempotency-Key"]; got != wantKey {
			t.Fatalf("request %d idempotency key = %q, want %q", index, got, wantKey)
		}
	}
	// The directory enable and the reconcile must name the revisions the server produced.
	var directoryRequest AuthProjectDirectoryRequest
	if err := json.Unmarshal(server.requests[1].Body, &directoryRequest); err != nil {
		t.Fatalf("decode directory body: %v", err)
	}
	if directoryRequest.ExpectedRevision != 1 {
		t.Fatalf("directory expected revision = %d, want 1 (derived from apply)", directoryRequest.ExpectedRevision)
	}
	var reconcileRequest AuthProjectClientSyncRequest
	if err := json.Unmarshal(server.requests[2].Body, &reconcileRequest); err != nil {
		t.Fatalf("decode reconcile body: %v", err)
	}
	if reconcileRequest.ExpectedRevision != 2 {
		t.Fatalf("reconcile expected revision = %d, want 2 (derived from directory enable)", reconcileRequest.ExpectedRevision)
	}
	if server.revision != 2 || server.observedAt != 2 {
		t.Fatalf("server revision = %d, observed = %d, want 2 and 2", server.revision, server.observedAt)
	}
	if receipt.FailedStep != "" {
		t.Fatalf("failed step = %q, want none", receipt.FailedStep)
	}
	if receipt.Apply == nil || receipt.Apply.Revision != 1 {
		t.Fatalf("apply receipt = %+v", receipt.Apply)
	}
	if receipt.Directory == nil || receipt.Directory.Revision != 2 {
		t.Fatalf("directory receipt = %+v", receipt.Directory)
	}
	if receipt.Reconcile == nil || receipt.Reconcile.Revision != 2 {
		t.Fatalf("reconcile receipt = %+v", receipt.Reconcile)
	}
	if receipt.Status == nil || !receipt.Status.Reconciled {
		t.Fatalf("status = %+v, want reconciled", receipt.Status)
	}
}

// TestAuthProjectProvisionRealmResumesWithSameBaseKeyAfterLateFailure proves the
// per-step keys make a retry resume: a status failure after every mutating step
// succeeded is retried with the same base key, the mutating steps replay instead of
// writing again, and the environment revision does not move.
func TestAuthProjectProvisionRealmResumesWithSameBaseKeyAfterLateFailure(t *testing.T) {
	server := newRealmWireServer("project-1", "environment-1")
	server.statusFailures = 1
	client := newAdminDoerClient(t, server)
	options := AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: "realm-provision-2"}

	first, err := client.Admin.AuthProjects.ProvisionRealm(
		context.Background(), "project-1", "environment-1", realmProvisionRequest(), options,
	)
	if err == nil {
		t.Fatal("first ProvisionRealm returned nil error on a failing status step")
	}
	if first == nil || first.FailedStep != AuthProjectRealmStepStatus {
		t.Fatalf("first receipt = %+v, want failed step status", first)
	}
	if server.revision != 2 {
		t.Fatalf("server revision after first call = %d, want 2", server.revision)
	}

	second, err := client.Admin.AuthProjects.ProvisionRealm(
		context.Background(), "project-1", "environment-1", realmProvisionRequest(), options,
	)
	if err != nil {
		t.Fatalf("retry ProvisionRealm returned error: %v", err)
	}
	if second.FailedStep != "" {
		t.Fatalf("retry failed step = %q, want none", second.FailedStep)
	}
	if second.Status == nil || !second.Status.Reconciled {
		t.Fatalf("retry status = %+v, want reconciled", second.Status)
	}
	// Two calls, four steps each, and the revision never moved past the first run.
	if len(server.requests) != 8 {
		t.Fatalf("requests = %d, want 8", len(server.requests))
	}
	if server.revision != 2 {
		t.Fatalf("server revision after retry = %d, want 2 (no duplicate writes)", server.revision)
	}
}

// TestAuthProjectProvisionRealmFailsWhenNotConverged refuses to report success for a
// realm whose provider registrations are not known to describe the current revision.
func TestAuthProjectProvisionRealmFailsWhenNotConverged(t *testing.T) {
	server := newRealmWireServer("project-1", "environment-1")
	server.skipObservation = true
	client := newAdminDoerClient(t, server)

	receipt, err := client.Admin.AuthProjects.ProvisionRealm(
		context.Background(), "project-1", "environment-1", realmProvisionRequest(),
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: "realm-provision-3"},
	)
	if err == nil {
		t.Fatal("ProvisionRealm returned nil error for a non-converged realm")
	}
	if !errors.Is(err, ErrAuthProjectRealmNotConverged) {
		t.Fatalf("error = %v, want ErrAuthProjectRealmNotConverged", err)
	}
	if receipt == nil || receipt.FailedStep != AuthProjectRealmStepStatus {
		t.Fatalf("receipt = %+v, want failed step status", receipt)
	}
	if receipt.Status == nil || receipt.Status.Reconciled {
		t.Fatalf("status = %+v, want not reconciled", receipt.Status)
	}
}

// TestAuthProjectProvisionRealmRejectsOutOfBoundsBaseKey keeps every derived key
// inside the server's 8..128 bound: a base shorter than the minimum, or one whose
// longest suffix would exceed the maximum, is refused before a request is sent.
func TestAuthProjectProvisionRealmRejectsOutOfBoundsBaseKey(t *testing.T) {
	for name, key := range map[string]string{
		"too short": strings.Repeat("k", 7),
		"too long":  strings.Repeat("k", 112),
	} {
		t.Run(name, func(t *testing.T) {
			server := newRealmWireServer("project-1", "environment-1")
			client := newAdminDoerClient(t, server)

			_, err := client.Admin.AuthProjects.ProvisionRealm(
				context.Background(), "project-1", "environment-1", realmProvisionRequest(),
				AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: key},
			)
			if err == nil {
				t.Fatal("ProvisionRealm accepted an out-of-bounds base key")
			}
			if len(server.requests) != 0 {
				t.Fatalf("a request was sent for an out-of-bounds base key: %d", len(server.requests))
			}
		})
	}
}

// TestAuthProjectProvisionRealmAcceptsBoundaryBaseKeys proves the longest accepted
// base keeps every derived key at or below the server maximum.
func TestAuthProjectProvisionRealmAcceptsBoundaryBaseKeys(t *testing.T) {
	server := newRealmWireServer("project-1", "environment-1")
	client := newAdminDoerClient(t, server)
	base := strings.Repeat("k", 111)

	if _, err := client.Admin.AuthProjects.ProvisionRealm(
		context.Background(), "project-1", "environment-1", realmProvisionRequest(),
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: base},
	); err != nil {
		t.Fatalf("ProvisionRealm rejected the maximum base key: %v", err)
	}
	if len(server.requests) != 4 {
		t.Fatalf("provision made %d requests, want 4", len(server.requests))
	}
	for index, request := range server.requests {
		key := request.Headers["Idempotency-Key"]
		if key == "" {
			continue
		}
		if len(key) < 8 || len(key) > 128 {
			t.Fatalf("request %d key length = %d, want 8..128", index, len(key))
		}
	}
}

func realmRequestPaths(server *realmWireServer) []string {
	paths := make([]string, 0, len(server.requests))
	for _, request := range server.requests {
		paths = append(paths, request.URL)
	}
	return paths
}
