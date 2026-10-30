package custd

// AdminSiteIngestTarget is the one data destination a browser site writes to.
//
// A site authenticates a browser, so its page views carry no canonical scope or
// resource selection of their own. The site states that destination once, here,
// and Custd grants write authority over exactly that resource, never the whole
// company.
//
// ScopeCode and ResourceCode are read back for display: the server resolves them
// from the UUIDs and ignores them on input.
type AdminSiteIngestTarget struct {
	ScopeUUID    string `json:"scopeUuid"`
	ResourceUUID string `json:"resourceUuid"`
	ScopeCode    string `json:"scopeCode,omitempty"`
	ResourceCode string `json:"resourceCode,omitempty"`
}

type AdminSiteCreate struct {
	CompanySlug    string   `json:"companySlug"`
	Name           string   `json:"name"`
	IdentityMode   string   `json:"identityMode"`
	Environment    string   `json:"environment,omitempty"`
	AllowedOrigins []string `json:"allowedOrigins"`
	// AllowedEnvironments restricts which environments this site may record. An empty
	// list allows every environment.
	AllowedEnvironments []string `json:"allowedEnvironments,omitempty"`
	RateLimitPerMinute  int      `json:"rateLimitPerMinute,omitempty"`
	RetentionDays       int      `json:"retentionDays,omitempty"`
	// IngestTarget is required: a site whose write key cannot write anywhere is not
	// a usable setup. It is a pointer so a missing target is sent as null and the
	// server refuses it, rather than being silently dropped.
	IngestTarget *AdminSiteIngestTarget `json:"ingestTarget"`
}

// AdminSiteUpdate replaces a site's recording configuration. AllowedEnvironments is
// always sent, because an omitted field must never silently widen a restriction.
type AdminSiteUpdate struct {
	IdentityMode        string   `json:"identityMode"`
	Environment         string   `json:"environment"`
	AllowedOrigins      []string `json:"allowedOrigins"`
	AllowedEnvironments []string `json:"allowedEnvironments"`
	// IngestTarget states the site's destination. An omitted target leaves the existing
	// destination and its write grant unchanged; submitting one is an explicit request
	// to route the site there and hold exactly that resource's write authority.
	IngestTarget *AdminSiteIngestTarget `json:"ingestTarget,omitempty"`
}

type AdminSite struct {
	SiteUUID            string   `json:"siteUuid"`
	CompanySlug         string   `json:"companySlug"`
	Name                string   `json:"name"`
	IdentityMode        string   `json:"identityMode"`
	Environment         string   `json:"environment,omitempty"`
	AllowedOrigins      []string `json:"allowedOrigins"`
	AllowedEnvironments []string `json:"allowedEnvironments"`
	RateLimitPerMinute  int      `json:"rateLimitPerMinute"`
	RetentionDays       int      `json:"retentionDays"`
	Enabled             bool     `json:"enabled"`
	// IngestTarget is empty for a site created before destinations existed. Such a
	// site keeps its previous authorisation and is not resource-isolated.
	IngestTarget *AdminSiteIngestTarget `json:"ingestTarget,omitempty"`
}

type AdminSiteCreateResponse struct {
	SiteUUID           string   `json:"siteUuid"`
	CompanySlug        string   `json:"companySlug"`
	Name               string   `json:"name"`
	IdentityMode       string   `json:"identityMode"`
	AllowedOrigins     []string `json:"allowedOrigins"`
	RateLimitPerMinute int      `json:"rateLimitPerMinute"`
	RetentionDays      int      `json:"retentionDays"`
	Enabled            bool     `json:"enabled"`
	WriteKey           string   `json:"writeKey"`
	// IngestTarget echoes the destination the site was created with, including the
	// codes the server resolved for display.
	IngestTarget *AdminSiteIngestTarget `json:"ingestTarget,omitempty"`
}

type AdminSiteList struct {
	Sites []AdminSite `json:"sites"`
}

type AdminSiteWriteKeyResponse struct {
	WriteKey string `json:"writeKey"`
}
