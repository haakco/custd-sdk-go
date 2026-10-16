package custd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// parserEndpoint is the tenant-owned parser collection. It is served by the
// schema registry behind the authenticated tenant, so the credential already
// names the tenant and the caller never supplies a tenant slug.
const parserEndpoint = "/api/v1/parsers"

// parserPresetsEndpoint serves the built-in reusable parser presets. The
// catalog is static, non-tenant data behind the same authentication.
const parserPresetsEndpoint = "/api/v1/parsers/presets"

// parserPreviewEndpoint runs a client-owned definition against a bounded sample
// through the same parser the ingestion path uses. It writes nothing.
const parserPreviewEndpoint = "/api/v1/parser-preview"

// ParserType names a parser family the ingestion dispatcher can execute.
type ParserType string

const (
	ParserTypeRegex  ParserType = "regex"
	ParserTypeJSON   ParserType = "json"
	ParserTypeXML    ParserType = "xml"
	ParserTypeCSV    ParserType = "csv"
	ParserTypeSyslog ParserType = "syslog"
)

// ParserFramingType names a stateless batch-framing mode applied before the
// application parser runs on each record.
type ParserFramingType string

const (
	ParserFramingNone         ParserFramingType = "none"
	ParserFramingDockerJSON   ParserFramingType = "docker-json"
	ParserFramingK8sCRI       ParserFramingType = "k8s-cri"
	ParserFramingPHPMultiline ParserFramingType = "php-multiline"
)

var parserFramingTypes = map[ParserFramingType]struct{}{
	ParserFramingNone:         {},
	ParserFramingDockerJSON:   {},
	ParserFramingK8sCRI:       {},
	ParserFramingPHPMultiline: {},
}

// Parser framing bounds. A caller may request a lower cap; these ceilings are
// hard engine limits a request cannot raise. Zero bounds select the defaults.
// They mirror libs/go/logpresets, the single declaration-and-validation
// authority the parser engine, Preview, the catalog and the SDK share.
const (
	// DefaultParserFramingMaxRecordBytes is the per-record cap when a caller requests none.
	DefaultParserFramingMaxRecordBytes = 64 << 10
	// DefaultParserFramingMaxLines is the per-record line cap when a caller requests none.
	DefaultParserFramingMaxLines = 32
	// MaxParserFramingRecordBytes is the hard ceiling on one assembled record.
	MaxParserFramingRecordBytes = 1 << 20
	// MaxParserFramingLines is the hard ceiling on lines joined into one record.
	MaxParserFramingLines = 256
)

// ParserFramingOutcome is the visible per-record result of one framing pass. It
// is populated on a preview row only for a non-none framing mode.
type ParserFramingOutcome string

const (
	// ParserFramingOutcomeComplete is one assembled record.
	ParserFramingOutcomeComplete ParserFramingOutcome = "complete"
	// ParserFramingOutcomeMalformed is a record the framing mode could not decode.
	ParserFramingOutcomeMalformed ParserFramingOutcome = "malformed"
	// ParserFramingOutcomeIncomplete is a trailing partial record.
	ParserFramingOutcomeIncomplete ParserFramingOutcome = "incomplete"
	// ParserFramingOutcomeOversize is a record past a framing bound.
	ParserFramingOutcomeOversize ParserFramingOutcome = "oversize"
)

// ParserFraming is the stateless framing configuration carried by a parser
// definition. Framing is revision-pinned with the parser and is the same
// configuration for preview and save. Multiline assembly is bounded by
// MaxRecordBytes/MaxLines; there is deliberately no idle timeout on stateless
// batch framing.
type ParserFraming struct {
	Type           ParserFramingType `json:"type"`
	MaxRecordBytes *int              `json:"maxRecordBytes,omitempty"`
	MaxLines       *int              `json:"maxLines,omitempty"`
}

// validate rejects a framing declaration the server would reject before a
// request is sent: an unsupported type, or a bound outside 0..hard ceiling. A
// nil framing is valid; the server applies its own default (64 KiB / 32 lines).
func (f *ParserFraming) validate() error {
	if f == nil {
		return nil
	}
	if _, ok := parserFramingTypes[f.Type]; !ok {
		return fmt.Errorf("custd: parser framing type %q is not supported", f.Type)
	}
	if f.MaxRecordBytes != nil && (*f.MaxRecordBytes < 0 || *f.MaxRecordBytes > MaxParserFramingRecordBytes) {
		return fmt.Errorf("custd: parser framing maxRecordBytes must be 0..%d", MaxParserFramingRecordBytes)
	}
	if f.MaxLines != nil && (*f.MaxLines < 0 || *f.MaxLines > MaxParserFramingLines) {
		return fmt.Errorf("custd: parser framing maxLines must be 0..%d", MaxParserFramingLines)
	}
	return nil
}

// GeneratedFieldConfig defines how a generated field is computed from source fields.
type GeneratedFieldConfig struct {
	Fields    []string `json:"fields"`
	Namespace string   `json:"namespace"`
}

// ParserDefinition is a tenant-owned parser returned by the parser routes.
type ParserDefinition struct {
	UUID            string                          `json:"uuid"`
	Code            string                          `json:"code"`
	Name            string                          `json:"name"`
	ParserType      ParserType                      `json:"parserType"`
	Pattern         *string                         `json:"pattern"`
	TimestampFormat string                          `json:"timestampFormat"`
	FieldMapping    map[string]string               `json:"fieldMapping"`
	GeneratedFields map[string]GeneratedFieldConfig `json:"generatedFields,omitempty"`
	// Framing is the stateless framing configuration applied before the parser.
	// It is revision-pinned with the parser; it is optional so a server that
	// predates it still decodes.
	Framing *ParserFraming `json:"framing,omitempty"`
	// Enabled and Revision are newer server fields. Both are pointers so a
	// response from a server that predates them still decodes; when they are
	// present they carry the server's value unchanged.
	Enabled  *bool `json:"enabled,omitempty"`
	Revision *int  `json:"revision,omitempty"`
}

// ParserListResponse is the named DTO for the parser collection. The route emits
// a bare JSON array; the SDK wraps it so a caller never handles an untyped list.
type ParserListResponse struct {
	Parsers []ParserDefinition `json:"parsers"`
}

// ParserPreset is one built-in reusable parser definition from the published
// catalog. Clients turn a preset into a tenant-owned parser through Create.
type ParserPreset struct {
	Code            string            `json:"code"`
	Name            string            `json:"name"`
	Family          string            `json:"family"`
	Description     string            `json:"description"`
	ParserType      ParserType        `json:"parserType"`
	Pattern         string            `json:"pattern"`
	TimestampFormat string            `json:"timestampFormat"`
	FieldMapping    map[string]string `json:"fieldMapping"`
	Sample          string            `json:"sample"`
	Variants        []string          `json:"variants"`
}

// ParserPresetListResponse is the named DTO for the preset catalog.
type ParserPresetListResponse struct {
	Presets []ParserPreset `json:"presets"`
}

// ParserSaveRequest is the create/update body for a tenant-owned parser.
//
// Enabled is optional. It is a pointer so that when the caller leaves it unset
// the field is omitted from the body and the server keeps its default (an
// enabled parser); that also keeps a request valid against a server that does
// not yet accept the field.
type ParserSaveRequest struct {
	Code            string            `json:"code"`
	Name            string            `json:"name"`
	ParserType      ParserType        `json:"parserType"`
	Pattern         *string           `json:"pattern"`
	TimestampFormat string            `json:"timestampFormat"`
	FieldMapping    map[string]string `json:"fieldMapping"`
	Framing         *ParserFraming    `json:"framing,omitempty"`
	Enabled         *bool             `json:"enabled,omitempty"`
}

// ParserPreviewDefinition is the parser shape a preview executes. It mirrors
// the saved definition minus the server-assigned uuid and the newer-only
// enabled/revision fields; the preview route decodes strictly, so it must not
// carry them.
type ParserPreviewDefinition struct {
	Code            string            `json:"code"`
	Name            string            `json:"name"`
	ParserType      ParserType        `json:"parserType"`
	Pattern         *string           `json:"pattern"`
	TimestampFormat string            `json:"timestampFormat"`
	FieldMapping    map[string]string `json:"fieldMapping"`
	// Framing is the same stateless framing configuration save carries, so a
	// preview exercises exactly what a saved parser would.
	Framing *ParserFraming `json:"framing,omitempty"`
}

// ParserPreviewRequest is the dry-run body. The service bounds the sample to
// 256 KiB and MaxLines to 100; a longer request is rejected as invalid.
type ParserPreviewRequest struct {
	Parser   ParserPreviewDefinition `json:"parser"`
	Sample   string                  `json:"sample"`
	MaxLines int                     `json:"maxLines"`
}

// ParserPreviewRow is one sample line's result. Line is 1-based in the original
// sample. Framing is the source-framing outcome and is populated only for a
// non-none framing mode. Error is set only when OK is false.
type ParserPreviewRow struct {
	Line          int                  `json:"line"`
	OK            bool                 `json:"ok"`
	Framing       ParserFramingOutcome `json:"framing,omitempty"`
	EventTypeSlug string               `json:"eventTypeSlug,omitempty"`
	Timestamp     string               `json:"timestamp,omitempty"`
	Context       map[string]any       `json:"context,omitempty"`
	Payload       map[string]any       `json:"payload,omitempty"`
	Error         string               `json:"error,omitempty"`
}

// ParserPreviewResponse is the dry-run result. Parsed and Failed summarize
// Results.
type ParserPreviewResponse struct {
	Results []ParserPreviewRow `json:"results"`
	Parsed  int                `json:"parsed"`
	Failed  int                `json:"failed"`
}

// ParserDeleteResponse is the body of a successful parser deletion.
type ParserDeleteResponse struct {
	Deleted bool `json:"deleted"`
}

// ParserClient owns the tenant parser routes: the built-in preset catalog, the
// tenant-owned CRUD collection and the bounded dry-run preview.
type ParserClient struct {
	client *CustdClient
}

func newParserClient(client *CustdClient) *ParserClient {
	return &ParserClient{client: client}
}

// ListPresets returns the built-in reusable parser catalog.
func (p *ParserClient) ListPresets(ctx context.Context) (*ParserPresetListResponse, error) {
	var out ParserPresetListResponse
	if err := p.request(ctx, http.MethodGet, parserPresetsEndpoint, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List returns the tenant's own parsers. The route emits a bare array; the
// result is wrapped in the named ParserListResponse so no caller handles a bare
// list.
func (p *ParserClient) List(ctx context.Context) (*ParserListResponse, error) {
	var parsers []ParserDefinition
	if err := p.request(ctx, http.MethodGet, parserEndpoint, nil, &parsers); err != nil {
		return nil, err
	}
	if parsers == nil {
		parsers = []ParserDefinition{}
	}
	return &ParserListResponse{Parsers: parsers}, nil
}

// Create adds a tenant-owned parser.
func (p *ParserClient) Create(ctx context.Context, req ParserSaveRequest) (*ParserDefinition, error) {
	if err := req.Framing.validate(); err != nil {
		return nil, err
	}
	var parser ParserDefinition
	if err := p.request(ctx, http.MethodPost, parserEndpoint, req, &parser); err != nil {
		return nil, err
	}
	return &parser, nil
}

// Update replaces a tenant-owned parser addressed by its explicit UUID.
func (p *ParserClient) Update(
	ctx context.Context, parserUUID string, req ParserSaveRequest,
) (*ParserDefinition, error) {
	if err := req.Framing.validate(); err != nil {
		return nil, err
	}
	var parser ParserDefinition
	path := parserEndpoint + "/" + url.PathEscape(parserUUID)
	if err := p.request(ctx, http.MethodPut, path, req, &parser); err != nil {
		return nil, err
	}
	return &parser, nil
}

// Delete removes a tenant-owned parser addressed by its explicit UUID.
func (p *ParserClient) Delete(ctx context.Context, parserUUID string) (*ParserDeleteResponse, error) {
	var deleted ParserDeleteResponse
	path := parserEndpoint + "/" + url.PathEscape(parserUUID)
	if err := p.request(ctx, http.MethodDelete, path, nil, &deleted); err != nil {
		return nil, err
	}
	return &deleted, nil
}

// Preview runs a parser definition against a bounded sample and returns the
// per-line result. It writes no parser and retains no sample.
func (p *ParserClient) Preview(
	ctx context.Context, req ParserPreviewRequest,
) (*ParserPreviewResponse, error) {
	if err := req.Parser.Framing.validate(); err != nil {
		return nil, err
	}
	var out ParserPreviewResponse
	if err := p.request(ctx, http.MethodPost, parserPreviewEndpoint, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (p *ParserClient) request(ctx context.Context, method, path string, payload any, out any) error {
	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("custd: marshal parser request: %w", err)
		}
	}
	if p.client.config.HTTPClient != nil {
		return p.requestViaDoer(method, path, body, out)
	}
	return p.requestViaHTTP(ctx, method, path, body, out)
}

func (p *ParserClient) requestViaDoer(method, path string, body []byte, out any) error {
	resp, err := p.client.config.HTTPClient.Do(&HTTPRequest{
		Method:  method,
		URL:     p.endpoint(path),
		Headers: p.client.headers(false),
		Body:    body,
	})
	if err != nil {
		return &RequestError{Message: "custd: parser request failed: " + err.Error(), Retryable: true, Cause: err}
	}
	if err := p.client.checkStatus(resp.StatusCode, resp.Body); err != nil {
		return err
	}
	return decodeParserResponse(resp.Body, out)
}

func (p *ParserClient) requestViaHTTP(ctx context.Context, method, path string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, p.endpoint(path), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("custd: create parser request: %w", err)
	}
	for k, v := range p.client.headers(false) {
		req.Header.Set(k, v)
	}
	resp, err := p.client.httpClient.Do(req)
	if err != nil {
		return &RequestError{Message: "custd: parser request failed: " + err.Error(), Retryable: true, Cause: err}
	}
	respBody, bodyErr := readResponseBody(resp.Body)
	if bodyErr != nil {
		return fmt.Errorf("custd: read parser response: %w", bodyErr)
	}
	if err := p.client.checkStatus(resp.StatusCode, respBody); err != nil {
		return err
	}
	return decodeParserResponse(respBody, out)
}

func (p *ParserClient) endpoint(path string) string {
	return strings.TrimRight(p.client.config.BaseURL, "/") + path
}

func decodeParserResponse(body []byte, out any) error {
	if out == nil || len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("custd: decode parser response: %w", err)
	}
	return nil
}
