package custd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAuthProjectDesiredStateCarriesSocialSignInOnlyWhenSet(t *testing.T) {
	desired := authProjectDesiredState()
	encoded, err := json.Marshal(desired)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(encoded), `"socialSignIn":`) {
		t.Fatalf("an unset social section must be omitted so the stored providers are kept: %s", encoded)
	}
	desired.SocialSignIn = &AuthProjectSocialSignIn{Providers: []AuthProjectSocialProvider{
		{Provider: "google", Mode: AuthProjectSocialModeShared},
		{Provider: "github", Mode: AuthProjectSocialModeOwn, ClientID: "Iv1.app"},
	}}
	encoded, err = json.Marshal(desired)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `"socialSignIn":{"providers":[{"provider":"google","mode":"shared"},{"provider":"github","mode":"own","clientId":"Iv1.app"}]}`
	if !strings.Contains(string(encoded), want) {
		t.Fatalf("desired state = %s\nwant it to contain %s (an omitted secret keeps the stored one)", encoded, want)
	}
}
