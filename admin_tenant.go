package custd

type AdminTenantCreate struct {
	Slug        string `json:"slug"`
	CompanyName string `json:"companyName"`
	// Purpose is required by the server: a tenant created without one cannot be told apart
	// from a generated proof tenant.
	Purpose string `json:"purpose,omitempty"`
}

type AdminTenant struct {
	Slug        string `json:"slug"`
	CompanyName string `json:"companyName"`
	Enabled     bool   `json:"enabled"`
	// Purpose names what the tenant was provisioned for: a real customer, or a generated
	// proof tenant that must not be treated as one.
	Purpose string `json:"purpose"`
	// AutoCreateEnvironments reports whether ingest may create an environment value the
	// first time it sees one.
	AutoCreateEnvironments bool `json:"autoCreateEnvironments"`
}

type AdminTenantList struct {
	Tenants []AdminTenant `json:"tenants"`
}
