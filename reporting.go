package custd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const reportingEndpoint = "/api/v1/reporting"

var subjectInsightTemplatePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,127}$`)
var preparedDataUUIDPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

type ReportingClient struct {
	client *CustdClient
}

type ReportingDashboard struct {
	Key            string            `json:"key"`
	Title          string            `json:"title"`
	Hidden         bool              `json:"hidden,omitempty"`
	DefaultRange   string            `json:"defaultRange"`
	RefreshSeconds int               `json:"refreshSeconds"`
	RequiredScopes []string          `json:"requiredScopes"`
	Widgets        []ReportingWidget `json:"widgets"`
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

type SubjectInsightRequest struct {
	Template  string `json:"template"`
	Subject   string `json:"subject"`
	From      string `json:"from,omitempty"`
	To        string `json:"to,omitempty"`
	RangeDays int    `json:"rangeDays,omitempty"`
}

type SubjectInsightResponse struct {
	Data RenderedWidgetData `json:"data"`
}

type PreparedDataStatus struct {
	TenantSlug      string                 `json:"tenantSlug"`
	ProcessingState string                 `json:"processingState"`
	Availability    string                 `json:"availability"`
	ObservedAt      string                 `json:"observedAt"`
	Watermark       string                 `json:"watermark,omitempty"`
	Provenance      PreparedDataProvenance `json:"provenance"`
	Retryability    string                 `json:"retryability"`
	Warnings        []PreparedDataWarning  `json:"warnings,omitempty"`
	NextAction      PreparedDataNextAction `json:"nextAction"`
}
type PreparedDataProvenance struct {
	Owner      string `json:"owner"`
	Generation string `json:"generation,omitempty"`
	SourceURN  string `json:"sourceUrn,omitempty"`
	Watermark  string `json:"watermark,omitempty"`
}
type PreparedDataWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type PreparedDataNextAction struct {
	Action           string `json:"action"`
	PollAfterSeconds int    `json:"pollAfterSeconds,omitempty"`
	MaxRetries       int    `json:"maxRetries,omitempty"`
}
type PreparedDataReceiptStatus struct {
	PreparedDataStatus
	ReceiptUUID string `json:"receiptUuid"`
}
type PreparedDataOutputStatus struct {
	PreparedDataStatus
	OutputUUID string `json:"outputUuid"`
}
type PreparedDataOutputList struct {
	Outputs []PreparedDataOutputStatus `json:"outputs"`
}
type PreparedDataQueryEnvelope struct {
	Output PreparedDataOutputStatus `json:"output"`
	Data   RenderedWidgetData       `json:"data"`
}

type RenderedWidgetData struct {
	Buckets         []RenderedWidgetBucket   `json:"buckets"`
	Value           RenderedMetricValue      `json:"value"`
	Metadata        *ReportingQueryMetadata  `json:"metadata,omitempty"`
	Sources         []ReportingSourceSummary `json:"sources,omitempty"`
	Warnings        []string                 `json:"warnings,omitempty"`
	QueryDurationMs int64                    `json:"queryDurationMs"`
	ParquetURICount int                      `json:"parquetUriCount,omitempty"`
	SnapshotAgeMs   int64                    `json:"snapshotAgeMs"`
	EventLagP95Ms   int64                    `json:"eventLagP95Ms"`
	Delta           *RenderedMetricValue     `json:"delta,omitempty"`
	DeltaPercent    float64                  `json:"deltaPercent,omitempty"`
	DeltaLabel      string                   `json:"deltaLabel,omitempty"`
	SecondaryLabel  string                   `json:"secondaryLabel,omitempty"`
	Trust           *RenderedReportingTrust  `json:"trust,omitempty"`
}

type RenderedWidgetBucket struct {
	Date            string               `json:"date"`
	Value           RenderedMetricValue  `json:"value"`
	Source          string               `json:"source"`
	QueryDurationMs int64                `json:"queryDurationMs"`
	ParquetURICount int                  `json:"parquetUriCount,omitempty"`
	Message         string               `json:"message,omitempty"`
	Secondary       *RenderedMetricValue `json:"secondary,omitempty"`
}

type RenderedMetricValue struct {
	Value           float64 `json:"value"`
	Unit            string  `json:"unit"`
	SampleCount     int     `json:"sampleCount"`
	DataSufficiency string  `json:"dataSufficiency"`
	Complete        bool    `json:"complete"`
	Truncated       bool    `json:"truncated,omitempty"`
}

type ReportingQueryMetadata struct {
	ResolvedTemplate string `json:"resolvedTemplate"`
	RangeStart       string `json:"rangeStart,omitempty"`
	RangeEnd         string `json:"rangeEnd,omitempty"`
	EffectiveMaxRows int    `json:"effectiveMaxRows"`
	ReturnedRows     int    `json:"returnedRows"`
	ReturnedBuckets  int    `json:"returnedBuckets"`
	CoveredWindows   int    `json:"coveredWindows"`
}

type ReportingSourceSummary struct {
	Kind          string `json:"kind"`
	Count         int    `json:"count"`
	CoverageStart string `json:"coverageStart,omitempty"`
	CoverageEnd   string `json:"coverageEnd,omitempty"`
	Completeness  string `json:"completeness"`
}

type RenderedReportingTrust struct {
	Status            string   `json:"status"`
	DataFreshness     string   `json:"dataFreshness"`
	LastExport        string   `json:"lastExport,omitempty"`
	SchemaVersion     string   `json:"schemaVersion,omitempty"`
	ContractVersion   string   `json:"contractVersion,omitempty"`
	RollupState       string   `json:"rollupState"`
	QueryWarnings     []string `json:"queryWarnings,omitempty"`
	Coverage          string   `json:"coverage"`
	PermissionClass   string   `json:"permissionClass,omitempty"`
	CaptureState      string   `json:"captureState"`
	ConsentState      string   `json:"consentState"`
	ExportState       string   `json:"exportState"`
	PartialReason     string   `json:"partialReason,omitempty"`
	UnavailableReason string   `json:"unavailableReason,omitempty"`
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

func (r *ReportingClient) SubjectInsight(ctx context.Context, req SubjectInsightRequest) (*SubjectInsightResponse, error) {
	if err := validateReportingSubjectInsightRequest(req); err != nil {
		return nil, err
	}
	var out SubjectInsightResponse
	err := r.request(ctx, http.MethodPost, "/insights/subject", req, &out)
	return &out, err
}

func (r *ReportingClient) Receipt(ctx context.Context, receiptUUID string) (*PreparedDataReceiptStatus, error) {
	if !preparedDataUUIDPattern.MatchString(receiptUUID) {
		return nil, fmt.Errorf("custd: receiptUuid must be a UUID")
	}
	var out PreparedDataReceiptStatus
	err := r.request(ctx, http.MethodGet, "/processing/"+url.PathEscape(receiptUUID), nil, &out)
	return &out, err
}
func (r *ReportingClient) Outputs(ctx context.Context) (*PreparedDataOutputList, error) {
	var out PreparedDataOutputList
	err := r.request(ctx, http.MethodGet, "/outputs", nil, &out)
	return &out, err
}
func (r *ReportingClient) Output(ctx context.Context, outputUUID string) (*PreparedDataOutputStatus, error) {
	if !preparedDataUUIDPattern.MatchString(outputUUID) {
		return nil, fmt.Errorf("custd: outputUuid must be a UUID")
	}
	var out PreparedDataOutputStatus
	err := r.request(ctx, http.MethodGet, "/outputs/"+url.PathEscape(outputUUID), nil, &out)
	return &out, err
}
func (r *ReportingClient) QueryOutput(ctx context.Context, outputUUID string, req ReportingQueryRequest) (*PreparedDataQueryEnvelope, error) {
	if !preparedDataUUIDPattern.MatchString(outputUUID) {
		return nil, fmt.Errorf("custd: outputUuid must be a UUID")
	}
	var out PreparedDataQueryEnvelope
	err := r.request(ctx, http.MethodPost, "/outputs/"+url.PathEscape(outputUUID)+"/query", req, &out)
	return &out, err
}

func validateReportingSubjectInsightRequest(req SubjectInsightRequest) error {
	if !subjectInsightTemplatePattern.MatchString(req.Template) {
		return fmt.Errorf("custd: reporting subject insight template is invalid")
	}
	if strings.TrimSpace(req.Subject) == "" || len(req.Subject) > 512 {
		return fmt.Errorf("custd: reporting subject insight subject is required")
	}
	hasRange := req.RangeDays != 0
	hasFrom := req.From != ""
	hasTo := req.To != ""
	if hasRange == (hasFrom || hasTo) {
		return fmt.Errorf("custd: reporting subject insight requires rangeDays or from and to")
	}
	if hasRange && (req.RangeDays < 1 || req.RangeDays > 366) {
		return fmt.Errorf("custd: reporting subject insight rangeDays must be between 1 and 366")
	}
	if !hasRange && (!hasFrom || !hasTo) {
		return fmt.Errorf("custd: reporting subject insight from and to are required together")
	}
	if !hasRange {
		from, fromErr := time.Parse(time.RFC3339, req.From)
		to, toErr := time.Parse(time.RFC3339, req.To)
		if fromErr != nil || toErr != nil || !to.After(from) || to.Sub(from) > 366*24*time.Hour {
			return fmt.Errorf("custd: reporting subject insight date range is invalid")
		}
	}
	return nil
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
	if strings.HasPrefix(path, "/processing/") {
		return strings.TrimRight(r.client.config.BaseURL, "/") + "/api/v1" + path
	}
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

func (r *SubjectInsightResponse) UnmarshalJSON(data []byte) error {
	type subjectInsightResponseAlias SubjectInsightResponse
	var decoded subjectInsightResponseAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	if raw, ok := fields["data"]; !ok || string(raw) == "null" {
		return fmt.Errorf("custd: subject insight response missing data")
	}
	*r = SubjectInsightResponse(decoded)
	return nil
}

func (d *RenderedWidgetData) UnmarshalJSON(data []byte) error {
	type renderedWidgetAlias RenderedWidgetData
	var decoded renderedWidgetAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, field := range []string{"buckets", "value", "queryDurationMs", "snapshotAgeMs", "eventLagP95Ms"} {
		if raw, ok := fields[field]; !ok || string(raw) == "null" {
			return fmt.Errorf("custd: rendered widget data missing %s", field)
		}
	}
	*d = RenderedWidgetData(decoded)
	return nil
}

func (b *RenderedWidgetBucket) UnmarshalJSON(data []byte) error {
	type renderedBucketAlias RenderedWidgetBucket
	var decoded renderedBucketAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, field := range []string{"date", "value", "source", "queryDurationMs"} {
		if raw, ok := fields[field]; !ok || string(raw) == "null" {
			return fmt.Errorf("custd: rendered widget bucket missing %s", field)
		}
	}
	*b = RenderedWidgetBucket(decoded)
	return nil
}

func (v *RenderedMetricValue) UnmarshalJSON(data []byte) error {
	type renderedValueAlias RenderedMetricValue
	var decoded renderedValueAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, field := range []string{"value", "unit", "sampleCount", "dataSufficiency", "complete"} {
		if raw, ok := fields[field]; !ok || string(raw) == "null" {
			return fmt.Errorf("custd: rendered metric value missing %s", field)
		}
	}
	*v = RenderedMetricValue(decoded)
	return nil
}

func (t *ReportingTrust) UnmarshalJSON(data []byte) error {
	if err := rejectUnsafeReportingTrust(data); err != nil {
		return err
	}
	type trustAlias ReportingTrust
	var decoded trustAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*t = ReportingTrust(decoded)
	return nil
}

func (t *RenderedReportingTrust) UnmarshalJSON(data []byte) error {
	if err := rejectUnsafeReportingTrust(data); err != nil {
		return err
	}
	type trustAlias RenderedReportingTrust
	var decoded trustAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*t = RenderedReportingTrust(decoded)
	if err := validateRenderedReportingTrustFields(data); err != nil {
		return err
	}
	return nil
}

func validateRenderedReportingTrustFields(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, field := range []string{
		"status", "dataFreshness", "rollupState", "coverage",
		"captureState", "consentState", "exportState",
	} {
		raw, ok := fields[field]
		if !ok || string(raw) == "null" {
			return fmt.Errorf("custd: rendered reporting trust missing %s", field)
		}
		var value string
		if err := json.Unmarshal(raw, &value); err != nil || value == "" {
			return fmt.Errorf("custd: rendered reporting trust missing %s", field)
		}
	}
	return nil
}

func (m *ReportingQueryMetadata) UnmarshalJSON(data []byte) error {
	type metadataAlias ReportingQueryMetadata
	var decoded metadataAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*m = ReportingQueryMetadata(decoded)
	if err := validateReportingQueryMetadataFields(data); err != nil {
		return err
	}
	return nil
}

func validateReportingQueryMetadataFields(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	raw, ok := fields["resolvedTemplate"]
	if !ok || string(raw) == "null" {
		return fmt.Errorf("custd: reporting query metadata missing resolvedTemplate")
	}
	var resolvedTemplate string
	if err := json.Unmarshal(raw, &resolvedTemplate); err != nil || resolvedTemplate == "" {
		return fmt.Errorf("custd: reporting query metadata missing resolvedTemplate")
	}
	for _, field := range []string{
		"effectiveMaxRows", "returnedRows", "returnedBuckets", "coveredWindows",
	} {
		raw, ok := fields[field]
		if !ok || string(raw) == "null" {
			return fmt.Errorf("custd: reporting query metadata missing %s", field)
		}
		var value int
		if err := json.Unmarshal(raw, &value); err != nil {
			return fmt.Errorf("custd: reporting query metadata missing %s", field)
		}
	}
	return nil
}

func (s *ReportingSourceSummary) UnmarshalJSON(data []byte) error {
	type sourceAlias ReportingSourceSummary
	var decoded sourceAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*s = ReportingSourceSummary(decoded)
	if err := validateReportingSourceSummaryFields(data); err != nil {
		return err
	}
	return nil
}

func validateReportingSourceSummaryFields(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, field := range []string{"kind", "completeness"} {
		raw, ok := fields[field]
		if !ok || string(raw) == "null" {
			return fmt.Errorf("custd: reporting source summary missing %s", field)
		}
		var value string
		if err := json.Unmarshal(raw, &value); err != nil || value == "" {
			return fmt.Errorf("custd: reporting source summary missing %s", field)
		}
	}
	raw, ok := fields["count"]
	if !ok || string(raw) == "null" {
		return fmt.Errorf("custd: reporting source summary missing count")
	}
	var count int
	if err := json.Unmarshal(raw, &count); err != nil {
		return fmt.Errorf("custd: reporting source summary missing count")
	}
	return nil
}

func rejectUnsafeReportingTrust(data []byte) error {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if containsUnsafeReportingTrustKey(raw) {
		return fmt.Errorf("custd: unsafe reporting trust diagnostics")
	}
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
