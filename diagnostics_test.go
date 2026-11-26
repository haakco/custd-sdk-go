package custd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Health is the root path, not a path under /api/v1, and it carries no credential: its whole
// purpose is to tell "unreachable" apart from "not authorized".
func TestDiagnosticsGetHealthUsesTheRootAndSendsNoCredential(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"status":"ok"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	if _, _, err := client.Admin.GetHealth(context.Background()); err != nil {
		t.Fatalf("GetHealth: %v", err)
	}
	request := doer.requests[0]
	if request.URL != "http://localhost:8080/health" {
		t.Fatalf("url = %s, want the root health path", request.URL)
	}
	if _, present := request.Headers[http.CanonicalHeaderKey("Authorization")]; present {
		t.Fatal("the liveness probe must not send a credential")
	}
}

// The account route is under /api/v1 but not under /api/v1/admin, so it is built explicitly.
func TestDiagnosticsGetAccountUsesTheAccountRoute(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"userUuid":"u1","companySlug":"acme","companyName":"Acme",`+
		`"email":"a@acme.test","principalKind":"human","roles":[],"permissions":[],"scopes":["admin"]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	restore := unmarshalResponse
	unmarshalResponse = unmarshalStrict
	defer func() { unmarshalResponse = restore }()

	account, _, err := client.Admin.GetAccount(context.Background())
	if err != nil {
		t.Fatalf("GetAccount: %v", err)
	}
	if account.UserUUID != "u1" || account.CompanyName != "Acme" || account.PrincipalKind != AccountPrincipalHuman ||
		len(account.Scopes) != 1 || account.Scopes[0] != "admin" {
		t.Fatalf("account = %+v, want every field the ingest-api account handler writes", account)
	}
	if url := doer.requests[0].URL; url != "http://localhost:8080/api/v1/account/me" {
		t.Fatalf("url = %s, want the account route", url)
	}
}

// A discovery document is untrusted input. An endpoint on another host would let it redirect
// the signing-key fetch to keys of its own choosing, so the whole origin is compared.
func TestIssuerEndpointMustStayOnTheIssuerOrigin(t *testing.T) {
	const issuer = "https://auth.example.com"
	for _, testCase := range []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{"same origin", "https://auth.example.com/keys", false},
		{"different port", "https://auth.example.com:8443/keys", true},
		{"different host", "https://evil.example.net/keys", true},
		{"same host different scheme", "http://auth.example.com/keys", true},
		{"subdomain", "https://tenant.auth.example.com/keys", true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := trustedIssuerEndpoint(testCase.raw, issuer)
			if testCase.wantErr && err == nil {
				t.Fatalf("%s was accepted, want refusal", testCase.raw)
			}
			if !testCase.wantErr && err != nil {
				t.Fatalf("%s was refused: %v", testCase.raw, err)
			}
		})
	}
}

// An issuer may publish its identifier with a trailing slash (Ory Hydra does); the requested
// issuer is compared without one, so the slash alone must not fail the inspection.
func TestInspectIssuerAcceptsAnIssuerPublishedWithATrailingSlash(t *testing.T) {
	server := httptest.NewServer(nil)
	t.Cleanup(server.Close)
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/oauth-authorization-server":
			_, _ = w.Write([]byte(`{"issuer":"` + server.URL + `/","jwks_uri":"` + server.URL + `/keys"}`))
		default:
			_, _ = w.Write([]byte(`{"keys":[{"kty":"RSA","kid":"k1","n":"AQAB","e":"AQAB"}]}`))
		}
	})
	client := NewClient(&ClientConfig{BaseURL: server.URL})

	if _, err := client.Admin.InspectIssuer(context.Background(), server.URL); err != nil {
		t.Fatalf("InspectIssuer: %v", err)
	}
}
