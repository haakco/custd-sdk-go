package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

// newSurfaceClient returns a client whose every answer carries a request
// identifier, so a test can prove the identifier reaches the caller.
func newSurfaceClient(t *testing.T, status int, body string) (*CustdClient, *captureDoer) {
	t.Helper()
	doer := newCaptureDoer(status, body)
	doer.headers = map[string]string{requestIDHeader: "req-123"}
	return newAdminTestClient(t, doer, "http://localhost:8080"), doer
}

// TestNewSurfacesSendTheirDocumentedRoute is the parity guard. Each row is the
// operation the CLI performs today; a route change here means a SDK consumer
// and custdctl have silently diverged.
func TestNewSurfacesSendTheirDocumentedRoute(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		call   func(*CustdClient) error
		method string
		path   string
	}{
		{
			name: "support access create",
			body: `{"supportAccessGrantUuid":"01957abc-0000-7000-8000-000000000001"}`,
			call: func(c *CustdClient) error {
				_, _, err := c.Admin.SupportAccess.Create(context.Background(),
					SupportAccessGrantCreateRequest{CompanySlug: "acme", Reason: "incident", ExpiresAt: "2026-01-01T00:00:00Z"})
				return err
			},
			method: http.MethodPost, path: "/api/v1/admin/support-access",
		},
		{
			name: "support access revoke",
			body: `{}`,
			call: func(c *CustdClient) error {
				_, _, err := c.Admin.SupportAccess.Revoke(context.Background(), "grant-1")
				return err
			},
			method: http.MethodPost, path: "/api/v1/admin/support-access/grant-1/revoke",
		},
		{
			name: "impersonation terminate",
			body: `{}`,
			call: func(c *CustdClient) error {
				_, _, err := c.Admin.Impersonation.Terminate(context.Background(), "session-1")
				return err
			},
			method: http.MethodPost, path: "/api/v1/admin/impersonation-sessions/session-1/terminate",
		},
		{
			name: "webhook delete is bodyless",
			body: ``,
			call: func(c *CustdClient) error {
				_, err := c.Admin.Webhooks.Delete(context.Background(), "acme", "hook-1")
				return err
			},
			method: http.MethodDelete, path: "/api/v1/admin/notification-webhooks/hook-1",
		},
		{
			name: "environment policy update",
			body: `{"autoCreateEnvironments":true}`,
			call: func(c *CustdClient) error {
				_, _, err := c.Admin.EnvironmentPolicy.Update(context.Background(), "",
					EnvironmentPolicyUpdateRequest{})
				return err
			},
			method: http.MethodPatch, path: "/api/v1/admin/environment-policy",
		},
		{
			name: "admin user reset password",
			body: ``,
			call: func(c *CustdClient) error {
				_, _, err := c.Admin.AdminUsers.ResetPassword(context.Background(), "user-1")
				return err
			},
			method: http.MethodPost, path: "/api/v1/admin/users/user-1/reset-password",
		},
		{
			name: "scheduler pause",
			body: ``,
			call: func(c *CustdClient) error {
				_, err := c.Admin.System.PauseScheduledTask(context.Background(), "rollup")
				return err
			},
			method: http.MethodPost, path: "/api/v1/admin/scheduled-tasks/rollup/pause",
		},
		{
			name: "system status",
			body: `{"overallStatus":"healthy"}`,
			call: func(c *CustdClient) error {
				_, _, err := c.Admin.System.Status(context.Background())
				return err
			},
			method: http.MethodGet, path: "/api/v1/admin/system/status",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, doer := newSurfaceClient(t, http.StatusOK, test.body)
			if err := test.call(client); err != nil {
				t.Fatalf("call returned error: %v", err)
			}
			if len(doer.requests) != 1 {
				t.Fatalf("requests = %d, want 1", len(doer.requests))
			}
			request := doer.requests[0]
			if request.Method != test.method {
				t.Fatalf("method = %s, want %s", request.Method, test.method)
			}
			if !strings.Contains(request.URL, test.path) {
				t.Fatalf("url = %s, want it to contain %s", request.URL, test.path)
			}
		})
	}
}

// TestSurfacesReportTheServerRequestID proves the identifier reaches the caller
// in the two shapes that have no body of their own. A support path is useless
// if the client cannot quote the request the server logged.
func TestSurfacesReportTheServerRequestID(t *testing.T) {
	client, _ := newSurfaceClient(t, http.StatusOK, `{"status":"ok"}`)

	_, metadata, err := client.Admin.AdminUsers.ResetPassword(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("ResetPassword returned error: %v", err)
	}
	if metadata.RequestID != "req-123" {
		t.Fatalf("requestId = %q, want req-123", metadata.RequestID)
	}

	status, _, err := client.Admin.System.Status(context.Background())
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if status.OverallStatus != "" && status.OverallStatus != "healthy" {
		t.Fatalf("overallStatus = %q", status.OverallStatus)
	}
}

// TestTenantManifestApplyRefusesWithoutIdempotencyKey proves the client refuses
// locally what the server would refuse anyway: a reconciling request with no
// key can be replayed twice and apply the same intent twice.
func TestTenantManifestApplyRefusesWithoutIdempotencyKey(t *testing.T) {
	client, doer := newSurfaceClient(t, http.StatusOK, `{}`)

	if _, _, err := client.Admin.TenantManifest.Apply(
		context.Background(), "acme", "  ", []byte(`{}`),
	); err == nil {
		t.Fatal("Apply accepted an empty idempotency key")
	}
	if len(doer.requests) != 0 {
		t.Fatalf("requests = %d, want 0: the key is checked before the call", len(doer.requests))
	}
}

// TestTenantManifestApplySendsIdempotencyKey proves the key travels as the
// header the server reads.
func TestTenantManifestApplySendsIdempotencyKey(t *testing.T) {
	client, doer := newSurfaceClient(t, http.StatusOK, `{"tenantSlug":"acme"}`)

	if _, _, err := client.Admin.TenantManifest.Apply(
		context.Background(), "acme", "01957abc-0000-7000-8000-0000000000ff", []byte(`{"version":1}`),
	); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if len(doer.requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(doer.requests))
	}
	if got := doer.requests[0].Headers["Idempotency-Key"]; got != "01957abc-0000-7000-8000-0000000000ff" {
		t.Fatalf("Idempotency-Key = %q", got)
	}
}

// TestDeliveryStatusCarriesTheFailureDetail proves the polling surface carries what a
// caller needs to decide between waiting, retrying and reporting. Dropping these four
// fields turns a diagnosable delivery into a bare state string.
func TestDeliveryStatusCarriesTheFailureDetail(t *testing.T) {
	client, _ := newSurfaceClient(t, http.StatusOK, `{
		"messageUuid": "msg-1",
		"deliveryUuid": "del-1",
		"templateId": "welcome",
		"channel": "email",
		"state": "failed",
		"attemptCount": 3,
		"lastErrorClass": "transport",
		"failureReason": "connection refused",
		"nextAction": "retry"
	}`)

	status, _, err := client.Admin.Webhooks.GetDelivery(context.Background(), "acme", "del-1")
	if err != nil {
		t.Fatalf("GetDelivery returned error: %v", err)
	}
	if status.AttemptCount != 3 || status.LastErrorClass != "transport" ||
		status.FailureReason != "connection refused" || status.NextAction != "retry" {
		t.Fatalf("status = %#v, want the failure detail carried", status)
	}
}

// TestAdminUserRemovalSendsTheActionAndReason proves the removal surface models the
// operation the server has: a named action on a tenant-scoped route, with the reason
// that lands in the tenant's audit stream.
func TestAdminUserRemovalSendsTheActionAndReason(t *testing.T) {
	client, doer := newSurfaceClient(t, http.StatusOK, `{"userUuid":"user-1"}`)

	if _, _, err := client.Admin.AdminUsers.Remove(
		context.Background(), "user-1", AdminUserRemovalSoftDelete, "acme", "left the company",
	); err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}
	request := doer.requests[0]
	if !strings.Contains(request.URL, "/users/user-1/soft-delete") {
		t.Fatalf("url = %s, want the soft-delete route", request.URL)
	}
	if !strings.Contains(request.URL, "companySlug=acme") {
		t.Fatalf("url = %s, want the tenant query", request.URL)
	}
	if !strings.Contains(string(request.Body), "left the company") {
		t.Fatalf("body = %s, want the reason carried", string(request.Body))
	}
}

// TestAdminUserRemovalRefusesAnUnknownAction proves the client refuses locally what the
// server would refuse anyway, rather than posting a route that does not exist.
func TestAdminUserRemovalRefusesAnUnknownAction(t *testing.T) {
	client, doer := newSurfaceClient(t, http.StatusOK, `{}`)

	if _, _, err := client.Admin.AdminUsers.Remove(
		context.Background(), "user-1", AdminUserRemoval("delete"), "acme", "x",
	); err == nil {
		t.Fatal("Remove accepted an unknown removal action")
	}
	if len(doer.requests) != 0 {
		t.Fatalf("requests = %d, want 0", len(doer.requests))
	}
}

// TestActingCompanyIsSentOnAdminRequests proves a system administrator can work on a
// tenant's behalf: the server reads this header to decide which company the request
// acts for, and without it the CLI's cross-tenant commands cannot be expressed.
func TestActingCompanyIsSentOnAdminRequests(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"scopes":[]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	client.config.ActingCompanySlug = "awthy"

	if _, err := client.Admin.DataScopes.ListScopes(context.Background(), false); err != nil {
		t.Fatalf("ListScopes returned error: %v", err)
	}
	if got := doer.requests[0].Headers[actingCompanyHeader]; got != "awthy" {
		t.Fatalf("%s = %q, want awthy", actingCompanyHeader, got)
	}
}

// TestActingCompanyIsOmittedWhenUnset proves an ordinary tenant command does not claim to
// act for anyone: an empty setting must leave the header off entirely.
func TestActingCompanyIsOmittedWhenUnset(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"scopes":[]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	if _, err := client.Admin.DataScopes.ListScopes(context.Background(), false); err != nil {
		t.Fatalf("ListScopes returned error: %v", err)
	}
	if got, present := doer.requests[0].Headers[actingCompanyHeader]; present {
		t.Fatalf("%s = %q, want it absent", actingCompanyHeader, got)
	}
}

// TestMetadataRecorderCapturesTheRequestID proves a caller can quote the request the
// server logged without any method returning metadata: the recorder travels on the
// context, so none of the 171 signatures change.
func TestMetadataRecorderCapturesTheRequestID(t *testing.T) {
	client, _ := newSurfaceClient(t, http.StatusOK, `{"scopes":[]}`)

	var recorder MetadataRecorder
	ctx := WithMetadataRecorder(context.Background(), &recorder)
	if _, err := client.Admin.DataScopes.ListScopes(ctx, false); err != nil {
		t.Fatalf("ListScopes returned error: %v", err)
	}
	if recorder.RequestID != "req-123" {
		t.Fatalf("requestId = %q, want req-123", recorder.RequestID)
	}
}

// TestAdminUserListSendsTheFilters proves the list surface carries the route the server
// has: a named collection, narrowed by tenant, optionally including ended memberships.
func TestAdminUserListSendsTheFilters(t *testing.T) {
	client, doer := newSurfaceClient(t, http.StatusOK, `{"users":[{"userUuid":"user-1","email":"a@b.c"}]}`)

	page, err := fetchAdminUserPage(client, "acme", true)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(page.Users) != 1 || page.Users[0].UserUUID != "user-1" {
		t.Fatalf("page = %#v", page)
	}
	request := doer.requests[0]
	if request.Method != http.MethodGet {
		t.Fatalf("method = %s, want GET", request.Method)
	}
	for _, want := range []string{"/users", "companySlug=acme", "includeEnded=true"} {
		if !strings.Contains(request.URL, want) {
			t.Fatalf("url = %s, want it to contain %s", request.URL, want)
		}
	}
}

// fetchAdminUserPage keeps the assertion above inside the statement budget.
func fetchAdminUserPage(
	client *CustdClient, company string, includeEnded bool,
) (*AdminUserListResponse, error) {
	return client.Admin.AdminUsers.List(context.Background(), company, includeEnded)
}

// TestActingCompanyReachesTheRouteNotJustTheHeader proves the tenant travels where the
// route actually reads it. Admin routes resolve the tenant from ?companySlug= and
// validate it, so sending only the header made a system administrator's cross-tenant
// call fail with "valid companySlug is required" — a bug the stubbed tests could not see
// because a stub accepts any request.
func TestActingCompanyReachesTheRouteNotJustTheHeader(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"scopes":[]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	client.config.ActingCompanySlug = "awthy"

	if _, err := client.Admin.DataScopes.ListScopes(context.Background(), false); err != nil {
		t.Fatalf("ListScopes returned error: %v", err)
	}
	request := doer.requests[0]
	if got := request.Headers[actingCompanyHeader]; got != "awthy" {
		t.Fatalf("%s = %q, want awthy", actingCompanyHeader, got)
	}
	if !strings.Contains(request.URL, "companySlug=awthy") {
		t.Fatalf("url = %s, want the tenant query parameter", request.URL)
	}
}

// TestActingCompanyDoesNotDuplicateAnExistingQuery proves a call that already names a
// company keeps its own value rather than gaining a second one.
func TestActingCompanyDoesNotDuplicateAnExistingQuery(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"grants":[]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	client.config.ActingCompanySlug = "awthy"

	if _, _, err := client.Admin.SupportAccess.List(context.Background()); err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if got := strings.Count(doer.requests[0].URL, "companySlug="); got != 1 {
		t.Fatalf("url = %s, want exactly one companySlug", doer.requests[0].URL)
	}
}

// TestAdminUserDisableSendsReasonAndScope proves the disable action carries the reason
// that lands in the tenant's audit stream and the globalUser scope that decides whether
// the identity or one membership is affected.
func TestAdminUserDisableSendsReasonAndScope(t *testing.T) {
	client, doer := newSurfaceClient(t, http.StatusOK, `{}`)

	if _, err := client.Admin.AdminUsers.Disable(
		context.Background(), "user-1", "left the company", true,
	); err != nil {
		t.Fatalf("Disable returned error: %v", err)
	}
	request := doer.requests[0]
	for _, want := range []string{"/users/user-1/disable", "globalUser=true"} {
		if !strings.Contains(request.URL, want) {
			t.Fatalf("url = %s, want it to contain %s", request.URL, want)
		}
	}
	if !strings.Contains(string(request.Body), "left the company") {
		t.Fatalf("body = %s, want the reason", string(request.Body))
	}
}

// TestAdminUserEnableSendsTheRoles proves a membership is never reopened without saying
// what authority it carries.
func TestAdminUserEnableSendsTheRoles(t *testing.T) {
	client, doer := newSurfaceClient(t, http.StatusOK, `{}`)

	if _, err := client.Admin.AdminUsers.Enable(
		context.Background(), "user-1", []string{"tenant_admin"}, false,
	); err != nil {
		t.Fatalf("Enable returned error: %v", err)
	}
	request := doer.requests[0]
	if strings.Contains(request.URL, "globalUser=") {
		t.Fatalf("url = %s, want no globalUser for a membership enable", request.URL)
	}
	if !strings.Contains(string(request.Body), "tenant_admin") {
		t.Fatalf("body = %s, want the roles", string(request.Body))
	}
}

// TestPasswordResetReportsTheNotificationNotAPassword proves the reset surface models
// what the server actually answers. It never returns a credential; it reports the
// one-time message it queued and the deliveries it made.
func TestPasswordResetReportsTheNotificationNotAPassword(t *testing.T) {
	client, _ := newSurfaceClient(t, http.StatusOK,
		`{"messageUuid":"msg-1","deliveryUuids":["del-1","del-2"]}`)

	reset, _, err := client.Admin.AdminUsers.ResetPassword(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("ResetPassword returned error: %v", err)
	}
	if reset.MessageUUID != "msg-1" || len(reset.DeliveryUUIDs) != 2 {
		t.Fatalf("reset = %#v, want the queued notification", reset)
	}
}

// TestRolesCarryTheSlugACallerSubmits proves the role list reports the value a caller
// sends, not a field the server never returns.
func TestRolesCarryTheSlugACallerSubmits(t *testing.T) {
	client, _ := newSurfaceClient(t, http.StatusOK,
		`{"roles":[{"slug":"tenant_admin","name":"Tenant admin","description":"Manages one tenant"}]}`)

	roles, _, err := client.Admin.AdminUsers.ListRoles(context.Background())
	if err != nil {
		t.Fatalf("ListRoles returned error: %v", err)
	}
	if len(roles.Roles) != 1 || roles.Roles[0].Slug != "tenant_admin" || roles.Roles[0].Name != "Tenant admin" {
		t.Fatalf("roles = %#v", roles)
	}
}

// TestSiteUpdateSendsEveryRestriction proves a site update always states its
// AllowedEnvironments, because an omitted field must never silently widen a restriction,
// and carries the environment the CLI names.
func TestSiteUpdateSendsEveryRestriction(t *testing.T) {
	client, doer := newSurfaceClient(t, http.StatusOK, `{"siteUuid":"site-1","environment":"prod"}`)

	site, err := client.Admin.Sites.Update(context.Background(), "site-1", AdminSiteUpdate{
		IdentityMode: "anonymous", Environment: "prod", AllowedOrigins: []string{"https://a.test"},
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if site.Environment != "prod" {
		t.Fatalf("site = %#v, want the environment carried", site)
	}
	request := doer.requests[0]
	if request.Method != http.MethodPatch || !strings.HasSuffix(request.URL, "/sites/site-1") {
		t.Fatalf("request = %s %s", request.Method, request.URL)
	}
	var body map[string]any
	if err := json.Unmarshal(request.Body, &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["environment"] != "prod" || body["identityMode"] != "anonymous" {
		t.Fatalf("body = %#v", body)
	}
	if _, present := body["allowedEnvironments"]; !present {
		t.Fatalf("body = %#v, want allowedEnvironments stated", body)
	}
}

// TestAuditLabelsReachTheServer proves a CLI call can still be attributed. The server
// credits `source: custdctl` and the command name only when both labels arrive, so
// dropping them during a client migration silently downgrades every operator action to
// unattributed.
func TestAuditLabelsReachTheServer(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	client.config.ClientLabel = "custdctl"
	client.config.CommandLabel = "sites update"

	if err := client.Admin.Sites.Delete(context.Background(), "site-1"); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	headers := doer.requests[0].Headers
	if headers[clientLabelHeader] != "custdctl" || headers[commandLabelHeader] != "sites update" {
		t.Fatalf("headers = %#v, want both audit labels", headers)
	}
}

// TestAuditLabelsAreOmittedWhenUnset proves a non-CLI caller never claims to be one.
func TestAuditLabelsAreOmittedWhenUnset(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	if err := client.Admin.Sites.Delete(context.Background(), "site-1"); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	headers := doer.requests[0].Headers
	if _, present := headers[clientLabelHeader]; present {
		t.Fatalf("client label present: %#v", headers)
	}
}

// TestOAuthMutationsCarryTheIdempotencyKey proves the retry identity reaches the header.
// The server refuses creation and secret rotation without one, and validates that it is a
// canonical UUIDv7, so a client that silently omits it cannot be used to create anything.
func TestOAuthMutationsCarryTheIdempotencyKey(t *testing.T) {
	key, err := NewIdempotencyKey()
	if err != nil {
		t.Fatalf("NewIdempotencyKey: %v", err)
	}
	// A canonical UUIDv7 is 36 characters and has version 7 and the RFC 4122 variant.
	if len(key) != 36 || key[14] != '7' {
		t.Fatalf("key = %q, want a canonical UUIDv7", key)
	}

	doer := newCaptureDoer(http.StatusOK, `{"clientId":"c","clientSecret":"s"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	options := OAuthClientRequestOptions{IdempotencyKey: key}

	if _, err := client.Admin.OAuthClients.Create(
		context.Background(), AdminOAuthClientCreate{ClientID: "c"}, options,
	); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := client.Admin.OAuthClients.RotateSecret(context.Background(), "c", options); err != nil {
		t.Fatalf("RotateSecret returned error: %v", err)
	}
	for _, request := range doer.requests {
		if request.Headers["Idempotency-Key"] != key {
			t.Fatalf("%s headers = %#v, want the idempotency key", request.URL, request.Headers)
		}
	}
}

// TestOAuthMutationRefusesAnEmptyKey proves a misuse fails at the caller with one clear
// error instead of becoming a service 400.
func TestOAuthMutationRefusesAnEmptyKey(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	if _, err := client.Admin.OAuthClients.Create(
		context.Background(), AdminOAuthClientCreate{ClientID: "c"}, OAuthClientRequestOptions{},
	); err == nil {
		t.Fatal("Create accepted an empty idempotency key")
	}
	if len(doer.requests) != 0 {
		t.Fatalf("a refused call still made %d requests", len(doer.requests))
	}
}
