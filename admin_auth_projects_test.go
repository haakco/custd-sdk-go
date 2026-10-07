package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

const authProjectOwningUser = "01957abc-0000-7000-8000-0000000000aa"

func assertAuthProjectRequest(
	t *testing.T,
	req *HTTPRequest,
	method string,
	url string,
	owningUser string,
	idempotencyKey string,
) {
	t.Helper()
	if req.Method != method {
		t.Fatalf("method = %s, want %s", req.Method, method)
	}
	if req.URL != url {
		t.Fatalf("url = %s, want %s", req.URL, url)
	}
	if req.Headers["Authorization"] != "Bearer admin-token" {
		t.Fatalf("authorization header = %q", req.Headers["Authorization"])
	}
	if req.Headers[authProjectOwningUserHeader] != owningUser {
		t.Fatalf("owning user header = %q, want %q", req.Headers[authProjectOwningUserHeader], owningUser)
	}
	if req.Headers["Idempotency-Key"] != idempotencyKey {
		t.Fatalf("idempotency key header = %q, want %q", req.Headers["Idempotency-Key"], idempotencyKey)
	}
}

func decodeAuthProjectBody(t *testing.T, req *HTTPRequest) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(req.Body, &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return body
}

func TestAuthProjectsCreateProjectPinsRequestAndDecodesCreation(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"project": {
			"projectId": "01957abc-0000-7000-8000-000000000001",
			"slug": "hosting-eu",
			"name": "Hosting EU",
			"environmentId": "01957abc-0000-7000-8000-000000000002",
			"environmentSlug": "production",
			"identityMode": "isolated"
		},
		"operationId": "op-1",
		"replayed": false,
		"runtimeReady": true
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	creation, err := client.Admin.AuthProjects.CreateProject(
		context.Background(),
		AuthProjectCreateRequest{
			Slug:            "hosting-eu",
			Name:            "Hosting EU",
			EnvironmentSlug: "production",
			IdentityMode:    AuthProjectIdentityIsolated,
		},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: "idem-1"},
	)
	if err != nil {
		t.Fatalf("CreateProject returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodPost,
		"http://localhost:8080/api/v1/admin/auth-projects", authProjectOwningUser, "idem-1")
	body := decodeAuthProjectBody(t, doer.requests[0])
	if body["slug"] != "hosting-eu" || body["name"] != "Hosting EU" ||
		body["environmentSlug"] != "production" || body["identityMode"] != "isolated" {
		t.Fatalf("body = %+v", body)
	}
	if creation.OperationID != "op-1" || creation.Replayed || !creation.RuntimeReady {
		t.Fatalf("creation = %+v", creation)
	}
	if creation.Project.ProjectID != "01957abc-0000-7000-8000-000000000001" ||
		creation.Project.EnvironmentID != "01957abc-0000-7000-8000-000000000002" ||
		creation.Project.IdentityMode != AuthProjectIdentityIsolated {
		t.Fatalf("project = %+v", creation.Project)
	}
}

func TestAuthProjectsListProjectsSendsCursorAndOwningUser(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"projects": [{
			"projectId": "01957abc-0000-7000-8000-000000000001",
			"slug": "hosting-eu",
			"name": "Hosting EU",
			"environmentId": "01957abc-0000-7000-8000-000000000002",
			"environmentSlug": "production",
			"identityMode": "shared"
		}],
		"nextAfter": "01957abc-0000-7000-8000-000000000003"
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	page, err := client.Admin.AuthProjects.ListProjects(
		context.Background(),
		"prev-cursor",
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err != nil {
		t.Fatalf("ListProjects returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodGet,
		"http://localhost:8080/api/v1/admin/auth-projects?after=prev-cursor", authProjectOwningUser, "")
	if len(page.Projects) != 1 || page.Projects[0].Slug != "hosting-eu" {
		t.Fatalf("projects = %+v", page.Projects)
	}
	if page.Projects[0].IdentityMode != AuthProjectIdentityShared {
		t.Fatalf("identity mode = %q", page.Projects[0].IdentityMode)
	}
	if page.NextAfter != "01957abc-0000-7000-8000-000000000003" {
		t.Fatalf("nextAfter = %q", page.NextAfter)
	}
}

func TestAuthProjectsCreateEnvironmentPinsRequest(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"projectId": "01957abc-0000-7000-8000-000000000001",
		"slug": "hosting-eu",
		"name": "Hosting EU",
		"environmentId": "01957abc-0000-7000-8000-000000000004",
		"environmentSlug": "staging",
		"identityMode": "isolated"
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	environment, err := client.Admin.AuthProjects.CreateEnvironment(
		context.Background(),
		"01957abc-0000-7000-8000-000000000001",
		AuthProjectEnvironmentCreateRequest{Slug: "staging", Name: "Staging", IdentityMode: AuthProjectIdentityIsolated},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err != nil {
		t.Fatalf("CreateEnvironment returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodPost,
		"http://localhost:8080/api/v1/admin/auth-projects/01957abc-0000-7000-8000-000000000001/environments",
		authProjectOwningUser, "")
	body := decodeAuthProjectBody(t, doer.requests[0])
	if body["slug"] != "staging" || body["name"] != "Staging" || body["identityMode"] != "isolated" {
		t.Fatalf("body = %+v", body)
	}
	if environment.EnvironmentSlug != "staging" {
		t.Fatalf("environment = %+v", environment)
	}
}

func TestAuthProjectsListPrincipalSessionsDecodesInventory(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"projectId": "project-1",
		"environmentId": "environment-1",
		"directoryId": "directory-1",
		"principalId": "principal-1",
		"sessions": [{
			"sessionId": "session-1",
			"active": true,
			"authenticatedAt": "2026-10-01T00:00:00Z",
			"authenticatorAssuranceLevel": "aal1",
			"expiresAt": "2026-10-01T00:04:30Z",
			"issuedAt": "2026-10-01T00:00:00Z"
		}]
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	inventory, err := client.Admin.AuthProjects.ListPrincipalSessions(
		context.Background(), "project-1", "environment-1", "directory-1", "subject-1",
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err != nil {
		t.Fatalf("ListPrincipalSessions returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodGet,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1"+
			"/directories/directory-1/principals/subject-1/sessions",
		authProjectOwningUser, "")
	if inventory.PrincipalID != "principal-1" || len(inventory.Sessions) != 1 {
		t.Fatalf("inventory = %+v", inventory)
	}
	if !inventory.Sessions[0].Active || inventory.Sessions[0].SessionID != "session-1" {
		t.Fatalf("sessions = %+v", inventory.Sessions)
	}
}

func TestAuthProjectsRevokePrincipalSessionPinsRequest(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"projectId": "project-1",
		"directoryId": "directory-1",
		"principalId": "principal-1",
		"revoked": 1,
		"sessionId": "session-1"
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	revocation, err := client.Admin.AuthProjects.RevokePrincipalSession(
		context.Background(), "project-1", "environment-1", "directory-1", "subject-1",
		ApplicationSessionRevokeRequest{SessionID: "session-1"},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: "idem-2"},
	)
	if err != nil {
		t.Fatalf("RevokePrincipalSession returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodPost,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1"+
			"/directories/directory-1/principals/subject-1/sessions/revoke",
		authProjectOwningUser, "idem-2")
	if body := decodeAuthProjectBody(t, doer.requests[0]); body["sessionId"] != "session-1" {
		t.Fatalf("body = %+v", body)
	}
	if revocation.Revoked != 1 || revocation.SessionID != "session-1" {
		t.Fatalf("revocation = %+v", revocation)
	}
}

func TestAuthProjectsRevokeAllSessionsSendsExplicitConfirmation(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"projectId": "project-1",
		"directoryId": "directory-1",
		"principalId": "principal-1",
		"revoked": 3
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	revocation, err := client.Admin.AuthProjects.RevokePrincipalSessions(
		context.Background(), "project-1", "environment-1", "directory-1", "subject-1",
		ApplicationSessionsRevokeAllRequest{Confirm: true},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: "idem-3"},
	)
	if err != nil {
		t.Fatalf("RevokePrincipalSessions returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodPost,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1"+
			"/directories/directory-1/principals/subject-1/sessions/revoke-all",
		authProjectOwningUser, "idem-3")
	if body := decodeAuthProjectBody(t, doer.requests[0]); body["confirm"] != true {
		t.Fatalf("body = %+v", body)
	}
	if revocation.Revoked != 3 {
		t.Fatalf("revocation = %+v", revocation)
	}
}

func TestAuthProjectsRevokeMembershipPinsRequest(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"projectId": "project-1",
		"environmentId": "environment-1",
		"directoryId": "directory-1",
		"principalId": "principal-1",
		"organisationId": "organisation-1",
		"removedAt": "2026-10-01T00:00:00Z",
		"removed": true
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	revocation, err := client.Admin.AuthProjects.RevokePrincipalMembership(
		context.Background(), "project-1", "environment-1", "directory-1", "subject-1",
		ApplicationMembershipRevokeRequest{OrganisationID: "organisation-1"},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err != nil {
		t.Fatalf("RevokePrincipalMembership returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodPost,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1"+
			"/directories/directory-1/principals/subject-1/memberships/revoke",
		authProjectOwningUser, "")
	body := decodeAuthProjectBody(t, doer.requests[0])
	if body["organisationId"] != "organisation-1" {
		t.Fatalf("body = %+v", body)
	}
	if !revocation.Removed || revocation.OrganisationID != "organisation-1" {
		t.Fatalf("revocation = %+v", revocation)
	}
}

func TestAuthProjectsRejectMissingIdempotencyKeyBeforeSending(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	_, err := client.Admin.AuthProjects.CreateProject(
		context.Background(),
		AuthProjectCreateRequest{Slug: "hosting-eu", Name: "Hosting EU", EnvironmentSlug: "production"},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err == nil {
		t.Fatal("CreateProject returned nil error without an idempotency key")
	}
	if len(doer.requests) != 0 {
		t.Fatalf("a request was sent without an idempotency key: %d", len(doer.requests))
	}
}

func TestAuthProjectsOmitOwningUserHeaderForHumanAdministrator(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"projects":[]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	if _, err := client.Admin.AuthProjects.ListProjects(
		context.Background(), "", AuthProjectRequestOptions{},
	); err != nil {
		t.Fatalf("ListProjects returned error: %v", err)
	}

	req := doer.requests[0]
	if req.URL != "http://localhost:8080/api/v1/admin/auth-projects" {
		t.Fatalf("url = %s", req.URL)
	}
	if _, present := req.Headers[authProjectOwningUserHeader]; present {
		t.Fatalf("owning user header was sent for a human administrator: %q",
			req.Headers[authProjectOwningUserHeader])
	}
}

// authProjectDesiredState is the nested desired-state document a consumer
// builds from the named types alone. Apply and preview share it.
func authProjectDesiredState() AuthProjectDesiredState {
	return AuthProjectDesiredState{
		IdentityMode:       AuthProjectIdentityIsolated,
		RegistrationPolicy: "invite_only",
		LoginPaused:        false,
		Audiences: []AuthProjectAudienceBinding{{
			Audience:               "hosting-edge",
			PublicClient:           true,
			RedirectURIs:           []string{"https://app.example.com/callback"},
			PostLogoutRedirectURIs: []string{"https://app.example.com/logout"},
			AllowedOrigins:         []string{"https://app.example.com"},
		}},
		ProfileFields: []AuthProjectProfileField{{
			Key:                  "contact_email",
			Required:             true,
			VisibleToApplication: true,
			EditableBy:           "user",
		}},
	}
}

func TestAuthProjectsApplyEnvironmentDesiredStatePinsRequestAndDecodesOperation(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"id": "op-1",
		"kind": "project_auth_config_update",
		"status": "applied",
		"revision": 7,
		"idempotencyKey": "idem-apply",
		"appliedAt": "2026-10-07T00:00:00Z",
		"replayed": false
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	request := AuthProjectDesiredStateRequest{DesiredState: authProjectDesiredState(), ExpectedRevision: 6}
	operation, err := client.Admin.AuthProjects.ApplyEnvironmentDesiredState(
		context.Background(), "project-1", "environment-1", request,
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: "idem-apply"},
	)
	if err != nil {
		t.Fatalf("ApplyEnvironmentDesiredState returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodPost,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/apply",
		authProjectOwningUser, "idem-apply")
	var sent AuthProjectDesiredStateRequest
	if err := json.Unmarshal(doer.requests[0].Body, &sent); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !reflect.DeepEqual(sent, request) {
		t.Fatalf("body = %+v, want %+v", sent, request)
	}
	if operation.ID != "op-1" || operation.Kind != "project_auth_config_update" ||
		operation.Revision != 7 || operation.Replayed {
		t.Fatalf("operation = %+v", operation)
	}
}

func TestAuthProjectsPreviewEnvironmentDesiredStateDecodesChanges(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"projectId": "project-1",
		"environmentId": "environment-1",
		"revision": 7,
		"changes": [{"field": "desired.audiences", "before": "[]", "after": "[hosting-edge]"}],
		"sideEffects": ["client_registration"],
		"noOp": false
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	preview, err := client.Admin.AuthProjects.PreviewEnvironmentDesiredState(
		context.Background(), "project-1", "environment-1",
		AuthProjectDesiredStateRequest{DesiredState: authProjectDesiredState(), ExpectedRevision: 6},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err != nil {
		t.Fatalf("PreviewEnvironmentDesiredState returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodPost,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/preview",
		authProjectOwningUser, "")
	if body := decodeAuthProjectBody(t, doer.requests[0]); body["expectedRevision"] != float64(6) {
		t.Fatalf("body = %+v", body)
	}
	if preview.ProjectID != "project-1" || preview.EnvironmentID != "environment-1" || preview.NoOp {
		t.Fatalf("preview = %+v", preview)
	}
	if len(preview.Changes) != 1 || preview.Changes[0].Field != "desired.audiences" ||
		preview.Changes[0].After != "[hosting-edge]" {
		t.Fatalf("changes = %+v", preview.Changes)
	}
	if len(preview.SideEffects) != 1 || preview.SideEffects[0] != "client_registration" {
		t.Fatalf("side effects = %+v", preview.SideEffects)
	}
}

func TestAuthProjectsGetEnvironmentStatusReadsBackAudienceBinding(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"projectId": "project-1",
		"environmentId": "environment-1",
		"revision": 7,
		"desired": {
			"identityMode": "isolated",
			"registrationPolicy": "invite_only",
			"loginPaused": false,
			"audiences": [{
				"audience": "hosting-edge",
				"publicClient": true,
				"redirectUris": ["https://app.example.com/callback"],
				"postLogoutRedirectUris": null,
				"allowedOrigins": null
			}],
			"profileFields": null
		},
		"reconciled": true,
		"reconcileNote": "",
		"loginReady": false,
		"loginNote": "",
		"clientSync": {
			"clientIds": ["custd-app-environment-1-hosting-edge"],
			"revision": 7,
			"checkedAt": "2026-10-07T00:00:00Z",
			"current": true,
			"registrations": [{
				"audience": "hosting-edge",
				"clientId": "custd-app-environment-1-hosting-edge",
				"observed": true
			}]
		}
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	status, err := client.Admin.AuthProjects.GetEnvironmentStatus(
		context.Background(), "project-1", "environment-1",
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err != nil {
		t.Fatalf("GetEnvironmentStatus returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodGet,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/status",
		authProjectOwningUser, "")
	if status.ProjectID != "project-1" || status.EnvironmentID != "environment-1" || !status.Reconciled {
		t.Fatalf("status = %+v", status)
	}
	if len(status.Desired.Audiences) != 1 || status.Desired.Audiences[0].Audience != "hosting-edge" ||
		!status.Desired.Audiences[0].PublicClient {
		t.Fatalf("audience binding = %+v", status.Desired.Audiences)
	}
	if status.ClientSync == nil || len(status.ClientSync.Registrations) != 1 ||
		status.ClientSync.Registrations[0].ClientID != "custd-app-environment-1-hosting-edge" {
		t.Fatalf("client sync = %+v", status.ClientSync)
	}
}

func TestAuthProjectsApplyRejectsMissingIdempotencyKeyBeforeSending(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	_, err := client.Admin.AuthProjects.ApplyEnvironmentDesiredState(
		context.Background(), "project-1", "environment-1",
		AuthProjectDesiredStateRequest{DesiredState: authProjectDesiredState(), ExpectedRevision: 6},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err == nil {
		t.Fatal("ApplyEnvironmentDesiredState returned nil error without an idempotency key")
	}
	if len(doer.requests) != 0 {
		t.Fatalf("a request was sent without an idempotency key: %d", len(doer.requests))
	}
}
