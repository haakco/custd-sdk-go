package custd

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// AdminSiteMaxCampaignLinks is how many active campaign links a domain keeps.
const AdminSiteMaxCampaignLinks = 500

// SiteCampaignLinkTakenCode is the problem code a 409 carries when the domain already has an
// active link with the same tagged URL.
const SiteCampaignLinkTakenCode = "site_campaign_link_taken"

// SiteCampaignLinkLimitCode is the problem code a 409 carries once the domain keeps
// AdminSiteMaxCampaignLinks active links; archive one first.
const SiteCampaignLinkLimitCode = "site_campaign_link_limit"

// AdminSiteCampaignLinkRequest tags a page on the domain's own host with UTM values. Source,
// Medium and Campaign are required, Term (the paid keyword) and Content optional; every value is
// at most 100 characters, and the server trims, lower-cases and joins spaces with "_".
type AdminSiteCampaignLinkRequest struct {
	URL      string `json:"url"`
	Source   string `json:"source"`
	Medium   string `json:"medium"`
	Campaign string `json:"campaign"`
	Term     string `json:"term,omitempty"`
	Content  string `json:"content,omitempty"`
}

// AdminSiteCampaignLink is one saved campaign link. TaggedURL is the link to share; Term and
// Content are empty when unset; ArchivedAt is nil while the link is active.
type AdminSiteCampaignLink struct {
	LinkUUID   string     `json:"linkUuid"`
	URL        string     `json:"url"`
	Source     string     `json:"source"`
	Medium     string     `json:"medium"`
	Campaign   string     `json:"campaign"`
	Term       string     `json:"term"`
	Content    string     `json:"content"`
	TaggedURL  string     `json:"taggedUrl"`
	CreatedAt  time.Time  `json:"createdAt"`
	ArchivedAt *time.Time `json:"archivedAt,omitempty"`
}

// AdminSiteCampaignLinkList is a domain's campaign links, newest first, how many active links it
// may keep, and the utm_source and utm_medium names to suggest.
type AdminSiteCampaignLinkList struct {
	Links            []AdminSiteCampaignLink `json:"links"`
	MaxLinks         int                     `json:"maxLinks"`
	SuggestedSources []string                `json:"suggestedSources"`
	SuggestedMediums []string                `json:"suggestedMediums"`
}

// ListCampaignLinks returns the domain's active campaign links, and archived ones too when
// includeArchived is set.
func (c *SiteDomainAdminClient) ListCampaignLinks(
	ctx context.Context, domainUUID string, includeArchived bool,
) (*AdminSiteCampaignLinkList, error) {
	path := siteCampaignLinksPath(domainUUID)
	if includeArchived {
		path += "?" + url.Values{"archived": {"true"}}.Encode()
	}
	var list AdminSiteCampaignLinkList
	err := c.request(ctx, http.MethodGet, path, nil, &list)
	return &list, err
}

// CreateCampaignLink saves a tagged link (201). A second active copy answers 409 with
// SiteCampaignLinkTakenCode, a full domain SiteCampaignLinkLimitCode, and a page on another host
// 400 with SitePageNotOnDomainCode.
func (c *SiteDomainAdminClient) CreateCampaignLink(
	ctx context.Context, domainUUID string, req AdminSiteCampaignLinkRequest,
) (*AdminSiteCampaignLink, error) {
	var link AdminSiteCampaignLink
	err := c.request(ctx, http.MethodPost, siteCampaignLinksPath(domainUUID), req, &link)
	return &link, err
}

// ArchiveCampaignLink archives a link, keeping it listed with archived links and freeing its
// tagged URL. Archiving an archived link returns it unchanged.
func (c *SiteDomainAdminClient) ArchiveCampaignLink(
	ctx context.Context, domainUUID, linkUUID string,
) (*AdminSiteCampaignLink, error) {
	var link AdminSiteCampaignLink
	path := siteCampaignLinksPath(domainUUID) + "/" + url.PathEscape(linkUUID) + "/archive"
	err := c.request(ctx, http.MethodPost, path, nil, &link)
	return &link, err
}

func siteCampaignLinksPath(domainUUID string) string {
	return siteDomainPath(domainUUID) + "/campaign-links"
}
