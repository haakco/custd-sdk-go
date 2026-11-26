package custd

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

// readContractFixture loads a shared contract fixture, resolving both layouts it
// can run under: the standalone custd-sdk-go module (the release split vendors
// the fixtures at the module root) and the monorepo (the shared fixtures live
// one directory up from sdk-go).
func readContractFixture(t *testing.T, name string) []byte {
	t.Helper()
	candidates := []string{
		filepath.Join("contract-fixtures", name),       // standalone custd-sdk-go module
		filepath.Join("..", "contract-fixtures", name), // monorepo layout
	}
	for _, path := range candidates {
		if data, err := os.ReadFile(path); err == nil {
			return data
		}
	}
	t.Fatalf("contract fixture %s not found (looked in %v)", name, candidates)
	return nil
}

// fakeRequest is what a scripted fake sees of one SDK request: the encoded body and headers.
type fakeRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    []byte
}

// Header reads one request header the way HTTP does, ignoring the name's case.
func (r *fakeRequest) Header(name string) string {
	return r.Headers[http.CanonicalHeaderKey(name)]
}

// fakeResponse is a scripted whole answer.
type fakeResponse struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
}

// fakeDoer is the whole-body fake most tests script; fakeTransport adapts it to HTTPDoer.
type fakeDoer interface {
	Do(req *fakeRequest) (*fakeResponse, error)
}

type fakeTransportAdapter struct{ fake fakeDoer }

func fakeTransport(fake fakeDoer) HTTPDoer { return fakeTransportAdapter{fake: fake} }

func (t fakeTransportAdapter) Do(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		read, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		body = read
	}
	headers := make(map[string]string, len(req.Header))
	for name := range req.Header {
		headers[name] = req.Header.Get(name)
	}
	answer, err := t.fake.Do(&fakeRequest{Method: req.Method, URL: req.URL.String(), Headers: headers, Body: body})
	if err != nil {
		return nil, err
	}
	header := http.Header{}
	for name, value := range answer.Headers {
		header.Set(name, value)
	}
	return &http.Response{
		StatusCode: answer.StatusCode, Header: header, Body: io.NopCloser(bytes.NewReader(answer.Body)), Request: req,
	}, nil
}
