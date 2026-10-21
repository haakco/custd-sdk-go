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
	// LoginSetup is present instead of ClientSecret when a login client is created: a
	// public client has no secret, and the caller needs the values to configure itself.
	LoginSetup *OAuthClientLoginSetup `json:"loginSetup,omitempty"`
}

// OAuthClientLoginSetup is what a login client must be configured with. It carries no
// secret by construction, because a public client has none.
type OAuthClientLoginSetup struct {
	Profile       string `json:"profile,omitempty"`
	APIURL        string `json:"apiUrl,omitempty"`
	IssuerURL     string `json:"issuerUrl,omitempty"`
	ClientID      string `json:"clientId,omitempty"`
	Audience      string `json:"audience,omitempty"`
	Scope         string `json:"scope,omitempty"`
	ConfigCommand string `json:"configCommand,omitempty"`
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
