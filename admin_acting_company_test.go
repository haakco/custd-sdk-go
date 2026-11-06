package custd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// supportContextHeader is how a caller explicitly asks the server for grant-backed
// support context. Ordinary system-administrator targeting names the tenant through
// the companySlug query instead, so this header must only appear when the caller
// supplies it itself.
const supportContextHeader = "X-Custd-Effective-Company"

// recordedAdminRequest is one request as it arrived at the server, whichever
// transport carried it.
type recordedAdminRequest struct {
	method  string
	url     string
	headers map[string]string
}

// recordDoerRequest returns the last request the capturing transport saw.
func recordDoerRequest(doer *captureDoer) recordedAdminRequest {
	request := doer.requests[len(doer.requests)-1]
	return recordedAdminRequest{method: request.Method, url: request.URL, headers: request.Headers}
}

// newNativeAdmin builds an admin client over the SDK's own HTTP client against a
// local server, so a test exercises the transport that owns none of the custom
// doer's wiring.
func newNativeAdmin(t *testing.T, body string) (*AdminClient, func() recordedAdminRequest) {
	t.Helper()
	var recorded recordedAdminRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := make(map[string]string, len(r.Header))
		for key, values := range r.Header {
			headers[key] = values[0]
		}
		recorded = recordedAdminRequest{method: r.Method, url: r.URL.String(), headers: headers}
		w.Header().Set(requestIDHeader, "req-native")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	client := NewClient(&ClientConfig{
		BaseURL:       server.URL,
		APIKey:        "admin-token",
		FlushInterval: time.Hour,
	})
	t.Cleanup(func() { _ = client.Close(context.Background()) })
	return client.Admin, func() recordedAdminRequest { return recorded }
}

// assertTenantTargetedByQueryOnly proves one request names the tenant through the
// companySlug query parameter and never through the support-context header. The
// header asks for a grant; the query only selects a tenant the caller may already read.
func assertTenantTargetedByQueryOnly(t *testing.T, request recordedAdminRequest, company string) {
	t.Helper()
	if got, present := request.headers[supportContextHeader]; present {
		t.Fatalf("%s = %q, want it absent: targeting a tenant must not request support context",
			supportContextHeader, got)
	}
	parsed, err := url.Parse(request.url)
	if err != nil {
		t.Fatalf("parse url %q: %v", request.url, err)
	}
	if got := parsed.Query()[companySlugQueryParameter]; len(got) != 1 || got[0] != company {
		t.Fatalf("%s = %v in %q, want exactly %s", companySlugQueryParameter, got, request.url, company)
	}
}

// assertAuditLabels proves the caller's command and client labels still reach the
// server, so a supporting operator action keeps its audit attribution.
func assertAuditLabels(t *testing.T, request recordedAdminRequest, client, command string) {
	t.Helper()
	if got := request.headers[clientLabelHeader]; got != client {
		t.Fatalf("%s = %q, want %q", clientLabelHeader, got, client)
	}
	if got := request.headers[commandLabelHeader]; got != command {
		t.Fatalf("%s = %q, want %q", commandLabelHeader, got, command)
	}
}

// TestActingCompanyTargetsByQueryOnOrdinaryRequests proves a system administrator's
// ordinary tenant targeting is the companySlug query alone. Synthesizing the
// support-context header from it made a legitimate `--company` call demand a support
// grant, which surfaced as a 403 "support access grant not found" while the same read
// chosen in the UI succeeded.
func TestActingCompanyTargetsByQueryOnOrdinaryRequests(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"scopes":[]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	client.config.ActingCompanySlug = "haakco"
	client.config.ClientLabel = "custdctl"
	client.config.CommandLabel = "data-scopes.grants.list"

	if _, err := client.Admin.DataScopes.ListScopes(context.Background(), false); err != nil {
		t.Fatalf("ListScopes returned error: %v", err)
	}
	request := recordDoerRequest(doer)
	assertTenantTargetedByQueryOnly(t, request, "haakco")
	assertAuditLabels(t, request, "custdctl", "data-scopes.grants.list")
}

// TestActingCompanyTargetsByQueryOnOrdinaryRequestsOverHTTP proves the native HTTP
// transport carries the same contract as the custom doer: the support-context header
// is never inferred from ordinary targeting.
func TestActingCompanyTargetsByQueryOnOrdinaryRequestsOverHTTP(t *testing.T) {
	admin, lastRequest := newNativeAdmin(t, `{"scopes":[]}`)
	admin.client.config.ActingCompanySlug = "haakco"

	if _, err := admin.DataScopes.ListScopes(context.Background(), false); err != nil {
		t.Fatalf("ListScopes returned error: %v", err)
	}
	assertTenantTargetedByQueryOnly(t, lastRequest(), "haakco")
}

// TestActingCompanyTargetsByQueryOnMetadataRequests proves the metadata request helper
// labels the call and names the tenant, but does not open support context.
func TestActingCompanyTargetsByQueryOnMetadataRequests(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"grants":[]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	client.config.ActingCompanySlug = "haakco"
	client.config.ClientLabel = "custdctl"
	client.config.CommandLabel = "support.access.list"

	if _, _, err := client.Admin.SupportAccess.List(context.Background()); err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	request := recordDoerRequest(doer)
	assertTenantTargetedByQueryOnly(t, request, "haakco")
	assertAuditLabels(t, request, "custdctl", "support.access.list")
}

// TestActingCompanyTargetsByQueryOnMetadataRequestsOverHTTP proves the metadata helper's
// native HTTP branch does not synthesize the support-context header either.
func TestActingCompanyTargetsByQueryOnMetadataRequestsOverHTTP(t *testing.T) {
	admin, lastRequest := newNativeAdmin(t, `{"grants":[]}`)
	admin.client.config.ActingCompanySlug = "haakco"

	if _, _, err := admin.SupportAccess.List(context.Background()); err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	assertTenantTargetedByQueryOnly(t, lastRequest(), "haakco")
}

// TestActingCompanyKeepsAnExplicitCompanySelector proves a method that names its own
// company keeps that name: the configured default fills a gap, it never overrides the
// caller's own selection.
func TestActingCompanyKeepsAnExplicitCompanySelector(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"users":[]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	client.config.ActingCompanySlug = "haakco"

	if _, err := client.Admin.AdminUsers.List(context.Background(), "awthy", false); err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	assertTenantTargetedByQueryOnly(t, recordDoerRequest(doer), "awthy")
}

// TestActingCompanyIsOmittedWhenUnset proves an ordinary tenant command does not claim
// to act for anyone: an empty setting leaves both the query parameter and the
// support-context header off.
func TestActingCompanyIsOmittedWhenUnset(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"scopes":[]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	if _, err := client.Admin.DataScopes.ListScopes(context.Background(), false); err != nil {
		t.Fatalf("ListScopes returned error: %v", err)
	}
	request := recordDoerRequest(doer)
	if got, present := request.headers[supportContextHeader]; present {
		t.Fatalf("%s = %q, want it absent", supportContextHeader, got)
	}
	if strings.Contains(request.url, companySlugQueryParameter) {
		t.Fatalf("url = %s, want no tenant query", request.url)
	}
}

// TestExplicitSupportHeaderIsPreserved proves a caller that supplies the
// support-context header still gets it forwarded on both request helpers. Support use
// stays possible; it is simply never inferred from ordinary target selection.
func TestExplicitSupportHeaderIsPreserved(t *testing.T) {
	tests := []struct {
		name    string
		request func(*AdminClient, map[string]string) error
	}{
		{
			name: "ordinary request",
			request: func(admin *AdminClient, headers map[string]string) error {
				var out DataScopeListResponse
				return admin.requestWithHeaders(
					context.Background(), http.MethodGet, "/data-scopes", nil, &out, headers,
				)
			},
		},
		{
			name: "metadata request",
			request: func(admin *AdminClient, headers map[string]string) error {
				var out SupportAccessGrantListResponse
				_, err := admin.requestWithMetadataHeaders(
					context.Background(), http.MethodGet, "/support-access", nil, &out, headers,
				)
				return err
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, `{"scopes":[],"grants":[]}`)
			client := newAdminTestClient(t, doer, "http://localhost:8080")
			client.config.ActingCompanySlug = "haakco"
			headers := map[string]string{supportContextHeader: "awthy"}

			if err := test.request(client.Admin, headers); err != nil {
				t.Fatalf("request returned error: %v", err)
			}
			request := recordDoerRequest(doer)
			if got := request.headers[supportContextHeader]; got != "awthy" {
				t.Fatalf("%s = %q, want the caller's own support header forwarded", supportContextHeader, got)
			}
			if !strings.Contains(request.url, "companySlug=haakco") {
				t.Fatalf("url = %s, want the tenant query preserved", request.url)
			}
		})
	}
}

// TestExplicitSupportHeaderIsPreservedOverHTTP proves the native HTTP transport forwards
// a caller-supplied support-context header: only the automatic synthesis is gone.
func TestExplicitSupportHeaderIsPreservedOverHTTP(t *testing.T) {
	admin, lastRequest := newNativeAdmin(t, `{"grants":[]}`)
	admin.client.config.ActingCompanySlug = "haakco"
	headers := map[string]string{supportContextHeader: "awthy"}

	var out SupportAccessGrantListResponse
	if _, err := admin.requestWithMetadataHeaders(
		context.Background(), http.MethodGet, "/support-access", nil, &out, headers,
	); err != nil {
		t.Fatalf("requestWithMetadataHeaders returned error: %v", err)
	}
	if got := lastRequest().headers[supportContextHeader]; got != "awthy" {
		t.Fatalf("%s = %q, want the caller's own support header forwarded", supportContextHeader, got)
	}
}

// TestAccountDiagnosticDoesNotSynthesizeSupportHeader proves the account read the CLI
// performs before a command is ordinary too: it carries audit labels, never an inferred
// support context.
func TestAccountDiagnosticDoesNotSynthesizeSupportHeader(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"companySlug":"haakco","email":"a@b.c"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	client.config.ActingCompanySlug = "haakco"
	client.config.ClientLabel = "custdctl"
	client.config.CommandLabel = "auth.whoami"

	if _, _, err := client.Admin.GetAccount(context.Background()); err != nil {
		t.Fatalf("GetAccount returned error: %v", err)
	}
	request := recordDoerRequest(doer)
	if got, present := request.headers[supportContextHeader]; present {
		t.Fatalf("%s = %q, want it absent", supportContextHeader, got)
	}
	assertAuditLabels(t, request, "custdctl", "auth.whoami")
}

// TestPreparedDataDiagnosticDoesNotSynthesizeSupportHeader proves the prepared-data root
// is ordinary tenant reading as well: it keeps the audit labels and does not open
// support context.
func TestPreparedDataDiagnosticDoesNotSynthesizeSupportHeader(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, preparedDataOutputListBody)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	client.config.ActingCompanySlug = "haakco"
	client.config.ClientLabel = "custdctl"
	client.config.CommandLabel = "reporting.outputs.list"
	prepared := NewPreparedDataClient(client.Admin)

	if _, _, err := prepared.ListPreparedDataOutputs(context.Background()); err != nil {
		t.Fatalf("ListPreparedDataOutputs returned error: %v", err)
	}
	request := recordDoerRequest(doer)
	if got, present := request.headers[supportContextHeader]; present {
		t.Fatalf("%s = %q, want it absent", supportContextHeader, got)
	}
	assertAuditLabels(t, request, "custdctl", "reporting.outputs.list")
}
