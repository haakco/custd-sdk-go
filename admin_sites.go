package custd

type AdminSiteCreate struct {
	CompanySlug        string   `json:"companySlug"`
	Name               string   `json:"name"`
	IdentityMode       string   `json:"identityMode"`
	AllowedOrigins     []string `json:"allowedOrigins"`
	RateLimitPerMinute int      `json:"rateLimitPerMinute,omitempty"`
	RetentionDays      int      `json:"retentionDays,omitempty"`
}

type AdminSite struct {
	SiteUUID           string   `json:"siteUuid"`
	CompanySlug        string   `json:"companySlug"`
	Name               string   `json:"name"`
	IdentityMode       string   `json:"identityMode"`
	AllowedOrigins     []string `json:"allowedOrigins"`
	RateLimitPerMinute int      `json:"rateLimitPerMinute"`
	RetentionDays      int      `json:"retentionDays"`
	Enabled            bool     `json:"enabled"`
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
