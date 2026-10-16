package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

const parserDefinitionFixture = `{
  "uuid": "11111111-1111-1111-1111-111111111111",
  "code": "nginx-combined",
  "name": "Nginx combined",
  "parserType": "regex",
  "pattern": "^\\S+",
  "timestampFormat": "2006-01-02T15:04:05Z07:00",
  "fieldMapping": {"client_ip": "context.ip"},
  "generatedFields": {"day": {"fields": ["timestamp"], "namespace": "generated"}},
  "enabled": true,
  "revision": 3
}`

const parserPresetListFixture = `{"presets":[
  {"code":"apache-combined","name":"Apache combined","family":"apache","description":"Access log",
   "parserType":"regex","pattern":"^\\S+","timestampFormat":"2006-01-02T15:04:05Z07:00",
   "fieldMapping":{"client_ip":"context.ip"},"sample":"127.0.0.1 - -","variants":["combined"]}
]}`

const parserPreviewFixture = `{"results":[
  {"line":1,"ok":true,"eventTypeSlug":"access","timestamp":"2026-10-14T00:00:00Z",
   "context":{"ip":"127.0.0.1"},"payload":{"status":200}},
  {"line":2,"ok":false,"error":"no match found"}
],"parsed":1,"failed":1}`

func TestParserListPresetsUsesTheSchemaRegistryRoute(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, parserPresetListFixture)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	presets, err := client.Parsers.ListPresets(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(doer.requests) != 1 || !strings.HasSuffix(doer.requests[0].URL, "/api/v1/parsers/presets") {
		t.Fatalf("request URL = %s", doer.requests[0].URL)
	}
	if len(presets.Presets) != 1 || presets.Presets[0].Code != "apache-combined" {
		t.Fatalf("presets = %+v", presets.Presets)
	}
	if presets.Presets[0].Variants[0] != "combined" {
		t.Fatalf("variants = %+v", presets.Presets[0].Variants)
	}
}

// The parser route emits a bare JSON array; the SDK wraps it in the named
// ParserListResponse so no caller handles an untyped list.
func TestParserListWrapsTheBareArrayResponse(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, "["+parserDefinitionFixture+"]")
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	list, err := client.Parsers.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(doer.requests) != 1 || !strings.HasSuffix(doer.requests[0].URL, "/api/v1/parsers") {
		t.Fatalf("request URL = %s", doer.requests[0].URL)
	}
	if len(list.Parsers) != 1 || list.Parsers[0].Code != "nginx-combined" {
		t.Fatalf("parsers = %+v", list.Parsers)
	}
	// Enabled and revision are newer optional fields; when present they carry the
	// server's value unchanged.
	if list.Parsers[0].Enabled == nil || !*list.Parsers[0].Enabled || list.Parsers[0].Revision == nil || *list.Parsers[0].Revision != 3 {
		t.Fatalf("enabled/revision = %+v/%+v", list.Parsers[0].Enabled, list.Parsers[0].Revision)
	}
}

func TestParserCreateSendsTheSaveBody(t *testing.T) {
	doer := newCaptureDoer(http.StatusCreated, parserDefinitionFixture)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	parser, err := client.Parsers.Create(context.Background(), ParserSaveRequest{
		Code:            "nginx-combined",
		Name:            "Nginx combined",
		ParserType:      ParserTypeRegex,
		Pattern:         strPtr(`^\S+`),
		TimestampFormat: "2006-01-02T15:04:05Z07:00",
		FieldMapping:    map[string]string{"client_ip": "context.ip"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if parser.UUID != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("parser = %+v", parser)
	}
	req := doer.requests[0]
	if req.Method != http.MethodPost || !strings.HasSuffix(req.URL, "/api/v1/parsers") {
		t.Fatalf("request = %s %s", req.Method, req.URL)
	}
	// Enabled is unset, so it is omitted: the server keeps its default and an
	// older server that does not yet accept the field still decodes the body.
	var body map[string]json.RawMessage
	if err := json.Unmarshal(req.Body, &body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body["enabled"]; ok {
		t.Fatalf("unset enabled must be omitted: %s", req.Body)
	}
	if string(body["code"]) != `"nginx-combined"` || string(body["parserType"]) != `"regex"` {
		t.Fatalf("body = %s", req.Body)
	}
}

func TestParserUpdateAndDeleteAddressTheExplicitUUID(t *testing.T) {
	uuid := "11111111-1111-1111-1111-111111111111"
	enabled := false

	doer := newCaptureDoer(http.StatusOK, parserDefinitionFixture)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	if _, err := client.Parsers.Update(context.Background(), uuid, ParserSaveRequest{
		Code: "nginx-combined", Name: "Nginx combined", ParserType: ParserTypeRegex,
		Pattern: strPtr(`^\S+`), FieldMapping: map[string]string{}, Enabled: &enabled,
	}); err != nil {
		t.Fatal(err)
	}
	if req := doer.requests[0]; req.Method != http.MethodPut || !strings.HasSuffix(req.URL, "/api/v1/parsers/"+uuid) {
		t.Fatalf("update request = %s %s", req.Method, req.URL)
	}
	var updateBody map[string]json.RawMessage
	if err := json.Unmarshal(doer.requests[0].Body, &updateBody); err != nil {
		t.Fatal(err)
	}
	if string(updateBody["enabled"]) != "false" {
		t.Fatalf("explicit enabled must be sent: %s", doer.requests[0].Body)
	}

	deleteDoer := newCaptureDoer(http.StatusOK, `{"deleted":true}`)
	deleteClient := newAdminTestClient(t, deleteDoer, "http://localhost:8080")
	deleted, err := deleteClient.Parsers.Delete(context.Background(), uuid)
	if err != nil {
		t.Fatal(err)
	}
	if !deleted.Deleted {
		t.Fatalf("deleted = %+v", deleted)
	}
	if req := deleteDoer.requests[0]; req.Method != http.MethodDelete || !strings.HasSuffix(req.URL, "/api/v1/parsers/"+uuid) {
		t.Fatalf("delete request = %s %s", req.Method, req.URL)
	}
}

func TestParserPreviewPostsTheBoundedRequest(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, parserPreviewFixture)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	result, err := client.Parsers.Preview(context.Background(), ParserPreviewRequest{
		Parser: ParserPreviewDefinition{
			Code: "nginx-combined", Name: "Nginx combined", ParserType: ParserTypeRegex,
			Pattern: strPtr(`^\S+`), FieldMapping: map[string]string{"client_ip": "context.ip"},
		},
		Sample:   "127.0.0.1",
		MaxLines: 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	req := doer.requests[0]
	if req.Method != http.MethodPost || !strings.HasSuffix(req.URL, "/api/v1/parser-preview") {
		t.Fatalf("request = %s %s", req.Method, req.URL)
	}
	if result.Parsed != 1 || result.Failed != 1 || len(result.Results) != 2 {
		t.Fatalf("result = %+v", result)
	}
	if !result.Results[0].OK || result.Results[1].OK || result.Results[1].Error != "no match found" {
		t.Fatalf("rows = %+v", result.Results)
	}
	// The preview route decodes strictly; the previewed definition must not carry
	// the create/update-only enabled field.
	var body map[string]json.RawMessage
	if err := json.Unmarshal(req.Body, &body); err != nil {
		t.Fatal(err)
	}
	var definition map[string]json.RawMessage
	if err := json.Unmarshal(body["parser"], &definition); err != nil {
		t.Fatal(err)
	}
	if _, ok := definition["enabled"]; ok {
		t.Fatalf("preview parser must not carry enabled: %s", body["parser"])
	}
}

func strPtr(value string) *string { return &value }

func TestParserFramingRoundTripsAndIsRevisionPinned(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
	  "uuid":"11111111-1111-1111-1111-111111111111","code":"docker","name":"Docker",
	  "parserType":"json","pattern":null,"timestampFormat":"2006-01-02T15:04:05Z07:00",
	  "fieldMapping":{},"framing":{"type":"docker-json","maxRecordBytes":262144,"maxLines":200},
	  "enabled":true,"revision":4}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	parser, err := client.Parsers.Create(context.Background(), ParserSaveRequest{
		Code: "docker", Name: "Docker", ParserType: ParserTypeJSON,
		TimestampFormat: "2006-01-02T15:04:05Z07:00", FieldMapping: map[string]string{},
		Framing: &ParserFraming{Type: ParserFramingDockerJSON, MaxRecordBytes: intPtr(262144), MaxLines: intPtr(200)},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Framing travels with the save body so the persisted revision pins it.
	var body map[string]json.RawMessage
	if err := json.Unmarshal(doer.requests[0].Body, &body); err != nil {
		t.Fatal(err)
	}
	if string(body["framing"]) != `{"type":"docker-json","maxRecordBytes":262144,"maxLines":200}` {
		t.Fatalf("framing body = %s", body["framing"])
	}
	if parser.Framing == nil || parser.Framing.Type != ParserFramingDockerJSON ||
		parser.Framing.MaxLines == nil || *parser.Framing.MaxLines != 200 {
		t.Fatalf("decoded framing = %+v", parser.Framing)
	}
	// The revision is what pins the framing; there is no separate framing revision.
	if parser.Revision == nil || *parser.Revision != 4 {
		t.Fatalf("revision = %+v", parser.Revision)
	}
}

func TestParserRejectsUnsupportedFramingBeforeSending(t *testing.T) {
	framing := &ParserFraming{Type: "idle"}
	cases := []struct {
		name string
		call func(*ParserClient) error
		doer *captureDoer
	}{
		{"create", func(p *ParserClient) error {
			_, err := p.Create(context.Background(), ParserSaveRequest{
				Code: "x", Name: "x", ParserType: ParserTypeJSON,
				TimestampFormat: "t", FieldMapping: map[string]string{}, Framing: framing,
			})
			return err
		}, newCaptureDoer(http.StatusCreated, parserDefinitionFixture)},
		{"preview", func(p *ParserClient) error {
			_, err := p.Preview(context.Background(), ParserPreviewRequest{
				Parser: ParserPreviewDefinition{
					Code: "x", Name: "x", ParserType: ParserTypeJSON,
					TimestampFormat: "t", FieldMapping: map[string]string{}, Framing: framing,
				},
				Sample: "x", MaxLines: 1,
			})
			return err
		}, newCaptureDoer(http.StatusOK, parserPreviewFixture)},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			client := newAdminTestClient(t, testCase.doer, "http://localhost:8080")
			if err := testCase.call(client.Parsers); err == nil {
				t.Fatal("expected an unsupported-framing error")
			}
			if len(testCase.doer.requests) != 0 {
				t.Fatalf("validation should not cost a request, sent %d", len(testCase.doer.requests))
			}
		})
	}
}

func intPtr(value int) *int { return &value }

func TestParserPreviewDecodesFramingOutcome(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"results":[
	  {"line":1,"ok":true,"framing":"complete","eventTypeSlug":"access","payload":{}},
	  {"line":2,"ok":false,"framing":"oversize","error":"record exceeds the framing bound"}
	],"parsed":1,"failed":1}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	result, err := client.Parsers.Preview(context.Background(), ParserPreviewRequest{
		Parser: ParserPreviewDefinition{
			Code: "php", Name: "PHP", ParserType: ParserTypeRegex, Pattern: strPtr(`^\S+`),
			FieldMapping: map[string]string{},
			Framing:      &ParserFraming{Type: ParserFramingPHPMultiline},
		},
		Sample: "x", MaxLines: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Results[0].Framing != ParserFramingOutcomeComplete {
		t.Fatalf("row 0 framing = %q", result.Results[0].Framing)
	}
	if result.Results[1].Framing != ParserFramingOutcomeOversize {
		t.Fatalf("row 1 framing = %q", result.Results[1].Framing)
	}
}

func TestParserFramingRejectsOutOfRangeBounds(t *testing.T) {
	cases := []struct {
		name    string
		framing *ParserFraming
	}{
		{"negative bytes", &ParserFraming{Type: ParserFramingDockerJSON, MaxRecordBytes: intPtr(-1)}},
		{"bytes above hard ceiling", &ParserFraming{Type: ParserFramingDockerJSON, MaxRecordBytes: intPtr(MaxParserFramingRecordBytes + 1)}},
		{"negative lines", &ParserFraming{Type: ParserFramingPHPMultiline, MaxLines: intPtr(-1)}},
		{"lines above hard ceiling", &ParserFraming{Type: ParserFramingPHPMultiline, MaxLines: intPtr(MaxParserFramingLines + 1)}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusCreated, parserDefinitionFixture)
			client := newAdminTestClient(t, doer, "http://localhost:8080")
			_, err := client.Parsers.Create(context.Background(), ParserSaveRequest{
				Code: "x", Name: "x", ParserType: ParserTypeJSON,
				TimestampFormat: "t", FieldMapping: map[string]string{}, Framing: testCase.framing,
			})
			if err == nil {
				t.Fatal("expected a framing bound error")
			}
			if len(doer.requests) != 0 {
				t.Fatalf("validation should not cost a request, sent %d", len(doer.requests))
			}
		})
	}
}
