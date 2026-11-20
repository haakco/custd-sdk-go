package custd

import (
	"context"
	"net/http"
	"time"
)

// AdminSiteStatusSelection chooses a verified domain and the label customers see.
type AdminSiteStatusSelection struct {
	DomainUUID string `json:"domainUuid"`
	Label      string `json:"label"`
}

// AdminSiteStatusDraft selects the title and components to preview or publish.
type AdminSiteStatusDraft struct {
	Title      string                     `json:"title"`
	Components []AdminSiteStatusSelection `json:"components"`
}

// AdminSiteStatusComponent is the public projection of a selected component.
type AdminSiteStatusComponent struct {
	Label         string     `json:"label"`
	Status        string     `json:"status"`
	LastCheckedAt *time.Time `json:"lastCheckedAt,omitempty"`
}

// AdminSiteStatusUpdate is an update approved for public display.
type AdminSiteStatusUpdate struct {
	Message     string    `json:"message"`
	PublishedAt time.Time `json:"publishedAt"`
}

// AdminSiteStatusPage is the preview of what customers can see.
type AdminSiteStatusPage struct {
	Title      string                     `json:"title"`
	Components []AdminSiteStatusComponent `json:"components"`
	Updates    []AdminSiteStatusUpdate    `json:"updates"`
}

// AdminSiteStatusPublication identifies a published status page.
type AdminSiteStatusPublication struct {
	PublicKey   string    `json:"publicKey"`
	PublishedAt time.Time `json:"publishedAt"`
}

// AdminSiteStatusSettings is the private editor state for the Site.
type AdminSiteStatusSettings struct {
	Draft       *AdminSiteStatusDraft       `json:"draft,omitempty"`
	Publication *AdminSiteStatusPublication `json:"publication,omitempty"`
}

// AdminSiteStatusUpdateRequest contains a reviewed customer message and a stable retry ID.
type AdminSiteStatusUpdateRequest struct {
	RequestUUID string `json:"requestUuid"`
	Message     string `json:"message"`
}

// AdminSiteStatusUpdateHistory lists the latest published updates.
type AdminSiteStatusUpdateHistory struct {
	Updates []AdminSiteStatusUpdate `json:"updates"`
}

// GetStatusPageSettings reads the saved draft and current publication, if any.
func (c *SiteDomainAdminClient) GetStatusPageSettings(ctx context.Context, domainUUID string) (*AdminSiteStatusSettings, error) {
	var settings AdminSiteStatusSettings
	err := c.request(ctx, http.MethodGet, siteStatusPagePath(domainUUID), nil, &settings)
	return &settings, err
}

// PreviewStatusPage shows selected components without publishing the page.
func (c *SiteDomainAdminClient) PreviewStatusPage(
	ctx context.Context, domainUUID string, draft AdminSiteStatusDraft,
) (*AdminSiteStatusPage, error) {
	var page AdminSiteStatusPage
	err := c.request(ctx, http.MethodPost, siteStatusPagePath(domainUUID)+"/preview", draft, &page)
	return &page, err
}

// PublishStatusPage publishes the selected components.
func (c *SiteDomainAdminClient) PublishStatusPage(
	ctx context.Context, domainUUID string, draft AdminSiteStatusDraft,
) (*AdminSiteStatusPublication, error) {
	var publication AdminSiteStatusPublication
	err := c.request(ctx, http.MethodPut, siteStatusPagePath(domainUUID), draft, &publication)
	return &publication, err
}

// UnpublishStatusPage removes public access and leaves the draft available.
func (c *SiteDomainAdminClient) UnpublishStatusPage(ctx context.Context, domainUUID string) error {
	return c.request(ctx, http.MethodDelete, siteStatusPagePath(domainUUID), nil, nil)
}

// PostStatusPageUpdate publishes a reviewed message and notifies confirmed subscribers.
func (c *SiteDomainAdminClient) PostStatusPageUpdate(
	ctx context.Context, domainUUID string, request AdminSiteStatusUpdateRequest,
) (*AdminSiteStatusUpdateHistory, error) {
	var history AdminSiteStatusUpdateHistory
	err := c.request(ctx, http.MethodPost, siteStatusPagePath(domainUUID)+"/updates", request, &history)
	return &history, err
}

func siteStatusPagePath(domainUUID string) string {
	return siteDomainPath(domainUUID) + "/status-page"
}
