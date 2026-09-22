package custd

type AdminOAuthClientCreate struct {
	ClientID    string   `json:"clientId"`
	CompanySlug string   `json:"companySlug"`
	Scopes      []string `json:"scopes"`
	// PurposeProfile lets Custd derive the least-privilege scope set and the
	// matching client profile, so a caller states intent rather than scopes.
	Environment    string `json:"environment,omitempty"`
	Profile        string `json:"profile,omitempty"`
	PurposeProfile string `json:"purposeProfile,omitempty"`
}

type AdminOAuthClient struct {
	ClientID       string   `json:"clientId"`
	CompanySlug    string   `json:"companySlug"`
	Scopes         []string `json:"scopes"`
	Environment    string   `json:"environment,omitempty"`
	Profile        string   `json:"profile,omitempty"`
	PurposeProfile string   `json:"purposeProfile,omitempty"`
	ClientSecret   string   `json:"clientSecret,omitempty"`
}

type AdminOAuthClientCreateResponse struct {
	ClientID       string   `json:"clientId"`
	CompanySlug    string   `json:"companySlug"`
	Scopes         []string `json:"scopes"`
	Environment    string   `json:"environment,omitempty"`
	Profile        string   `json:"profile,omitempty"`
	PurposeProfile string   `json:"purposeProfile,omitempty"`
	ClientSecret   string   `json:"clientSecret"`
}

type AdminOAuthClientList struct {
	Clients []AdminOAuthClient `json:"clients"`
}

type AdminOAuthClientSecretResponse struct {
	ClientSecret string `json:"clientSecret"`
}

// AdminOAuthClientUpdateScopesRequest matches Stage 1 contract: scope profile
// governs the scope set; explicit Scopes overrides when provided.
type AdminOAuthClientUpdateScopesRequest struct {
	CompanySlug string   `json:"companySlug,omitempty"`
	Profile     string   `json:"profile,omitempty"`
	Scopes      []string `json:"scopes"`
}
