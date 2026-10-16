package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestDataScopeCRUDUsesAdminRoutes(t *testing.T) {
	scopeFixture := `{"uuid":"11111111-1111-1111-1111-111111111111","code":"acme-store","name":"Acme Store",
		"status":"active","createdAt":"2026-10-15T00:00:00Z","updatedAt":"2026-10-15T00:00:00Z"}`

	doer := newCaptureDoer(http.StatusCreated, scopeFixture)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	scope, err := client.Admin.DataScopes.CreateScope(context.Background(), DataScopeCreateRequest{
		Code: "acme-store", Name: "Acme Store",
	})
	if err != nil {
		t.Fatal(err)
	}
	if doer.requests[0].Method != http.MethodPost || !strings.HasSuffix(doer.requests[0].URL, "/api/v1/admin/data-scopes") {
		t.Fatalf("create request = %s %s", doer.requests[0].Method, doer.requests[0].URL)
	}
	if scope.UUID != "11111111-1111-1111-1111-111111111111" || scope.Status != DataScopeStatusActive {
		t.Fatalf("scope = %+v", scope)
	}
}

func TestDataScopeListCarriesIncludeDisabled(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"scopes":[]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	list, err := client.Admin.DataScopes.ListScopes(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}
	if got := doer.requests[0].URL; !strings.HasSuffix(got, "/api/v1/admin/data-scopes?includeDisabled=true") {
		t.Fatalf("list request URL = %s", got)
	}
	if list.Scopes == nil || len(list.Scopes) != 0 {
		t.Fatalf("scopes = %+v", list.Scopes)
	}
}

func TestDataScopeDisableAndGrantsUseNoContentRoutes(t *testing.T) {
	uuid := "11111111-1111-1111-1111-111111111111"
	doer := newCaptureDoer(http.StatusNoContent, "")
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	if err := client.Admin.DataScopes.DisableScope(context.Background(), uuid); err != nil {
		t.Fatal(err)
	}
	if req := doer.requests[0]; req.Method != http.MethodPost || !strings.HasSuffix(req.URL, "/api/v1/admin/data-scopes/"+uuid+"/disable") {
		t.Fatalf("disable request = %s %s", req.Method, req.URL)
	}

	grant := DataScopeGrant{
		Principal: DataScopePrincipalRef{Kind: DataScopePrincipalMachine, ClientID: "custd-producer"},
		Target:    DataScopeGrantTarget{Kind: DataScopeGrantTargetSubClient, ScopeUUID: uuid},
		Action:    DataScopeActionRead,
	}
	grantDoer := newCaptureDoer(http.StatusNoContent, "")
	grantClient := newAdminTestClient(t, grantDoer, "http://localhost:8080")
	if err := grantClient.Admin.DataScopes.RevokeGrant(context.Background(), grant); err != nil {
		t.Fatal(err)
	}
	if req := grantDoer.requests[0]; req.Method != http.MethodPost || !strings.HasSuffix(req.URL, "/api/v1/admin/data-scope-grants/revoke") {
		t.Fatalf("revoke request = %s %s", req.Method, req.URL)
	}
}

func TestDataScopeListGrantsSendsPrincipalQuery(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"grants":[]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	_, err := client.Admin.DataScopes.ListGrants(context.Background(), DataScopePrincipalRef{
		Kind: DataScopePrincipalHuman, UserUUID: "22222222-2222-2222-2222-222222222222",
	})
	if err != nil {
		t.Fatal(err)
	}
	got := doer.requests[0].URL
	if !strings.Contains(got, "/api/v1/admin/data-scope-grants?") ||
		!strings.Contains(got, "kind=human") ||
		!strings.Contains(got, "userUuid=22222222-2222-2222-2222-222222222222") {
		t.Fatalf("grants request URL = %s", got)
	}
}

func TestDataScopeCreateGrantSendsClosedShape(t *testing.T) {
	grantFixture := `{"principal":{"kind":"machine","clientId":"custd-producer"},
		"target":{"kind":"resource","scopeUuid":"11111111-1111-1111-1111-111111111111",
		"resourceUuid":"33333333-3333-3333-3333-333333333333"},"action":"write"}`
	doer := newCaptureDoer(http.StatusCreated, grantFixture)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	grant, err := client.Admin.DataScopes.CreateGrant(context.Background(), DataScopeGrant{
		Principal: DataScopePrincipalRef{Kind: DataScopePrincipalMachine, ClientID: "custd-producer"},
		Target: DataScopeGrantTarget{
			Kind:         DataScopeGrantTargetResource,
			ScopeUUID:    "11111111-1111-1111-1111-111111111111",
			ResourceUUID: "33333333-3333-3333-3333-333333333333",
		},
		Action: DataScopeActionWrite,
	})
	if err != nil {
		t.Fatal(err)
	}
	if grant.Action != DataScopeActionWrite || grant.Target.Kind != DataScopeGrantTargetResource {
		t.Fatalf("grant = %+v", grant)
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(doer.requests[0].Body, &body); err != nil {
		t.Fatal(err)
	}
	// The grant carries the closed action and target-kind spellings, not a policy decision.
	if string(body["action"]) != `"write"` {
		t.Fatalf("body = %s", doer.requests[0].Body)
	}
}

func TestDataScopeClientValidatesBeforeSending(t *testing.T) {
	cases := []struct {
		name string
		call func(*DataScopeAdminClient) error
	}{
		{"bad scope code", func(c *DataScopeAdminClient) error {
			_, err := c.CreateScope(context.Background(), DataScopeCreateRequest{Code: "bad code", Name: "x"})
			return err
		}},
		{"empty update", func(c *DataScopeAdminClient) error {
			_, err := c.UpdateScope(context.Background(), "11111111-1111-1111-1111-111111111111", DataScopeUpdateRequest{})
			return err
		}},
		{"human with client id", func(c *DataScopeAdminClient) error {
			_, err := c.CreateGrant(context.Background(), DataScopeGrant{
				Principal: DataScopePrincipalRef{Kind: DataScopePrincipalHuman, UserUUID: "22222222-2222-2222-2222-222222222222", ClientID: "x"},
				Target:    DataScopeGrantTarget{Kind: DataScopeGrantTargetParent},
				Action:    DataScopeActionRead,
			})
			return err
		}},
		{"parent with scope", func(c *DataScopeAdminClient) error {
			_, err := c.CreateGrant(context.Background(), DataScopeGrant{
				Principal: DataScopePrincipalRef{Kind: DataScopePrincipalMachine, ClientID: "custd-producer"},
				Target:    DataScopeGrantTarget{Kind: DataScopeGrantTargetParent, ScopeUUID: "11111111-1111-1111-1111-111111111111"},
				Action:    DataScopeActionRead,
			})
			return err
		}},
		{"unknown action", func(c *DataScopeAdminClient) error {
			_, err := c.CreateGrant(context.Background(), DataScopeGrant{
				Principal: DataScopePrincipalRef{Kind: DataScopePrincipalMachine, ClientID: "custd-producer"},
				Target:    DataScopeGrantTarget{Kind: DataScopeGrantTargetParent},
				Action:    DataScopeAction("admin"),
			})
			return err
		}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, `{}`)
			client := newAdminTestClient(t, doer, "http://localhost:8080")
			if err := testCase.call(client.Admin.DataScopes); err == nil {
				t.Fatal("expected a validation error")
			}
			if len(doer.requests) != 0 {
				t.Fatalf("validation should not cost a request, sent %d", len(doer.requests))
			}
		})
	}
}
