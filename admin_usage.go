package custd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// usageCurrentTenantPath is the client-visible usage report for the token's own
// tenant. The system-admin surfaces (list, export) have their own methods so a
// caller can tell a tenant-scoped read from a filtered one.
const usageCurrentTenantPath = "/usage/me"

// The system-admin usage surfaces. They are relative to the admin endpoint, so
// a call to /usage reaches /api/v1/admin/usage.
const (
	// adminUsageListPath lists reconciled usage across the companies a system
	// administrator may see.
	adminUsageListPath = "/usage"
	// adminUsageExportPath exports the same reconciled usage as the list route.
	adminUsageExportPath = "/usage/export"
	// adminRetainedUsagePath reports the current tenant's retained storage.
	adminRetainedUsagePath = "/storage-usage/me"
)

// UsageDefaultLimit and UsageMaxLimit bound one usage request's row count. The
// service applies the default when the request omits a limit and rejects a
// value outside [1, UsageMaxLimit] before loading anything.
const (
	UsageDefaultLimit = 500
	UsageMaxLimit     = 5000
)

// UsageCompletenessState names how final one usage row is. The server derives
// it and owns its meaning; the SDK surfaces it unchanged.
type UsageCompletenessState string

const (
	UsageCompletenessProvisional UsageCompletenessState = "provisional"
	UsageCompletenessFinal       UsageCompletenessState = "final"
	UsageCompletenessCorrected   UsageCompletenessState = "corrected"
	UsageCompletenessIncomplete  UsageCompletenessState = "incomplete"
)

// UsageRow is one attributed usage quantity for one half-open UTC window.
type UsageRow struct {
	AccountCompanySlug   string                 `json:"accountCompanySlug"`
	DataSpaceCompanySlug string                 `json:"dataSpaceCompanySlug"`
	MeterSlug            string                 `json:"meterSlug"`
	MeterVersion         int                    `json:"meterVersion"`
	Unit                 string                 `json:"unit"`
	WindowStart          string                 `json:"windowStart"`
	WindowEnd            string                 `json:"windowEnd"`
	Quantity             int64                  `json:"quantity"`
	SourceWatermark      int64                  `json:"sourceWatermark"`
	CompletenessState    UsageCompletenessState `json:"completenessState"`
	CorrectionGeneration int                    `json:"correctionGeneration"`
	CalculationVersion   int                    `json:"calculationVersion"`
}

// UsageTotal is a meter total across every row returned for the window.
type UsageTotal struct {
	AccountCompanySlug   string `json:"accountCompanySlug"`
	DataSpaceCompanySlug string `json:"dataSpaceCompanySlug"`
	MeterSlug            string `json:"meterSlug"`
	Unit                 string `json:"unit"`
	Quantity             int64  `json:"quantity"`
}

// UsageReport is the attributed usage for the token's own tenant.
//
// ContainsProvisional and ContainsIncomplete are the server's own assessment of
// the returned rows; a caller deciding whether a number is final must read them
// rather than assume every row is settled.
type UsageReport struct {
	SchemaVersion       string       `json:"schemaVersion"`
	CompanySlug         string       `json:"companySlug,omitempty"`
	Start               string       `json:"start"`
	End                 string       `json:"end"`
	Rows                []UsageRow   `json:"rows"`
	Totals              []UsageTotal `json:"totals"`
	SourceWatermark     int64        `json:"sourceWatermark"`
	ContainsProvisional bool         `json:"containsProvisional"`
	ContainsIncomplete  bool         `json:"containsIncomplete"`
}

// UsageQuery narrows and bounds the usage window.
//
// A zero Start and End use the service default, the trailing 30 days ending
// now; a zero Limit uses UsageDefaultLimit.
type UsageQuery struct {
	MeterSlug string
	Start     time.Time
	End       time.Time
	Limit     int
}

// RetainedUsage is the current tenant's retained storage position.
//
// It reports what the tenant holds, not what the platform is charged for: the
// limit is enforced on the attributed footprint, which this read does not
// carry.
type RetainedUsage struct {
	CompanySlug          string                        `json:"companySlug"`
	CompanyName          string                        `json:"companyName"`
	MeasuredAt           string                        `json:"measuredAt"`
	RetainedLogicalBytes int64                         `json:"retainedLogicalBytes"`
	RetainedEventCount   int64                         `json:"retainedEventCount"`
	MeasurementState     RetainedUsageMeasurementState `json:"measurementState"`
}

// RetainedUsageMeasurementState is the provenance of one retained-usage
// reading: the watermark it was measured at and how final that measurement is.
type RetainedUsageMeasurementState struct {
	SourceWatermark    int64  `json:"sourceWatermark"`
	CompletenessState  string `json:"completenessState"`
	CalculationVersion int    `json:"calculationVersion"`
}

// UsageExport is a streamed usage export.
//
// The export is CSV rather than a JSON document, so the body is a stream the
// caller closes instead of a decoded value.
type UsageExport struct {
	// Body carries the exported rows. The caller closes it.
	Body io.ReadCloser
	// Metadata carries the server's request identifier.
	Metadata ResponseMetadata
}

// UsageAdminClient reads attributed usage for the authenticated tenant.
type UsageAdminClient struct {
	admin *AdminClient
}

// Get returns the attributed usage for the token's own tenant. The tenant is
// derived from the credential, so the caller never supplies a company slug.
//
// It is the tenant-facing name for the same route ListCurrentUsage reads.
func (c *UsageAdminClient) Get(ctx context.Context, query UsageQuery) (*UsageReport, error) {
	return c.ListCurrentUsage(ctx, query)
}

// ListCurrentUsage reads the reconciled usage of the tenant the credential
// belongs to. The tenant is derived from the credential, so the caller never
// supplies a company slug.
func (c *UsageAdminClient) ListCurrentUsage(ctx context.Context, query UsageQuery) (*UsageReport, error) {
	path, err := adminUsagePath(usageCurrentTenantPath, "", query)
	if err != nil {
		return nil, err
	}
	return c.readUsageReport(ctx, path)
}

// ListUsage reads the reconciled usage a system administrator may see.
//
// companySlug narrows the read to one company; an empty slug reads every
// company the credential may see. The window and meter come from query.
func (c *UsageAdminClient) ListUsage(
	ctx context.Context, companySlug string, query UsageQuery,
) (*UsageReport, error) {
	path, err := adminUsagePath(adminUsageListPath, companySlug, query)
	if err != nil {
		return nil, err
	}
	return c.readUsageReport(ctx, path)
}

// GetRetainedUsage reads the current tenant's retained storage position.
func (c *UsageAdminClient) GetRetainedUsage(ctx context.Context) (*RetainedUsage, error) {
	var body json.RawMessage
	if err := c.admin.request(ctx, http.MethodGet, adminRetainedUsagePath, nil, &body); err != nil {
		return nil, err
	}
	return decodeRetainedUsage(body)
}

// ExportUsage streams the reconciled usage as CSV.
//
// The export is CSV rather than a JSON document, so the body is a stream the
// caller closes instead of a decoded value. The request always asks for the
// CSV format the admin API serves for this route.
func (c *UsageAdminClient) ExportUsage(
	ctx context.Context, companySlug string, query UsageQuery,
) (*UsageExport, error) {
	path, err := adminUsagePath(adminUsageExportPath, companySlug, query)
	if err != nil {
		return nil, err
	}
	return c.openUsageExport(ctx, path)
}

// readUsageReport reads one usage report and decodes it.
func (c *UsageAdminClient) readUsageReport(ctx context.Context, path string) (*UsageReport, error) {
	var body json.RawMessage
	if err := c.admin.request(ctx, http.MethodGet, path, nil, &body); err != nil {
		return nil, err
	}
	return decodeUsageReport(body)
}

// adminUsagePath builds an admin usage path with its query string.
//
// The parameter names are the ones the admin API reads: companySlug, meter,
// start, end and limit. A company slug is only sent when the caller names
// one, because the current-tenant route derives the company from the
// credential.
func adminUsagePath(path, companySlug string, query UsageQuery) (string, error) {
	params, err := usageQueryParams(query)
	if err != nil {
		return "", err
	}
	if companySlug != "" {
		params.Set(companySlugQueryParameter, companySlug)
	}
	if encoded := params.Encode(); encoded != "" {
		return path + "?" + encoded, nil
	}
	return path, nil
}

// decodeUsageReport decodes and validates the GET /usage/me response. The
// completeness flags and both collections are required: a report that omits them
// is rejected rather than becoming a zero-valued report. Only companySlug is an
// optional wire field, because the owner emits it with omitempty.
func decodeUsageReport(data []byte) (*UsageReport, error) {
	const context = "usage response"
	fields, err := decodeJSONObject(data, context)
	if err != nil {
		return nil, err
	}
	if err := requireJSONFields(fields, context,
		"schemaVersion", "start", "end", "sourceWatermark",
		"containsProvisional", "containsIncomplete",
	); err != nil {
		return nil, err
	}
	if err := requireJSONObjectList(fields, context, "rows",
		"accountCompanySlug", "dataSpaceCompanySlug", "meterSlug", "meterVersion", "unit",
		"windowStart", "windowEnd", "quantity", "sourceWatermark", "completenessState",
		"correctionGeneration", "calculationVersion",
	); err != nil {
		return nil, err
	}
	if err := requireJSONObjectList(fields, context, "totals",
		"accountCompanySlug", "dataSpaceCompanySlug", "meterSlug", "unit", "quantity",
	); err != nil {
		return nil, err
	}
	var report UsageReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, validationWrap(fmt.Sprintf("custd: decode %s", context), err)
	}
	return &report, nil
}

// usageQueryParams builds the query string and rejects locally what the service
// would reject anyway, so an invalid window never costs a round trip.
func usageQueryParams(query UsageQuery) (url.Values, error) {
	if !query.Start.IsZero() && !query.End.IsZero() && !query.Start.Before(query.End) {
		return nil, fmt.Errorf("custd: usage start must be before end")
	}
	if query.Limit != 0 && (query.Limit < 1 || query.Limit > UsageMaxLimit) {
		return nil, fmt.Errorf("custd: usage limit must be between 1 and %d", UsageMaxLimit)
	}
	params := url.Values{}
	if query.MeterSlug != "" {
		params.Set("meter", query.MeterSlug)
	}
	if !query.Start.IsZero() {
		params.Set("start", query.Start.UTC().Format(time.RFC3339Nano))
	}
	if !query.End.IsZero() {
		params.Set("end", query.End.UTC().Format(time.RFC3339Nano))
	}
	if query.Limit != 0 {
		params.Set("limit", strconv.Itoa(query.Limit))
	}
	return params, nil
}

// decodeRetainedUsage decodes the retained-storage position. The measurement state is
// required rather than optional: a caller has to be able to tell a final reading from a
// provisional one, and a missing state would read as final.
func decodeRetainedUsage(data []byte) (*RetainedUsage, error) {
	const context = "retained usage response"
	fields, err := decodeJSONObject(data, context)
	if err != nil {
		return nil, err
	}
	if err := requireJSONFields(fields, context,
		"companySlug", "companyName", "measuredAt",
		"retainedLogicalBytes", "retainedEventCount", "measurementState",
	); err != nil {
		return nil, err
	}
	var usage RetainedUsage
	if err := json.Unmarshal(data, &usage); err != nil {
		return nil, validationWrap(fmt.Sprintf("custd: decode %s", context), err)
	}
	return &usage, nil
}

// openUsageExport starts a CSV export and hands back the unread body, so a large export
// costs one stream rather than one whole document in memory.
func (c *UsageAdminClient) openUsageExport(ctx context.Context, path string) (*UsageExport, error) {
	if c.admin.client.config.HTTPClient != nil {
		return c.openUsageExportViaDoer(ctx, path)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.admin.endpoint(path), nil)
	if err != nil {
		return nil, fmt.Errorf("custd: create usage export request: %w", err)
	}
	for key, value := range c.admin.client.headers(false) {
		req.Header.Set(key, value)
	}
	resp, err := c.admin.client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("custd: usage export failed: %w", err)
	}
	return c.streamUsageExport(ctx, resp)
}

// openUsageExportViaDoer answers from a caller-supplied doer, which returns a complete body
// rather than a stream, so the stream is that body.
func (c *UsageAdminClient) openUsageExportViaDoer(
	ctx context.Context, path string,
) (*UsageExport, error) {
	resp, err := c.admin.client.config.HTTPClient.Do(&HTTPRequest{
		Method: http.MethodGet, URL: c.admin.endpoint(path), Headers: c.admin.client.headers(false),
	})
	if err != nil {
		return nil, fmt.Errorf("custd: usage export failed: %w", err)
	}
	if err := c.admin.client.checkStatus(resp.StatusCode, resp.Body); err != nil {
		return nil, err
	}
	requestID := resp.Headers[requestIDHeader]
	recordMetadata(ctx, requestID)
	return &UsageExport{
		Body:     io.NopCloser(bytes.NewReader(resp.Body)),
		Metadata: ResponseMetadata{RequestID: requestID},
	}, nil
}

// streamUsageExport checks the answer before handing the body over. A refused export is read
// into memory so the caller sees the server's reason instead of an unread stream.
func (c *UsageAdminClient) streamUsageExport(
	ctx context.Context, resp *http.Response,
) (*UsageExport, error) {
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, readErr := readResponseBody(resp.Body)
		closeErr := resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("custd: read usage export response: %w", readErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("custd: close usage export response: %w", closeErr)
		}
		return nil, c.admin.client.checkStatus(resp.StatusCode, body)
	}
	requestID := resp.Header.Get(requestIDHeader)
	recordMetadata(ctx, requestID)
	return &UsageExport{Body: resp.Body, Metadata: ResponseMetadata{RequestID: requestID}}, nil
}
