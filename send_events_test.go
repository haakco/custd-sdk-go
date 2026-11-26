package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSendEventsReturnsConfirmedPartialAcceptance(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Bearer relay-token" {
			t.Error("missing existing SDK auth")
		}
		var request SendEventsRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if len(request.Events) != 2 {
			t.Error("wrong batch")
		}
		w.WriteHeader(202)
		_, _ = w.Write([]byte(`{"success":false,"results":[{"eventUuid":"a1b2c3d4-e5f6-7890-abcd-ef1234567890","success":true,"status":202},{"eventUuid":"second-event","success":false,"status":422,"error":{"code":"schema-invalid"}}]}`))
	}))
	defer server.Close()
	client := NewClient(&ClientConfig{BaseURL: server.URL, APIKey: "relay-token"})
	defer func() { _ = client.Close(context.Background()) }()
	second := *validEvent()
	second.EventUUID = "second-event"
	result, err := client.SendEvents(context.Background(), SendEventsRequest{Events: []EventEnvelope{*validEvent(), second}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.AcceptedEventIDs) != 1 || result.AcceptedEventIDs[0] != validEvent().EventUUID || len(result.Results) != 2 || result.Results[1].Error.Code != "schema-invalid" {
		t.Fatalf("wrong partial result: %+v", result)
	}
	if err := client.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("relay sends must not queue/retry rejected items")
	}
}

func TestSendEventsRejectsUnprovenAcceptance(t *testing.T) {
	for _, body := range []string{``, `{"success":true}`, `{"success":true,"results":[]}`, `{"results":[{"eventUuid":"a1b2c3d4-e5f6-7890-abcd-ef1234567890","success":false,"status":422}]}`, `{"success":true,"results":[{"eventUuid":"unknown","success":true,"status":202}]}`, `{"success":true,"results":[{"eventUuid":"a1b2c3d4-e5f6-7890-abcd-ef1234567890","success":true,"status":500}]}`} {
		t.Run(body, func(t *testing.T) {
			doer := newCaptureDoer(202, body)
			client := NewClient(&ClientConfig{BaseURL: "https://custd.example.test", APIKey: "token", HTTPClient: fakeTransport(doer)})
			defer func() { _ = client.Close(context.Background()) }()
			result, err := client.SendEvents(context.Background(), SendEventsRequest{Events: []EventEnvelope{*validEvent()}})
			if err == nil || result != nil {
				t.Fatal("unproven events acknowledged")
			}
		})
	}
	doer := newCaptureDoer(202, `{"success":true}`)
	client := NewClient(&ClientConfig{BaseURL: "https://custd.example.test", APIKey: "token", HTTPClient: fakeTransport(doer)})
	defer func() { _ = client.Close(context.Background()) }()
	invalid := *validEvent()
	invalid.EventUUID = ""
	if _, err := client.SendEvents(context.Background(), SendEventsRequest{Events: []EventEnvelope{invalid}}); err == nil || !strings.Contains(err.Error(), "eventUuid") {
		t.Fatal("missing supplied retry identity accepted")
	}
}

func TestSendEventsDeadlineIncludesOAuthTokenExchange(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-time.After(time.Second):
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"synthetic-token","token_type":"Bearer","expires_in":3600}`))
	}))
	defer tokenServer.Close()
	client := NewClient(&ClientConfig{BaseURL: tokenServer.URL, ClientID: "synthetic-client", ClientSecret: "synthetic-secret", TokenURL: tokenServer.URL})
	defer func() { _ = client.Close(context.Background()) }()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	started := time.Now()
	result, err := client.SendEvents(ctx, SendEventsRequest{Events: []EventEnvelope{*validEvent()}})
	if err == nil || result != nil {
		t.Fatal("expired token exchange acknowledged events")
	}
	if time.Since(started) > 500*time.Millisecond {
		t.Fatal("token exchange ignored caller deadline")
	}
}
