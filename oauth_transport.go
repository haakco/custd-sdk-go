package custd

import (
	"context"
	"net/http"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// clientCredentialsTransport keeps token reuse while deriving refreshes from
// the actual request context. oauthCfg.Client(background) otherwise detaches
// token acquisition from the deadline of the operation needing the token.
type clientCredentialsTransport struct {
	config clientcredentials.Config
	mu     sync.Mutex
	token  *oauth2.Token
}

func (t *clientCredentialsTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	t.mu.Lock()
	token := t.token
	t.mu.Unlock()
	if !token.Valid() {
		// Refresh outside the mutex: another request must not wait beyond its own
		// deadline behind a slow token exchange. Concurrent refreshes are harmless.
		ctx := context.WithValue(request.Context(), oauth2.HTTPClient, &http.Client{
			Timeout:       defaultHTTPTimeout,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
		})
		refreshed, err := t.config.Token(ctx)
		if err != nil {
			return nil, err
		}
		token = refreshed
		t.mu.Lock()
		t.token = token
		t.mu.Unlock()
	}
	transport := oauth2.Transport{Source: oauth2.StaticTokenSource(token), Base: http.DefaultTransport}
	return transport.RoundTrip(request)
}
