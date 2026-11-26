package custd

// AuthProjectHostingMode is whether an environment is served as a hosted
// application. An absent mode reads as disabled on the server.
type AuthProjectHostingMode string

const (
	AuthProjectHostingDisabled AuthProjectHostingMode = "disabled"
	AuthProjectHostingEnabled  AuthProjectHostingMode = "enabled"
)

// AuthProjectHostingCapability is one capability a hosted application turns on.
// The discovery document lists the enabled ones so the app wires only those.
type AuthProjectHostingCapability string

const (
	AuthProjectHostingTracking     AuthProjectHostingCapability = "tracking"
	AuthProjectHostingErrorCapture AuthProjectHostingCapability = "error_capture"
	AuthProjectHostingAppData      AuthProjectHostingCapability = "app_data"
	AuthProjectHostingSignIn       AuthProjectHostingCapability = "sign_in"
)

// AuthProjectHosting is the environment's application hosting section. The host
// name is not part of it: the server derives it from the project and
// environment slugs and reports it on AuthProjectStatus.Hosting.
type AuthProjectHosting struct {
	Mode         AuthProjectHostingMode         `json:"mode"`
	Capabilities []AuthProjectHostingCapability `json:"capabilities"`
	// IngestTarget is where the hosted Site's events land. It is required while
	// hosting is enabled.
	IngestTarget *AuthProjectHostingIngestTarget `json:"ingestTarget,omitempty"`
}

// AuthProjectHostingIngestTarget names the tenant data scope and resource the
// hosted Site writes to.
type AuthProjectHostingIngestTarget struct {
	ScopeUUID    string `json:"scopeUuid"`
	ResourceUUID string `json:"resourceUuid"`
}

// AuthProjectHostingStatus is a hosted environment's derived serving identity.
// SignIn is nil unless the sign_in capability is on.
type AuthProjectHostingStatus struct {
	Host   string                   `json:"host"`
	Origin string                   `json:"origin"`
	SignIn *AuthProjectHostedSignIn `json:"signIn,omitempty"`
}

// AuthProjectHostedSignIn is the public client a hosted app signs people in with.
type AuthProjectHostedSignIn struct {
	ClientID    string `json:"clientId"`
	Audience    string `json:"audience"`
	RedirectURI string `json:"redirectUri"`
	// PostLogoutRedirectURI is where the hosted sign-out returns the browser.
	PostLogoutRedirectURI string `json:"postLogoutRedirectUri"`
}
