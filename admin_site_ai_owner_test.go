package custd

import (
	"context"
	"net/http"
	"testing"
	"time"
)

type scriptedDoer struct {
	requests  []*fakeRequest
	responses []*fakeResponse
}

func (d *scriptedDoer) Do(req *fakeRequest) (*fakeResponse, error) {
	d.requests = append(d.requests, req)
	response := d.responses[0]
	d.responses = d.responses[1:]
	return response, nil
}

// TestSiteAIOwnerSessionRedeemsThenWritesWithBinding proves the browser-session
// contract: a form redeem from the hub origin, the cookie on every call, and a
// write that first reads the binding and then sends it with Custd's origin.
func TestSiteAIOwnerSessionRedeemsThenWritesWithBinding(t *testing.T) {
	doer := &scriptedDoer{responses: []*fakeResponse{
		{StatusCode: http.StatusSeeOther, Headers: map[string]string{
			"Set-Cookie": "__Host-custd-site-ai-owner=handle1; Path=/; Secure; HttpOnly; SameSite=Strict",
		}},
		{StatusCode: http.StatusOK, Body: []byte(`{"tenantSlug":"t1","sessionBinding":"bind1","settings":{"aiMode":"off","monthlyCapUsdMicros":0,"keyConfigured":false}}`)},
		{StatusCode: http.StatusOK, Body: []byte(`{"aiMode":"platform","monthlyCapUsdMicros":5,"keyConfigured":false}`)},
	}}
	client := NewClient(&ClientConfig{BaseURL: "https://api.custd.test", APIKey: "admin-token", HTTPClient: fakeTransport(doer), FlushInterval: time.Hour})
	t.Cleanup(func() { _ = client.Close(context.Background()) })

	session, err := client.RedeemSiteAIOwnerHandoff(context.Background(),
		SiteAIOwnerHandoff{Code: "code1", RedeemURL: "https://custd.test/api/v1/site-ai-owner/redeem"}, "https://hub.test")
	if err != nil {
		t.Fatalf("redeem: %v", err)
	}
	settings, err := session.SetSettings(context.Background(), SiteAISettingsRequest{Mode: SiteAIPlatform, MonthlyCapUSDMicros: 5})
	if err != nil || settings.Mode != SiteAIPlatform {
		t.Fatalf("set settings = %+v, %v", settings, err)
	}
	redeem, read, write := doer.requests[0], doer.requests[1], doer.requests[2]
	if string(redeem.Body) != "code=code1" || redeem.Header("Origin") != "https://hub.test" ||
		redeem.Header("Content-Type") != "application/x-www-form-urlencoded" || redeem.Header("Authorization") != "" {
		t.Fatalf("redeem request = %+v", redeem)
	}
	if read.Method != http.MethodGet || read.URL != "https://custd.test/api/v1/site-ai-owner/settings" ||
		read.Header("Cookie") != "__Host-custd-site-ai-owner=handle1" {
		t.Fatalf("settings read = %+v", read)
	}
	if write.Method != http.MethodPut || write.Header("Origin") != "https://custd.test" ||
		write.Header("X-Custd-Site-AI-Session") != "bind1" || write.Header("Cookie") != read.Header("Cookie") {
		t.Fatalf("settings write = %+v", write)
	}
}
