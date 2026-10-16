package custd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// analyticsEventsEndpoint is the tenant event query. It reads a tenant's own ingested
// events, including payloads, behind the dedicated events.read scope. Effective-tenant
// authority is enforced server-side, so the caller never supplies a tenant slug.
const analyticsEventsEndpoint = "/api/v1/analytics/query"

// analyticsEventsRangeEndpoint is the tenant event range query. It shares the tenant
// and scope rules of analyticsEventsEndpoint; only the addressable window differs.
const analyticsEventsRangeEndpoint = "/api/v1/analytics/query-range"

// MaxAnalyticsLabelFilters is the most label filters the service accepts on one query.
const MaxAnalyticsLabelFilters = 4

// AnalyticsMaxRangeDays is the longest inclusive range the service accepts. A longer
// request is rejected locally rather than spending a round trip on a 400.
const AnalyticsMaxRangeDays = 120

// AnalyticsRangeGroupByDay is the only bucket granularity the service supports today; a
// request that omits groupBy is accepted and means the same thing.
const AnalyticsRangeGroupByDay = "day"

// AnalyticsQuerySource names a source the server may answer a query from.
type AnalyticsQuerySource string

const (
	AnalyticsSourceAuto         AnalyticsQuerySource = "auto"
	AnalyticsSourcePostgres     AnalyticsQuerySource = "postgres"
	AnalyticsSourceDuckDB       AnalyticsQuerySource = "duckdb"
	AnalyticsSourceRollup       AnalyticsQuerySource = "rollup"
	AnalyticsSourceMaterialized AnalyticsQuerySource = "materialized"
)

// AnalyticsRangeQuerySource names a source a range query may select. The range
// API accepts only omitted/empty, auto, or duckdb; the single-day query's
// retired postgres, rollup, and materialized sources are not part of the range
// contract, so they are rejected before a request is sent.
type AnalyticsRangeQuerySource string

const (
	AnalyticsRangeSourceAuto   AnalyticsRangeQuerySource = "auto"
	AnalyticsRangeSourceDuckDB AnalyticsRangeQuerySource = "duckdb"
)

// AnalyticsLabelFilter is one exact tenant-vocabulary key/value filter.
type AnalyticsLabelFilter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// AnalyticsEventQueryRequest is the public query body.
//
// The service also accepts an internal anonymousId exact-subject predicate and a
// countOnly flag, but both are absent from its public JSON contract. They are therefore
// not representable here, which keeps them off the wire by construction.
type AnalyticsEventQueryRequest struct {
	// Date is the UTC day to query, formatted YYYY-MM-DD. The service requires it.
	Date string `json:"date"`
	// EventType is an exact event-type slug.
	EventType string `json:"eventType,omitempty"`
	// Limit caps returned rows. The service clamps it and reports the applied count.
	Limit  int                  `json:"limit,omitempty"`
	Source AnalyticsQuerySource `json:"source,omitempty"`
	// LabelFilters accepts at most MaxAnalyticsLabelFilters entries.
	LabelFilters []AnalyticsLabelFilter `json:"labelFilters,omitempty"`
}

// AnalyticsEventRow is one tenant event.
//
// The wire shape is a column bag taken from the underlying parquet schema rather than a
// fixed struct: the service serialises whatever the writer produced. Decoding into a map
// surfaces new columns without an SDK change.
type AnalyticsEventRow map[string]any

// AnalyticsEventSourceSummary is one source's contribution to a query.
//
// Complete and Fresh are the server's own assessment of whether that source answered the
// whole request and whether its data sat inside the freshness window. They are the
// authority for how far a result may be trusted; the SDK must not substitute its own
// heuristic.
type AnalyticsEventSourceSummary struct {
	Name            AnalyticsQuerySource `json:"name"`
	Count           int                  `json:"count"`
	ParquetURICount int                  `json:"parquetUriCount,omitempty"`
	Complete        bool                 `json:"complete"`
	Fresh           bool                 `json:"fresh"`
	QueryDurationMs int64                `json:"queryDurationMs"`
	FreshnessLagMs  int64                `json:"freshnessLagMs"`
	Message         string               `json:"message,omitempty"`
}

// AnalyticsEventTiming is the server-measured lag and coverage for a query.
type AnalyticsEventTiming struct {
	EventLagP50Ms        int64  `json:"eventLagP50Ms"`
	EventLagP95Ms        int64  `json:"eventLagP95Ms"`
	EventLagMaxMs        int64  `json:"eventLagMaxMs"`
	QueryDurationMs      int64  `json:"queryDurationMs"`
	OldestEventTimestamp string `json:"oldestEventTimestamp,omitempty"`
	NewestEventTimestamp string `json:"newestEventTimestamp,omitempty"`
	SnapshotAgeMs        int64  `json:"snapshotAgeMs"`
}

// AnalyticsEventQueryResponse is the query result, surfaced verbatim.
type AnalyticsEventQueryResponse struct {
	Results []AnalyticsEventRow           `json:"results"`
	Count   int                           `json:"count"`
	Sources []AnalyticsEventSourceSummary `json:"sources"`
	Timing  AnalyticsEventTiming          `json:"timing"`
}

// AnalyticsEventRangeQueryRequest is the public range query body.
//
// The service also accepts an internal anonymousId exact-subject predicate and a
// countOnly flag, but both are absent from its public JSON contract. They are therefore
// not representable here, which keeps them off the wire by construction.
type AnalyticsEventRangeQueryRequest struct {
	// From and To bound the inclusive range, formatted YYYY-MM-DD. Both are required.
	From string `json:"from"`
	To   string `json:"to"`
	// EventType is an exact event-type slug.
	EventType string `json:"eventType,omitempty"`
	// Limit caps returned rows across the whole range. The service clamps it and
	// reports the applied count.
	Limit  int                       `json:"limit,omitempty"`
	Source AnalyticsRangeQuerySource `json:"source,omitempty"`
	// GroupBy selects the bucket granularity; AnalyticsRangeGroupByDay is the only
	// accepted value.
	GroupBy string `json:"groupBy,omitempty"`
	// LabelFilters accepts at most MaxAnalyticsLabelFilters entries.
	LabelFilters []AnalyticsLabelFilter `json:"labelFilters,omitempty"`
}

// AnalyticsEventRangeBucket is one day's coverage inside a range query.
//
// Complete is the server's own assessment of whether that day's source answered the
// whole request, so it is surfaced unchanged; the SDK must not substitute its own
// heuristic.
type AnalyticsEventRangeBucket struct {
	Date            string               `json:"date"`
	Count           int                  `json:"count"`
	Source          AnalyticsQuerySource `json:"source"`
	Complete        bool                 `json:"complete"`
	QueryDurationMs int64                `json:"queryDurationMs"`
	ParquetURICount int                  `json:"parquetUriCount,omitempty"`
	Message         string               `json:"message,omitempty"`
}

// AnalyticsEventRangeQueryResponse is the range query result, surfaced verbatim.
type AnalyticsEventRangeQueryResponse struct {
	Rows    []AnalyticsEventRow           `json:"rows"`
	Count   int                           `json:"count"`
	Buckets []AnalyticsEventRangeBucket   `json:"buckets"`
	Sources []AnalyticsEventSourceSummary `json:"sources"`
	Timing  AnalyticsEventTiming          `json:"timing"`
}

// AnalyticsEventClient queries this tenant's own events.
type AnalyticsEventClient struct {
	client *CustdClient
}

func newAnalyticsEventClient(client *CustdClient) *AnalyticsEventClient {
	return &AnalyticsEventClient{client: client}
}

// Query reads this tenant's own events for a single UTC day.
func (a *AnalyticsEventClient) Query(ctx context.Context, req AnalyticsEventQueryRequest) (*AnalyticsEventQueryResponse, error) {
	if err := validateAnalyticsEventQuery(req); err != nil {
		return nil, err
	}
	var out AnalyticsEventQueryResponse
	if err := a.request(ctx, http.MethodPost, analyticsEventsEndpoint, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// QueryRange reads this tenant's own events across an inclusive date range.
func (a *AnalyticsEventClient) QueryRange(
	ctx context.Context,
	req AnalyticsEventRangeQueryRequest,
) (*AnalyticsEventRangeQueryResponse, error) {
	if err := validateAnalyticsEventRangeQuery(req); err != nil {
		return nil, err
	}
	var body json.RawMessage
	if err := a.request(ctx, http.MethodPost, analyticsEventsRangeEndpoint, req, &body); err != nil {
		return nil, err
	}
	return decodeAnalyticsRangeResponse(body)
}

// validateAnalyticsEventQuery rejects locally what the service would reject anyway, so an
// over-limit request never costs a round trip.
func validateAnalyticsEventQuery(req AnalyticsEventQueryRequest) error {
	if req.Date == "" {
		return fmt.Errorf("custd: analytics query requires a date (YYYY-MM-DD)")
	}
	if len(req.LabelFilters) > MaxAnalyticsLabelFilters {
		return fmt.Errorf(
			"custd: analytics query accepts at most %d label filters, received %d",
			MaxAnalyticsLabelFilters,
			len(req.LabelFilters),
		)
	}
	return nil
}

// validateAnalyticsEventRangeQuery rejects locally what the service would reject anyway,
// so an over-long or malformed window never costs a round trip.
func validateAnalyticsEventRangeQuery(req AnalyticsEventRangeQueryRequest) error {
	if req.From == "" || req.To == "" {
		return fmt.Errorf("custd: analytics range query requires from and to (YYYY-MM-DD)")
	}
	from, err := time.Parse("2006-01-02", req.From)
	if err != nil {
		return fmt.Errorf("custd: analytics range query from must be YYYY-MM-DD: %w", err)
	}
	to, err := time.Parse("2006-01-02", req.To)
	if err != nil {
		return fmt.Errorf("custd: analytics range query to must be YYYY-MM-DD: %w", err)
	}
	if to.Before(from) {
		return fmt.Errorf("custd: analytics range query to must not be before from")
	}
	if days := int(to.Sub(from).Hours()/24) + 1; days > AnalyticsMaxRangeDays {
		return fmt.Errorf(
			"custd: analytics range query spans %d days, the maximum is %d",
			days, AnalyticsMaxRangeDays,
		)
	}
	if len(req.LabelFilters) > MaxAnalyticsLabelFilters {
		return fmt.Errorf(
			"custd: analytics range query accepts at most %d label filters, received %d",
			MaxAnalyticsLabelFilters,
			len(req.LabelFilters),
		)
	}
	if req.GroupBy != "" && req.GroupBy != AnalyticsRangeGroupByDay {
		return fmt.Errorf("custd: analytics range query groupBy must be %q", AnalyticsRangeGroupByDay)
	}
	switch req.Source {
	case "", AnalyticsRangeSourceAuto, AnalyticsRangeSourceDuckDB:
	default:
		return fmt.Errorf(
			"custd: analytics range query source must be %q or %q",
			AnalyticsRangeSourceAuto,
			AnalyticsRangeSourceDuckDB,
		)
	}
	return nil
}

func (a *AnalyticsEventClient) request(ctx context.Context, method, path string, payload any, out any) error {
	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("custd: marshal analytics request: %w", err)
		}
	}
	if a.client.config.HTTPClient != nil {
		return a.requestViaDoer(method, path, body, out)
	}
	return a.requestViaHTTP(ctx, method, path, body, out)
}

func (a *AnalyticsEventClient) requestViaDoer(method, path string, body []byte, out any) error {
	resp, err := a.client.config.HTTPClient.Do(&HTTPRequest{
		Method:  method,
		URL:     a.endpoint(path),
		Headers: a.client.headers(false),
		Body:    body,
	})
	if err != nil {
		return &RequestError{Message: "custd: analytics request failed: " + err.Error(), Retryable: true, Cause: err}
	}
	if err := a.client.checkStatus(resp.StatusCode, resp.Body); err != nil {
		return err
	}
	return decodeAnalyticsResponse(resp.Body, out)
}

func (a *AnalyticsEventClient) requestViaHTTP(ctx context.Context, method, path string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, a.endpoint(path), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("custd: create analytics request: %w", err)
	}
	for k, v := range a.client.headers(false) {
		req.Header.Set(k, v)
	}
	resp, err := a.client.httpClient.Do(req)
	if err != nil {
		return &RequestError{Message: "custd: analytics request failed: " + err.Error(), Retryable: true, Cause: err}
	}
	respBody, bodyErr := readResponseBody(resp.Body)
	if bodyErr != nil {
		return fmt.Errorf("custd: read analytics response: %w", bodyErr)
	}
	if err := a.client.checkStatus(resp.StatusCode, respBody); err != nil {
		return err
	}
	return decodeAnalyticsResponse(respBody, out)
}

func (a *AnalyticsEventClient) endpoint(path string) string {
	return strings.TrimRight(a.client.config.BaseURL, "/") + path
}

func decodeAnalyticsResponse(body []byte, out any) error {
	if out == nil || len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("custd: decode analytics response: %w", err)
	}
	return nil
}

// decodeAnalyticsRangeResponse decodes and validates the POST /query-range
// response. The collections, timing, and per-source/per-bucket completeness
// flags are required so a partial success body cannot pass as a complete range.
// parquetUriCount, message, and the two event timestamps are the optional wire
// fields the owner emits with omitempty.
func decodeAnalyticsRangeResponse(data []byte) (*AnalyticsEventRangeQueryResponse, error) {
	const context = "analytics range response"
	fields, err := decodeJSONObject(data, context)
	if err != nil {
		return nil, err
	}
	if err := requireJSONFields(fields, context, "count"); err != nil {
		return nil, err
	}
	if err := requireJSONObjectList(fields, context, "rows"); err != nil {
		return nil, err
	}
	if err := requireJSONObjectList(fields, context, "buckets",
		"date", "count", "source", "complete", "queryDurationMs",
	); err != nil {
		return nil, err
	}
	if err := requireJSONObjectList(fields, context, "sources",
		"name", "count", "complete", "fresh", "queryDurationMs", "freshnessLagMs",
	); err != nil {
		return nil, err
	}
	if err := requireJSONObject(fields, context, "timing",
		"eventLagP50Ms", "eventLagP95Ms", "eventLagMaxMs", "queryDurationMs", "snapshotAgeMs",
	); err != nil {
		return nil, err
	}
	var response AnalyticsEventRangeQueryResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, validationWrap(fmt.Sprintf("custd: decode %s", context), err)
	}
	return &response, nil
}
