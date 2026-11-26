package custd

import (
	"context"
	"net/http"
	"testing"
)

// An item call carries only the app's Site key, its origin and the caller's
// link and user tokens: never the client's own API credential, and never a
// token in the URL.
func TestAppItemSendsOnlyAppCredentials(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"item":{"itemUuid":"i1"}}`)
	items, err := NewAppItemClient(newAdminTestClient(t, doer, "https://api.example"),
		AppItemClientConfig{SiteKey: "site_pk_1", Origin: "https://app.example"})
	if err != nil {
		t.Fatal(err)
	}
	access := AppItemAccess{LinkToken: "link-token", UserToken: "user-token", Item: "i2"}
	if _, err := items.Get(context.Background(), access); err != nil {
		t.Fatal(err)
	}
	request := doer.requests[0]
	if request.URL != "https://api.example/api/v1/app/item?item=i2" {
		t.Fatalf("url = %s", request.URL)
	}
	want := map[string]string{
		"Authorization": "Bearer link-token", AppItemUserTokenHeader: "user-token",
		AppItemSiteKeyHeader: "site_pk_1", "Origin": "https://app.example",
	}
	for name, value := range want {
		if request.Header(name) != value {
			t.Fatalf("%s = %q, want %q", name, request.Header(name), value)
		}
	}
	if _, err := NewAppItemClient(newAdminTestClient(t, doer, "https://api.example"), AppItemClientConfig{}); err == nil {
		t.Fatal("a client without a site key and origin must be refused")
	}
}
