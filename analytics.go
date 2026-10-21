package custd

import (
	"context"
	"fmt"
	"net/http"
)

// analyticsQueryPath is the tenant-facing query route. It sits under /api/v1/analytics rather
// than /api/v1/admin, so it is issued through the non-admin path builder and the tenant comes
// from the credential rather than from a slug the caller supplies.
const analyticsQueryPath = "/analytics/query"

// AnalyticsEventLabelFilter is one exact label predicate on an event query.
type AnalyticsEventLabelFilter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// AnalyticsEventQuery asks how many events match. Only the fields the server reads are
// modelled, so a change on the server surfaces as a failing decode instead of a silently
// ignored predicate.
type AnalyticsEventQuery struct {
	Date         string                      `json:"date"`
	EventType    string                      `json:"eventType,omitempty"`
	Limit        int                         `json:"limit"`
	Source       string                      `json:"source"`
	LabelFilters []AnalyticsEventLabelFilter `json:"labelFilters,omitempty"`
}

// AnalyticsEventCount is the count a query reports, which answers "how many events" without
// pulling the rows back.
type AnalyticsEventCount struct {
	Count int `json:"count"`
}

// CountEvents answers how many events match a query.
//
// The query is scoped by the credential's tenant; there is no company slug to pass and none
// is sent, because the route reads the tenant from the token.
func (c *AdminClient) CountEvents(
	ctx context.Context, query AnalyticsEventQuery,
) (*AnalyticsEventCount, error) {
	if query.Date == "" {
		return nil, fmt.Errorf("custd: analytics query date is required")
	}
	if query.Source == "" {
		return nil, fmt.Errorf("custd: analytics query source is required")
	}
	var out AnalyticsEventCount
	if err := c.requestNonAdmin(ctx, http.MethodPost, analyticsQueryPath, query, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
