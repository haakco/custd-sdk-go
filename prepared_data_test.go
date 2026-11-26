package custd

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

const (
	preparedDataBootstrapBody = `{"tenantSlug":"acme","state":"ready","observedAt":"2026-07-18T12:00:00Z",` +
		`"provenance":{"owner":"identity-bootstrap"},"retryability":"none","nextAction":{"action":"none"}}`
	preparedDataReceiptBody = `{"receiptUuid":"11111111-1111-4111-8111-111111111111","tenantSlug":"acme",` +
		`"processingState":"accepted","availability":"complete","observedAt":"2026-07-18T12:00:00Z",` +
		`"provenance":{"owner":"segments"},"retryability":"none","nextAction":{"action":"none"}}`
	preparedDataOutputBody = `{"outputUuid":"22222222-2222-4222-8222-222222222222","tenantSlug":"acme",` +
		`"processingState":"ready","availability":"complete","observedAt":"2026-07-18T12:00:00Z",` +
		`"provenance":{"owner":"rollups"},"retryability":"none","nextAction":{"action":"none"}}`
	preparedDataOutputListBody = `{"outputs":[` + preparedDataOutputBody + `]}`
	preparedDataQueryBody      = `{"output":` + preparedDataOutputBody + `,"data":{"buckets":[{"date":"2026-07-18",` +
		`"value":{"value":1,"unit":"count","sampleCount":1,"dataSufficiency":"complete","complete":true},` +
		`"source":"prepared","queryDurationMs":5}],"value":{"value":1,"unit":"count","sampleCount":1,` +
		`"dataSufficiency":"complete","complete":true},"queryDurationMs":5,"snapshotAgeMs":0,"eventLagP95Ms":0}}`
)

// newPreparedDataTestClient builds a prepared-data client over a capturing
// transport so a test can assert the exact method, path and body sent.
func newPreparedDataTestClient(t *testing.T, doer *captureDoer) *PreparedDataClient {
	t.Helper()
	client := newAdminTestClient(t, doer, "http://localhost:8080/")
	return NewPreparedDataClient(client.Admin)
}

// assertPreparedDataRequest asserts the method, path and body of the one
// request the client made.
func assertPreparedDataRequest(
	t *testing.T, doer *captureDoer, wantMethod, wantPath, wantBody string,
) {
	t.Helper()
	if len(doer.requests) != 1 {
		t.Fatalf("requests = %d, want 1", len(doer.requests))
	}
	request := doer.requests[0]
	if request.Method != wantMethod {
		t.Fatalf("method = %s, want %s", request.Method, wantMethod)
	}
	if request.URL != "http://localhost:8080"+wantPath {
		t.Fatalf("url = %s, want %s", request.URL, "http://localhost:8080"+wantPath)
	}
	if query := requestURLQuery(t, request); query != "" {
		t.Fatalf("query = %q, want empty", query)
	}
	if wantBody == "" {
		if len(request.Body) != 0 {
			t.Fatalf("body = %s, want empty", request.Body)
		}
		return
	}
	if string(request.Body) != wantBody {
		t.Fatalf("body = %s, want %s", request.Body, wantBody)
	}
}

// requestURLQuery returns the query string of a captured request.
func requestURLQuery(t *testing.T, request *HTTPRequest) string {
	t.Helper()
	_, after, found := strings.Cut(request.URL, "?")
	if !found {
		return ""
	}
	return after
}

func TestPreparedDataGetClientBootstrap(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{name: "ready", body: preparedDataBootstrapBody},
		{name: "attention required", body: `{"tenantSlug":"acme","state":"attention_required",` +
			`"observedAt":"2026-07-18T12:00:00Z","provenance":{"owner":"identity-bootstrap"},` +
			`"retryability":"bounded","warnings":[{"code":"missing_domain","message":"domain missing"}],` +
			`"nextAction":{"action":"retry","maxRetries":3}}`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, testCase.body)
			client := newPreparedDataTestClient(t, doer)

			status, _, err := client.GetClientBootstrap(context.Background())
			if err != nil {
				t.Fatalf("GetClientBootstrap error: %v", err)
			}
			if status.TenantSlug != "acme" || status.State == "" {
				t.Fatalf("status = %+v", status)
			}
			assertPreparedDataRequest(t, doer, http.MethodGet, "/api/v1/admin/bootstrap", "")
		})
	}
}

func TestPreparedDataSubmitClientBootstrap(t *testing.T) {
	cases := []struct {
		name     string
		request  PreparedDataBootstrapSubmitRequest
		wantBody string
	}{
		{
			name:     "note only",
			request:  PreparedDataBootstrapSubmitRequest{Note: "primary domain added"},
			wantBody: `{"note":"primary domain added"}`,
		},
		{
			name: "with proofs",
			request: PreparedDataBootstrapSubmitRequest{
				ContactEmail: "ops@acme.test",
				Proofs: []PreparedDataBootstrapProof{{
					Kind:       PreparedDataBootstrapProofKindDNSTXT,
					Value:      "custd-verify=abc",
					ObservedAt: "2026-07-18T12:00:00Z",
				}},
			},
			wantBody: `{"contactEmail":"ops@acme.test","proofs":[{"kind":"dns_txt","value":"custd-verify=abc",` +
				`"observedAt":"2026-07-18T12:00:00Z"}]}`,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, preparedDataBootstrapBody)
			client := newPreparedDataTestClient(t, doer)

			status, _, err := client.SubmitClientBootstrap(context.Background(), testCase.request)
			if err != nil {
				t.Fatalf("SubmitClientBootstrap error: %v", err)
			}
			if status.TenantSlug != "acme" {
				t.Fatalf("status = %+v", status)
			}
			assertPreparedDataRequest(t, doer, http.MethodPost, "/api/v1/admin/bootstrap", testCase.wantBody)
		})
	}
}
