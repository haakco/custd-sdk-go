package custd

type AdminTenantCreate struct {
	Slug        string `json:"slug"`
	CompanyName string `json:"companyName"`
}

type AdminTenant struct {
	Slug        string `json:"slug"`
	CompanyName string `json:"companyName"`
	Enabled     bool   `json:"enabled"`
}

type AdminTenantList struct {
	Tenants []AdminTenant `json:"tenants"`
}
