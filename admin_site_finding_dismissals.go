package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// AdminSiteFindingDismissalScope selects the rule or the finding's individual URL.
type AdminSiteFindingDismissalScope string

const (
	AdminSiteFindingDismissalRule AdminSiteFindingDismissalScope = "rule"
	AdminSiteFindingDismissalURL  AdminSiteFindingDismissalScope = "url"
)

// AdminSiteFindingDismissalRequest records why an observed finding should be hidden.
type AdminSiteFindingDismissalRequest struct {
	FindingUUID string                         `json:"findingUuid"`
	Scope       AdminSiteFindingDismissalScope `json:"scope"`
	Reason      string                         `json:"reason"`
}

// AdminSiteFindingDismissal is a current dismissal. URL is empty for a rule-wide dismissal.
type AdminSiteFindingDismissal struct {
	DismissalUUID string    `json:"dismissalUuid"`
	Rule          string    `json:"rule"`
	URL           string    `json:"url,omitempty"`
	Title         string    `json:"title"`
	Reason        string    `json:"reason"`
	DismissedAt   time.Time `json:"dismissedAt"`
}

// AdminSiteFindingDismissals lists a domain's current audit dismissals.
type AdminSiteFindingDismissals struct {
	Dismissals []AdminSiteFindingDismissal `json:"dismissals"`
}

// ListFindingDismissals returns the current rule- and URL-scoped overrides.
func (c *SiteDomainAdminClient) ListFindingDismissals(
	ctx context.Context, domainUUID string,
) (*AdminSiteFindingDismissals, error) {
	var result AdminSiteFindingDismissals
	err := c.request(ctx, http.MethodGet, siteFindingDismissalsPath(domainUUID), nil, &result)
	return &result, err
}

// DismissFinding hides a currently observed audit finding at the requested scope.
func (c *SiteDomainAdminClient) DismissFinding(
	ctx context.Context, domainUUID string, request AdminSiteFindingDismissalRequest,
) (*AdminSiteFindingDismissal, error) {
	var result AdminSiteFindingDismissal
	err := c.request(ctx, http.MethodPost, siteFindingDismissalsPath(domainUUID), request, &result)
	return &result, err
}

// RestoreFinding removes an override so a still-observed finding returns to active reporting.
func (c *SiteDomainAdminClient) RestoreFinding(ctx context.Context, domainUUID, dismissalUUID string) error {
	path := siteFindingDismissalsPath(domainUUID) + "/" + url.PathEscape(dismissalUUID)
	return c.request(ctx, http.MethodDelete, path, nil, nil)
}

func siteFindingDismissalsPath(domainUUID string) string {
	return siteDomainPath(domainUUID) + "/finding-dismissals"
}
