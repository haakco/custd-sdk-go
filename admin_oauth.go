package custd

type AdminOAuthClientCreate struct {
	ClientID    string   `json:"clientId"`
	CompanySlug string   `json:"companySlug"`
	Scopes      []string `json:"scopes"`
}

type AdminOAuthClient struct {
	ClientID     string   `json:"clientId"`
	CompanySlug  string   `json:"companySlug"`
	Scopes       []string `json:"scopes"`
	ClientSecret string   `json:"clientSecret,omitempty"`
}

type AdminOAuthClientCreateResponse struct {
	ClientID     string   `json:"clientId"`
	CompanySlug  string   `json:"companySlug"`
	Scopes       []string `json:"scopes"`
	ClientSecret string   `json:"clientSecret"`
}

type AdminOAuthClientList struct {
	Clients []AdminOAuthClient `json:"clients"`
}

type AdminOAuthClientSecretResponse struct {
	ClientSecret string `json:"clientSecret"`
}
