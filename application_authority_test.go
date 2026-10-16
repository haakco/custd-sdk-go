package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestApplicationAuthorityUsesOnlyApplicationTokenAndParsesDenial(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/oauth/authority-check" || r.Method != http.MethodPost {
			t.Errorf("unexpected authority route: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer application-token" || r.Header.Get(authProjectOwningUserHeader) != "" {
			t.Error("authority request must carry only the application credential")
		}
		var body ApplicationAuthorityRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.AccessToken != "application-token" {
			t.Error("authority request must use the camelCase application access token body")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"authorized":false,"reason":"authority_withdrawn"}`))
	}))
	defer server.Close()
	client := NewClient(&ClientConfig{
		BaseURL: server.URL, APIKey: "platform-token-must-not-leak",
		ClientID: "machine-client", ClientSecret: "machine-secret",
		TokenURL: server.URL + "/oauth/token", FlushInterval: time.Hour,
	})
	t.Cleanup(func() { _ = client.Close(context.Background()) })

	decision, err := client.CheckApplicationAuthority(context.Background(), ApplicationAuthorityRequest{AccessToken: "application-token"})
	if err != nil || decision.Authorized || decision.Reason != "authority_withdrawn" {
		t.Fatalf("HTTP-200 denial must remain a denial: decision=%+v err=%v", decision, err)
	}
	if calls.Load() != 1 {
		t.Fatalf("application authority must not mint a platform token: calls=%d", calls.Load())
	}
}

func TestApplicationAuthorityParsesPositiveEvidenceWithoutMachineCredential(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"authorized":true,"reason":"authorized"}`)
	client := NewClient(&ClientConfig{BaseURL: "http://localhost:8080", HTTPClient: doer, FlushInterval: time.Hour})
	t.Cleanup(func() { _ = client.Close(context.Background()) })
	decision, err := client.CheckApplicationAuthority(context.Background(), ApplicationAuthorityRequest{AccessToken: "application-token"})
	if err != nil || !decision.Authorized || len(doer.requests) != 1 {
		t.Fatalf("positive authority decision: decision=%+v err=%v", decision, err)
	}
	if doer.requests[0].Headers["Authorization"] != "Bearer application-token" {
		t.Error("custom transport must also receive the application credential, not platform auth")
	}
}

func TestApplicationAuthorityRejectsMalformedSuccess(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"reason":"authorized"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	decision, err := client.CheckApplicationAuthority(context.Background(), ApplicationAuthorityRequest{AccessToken: "application-token"})
	if err == nil || decision.Authorized {
		t.Fatal("a malformed HTTP success cannot become positive evidence")
	}
}

func TestApplicationAuthorityRejectsEmptyTokenBeforeTransport(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"authorized":true,"reason":"authorized"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	_, err := client.CheckApplicationAuthority(context.Background(), ApplicationAuthorityRequest{})
	if err == nil || len(doer.requests) != 0 {
		t.Fatal("empty application token must fail before any request")
	}
}
