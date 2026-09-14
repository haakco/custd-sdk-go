package custd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// analyticsEventsEndpoint is the tenant event query. It reads a tenant's own ingested
// events, including payloads, behind the dedicated events.read scope. Effective-tenant
// authority is enforced server-side, so the caller never supplies a tenant slug.
const analyticsEventsEndpoint = "/api/v1/analytics/query"

// MaxAnalyticsLabelFilters is the most label filters the service accepts on one query.
const MaxAnalyticsLabelFilters = 4

// AnalyticsQuerySource names a source the server may answer a query from.
type AnalyticsQuerySource string

const (
	AnalyticsSourceAuto         AnalyticsQuerySource = "auto"
	AnalyticsSourcePostgres     AnalyticsQuerySource = "postgres"
	AnalyticsSourceDuckDB       AnalyticsQuerySource = "duckdb"
	AnalyticsSourceRollup       AnalyticsQuerySource = "rollup"
	AnalyticsSourceMaterialized AnalyticsQuerySource = "materialized"
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
	if err := a.request(ctx, http.MethodPost, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
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

func (a *AnalyticsEventClient) request(ctx context.Context, method string, payload any, out any) error {
	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("custd: marshal analytics request: %w", err)
		}
	}
	if a.client.config.HTTPClient != nil {
		return a.requestViaDoer(method, body, out)
	}
	return a.requestViaHTTP(ctx, method, body, out)
}

func (a *AnalyticsEventClient) requestViaDoer(method string, body []byte, out any) error {
	resp, err := a.client.config.HTTPClient.Do(&HTTPRequest{
		Method:  method,
		URL:     a.endpoint(),
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

func (a *AnalyticsEventClient) requestViaHTTP(ctx context.Context, method string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, a.endpoint(), bytes.NewReader(body))
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

func (a *AnalyticsEventClient) endpoint() string {
	return strings.TrimRight(a.client.config.BaseURL, "/") + analyticsEventsEndpoint
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
