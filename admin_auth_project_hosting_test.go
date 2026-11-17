package custd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAuthProjectDesiredStateOmitsAnUnsetHostingSection(t *testing.T) {
	encoded, err := json.Marshal(authProjectDesiredState())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(encoded), `"hosting":`) {
		t.Fatalf("an unset hosting section must be omitted so the stored one is kept: %s", encoded)
	}
}

func TestAuthProjectDesiredStateSendsTheHostingSection(t *testing.T) {
	desired := authProjectDesiredState()
	desired.Hosting = &AuthProjectHosting{
		Mode:         AuthProjectHostingEnabled,
		Capabilities: []AuthProjectHostingCapability{AuthProjectHostingTracking, AuthProjectHostingSignIn},
		IngestTarget: &AuthProjectHostingIngestTarget{ScopeUUID: "scope-1", ResourceUUID: "resource-1"},
	}
	encoded, err := json.Marshal(desired)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `"hosting":{"mode":"enabled","capabilities":["tracking","sign_in"],` +
		`"ingestTarget":{"scopeUuid":"scope-1","resourceUuid":"resource-1"}}`
	if !strings.Contains(string(encoded), want) {
		t.Fatalf("desired state = %s\nwant it to contain %s", encoded, want)
	}
}

func TestAuthProjectStatusDecodesTheHostedServingIdentity(t *testing.T) {
	var status AuthProjectStatus
	body := `{"projectId":"p","environmentId":"e","hosting":{"host":"shop.apps.example","origin":"https://shop.apps.example",` +
		`"signIn":{"clientId":"c1","audience":"shop","redirectUri":"https://shop.apps.example/auth/callback"}}}`
	if err := json.Unmarshal([]byte(body), &status); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	hosting := status.Hosting
	if hosting == nil || hosting.Host != "shop.apps.example" || hosting.Origin != "https://shop.apps.example" ||
		hosting.SignIn == nil || hosting.SignIn.ClientID != "c1" || hosting.SignIn.Audience != "shop" ||
		hosting.SignIn.RedirectURI != "https://shop.apps.example/auth/callback" {
		t.Fatalf("hosting = %+v", hosting)
	}
}
