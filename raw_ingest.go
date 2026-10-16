package custd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// rawIngestEndpoint is the tenant raw-line ingest route. It accepts one raw,
// line-delimited body and applies a parser to every line. Direct collectors
// already use this route; this client adds the declarative source target to it
// and changes nothing else about it.
const rawIngestEndpoint = "/api/v1/raw"

// IdempotencyKeyHeader is the standard request header that opts a raw upload
// into stable, retry-safe admission. Absent, the existing one-shot unkeyed
// behavior (partial 202, no safe replay) is unchanged.
const IdempotencyKeyHeader = "Idempotency-Key"

// rawIdempotencyKeyPattern is the canonical, lowercase UUIDv4 body; the version
// and RFC 4122 variant nibbles are pinned.
var rawIdempotencyKeyPattern = regexp.MustCompile(
	`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// RawRequestIdempotencyKey is a standard raw-upload idempotency key:
// "<unix-milliseconds>:<uuid-v4>". The same key and the same body/parser/target
// let the backend return the accepted receipt again instead of landing a second
// copy; a changed intent under the same key is refused with a conflict. It
// carries no payload or credential.
type RawRequestIdempotencyKey string

// NewRawRequestIdempotencyKey mints a fresh key from the current time and a
// random UUIDv4. Mint it ONCE per request object so a resend of the same request
// reuses the key; a genuine identical occurrence needs a new request object and
// therefore a new key.
func NewRawRequestIdempotencyKey() RawRequestIdempotencyKey {
	return RawRequestIdempotencyKey(fmt.Sprintf("%d:%s", time.Now().UnixMilli(), randomUUID()))
}

// Validate reports whether the key is "<unix-milliseconds>:<uuid-v4>" with a
// positive timestamp and a canonical lowercase v4 UUID.
func (k RawRequestIdempotencyKey) Validate() error {
	millisPart, uuidPart, ok := strings.Cut(string(k), ":")
	if !ok {
		return fmt.Errorf("custd: raw idempotency key must be <unix-milliseconds>:<uuid-v4>")
	}
	millis, err := strconv.ParseInt(millisPart, 10, 64)
	if err != nil || millis <= 0 || !rawIdempotencyKeyPattern.MatchString(uuidPart) {
		return fmt.Errorf("custd: raw idempotency key must be <unix-milliseconds>:<uuid-v4>")
	}
	return nil
}

// RawSourceTarget is the declarative default data-scope target for one raw
// upload. Both fields are registered immutable CODES supplied as a pair, never
// numeric platform identities: the backend fills a record that carries no
// canonical selection and refuses a record whose explicit selection disagrees.
type RawSourceTarget struct {
	SubClient string
	Resource  string
}

// IsZero reports that no target was declared.
func (t RawSourceTarget) IsZero() bool {
	return t.SubClient == "" && t.Resource == ""
}

// Validate rejects a partial or malformed target before a request is sent. Both
// codes must be present and well-formed, or neither. It is a cheap shape check;
// the backend remains the authority on whether the codes are registered.
func (t RawSourceTarget) Validate() error {
	if t.IsZero() {
		return nil
	}
	if !ValidDataScopeCode(t.SubClient) || !ValidDataScopeCode(t.Resource) {
		return fmt.Errorf("custd: raw source target requires subClient and resource as a pair of valid codes")
	}
	return nil
}

// RawIngestInput is the caller-supplied raw upload intent.
type RawIngestInput struct {
	// Content is the raw, line-delimited body to parse, sent verbatim.
	Content string
	// ParserUUID binds the parser executed for every line (X-Parser-UUID).
	ParserUUID string
	// Target is the optional declarative default data-scope target.
	Target *RawSourceTarget
}

// RawIngestRequest is one prepared POST /api/v1/raw upload: the intent plus the
// idempotency key retained for its lifetime. Build it with NewRawIngestRequest
// so the key is minted once; because it is a value, resending the same request
// reuses the key, while a genuinely repeated occurrence needs a new object.
type RawIngestRequest struct {
	// Content is the raw, line-delimited body to parse, sent verbatim.
	Content string
	// ParserUUID binds the parser executed for every line (X-Parser-UUID). The
	// existing parser-identifier binding is unchanged.
	ParserUUID string
	// Target is the optional declarative default data-scope target. Nil means no
	// declared target, so the backend refuses a record with no canonical
	// selection as missing.
	Target *RawSourceTarget
	// IdempotencyKey opts this upload into stable admission. Empty means the
	// existing one-shot unkeyed behavior with no safe replay.
	IdempotencyKey RawRequestIdempotencyKey
}

// NewRawIngestRequest mints one prepared request with a fresh idempotency key.
func NewRawIngestRequest(input RawIngestInput) RawIngestRequest {
	return RawIngestRequest{
		Content:        input.Content,
		ParserUUID:     input.ParserUUID,
		Target:         input.Target,
		IdempotencyKey: NewRawRequestIdempotencyKey(),
	}
}

// RawIngestResult is the POST /api/v1/raw acceptance body. The route always
// answers 202 Accepted, so the HTTP status alone is not per-record success:
// read Processed and Failed. A 202 with Failed > 0 is a partial failure. A body
// that is not an object carrying a boolean success and non-negative integer
// Processed and Failed is rejected rather than defaulted.
type RawIngestResult struct {
	Success   bool `json:"success"`
	Processed int  `json:"processed"`
	Failed    int  `json:"failed"`
}

// RawAcceptanceUnknownError reports that a raw upload's acceptance could not be
// determined: the connection failed before a response arrived, the 202
// acceptance body was missing or malformed, or the response carried a status
// that does not prove rejection. The request may or may not have landed, so the
// SDK does not retry it and does not recommend replaying the POST; a resubmission
// could double-land accepted records. When the request carried an idempotency
// key it is surfaced here as the original operation key, so a caller can retry
// with the SAME key once the backend dedup protocol is proven. Match with
// errors.As. The wrapped cause never carries the request payload or a token.
type RawAcceptanceUnknownError struct {
	Message string
	Cause   error
	// IdempotencyKey is the original operation key, when the request carried one.
	IdempotencyKey RawRequestIdempotencyKey
}

// Error renders the unknown-acceptance failure and, when present, its
// non-secret cause.
func (e *RawAcceptanceUnknownError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Unwrap exposes the non-secret cause for errors.Is/As; it never carries the
// request payload or a token.
func (e *RawAcceptanceUnknownError) Unwrap() error { return e.Cause }

func newRawAcceptanceUnknownError(cause error, key RawRequestIdempotencyKey) *RawAcceptanceUnknownError {
	return &RawAcceptanceUnknownError{
		Message:        "custd: raw upload acceptance is unknown; the request may or may not have landed and must not be replayed",
		Cause:          cause,
		IdempotencyKey: key,
	}
}

// IngestRawLines uploads raw line-delimited content for the bound parser. When
// a SourceTarget is supplied it appends the PAIRED subClient/resource query
// parameters to the existing raw endpoint; the parser-identifier header binding
// is otherwise unchanged. A non-empty IdempotencyKey is sent as the standard
// `Idempotency-Key` header, so a resend of the same request (same key, body,
// parser UUID and target codes) is deduplicated by the backend rather than
// landing twice.
//
// A transport failure or an unusable 202 body cannot prove whether the upload
// landed, so they surface as a *RawAcceptanceUnknownError rather than a
// retryable error: the SDK never retries this operation automatically. Do not
// treat the error as safe to replay until the backend's keyed dedup protocol is
// proven; the original operation key is carried on the error when present.
func (c *CustdClient) IngestRawLines(
	ctx context.Context, req RawIngestRequest,
) (*RawIngestResult, error) {
	path, err := rawIngestPath(req)
	if err != nil {
		return nil, err
	}
	if req.IdempotencyKey != "" {
		if err := req.IdempotencyKey.Validate(); err != nil {
			return nil, err
		}
	}
	headers := c.headers(false)
	headers["Content-Type"] = "text/plain; charset=utf-8"
	headers["X-Parser-UUID"] = req.ParserUUID
	if req.IdempotencyKey != "" {
		headers[IdempotencyKeyHeader] = string(req.IdempotencyKey)
	}

	if c.config.HTTPClient != nil {
		resp, err := c.config.HTTPClient.Do(&HTTPRequest{
			Method: http.MethodPost, URL: c.rawEndpoint(path), Headers: headers, Body: []byte(req.Content),
		})
		if err != nil {
			return nil, newRawAcceptanceUnknownError(err, req.IdempotencyKey)
		}
		if err := c.checkRawStatus(resp.StatusCode, resp.Body, req.IdempotencyKey); err != nil {
			return nil, err
		}
		return rawIngestResultFromBody(resp.Body, req.IdempotencyKey)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx, http.MethodPost, c.rawEndpoint(path), strings.NewReader(req.Content),
	)
	if err != nil {
		return nil, fmt.Errorf("custd: create raw ingest request: %w", err)
	}
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, newRawAcceptanceUnknownError(err, req.IdempotencyKey)
	}
	body, bodyErr := readResponseBody(resp.Body)
	if bodyErr != nil {
		return nil, newRawAcceptanceUnknownError(bodyErr, req.IdempotencyKey)
	}
	if err := c.checkRawStatus(resp.StatusCode, body, req.IdempotencyKey); err != nil {
		return nil, err
	}
	return rawIngestResultFromBody(body, req.IdempotencyKey)
}

// checkRawStatus classifies a raw response status. A status in the retryable
// set does not prove the upload was rejected, so it surfaces as an
// unknown-acceptance error rather than a retry hint; a non-retryable status
// (a 4xx rejected before processing) keeps the shared problem error.
func (c *CustdClient) checkRawStatus(statusCode int, body []byte, key RawRequestIdempotencyKey) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}
	if isRetryableStatus(statusCode, c.retrySet) {
		return newRawAcceptanceUnknownError(fmt.Errorf("custd: raw upload returned status %d", statusCode), key)
	}
	return c.checkStatus(statusCode, body)
}

// rawIngestResultFromBody decodes the acceptance body and, when it is empty or
// malformed, reports an unknown acceptance rather than an implied success.
func rawIngestResultFromBody(body []byte, key RawRequestIdempotencyKey) (*RawIngestResult, error) {
	result, err := decodeRawIngestResult(body)
	if err != nil {
		return nil, newRawAcceptanceUnknownError(err, key)
	}
	return result, nil
}

// rawIngestPath validates the client-supplied inputs and builds the request
// path. A partial or malformed target and a missing/malformed parser UUID are
// rejected locally rather than costing a round trip; no backend parsing is
// shadowed.
func rawIngestPath(req RawIngestRequest) (string, error) {
	if req.ParserUUID == "" {
		return "", fmt.Errorf("custd: raw ingest requires a parser UUID")
	}
	if !isDataScopeUUID(req.ParserUUID) {
		return "", fmt.Errorf("custd: raw ingest parser UUID must be a UUID")
	}
	if req.Target != nil {
		if err := req.Target.Validate(); err != nil {
			return "", err
		}
	}
	path := rawIngestEndpoint
	if req.Target != nil && !req.Target.IsZero() {
		params := url.Values{}
		params.Set("subClient", req.Target.SubClient)
		params.Set("resource", req.Target.Resource)
		path += "?" + params.Encode()
	}
	return path, nil
}

func (c *CustdClient) rawEndpoint(path string) string {
	return strings.TrimRight(c.config.BaseURL, "/") + path
}

// decodeRawIngestResult decodes the acceptance body strictly. The 202 status is
// checked before this; the body is what distinguishes accepted lines from
// refused ones, so an empty, null, or non-object body and a missing, null, or
// wrongly typed `success`/`processed`/`failed` are all errors rather than
// defaults. `processed` and `failed` must be non-negative integers; the
// legitimate partial counts the route emits are accepted unchanged.
func decodeRawIngestResult(data []byte) (*RawIngestResult, error) {
	const context = "raw ingest response"
	fields, err := decodeJSONObject(data, context)
	if err != nil {
		return nil, err
	}
	success, err := requireJSONBoolean(fields, context, "success")
	if err != nil {
		return nil, err
	}
	processed, err := requireJSONNonNegativeInteger(fields, context, "processed")
	if err != nil {
		return nil, err
	}
	failed, err := requireJSONNonNegativeInteger(fields, context, "failed")
	if err != nil {
		return nil, err
	}
	return &RawIngestResult{Success: success, Processed: processed, Failed: failed}, nil
}

// requireJSONBoolean decodes a required boolean field. A missing or null field,
// or a present number/string, is rejected rather than coerced.
func requireJSONBoolean(fields map[string]json.RawMessage, context, name string) (bool, error) {
	raw, ok := fields[name]
	if !ok || string(raw) == "null" {
		return false, validationErrorf("custd: %s field %s is required", context, name)
	}
	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		return false, validationWrap(fmt.Sprintf("custd: %s field %s must be a boolean", context, name), err)
	}
	return value, nil
}

// requireJSONNonNegativeInteger decodes a required non-negative integer field.
// A missing or null field, a fractional or non-numeric value, and a negative
// number are all rejected; no default is substituted for an omitted count.
func requireJSONNonNegativeInteger(fields map[string]json.RawMessage, context, name string) (int, error) {
	raw, ok := fields[name]
	if !ok || string(raw) == "null" {
		return 0, validationErrorf("custd: %s field %s is required", context, name)
	}
	// json.Number accepts a quoted numeric string, so require an actual JSON
	// number token before decoding; a string, boolean, or array is rejected.
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || (trimmed[0] != '-' && (trimmed[0] < '0' || trimmed[0] > '9')) {
		return 0, validationErrorf("custd: %s field %s must be an integer", context, name)
	}
	var number json.Number
	if err := json.Unmarshal(raw, &number); err != nil {
		return 0, validationWrap(fmt.Sprintf("custd: %s field %s must be an integer", context, name), err)
	}
	value, err := number.Int64()
	if err != nil || value < 0 {
		return 0, validationErrorf("custd: %s field %s must be a non-negative integer", context, name)
	}
	return int(value), nil
}
