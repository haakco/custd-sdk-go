package custd

import (
	"context"
	"net/http"
	"testing"
)

func TestPrincipalExportPreservesAllIdentityTraitsAndCompleteness(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"projectId":"project/1","environmentId":"environment/1","directoryId":"directory/1","principalId":"principal-1","enabled":true,"createdAt":"2026-10-08T00:00:00Z","identities":[{"providerIssuer":"store-one","providerSubject":"subject-one","traitsStatus":"included","traits":[{"name":"verified","value":true},{"name":"preferences","value":{"nested":["plain",null,false,2]}}]},{"providerIssuer":"store-two","providerSubject":"subject-two","traitsStatus":"incomplete","traits":null}],"memberships":[],"profileValues":[]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	target := ApplicationPrincipalTarget{ProjectID: "project/1", EnvironmentID: "environment/1", DirectoryID: "directory/1", ProviderSubject: "subject/1"}
	exported, err := client.Admin.AuthProjects.ExportPrincipal(context.Background(), target, AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser})
	if err != nil {
		t.Fatal(err)
	}
	assertAuthProjectRequest(t, doer.requests[0], http.MethodGet, "http://localhost:8080/api/v1/admin/auth-projects/project%2F1/environments/environment%2F1/directories/directory%2F1/principals/subject%2F1/export", authProjectOwningUser, "")
	if len(exported.Identities) != 2 || exported.Identities[0].TraitsStatus != ApplicationIdentityTraitsIncluded || exported.Identities[1].TraitsStatus != ApplicationIdentityTraitsIncomplete {
		t.Fatalf("identity completeness = %+v", exported.Identities)
	}
	traits := exported.Identities[0].Traits
	if len(traits) != 2 || string(traits[0].Value) != "true" || string(traits[1].Value) != `{"nested":["plain",null,false,2]}` {
		t.Fatalf("JSON traits = %+v", traits)
	}
	if exported.Identities[1].Traits != nil {
		t.Fatal("incomplete traits must not be presented as included")
	}
}
