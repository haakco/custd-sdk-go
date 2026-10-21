package custd

import (
	"context"
	"net/http"
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
	if _, present := request.Headers["Authorization"]; present {
		t.Fatal("the liveness probe must not send a credential")
	}
}

// The account route is under /api/v1 but not under /api/v1/admin, so it is built explicitly.
func TestDiagnosticsGetAccountUsesTheAccountRoute(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"companySlug":"acme"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	if _, _, err := client.Admin.GetAccount(context.Background()); err != nil {
		t.Fatalf("GetAccount: %v", err)
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
