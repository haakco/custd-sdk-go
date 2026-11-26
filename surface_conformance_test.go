package custd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// surfaceRow proves one public SDK call sends one Custd operation. The id is
// the operation in contract-fixtures/api/operations.json; method is set only
// for ANY operations, where the row pins the method the SDK actually sends.
// skipExample names the server example ("minimal", "maximal" or "both") a
// row cannot replay, with skipReason saying why; the reason is logged.
type surfaceRow struct {
	id          string
	path        string
	method      string
	skipExample string
	skipReason  string
	call        func(*surfaceEnv) error
}

// surfaceEnv is the fake transport a row calls through. Client calls go
// through the capture doer; serverURL serves clients that take a base URL and
// their own *http.Client, recording into the same request list.
type surfaceEnv struct {
	t      *testing.T
	client *CustdClient
	doer   *captureDoer
	status int
	body   string
}

type surfaceOperation struct {
	ID      string                  `json:"id"`
	Method  string                  `json:"method"`
	Path    string                  `json:"path"`
	SDK     string                  `json:"sdk"`
	Example string                  `json:"example"`
	Query   []surfaceQueryParameter `json:"query"`
	pattern *regexp.Regexp
}

type surfaceQueryParameter struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
}

type surfaceSnapshot struct {
	Operations []surfaceOperation `json:"operations"`
}

type surfaceCoverage struct {
	Operations []string `json:"operations"`
}

type surfaceExample struct {
	Minimal json.RawMessage `json:"minimal"`
	Maximal json.RawMessage `json:"maximal"`
}

var surfacePathParameter = regexp.MustCompile(`\\\{[^}]+\\\}`)

func allSurfaceRows() []surfaceRow {
	rows := append([]surfaceRow{}, clientSurfaceRows...)
	rows = append(rows, adminSurfaceRows...)
	rows = append(rows, siteDomainSurfaceRows...)
	rows = append(rows, authProjectSurfaceRows...)
	rows = append(rows, uncertaintySurfaceRows...)
	rows = append(rows, schemaRegistrySurfaceRows...)
	rows = append(rows, measurementSurfaceRows...)
	rows = append(rows, measurementStorySurfaceRows...)
	rows = append(rows, predictionReadSurfaceRows...)
	rows = append(rows, appItemSurfaceRows...)
	rows = append(rows, emailSurfaceRows...)
	rows = append(rows, operationsSurfaceRows...)
	rows = append(rows, publicReadSurfaceRows...)
	rows = append(rows, analyticsReadSurfaceRows...)
	rows = append(rows, ingestSurfaceRows...)
	return append(rows, argumentSurfaceRows...)
}

// TestSurfaceConformance calls every operation the Go SDK implements and
// proves the route it sends, that the server's example responses decode into
// the SDK's response type, and that coverage/go.json claims exactly these rows.
func TestSurfaceConformance(t *testing.T) {
	operations := loadSurfaceOperations(t)
	covered := map[string]bool{}
	sent := map[string]url.Values{}
	for _, row := range allSurfaceRows() {
		operation, ok := operations[row.id]
		if !ok || operation.SDK != "required" {
			t.Fatalf("row %q is not a required operation in operations.json", row.id)
		}
		covered[row.id] = true
		t.Run(row.id, func(t *testing.T) {
			assertSurfaceRoute(t, row, operation, operations)
			if sent[row.id] == nil {
				sent[row.id] = url.Values{}
			}
			for name := range surfaceQuerySent(t, row) {
				sent[row.id].Set(name, "")
			}
			assertSurfaceExampleDecodes(t, row, operation)
		})
	}
	assertSurfaceQuerySendable(t, operations, sent)
	assertSurfaceCoverageFile(t, covered)
}

func assertSurfaceRoute(t *testing.T, row surfaceRow, operation surfaceOperation, operations map[string]surfaceOperation) {
	t.Helper()
	env := newSurfaceEnv(t, `{}`)
	// The route is proven by the request sent. Whether the reply decodes is
	// the example check's question, so a reply error alone does not fail here.
	err := row.call(env)
	if len(env.doer.requests) != 1 {
		t.Fatalf("requests = %d, want 1 (call error: %v)", len(env.doer.requests), err)
	}
	request := env.doer.requests[0]
	wantMethod := operation.Method
	if wantMethod == "ANY" {
		wantMethod = row.method
	}
	if request.Method != wantMethod {
		t.Fatalf("method = %s, want %s", request.Method, wantMethod)
	}
	parsed, parseErr := url.Parse(request.URL)
	if parseErr != nil {
		t.Fatalf("parse url %q: %v", request.URL, parseErr)
	}
	if parsed.EscapedPath() != row.path {
		t.Fatalf("path = %s, want %s", parsed.EscapedPath(), row.path)
	}
	if resolved := resolveSurfaceOperation(operations, request.Method, row.path); resolved != row.id {
		t.Fatalf("%s %s resolves to %q, want %q", request.Method, row.path, resolved, row.id)
	}
	for _, parameter := range operation.Query {
		if parameter.Required && !parsed.Query().Has(parameter.Name) &&
			surfaceRequiredByActingCompany[row.id+" ?"+parameter.Name] == "" {
			t.Errorf("required query parameter %q is not sent", parameter.Name)
		}
	}
}

// surfaceQuerySent runs the row with an acting company, the Go SDK's carrier
// for companySlug, and returns the query parameter names the call sends.
func surfaceQuerySent(t *testing.T, row surfaceRow) url.Values {
	t.Helper()
	env := newSurfaceEnv(t, `{}`)
	env.client.config.ActingCompanySlug = "acme"
	_ = row.call(env)
	parsed, err := url.Parse(env.doer.requests[0].URL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	return parsed.Query()
}

// assertSurfaceQuerySendable proves that, across an operation's rows, the Go
// methods can send every query parameter the operation declares. A parameter
// no method can send must be in surfaceUnsentQuery with its reason.
func assertSurfaceQuerySendable(t *testing.T, operations map[string]surfaceOperation, sent map[string]url.Values) {
	t.Helper()
	for id, names := range sent {
		for _, parameter := range operations[id].Query {
			reason, exempt := surfaceUnsentQuery[id+" ?"+parameter.Name]
			switch {
			case exempt && names.Has(parameter.Name):
				t.Errorf("%s: query parameter %q is sent, so its surfaceUnsentQuery entry is stale", id, parameter.Name)
			case exempt:
				t.Logf("%s: query parameter %q not sent: %s", id, parameter.Name, reason)
			case !names.Has(parameter.Name):
				t.Errorf("%s: declared query parameter %q is not sent by any row; pass it or list it in surfaceUnsentQuery",
					id, parameter.Name)
			}
		}
	}
}

// surfaceRequiredByActingCompany lists required query parameters, keyed like
// surfaceUnsentQuery, that the method takes from ClientConfig.ActingCompanySlug
// rather than an argument; surfaceQuerySent proves the acting company sends them.
var surfaceRequiredByActingCompany = map[string]string{
	"admin-api POST /api/v1/admin/subject-exports/{requestId}/force ?companySlug": "Force takes no company argument",
}

// surfaceUnsentQuery lists declared query parameters, keyed "<operation id>
// ?<name>", that no Go method sends, with the reason.
var surfaceUnsentQuery = map[string]string{}

// assertSurfaceExampleDecodes answers the call with the server's minimal and
// maximal example bodies; a Go response type that cannot hold them, or that
// lacks a field they carry, fails.
func assertSurfaceExampleDecodes(t *testing.T, row surfaceRow, operation surfaceOperation) {
	t.Helper()
	if operation.Example == "" {
		return
	}
	var example surfaceExample
	if err := json.Unmarshal(readContractFixture(t, "api/"+operation.Example), &example); err != nil {
		t.Fatalf("parse example: %v", err)
	}
	if (row.skipExample == "") != (row.skipReason == "") {
		t.Fatalf("skipExample %q needs a skipReason", row.skipExample)
	}
	restore := unmarshalResponse
	unmarshalResponse = unmarshalStrict
	defer func() { unmarshalResponse = restore }()
	for name, body := range map[string]json.RawMessage{"minimal": example.Minimal, "maximal": example.Maximal} {
		if name == row.skipExample || row.skipExample == "both" {
			t.Logf("%s example not replayed: %s", name, row.skipReason)
			continue
		}
		if err := row.call(newSurfaceEnv(t, string(body))); err != nil {
			t.Errorf("%s example does not decode: %v", name, err)
		}
	}
}

// unmarshalStrict rejects a field the SDK type lacks, so a server field the
// Go SDK drops fails the example replay instead of being silently ignored.
func unmarshalStrict(data []byte, out any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(out)
}

func assertSurfaceCoverageFile(t *testing.T, covered map[string]bool) {
	t.Helper()
	var coverage surfaceCoverage
	if err := json.Unmarshal(readContractFixture(t, "api/coverage/go.json"), &coverage); err != nil {
		t.Fatalf("parse coverage: %v", err)
	}
	want := make([]string, 0, len(covered))
	for id := range covered {
		want = append(want, id)
	}
	sort.Strings(want)
	if strings.Join(coverage.Operations, "\n") != strings.Join(want, "\n") {
		t.Fatalf("coverage/go.json must list exactly the %d tested operations, sorted:\n%s",
			len(want), strings.Join(want, "\n"))
	}
}

func loadSurfaceOperations(t *testing.T) map[string]surfaceOperation {
	t.Helper()
	var snapshot surfaceSnapshot
	if err := json.Unmarshal(readContractFixture(t, "api/operations.json"), &snapshot); err != nil {
		t.Fatalf("parse operations: %v", err)
	}
	operations := make(map[string]surfaceOperation, len(snapshot.Operations))
	for _, operation := range snapshot.Operations {
		pattern := surfacePathParameter.ReplaceAllString(regexp.QuoteMeta(operation.Path), `[^/]+`)
		operation.pattern = regexp.MustCompile("^" + pattern + "$")
		operations[operation.ID] = operation
	}
	return operations
}

// resolveSurfaceOperation finds the operation a request lands on, preferring
// the most literal template, so a row cannot claim a neighbouring route.
func resolveSurfaceOperation(operations map[string]surfaceOperation, method, path string) string {
	best := ""
	bestParameters := -1
	for id, operation := range operations {
		if (operation.Method != method && operation.Method != "ANY") || !operation.pattern.MatchString(path) {
			continue
		}
		parameters := strings.Count(operation.Path, "{")
		if bestParameters == -1 || parameters < bestParameters || (parameters == bestParameters && id < best) {
			best, bestParameters = id, parameters
		}
	}
	return best
}

func newSurfaceEnv(t *testing.T, body string) *surfaceEnv {
	t.Helper()
	doer := newCaptureDoer(http.StatusOK, body)
	return &surfaceEnv{
		t:      t,
		client: newAdminTestClient(t, doer, "http://localhost:8080"),
		doer:   doer,
		status: http.StatusOK,
		body:   body,
	}
}

// serverURL starts a server answering like the capture doer, for clients that
// take a base URL and their own *http.Client.
func (e *surfaceEnv) serverURL() string {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		e.doer.requests = append(e.doer.requests, &HTTPRequest{Method: r.Method, URL: r.URL.String()})
		w.WriteHeader(e.status)
		_, _ = io.WriteString(w, e.body)
	}))
	e.t.Cleanup(server.Close)
	return server.URL
}
