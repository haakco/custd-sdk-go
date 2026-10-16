package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

func TestPrincipalAudiencesList(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"audienceEntitlementMode":"explicit_grants","audiences":[{"audience":"app-web","granted":true,"managed":true}]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	target := ApplicationPrincipalTarget{ProjectID: "project/1", EnvironmentID: "env/1", DirectoryID: "dir/1", ProviderSubject: "provider/subject"}
	got, err := client.Admin.AuthProjects.ListPrincipalAudiences(context.Background(), target, AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser})
	if err != nil {
		t.Fatal(err)
	}
	assertAuthProjectRequest(t, doer.requests[0], http.MethodGet, "http://localhost:8080/api/v1/admin/auth-projects/project%2F1/environments/env%2F1/principals/provider%2Fsubject/audiences?directoryId=dir%2F1", authProjectOwningUser, "")
	if len(doer.requests[0].Body) != 0 {
		t.Fatal("GET has body")
	}
	want := &PrincipalAudiences{AudienceEntitlementMode: "explicit_grants", Audiences: []PrincipalAudience{{Audience: "app-web", Granted: true, Managed: true}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v", got)
	}
}

func TestPrincipalAudienceGrantAndRevokeOnlyNamedAudience(t *testing.T) {
	for _, action := range []string{"grant", "revoke"} {
		t.Run(action, func(t *testing.T) {
			granted := action == "grant"
			response := AudienceGrant{PrincipalID: "principal-1", OrganisationID: "org-1", Audience: "app-web", Granted: granted, Managed: true}
			encoded, _ := json.Marshal(response)
			doer := newCaptureDoer(http.StatusOK, string(encoded))
			client := newAdminTestClient(t, doer, "http://localhost:8080")
			target := ApplicationPrincipalTarget{ProjectID: "project/1", EnvironmentID: "env/1", DirectoryID: "dir/1", ProviderSubject: "provider/subject"}
			body := AudienceGrantRequest{DirectoryID: "dir/1", Audience: "app-web", OrganisationID: "org-1"}
			method := client.Admin.AuthProjects.GrantPrincipalAudience
			if !granted {
				method = client.Admin.AuthProjects.RevokePrincipalAudience
			}
			for _, key := range []string{"", "   "} {
				if _, err := method(context.Background(), target, body, AuthProjectRequestOptions{IdempotencyKey: key}); err == nil {
					t.Fatal("accepted absent key")
				}
			}
			if len(doer.requests) != 0 {
				t.Fatal("missing key reached transport")
			}
			got, err := method(context.Background(), target, body, AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: "audience-key"})
			if err != nil {
				t.Fatal(err)
			}
			assertAuthProjectRequest(t, doer.requests[0], http.MethodPost, "http://localhost:8080/api/v1/admin/auth-projects/project%2F1/environments/env%2F1/principals/provider%2Fsubject/audiences/"+action, authProjectOwningUser, "audience-key")
			var payload map[string]any
			if err := json.Unmarshal(doer.requests[0].Body, &payload); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(payload, map[string]any{"directoryId": "dir/1", "audience": "app-web", "organisationId": "org-1"}) {
				t.Fatalf("payload %+v", payload)
			}
			if !reflect.DeepEqual(got, &response) {
				t.Fatalf("got %+v", got)
			}
		})
	}
}
