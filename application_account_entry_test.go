package custd

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestAppAccountEntryBindsExactIssuerAndAppOrigin(t *testing.T) {
	action := "https://auth.canvas.test/settings/entry"
	ticket := strings.Repeat("A", 43)
	doer := newCaptureDoer(http.StatusOK, `{"actionUrl":"`+action+`","ticket":"`+ticket+`"}`)
	client := newAdminTestClient(t, doer, "https://api.custd.test")
	request := AppAccountEntryRequest{Issuer: "https://auth.canvas.test/", ApplicationOrigin: "https://canvas.test"}
	entry, err := client.CreateAppAccountEntry(t.Context(), "application-token", request)
	if err != nil || entry.ActionURL != action || entry.Ticket != ticket {
		t.Fatal("entry must preserve the selected issuer action and the opaque ticket")
	}
	if len(doer.requests) != 1 {
		t.Fatal("expected one entry ticket request")
	}
	wire := doer.requests[0]
	if wire.Method != http.MethodPost || wire.URL != action+"/ticket" ||
		wire.Header("Authorization") != "Bearer application-token" {
		t.Fatal("entry request must use the configured issuer and current application token")
	}
	var body struct {
		ApplicationOrigin string `json:"applicationOrigin"`
	}
	if err := json.Unmarshal(wire.Body, &body); err != nil || body.ApplicationOrigin != request.ApplicationOrigin {
		t.Fatal("entry request must carry the registered origin")
	}
}

func TestAppAccountEntryRejectsUntrustedAction(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"actionUrl":"https://other.test/settings/entry","ticket":"`+strings.Repeat("A", 43)+`"}`)
	client := newAdminTestClient(t, doer, "https://api.custd.test")
	entry, err := client.CreateAppAccountEntry(t.Context(), "application-token", AppAccountEntryRequest{
		Issuer: "https://auth.canvas.test", ApplicationOrigin: "https://canvas.test",
	})
	if err == nil || entry != nil {
		t.Fatal("untrusted form action must be refused")
	}
}
