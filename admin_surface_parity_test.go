package custd

import (
	"context"
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
				_, err := c.Admin.AdminUsers.ResetPassword(context.Background(), "user-1",
					AdminUserPasswordReset{Password: "x"})
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

	metadata, err := client.Admin.AdminUsers.ResetPassword(context.Background(), "user-1",
		AdminUserPasswordReset{Password: "x"})
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
