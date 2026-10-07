package custd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// usageCurrentTenantPath is the client-visible usage report for the token's own
// tenant. The system-admin /usage and /usage/export surfaces are deliberately
// not exposed here: a tenant client must not depend on system-admin filtering.
const usageCurrentTenantPath = "/usage/me"

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

// UsageAdminClient reads attributed usage for the authenticated tenant.
type UsageAdminClient struct {
	admin *AdminClient
}

// Get returns the attributed usage for the token's own tenant. The tenant is
// derived from the credential, so the caller never supplies a company slug.
func (c *UsageAdminClient) Get(ctx context.Context, query UsageQuery) (*UsageReport, error) {
	params, err := usageQueryParams(query)
	if err != nil {
		return nil, err
	}
	path := usageCurrentTenantPath
	if encoded := params.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var body json.RawMessage
	if err := c.admin.request(ctx, http.MethodGet, path, nil, &body); err != nil {
		return nil, err
	}
	return decodeUsageReport(body)
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
