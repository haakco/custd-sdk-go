package custd

import (
	"context"
	"net/http"
	"testing"
)

const appsURL = "http://localhost:8080/api/v1/admin/apps"

func TestAppOperationsSendTheirRoutes(t *testing.T) {
	options := AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: "create-done-when"}
	summary := `{"appId":"p1","slug":"done-when","name":"Done When","environmentId":"e1","host":"done-when.apps.example",` +
		`"url":"https://done-when.apps.example","status":"needs_attention","reason":{"code":"sign_in_not_ready","message":"m"},"signIn":true}`
	tests := []struct {
		name, body, method, url, key string
		call                         func(*AuthProjectAdminClient) error
	}{
		{
			name: "create", body: `{"app":` + summary + `,"replayed":false,"trafficReporting":"ready"}`,
			method: http.MethodPost, url: appsURL, key: "create-done-when",
			call: func(c *AuthProjectAdminClient) error {
				created, err := c.CreateApp(context.Background(), AppCreateRequest{Name: "Done When", SignIn: true}, options)
				if err == nil && (created.App.Status != AppStatusNeedsAttention || created.App.Reason.Code != AppReasonSignInNotReady ||
					created.TrafficReporting != AppTrafficReportingReady) {
					t.Fatalf("created = %+v", created)
				}
				return err
			},
		},
		{
			name: "list", body: `{"apps":[` + summary + `],"nextAfter":"p1"}`,
			method: http.MethodGet, url: appsURL + "?after=p0",
			call: func(c *AuthProjectAdminClient) error {
				list, err := c.ListApps(context.Background(), "p0", options)
				if err == nil && (len(list.Apps) != 1 || list.NextAfter != "p1") {
					t.Fatalf("list = %+v", list)
				}
				return err
			},
		},
		{
			name: "get", body: `{"app":` + summary + `,"latestRelease":{"releaseId":"r1","state":"verified","active":true,` +
				`"label":"v1","fileCount":2,"totalBytes":10,"createdAt":"2026-11-17T00:00:00Z"},"releaseCount":1,"ownerLinks":2,"shareLinks":3}`,
			method: http.MethodGet, url: appsURL + "/done%20when",
			call: func(c *AuthProjectAdminClient) error {
				overview, err := c.GetApp(context.Background(), "done when", options)
				if err == nil && (overview.LatestRelease.Label != "v1" || overview.ShareLinks != 3) {
					t.Fatalf("overview = %+v", overview)
				}
				return err
			},
		},
		{
			name: "deploy credential create", body: `{"clientId":"app-deploy-1","clientSecret":"s","app":"done-when","scope":"apps.releases.write"}`,
			method: http.MethodPost, url: appsURL + "/done-when/deploy-credentials",
			call: func(c *AuthProjectAdminClient) error {
				credential, err := c.CreateAppDeployCredential(context.Background(), "done-when", options)
				if err == nil && credential.ClientSecret != "s" {
					t.Fatalf("credential = %+v", credential)
				}
				return err
			},
		},
		{
			name: "deploy credential delete", method: http.MethodDelete, url: appsURL + "/done-when/deploy-credentials/app-deploy-1",
			call: func(c *AuthProjectAdminClient) error {
				return c.DeleteAppDeployCredential(context.Background(), "done-when", "app-deploy-1", options)
			},
		},
		{
			name: "list handoff credentials", body: `{"credentials":[{"clientId":"seller-app"}]}`,
			method: http.MethodGet, url: appsURL + "/shop/purchase-handoff-credentials",
			call: func(c *AuthProjectAdminClient) error {
				credentials, err := c.ListPurchaseHandoffCredentials(t.Context(), "shop", options)
				if err == nil && (len(credentials.Credentials) != 1 || credentials.Credentials[0].ClientID != "seller-app") {
					t.Fatal("handoff credential list did not preserve the client ID")
				}
				return err
			},
		},
		{
			name: "create handoff credential", body: `{"clientId":"seller-app","clientSecret":"fixture","app":"shop","scope":"billing.purchase_handoffs"}`,
			method: http.MethodPost, url: appsURL + "/shop/purchase-handoff-credentials",
			call: func(c *AuthProjectAdminClient) error {
				credential, err := c.CreatePurchaseHandoffCredential(t.Context(), "shop", options)
				if err == nil && (credential.ClientSecret != "fixture" || credential.Scope != "billing.purchase_handoffs") {
					t.Fatal("handoff credential did not preserve its receipt")
				}
				return err
			},
		},
		{
			name:   "delete handoff credential",
			method: http.MethodDelete, url: appsURL + "/shop/purchase-handoff-credentials/seller-app",
			call: func(c *AuthProjectAdminClient) error {
				return c.DeletePurchaseHandoffCredential(t.Context(), "shop", "seller-app", options)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, test.body)
			client := newAdminTestClient(t, doer, "http://localhost:8080")
			if err := test.call(client.Admin.AuthProjects); err != nil {
				t.Fatalf("%s returned error: %v", test.name, err)
			}
			req := doer.requests[0]
			if req.Method != test.method || req.URL != test.url {
				t.Fatalf("request = %s %s, want %s %s", req.Method, req.URL, test.method, test.url)
			}
			if req.Header(authProjectOwningUserHeader) != authProjectOwningUser {
				t.Fatalf("owning user header = %q", req.Header(authProjectOwningUserHeader))
			}
			if test.key != "" && req.Header("Idempotency-Key") != test.key {
				t.Fatalf("idempotency key = %q, want %q", req.Header("Idempotency-Key"), test.key)
			}
		})
	}
}

func TestCreateAppRequiresAnIdempotencyKey(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	if _, err := client.Admin.AuthProjects.CreateApp(context.Background(), AppCreateRequest{Name: "Done When"},
		AuthProjectRequestOptions{}); err == nil || len(doer.requests) != 0 {
		t.Fatalf("CreateApp without a key: err = %v, requests = %d", err, len(doer.requests))
	}
}
