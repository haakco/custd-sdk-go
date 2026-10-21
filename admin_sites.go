package custd

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
}

// AdminSiteUpdate replaces a site's recording configuration. AllowedEnvironments is
// always sent, because an omitted field must never silently widen a restriction.
type AdminSiteUpdate struct {
	IdentityMode        string   `json:"identityMode"`
	Environment         string   `json:"environment"`
	AllowedOrigins      []string `json:"allowedOrigins"`
	AllowedEnvironments []string `json:"allowedEnvironments"`
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
}

type AdminSiteList struct {
	Sites []AdminSite `json:"sites"`
}

type AdminSiteWriteKeyResponse struct {
	WriteKey string `json:"writeKey"`
}
