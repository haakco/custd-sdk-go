package custd

import (
	"context"
	"net/http"
	"testing"
)

func TestNotificationWebhookCreateAddsASlackSubscription(t *testing.T) {
	doer := newCaptureDoer(http.StatusCreated, `{"uuid":"hook-1","name":"ops","kind":"slack",`+
		`"endpoint":"hooks.slack.com/services/T0/…","enabled":true,"signingSecret":""}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	created, _, err := client.Admin.Webhooks.Create(context.Background(), NotificationWebhookCreateRequest{
		Name: "ops", Kind: NotificationWebhookKindSlack, Endpoint: "https://hooks.slack.com/services/T0/B0/x", Enabled: true,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	want := `{"name":"ops","kind":"slack","endpoint":"https://hooks.slack.com/services/T0/B0/x","enabled":true}`
	if got := string(doer.requests[0].Body); got != want {
		t.Fatalf("body = %s\nwant  %s", got, want)
	}
	if created.UUID != "hook-1" || created.Kind != NotificationWebhookKindSlack {
		t.Fatalf("created = %+v", created)
	}
}
