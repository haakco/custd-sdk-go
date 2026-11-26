package custd

import (
	"context"
	"net/http"
	"time"
)

// AdminSiteIndexNowStatus is how an IndexNow submission went.
type AdminSiteIndexNowStatus string

// The states of AdminSiteIndexNowSubmission.Status. Compare against these rather than assuming
// the set is closed.
const (
	AdminSiteIndexNowQueued   AdminSiteIndexNowStatus = "queued"
	AdminSiteIndexNowAccepted AdminSiteIndexNowStatus = "accepted"
	AdminSiteIndexNowRejected AdminSiteIndexNowStatus = "rejected"
	AdminSiteIndexNowFailed   AdminSiteIndexNowStatus = "failed"
)

// SiteIndexNowDomainUnverifiedCode is the problem code a 409 carries when IndexNow is enabled on
// a domain that is not verified.
const SiteIndexNowDomainUnverifiedCode = "site_indexnow_domain_unverified"

// SiteIndexNowNotEnabledCode is the problem code a 409 carries when the domain has not enabled
// IndexNow.
const SiteIndexNowNotEnabledCode = "site_indexnow_not_enabled"

// SiteIndexNowKeyUnverifiedCode is the problem code a 409 carries when URLs are submitted before
// the key file is hosted and verified.
const SiteIndexNowKeyUnverifiedCode = "site_indexnow_key_unverified"

// SiteIndexNowURLsInvalidCode is the problem code a 400 carries unless the URLs are 1 to 100
// distinct http(s) URLs on the domain's own host.
const SiteIndexNowURLsInvalidCode = "site_indexnow_urls_invalid"

// SiteIndexNowLimitCode is the problem code a 429 carries once the tenant has used its daily
// IndexNow URLs.
const SiteIndexNowLimitCode = "site_indexnow_limit"

// AdminSiteIndexNowSubmission is one batch of URLs sent to IndexNow. ProviderStatus is IndexNow's
// HTTP status, nil until it answered; SubmittedAt is nil until it was sent.
type AdminSiteIndexNowSubmission struct {
	SubmissionUUID string                  `json:"submissionUuid"`
	Status         AdminSiteIndexNowStatus `json:"status"`
	URLs           []string                `json:"urls"`
	ProviderStatus *int                    `json:"providerStatus,omitempty"`
	Failure        string                  `json:"failure,omitempty"`
	RequestedAt    time.Time               `json:"requestedAt"`
	SubmittedAt    *time.Time              `json:"submittedAt,omitempty"`
}

// AdminSiteIndexNowState is a domain's IndexNow setup, the tenant's daily use and its latest
// submissions. Serve a text file at KeyLocation holding only Key, then call VerifyIndexNowKey;
// a failed check says why in LastCheckFailure. VerifiedAt is nil until the key is verified.
type AdminSiteIndexNowState struct {
	Enabled          bool                          `json:"enabled"`
	Key              string                        `json:"key,omitempty"`
	KeyLocation      string                        `json:"keyLocation,omitempty"`
	Verified         bool                          `json:"verified"`
	VerifiedAt       *time.Time                    `json:"verifiedAt,omitempty"`
	LastCheckFailure string                        `json:"lastCheckFailure,omitempty"`
	DailyURLLimit    int                           `json:"dailyUrlLimit"`
	URLsToday        int                           `json:"urlsToday"`
	Submissions      []AdminSiteIndexNowSubmission `json:"submissions"`
}

type adminSiteIndexNowSubmissionRequest struct {
	URLs []string `json:"urls"`
}

// GetIndexNow returns the domain's IndexNow state.
func (c *SiteDomainAdminClient) GetIndexNow(ctx context.Context, domainUUID string) (*AdminSiteIndexNowState, error) {
	return c.indexNowState(ctx, http.MethodGet, siteIndexNowPath(domainUUID))
}

// EnableIndexNow enables IndexNow on a verified domain and returns its key to host. An unverified
// domain is refused with SiteIndexNowDomainUnverifiedCode.
func (c *SiteDomainAdminClient) EnableIndexNow(ctx context.Context, domainUUID string) (*AdminSiteIndexNowState, error) {
	return c.indexNowState(ctx, http.MethodPut, siteIndexNowPath(domainUUID))
}

// DisableIndexNow turns IndexNow off for the domain.
func (c *SiteDomainAdminClient) DisableIndexNow(ctx context.Context, domainUUID string) error {
	return c.request(ctx, http.MethodDelete, siteIndexNowPath(domainUUID), nil, nil)
}

// VerifyIndexNowKey checks the hosted key file now. A failed check is not an error: the state
// comes back with Verified false and the reason in LastCheckFailure.
func (c *SiteDomainAdminClient) VerifyIndexNowKey(ctx context.Context, domainUUID string) (*AdminSiteIndexNowState, error) {
	return c.indexNowState(ctx, http.MethodPost, siteIndexNowPath(domainUUID)+"/verify")
}

// SubmitIndexNow queues 1 to 100 URLs on the domain's host for IndexNow (202).
func (c *SiteDomainAdminClient) SubmitIndexNow(
	ctx context.Context, domainUUID string, urls []string,
) (*AdminSiteIndexNowSubmission, error) {
	var submission AdminSiteIndexNowSubmission
	err := c.request(ctx, http.MethodPost, siteIndexNowPath(domainUUID)+"/submissions",
		adminSiteIndexNowSubmissionRequest{URLs: nonNilSlice(urls)}, &submission)
	return &submission, err
}

func (c *SiteDomainAdminClient) indexNowState(ctx context.Context, method, path string) (*AdminSiteIndexNowState, error) {
	var state AdminSiteIndexNowState
	err := c.request(ctx, method, path, nil, &state)
	return &state, err
}

func siteIndexNowPath(domainUUID string) string {
	return siteDomainPath(domainUUID) + "/indexnow"
}
