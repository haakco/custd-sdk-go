package custd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const reportingEndpoint = "/api/v1/reporting"

type ReportingClient struct {
	client *CustdClient
}

type ReportingDashboard struct {
	Key     string            `json:"key"`
	Title   string            `json:"title"`
	Hidden  bool              `json:"hidden,omitempty"`
	Widgets []ReportingWidget `json:"widgets"`
}

type ReportingWidget struct {
	Key        string   `json:"key"`
	Title      string   `json:"title"`
	Kind       string   `json:"kind"`
	Template   string   `json:"template"`
	Metrics    []string `json:"metrics"`
	Dimensions []string `json:"dimensions,omitempty"`
}

type ReportingQueryRequest struct {
	Template   string            `json:"template"`
	Metrics    []string          `json:"metrics"`
	Dimensions []string          `json:"dimensions,omitempty"`
	Filters    []ReportingFilter `json:"filters,omitempty"`
	From       string            `json:"from,omitempty"`
	To         string            `json:"to,omitempty"`
	RangeDays  int               `json:"rangeDays,omitempty"`
	MaxRows    int               `json:"maxRows,omitempty"`
	CountOnly  bool              `json:"countOnly,omitempty"`
}

type ReportingFilter struct {
	Dimension string `json:"dimension"`
	Operator  string `json:"operator"`
	Value     string `json:"value,omitempty"`
}

type ReportingWidgetData struct {
	Buckets         []ReportingWidgetBucket `json:"buckets"`
	Count           int                     `json:"count"`
	Complete        bool                    `json:"complete"`
	Truncated       bool                    `json:"truncated"`
	QueryDurationMs int64                   `json:"queryDurationMs"`
	ParquetURICount int                     `json:"parquetUriCount,omitempty"`
	SnapshotAgeMs   int64                   `json:"snapshotAgeMs"`
	EventLagP95Ms   int64                   `json:"eventLagP95Ms"`
	DeltaCount      int                     `json:"deltaCount,omitempty"`
	DeltaPercent    float64                 `json:"deltaPercent,omitempty"`
	DeltaLabel      string                  `json:"deltaLabel,omitempty"`
	SecondaryLabel  string                  `json:"secondaryLabel,omitempty"`
	Trust           *ReportingTrust         `json:"trust,omitempty"`
}

type ReportingWidgetBucket struct {
	Date            string `json:"date"`
	Count           int    `json:"count"`
	Source          string `json:"source"`
	Complete        bool   `json:"complete"`
	QueryDurationMs int64  `json:"queryDurationMs"`
	ParquetURICount int    `json:"parquetUriCount,omitempty"`
	Message         string `json:"message,omitempty"`
	SecondaryCount  int    `json:"secondaryCount,omitempty"`
}

type ReportingTrust struct {
	Status            string   `json:"status"`
	DataFreshness     string   `json:"dataFreshness"`
	LastAwthyExport   string   `json:"lastAwthyExport,omitempty"`
	SchemaVersion     string   `json:"schemaVersion,omitempty"`
	ContractVersion   string   `json:"contractVersion,omitempty"`
	RollupState       string   `json:"rollupState"`
	QueryWarnings     []string `json:"queryWarnings,omitempty"`
	Coverage          string   `json:"coverage"`
	PermissionClass   string   `json:"permissionClass,omitempty"`
	DataSufficiency   string   `json:"dataSufficiency"`
	CaptureState      string   `json:"captureState"`
	ConsentState      string   `json:"consentState"`
	ExportState       string   `json:"exportState"`
	PartialReason     string   `json:"partialReason,omitempty"`
	UnavailableReason string   `json:"unavailableReason,omitempty"`
}

func newReportingClient(client *CustdClient) *ReportingClient {
	return &ReportingClient{client: client}
}

func (r *ReportingClient) Dashboard(ctx context.Context, key string) (*ReportingDashboard, error) {
	var out ReportingDashboard
	err := r.request(ctx, http.MethodGet, "/dashboards/"+url.PathEscape(key), nil, &out)
	return &out, err
}

func (r *ReportingClient) Query(ctx context.Context, req ReportingQueryRequest) (*ReportingWidgetData, error) {
	var out ReportingWidgetData
	err := r.request(ctx, http.MethodPost, "/query", req, &out)
	return &out, err
}

func (r *ReportingClient) request(ctx context.Context, method string, path string, payload any, out any) error {
	var body []byte
	var err error
	if payload != nil {
		body, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("custd: marshal reporting request: %w", err)
		}
	}
	if r.client.config.HTTPClient != nil {
		return r.requestViaDoer(method, path, body, out)
	}
	return r.requestViaHTTP(ctx, method, path, body, out)
}

func (r *ReportingClient) requestViaDoer(method string, path string, body []byte, out any) error {
	resp, err := r.client.config.HTTPClient.Do(&HTTPRequest{
		Method:  method,
		URL:     r.endpoint(path),
		Headers: r.client.headers(false),
		Body:    body,
	})
	if err != nil {
		return fmt.Errorf("custd: reporting request failed: %w", err)
	}
	if err := r.client.checkStatus(resp.StatusCode, resp.Body); err != nil {
		return err
	}
	return decodeReportingResponse(resp.Body, out)
}

func (r *ReportingClient) requestViaHTTP(ctx context.Context, method string, path string, body []byte, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, r.endpoint(path), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("custd: create reporting request: %w", err)
	}
	for k, v := range r.client.headers(false) {
		req.Header.Set(k, v)
	}
	resp, err := r.client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("custd: reporting request failed: %w", err)
	}
	// nolint:errcheck // response body fully read below; a close error cannot affect the already-read reporting response
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	if err := r.client.checkStatus(resp.StatusCode, respBody); err != nil {
		return err
	}
	return decodeReportingResponse(respBody, out)
}

func (r *ReportingClient) endpoint(path string) string {
	return strings.TrimRight(r.client.config.BaseURL, "/") + reportingEndpoint + path
}

func decodeReportingResponse(body []byte, out any) error {
	if out == nil || len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("custd: decode reporting response: %w", err)
	}
	return nil
}

func (t *ReportingTrust) UnmarshalJSON(data []byte) error {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if containsUnsafeReportingTrustKey(raw) {
		return fmt.Errorf("custd: unsafe reporting trust diagnostics")
	}
	type trustAlias ReportingTrust
	var decoded trustAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*t = ReportingTrust(decoded)
	return nil
}

func containsUnsafeReportingTrustKey(value any) bool {
	switch v := value.(type) {
	case []any:
		for _, item := range v {
			if containsUnsafeReportingTrustKey(item) {
				return true
			}
		}
	case map[string]any:
		for key, child := range v {
			if forbiddenReportingTrustKeys[strings.ToLower(key)] || containsUnsafeReportingTrustKey(child) {
				return true
			}
		}
	}
	return false
}

var forbiddenReportingTrustKeys = map[string]bool{
	"rawpayload": true,
	"sql":        true,
	"token":      true,
	"secret":     true,
	"stack":      true,
	"email":      true,
	"ipaddress":  true,
	"hostname":   true,
	"orderid":    true,
	"carttoken":  true,
}
