package custd

import (
	"context"
	"net/http"
	"net/url"
)

// AdminSiteAlertDestination is where one Site's finding and outage alerts go. An empty
// EmailRecipient means the tenant's own alert destinations; a set one emails that Site's alerts
// there instead, so a reseller can alert the customer whose site it is.
type AdminSiteAlertDestination struct {
	SiteUUID       string `json:"siteUuid"`
	EmailRecipient string `json:"emailRecipient,omitempty"`
}

type adminSiteAlertDestinationRequest struct {
	EmailRecipient string `json:"emailRecipient"`
}

const siteAlertsPath = siteDomainsPath + "/alert-destination"

// GetAlertDestination returns a Site's alert address. A site-bound credential may pass an empty
// siteUUID for its own Site.
func (c *SiteDomainAdminClient) GetAlertDestination(ctx context.Context, siteUUID string) (*AdminSiteAlertDestination, error) {
	var destination AdminSiteAlertDestination
	err := c.request(ctx, http.MethodGet, siteAlertsPathFor(siteUUID), nil, &destination)
	return &destination, err
}

// SetAlertDestination emails a Site's alerts to emailRecipient, or sends them to the tenant's
// destinations again when it is empty. A site-bound credential may pass an empty siteUUID.
func (c *SiteDomainAdminClient) SetAlertDestination(
	ctx context.Context, siteUUID, emailRecipient string,
) (*AdminSiteAlertDestination, error) {
	var destination AdminSiteAlertDestination
	err := c.request(ctx, http.MethodPut, siteAlertsPathFor(siteUUID),
		adminSiteAlertDestinationRequest{EmailRecipient: emailRecipient}, &destination)
	return &destination, err
}

func siteAlertsPathFor(siteUUID string) string {
	if siteUUID == "" {
		return siteAlertsPath
	}
	return siteAlertsPath + "?" + url.Values{"siteUuid": {siteUUID}}.Encode()
}
