package custd

import (
	"context"
	"net/http"
	"net/url"
)

// MachineAccessAdminClient owns what a company's machine credentials (OAuth clients) may do, per
// feature area, for the whole company and for each machine.
//
// The policy only narrows: an allowed area still needs the client's own scopes, and a blocked area
// refuses that company's machines with a problem coded ProblemCodeMachineAccessBlocked.
type MachineAccessAdminClient struct {
	admin *AdminClient
}

// MachineAccessArea is one feature area a machine access setting controls.
type MachineAccessArea string

const (
	MachineAccessSites           MachineAccessArea = "sites"
	MachineAccessReporting       MachineAccessArea = "reporting"
	MachineAccessPrivacyErasure  MachineAccessArea = "privacy_erasure"
	MachineAccessDestructiveJobs MachineAccessArea = "destructive_jobs"
	MachineAccessUsersAccess     MachineAccessArea = "users_access"
	MachineAccessProvisioning    MachineAccessArea = "provisioning_credentials"
	MachineAccessIngestSchemas   MachineAccessArea = "ingest_schemas"
	MachineAccessUsageBilling    MachineAccessArea = "usage_billing"
)

// MachineAccessState is one area's setting. A company setting is allowed or blocked; a machine
// override may also inherit, which defers to the company setting.
type MachineAccessState string

const (
	MachineAccessAllowed MachineAccessState = "allowed"
	MachineAccessBlocked MachineAccessState = "blocked"
	MachineAccessInherit MachineAccessState = "inherit"
)

// MachineAccessDecidedBy names the setting that decided an area, or that refused a request.
type MachineAccessDecidedBy string

const (
	// MachineAccessDecidedByCompany is the company's default for the area.
	MachineAccessDecidedByCompany MachineAccessDecidedBy = "company"
	// MachineAccessDecidedByMachine is the machine's own override for the area.
	MachineAccessDecidedByMachine MachineAccessDecidedBy = "machine"
	// MachineAccessDecidedByScope means the effective view found no matching OAuth scope.
	MachineAccessDecidedByScope MachineAccessDecidedBy = "scope"
	// MachineAccessDecidedBySupportAutomation is the target company's "Allow Custd support
	// automation" setting, which refuses Custd's platform operator machine acting inside it.
	MachineAccessDecidedBySupportAutomation MachineAccessDecidedBy = "support_automation"
)

const (
	// ProblemCodeMachineAccessBlocked is the problem code of a request a machine access setting
	// refused; Problem.MachineAccess says which.
	ProblemCodeMachineAccessBlocked = "machine_access_blocked"
	// ProblemCodeMachineAccessUnavailable is the problem code when the policy could not be read.
	// It is a bounded retry, not a setting to change.
	ProblemCodeMachineAccessUnavailable = "machine_access_unavailable"
)

// MachineAccessBlock is the machine-readable reason a request was refused.
type MachineAccessBlock struct {
	Area        MachineAccessArea      `json:"area"`
	BlockedBy   MachineAccessDecidedBy `json:"blockedBy"`
	CompanySlug string                 `json:"companySlug"`
	ClientID    string                 `json:"clientId"`
}

// MachineAccessAreaSetting is one feature area's setting.
type MachineAccessAreaSetting struct {
	Area  MachineAccessArea  `json:"area"`
	State MachineAccessState `json:"state"`
}

// MachineAccessCompanyPolicy is a company's default for every machine, and whether Custd's
// platform operator machine may act inside the company.
type MachineAccessCompanyPolicy struct {
	CompanySlug            string                     `json:"companySlug"`
	AllowSupportAutomation bool                       `json:"allowSupportAutomation"`
	Areas                  []MachineAccessAreaSetting `json:"areas"`
}

// MachineAccessCompanyUpdateRequest changes the settings it names and leaves the rest. Each area
// is allowed or blocked; a nil AllowSupportAutomation leaves that setting unchanged.
type MachineAccessCompanyUpdateRequest struct {
	AllowSupportAutomation *bool                      `json:"allowSupportAutomation,omitempty"`
	Areas                  []MachineAccessAreaSetting `json:"areas,omitempty"`
}

// MachineAccessMachinePolicy is one machine's override: each area inherit, allowed or blocked.
type MachineAccessMachinePolicy struct {
	CompanySlug string                     `json:"companySlug"`
	ClientID    string                     `json:"clientId"`
	Areas       []MachineAccessAreaSetting `json:"areas"`
}

// MachineAccessMachineUpdateRequest changes the areas it names for one machine.
type MachineAccessMachineUpdateRequest struct {
	Areas []MachineAccessAreaSetting `json:"areas"`
}

// MachineAccessEffectiveArea is what one machine may do in one area and which setting decided it.
type MachineAccessEffectiveArea struct {
	Area         MachineAccessArea      `json:"area"`
	Allowed      bool                   `json:"allowed"`
	DecidedBy    MachineAccessDecidedBy `json:"decidedBy"`
	CompanyState MachineAccessState     `json:"companyState"`
	MachineState MachineAccessState     `json:"machineState"`
}

// MachineAccessEffective is one machine's effective access next to the scopes it was issued. An
// allowed area still needs the matching scope.
type MachineAccessEffective struct {
	CompanySlug  string                       `json:"companySlug"`
	ClientID     string                       `json:"clientId"`
	Scopes       []string                     `json:"scopes"`
	AllowedAreas int                          `json:"allowedAreas"`
	TotalAreas   int                          `json:"totalAreas"`
	Areas        []MachineAccessEffectiveArea `json:"areas"`
}

const machineAccessPath = "/machine-access"

// Get reads the company policy. A tenant administrator leaves companySlug empty; platform
// authority names the company.
func (c *MachineAccessAdminClient) Get(
	ctx context.Context, companySlug string,
) (*MachineAccessCompanyPolicy, ResponseMetadata, error) {
	var out MachineAccessCompanyPolicy
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodGet, machineAccessPath+optionalCompanySlug(companySlug), nil, &out,
	)
	return &out, metadata, err
}

// Update changes the company policy and returns the policy now in force.
func (c *MachineAccessAdminClient) Update(
	ctx context.Context, companySlug string, req MachineAccessCompanyUpdateRequest,
) (*MachineAccessCompanyPolicy, ResponseMetadata, error) {
	var out MachineAccessCompanyPolicy
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodPut, machineAccessPath+optionalCompanySlug(companySlug), req, &out,
	)
	return &out, metadata, err
}

// GetOverride reads one machine's per-area override. The machine's own company is the scope.
func (c *MachineAccessAdminClient) GetOverride(
	ctx context.Context, clientID string,
) (*MachineAccessMachinePolicy, ResponseMetadata, error) {
	var out MachineAccessMachinePolicy
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodGet, machineAccessClientPath(clientID, ""), nil, &out)
	return &out, metadata, err
}

// UpdateOverride changes one machine's override and returns the override now in force.
func (c *MachineAccessAdminClient) UpdateOverride(
	ctx context.Context, clientID string, req MachineAccessMachineUpdateRequest,
) (*MachineAccessMachinePolicy, ResponseMetadata, error) {
	var out MachineAccessMachinePolicy
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodPut, machineAccessClientPath(clientID, ""), req, &out)
	return &out, metadata, err
}

// GetEffective reads what one machine may do in each area and which setting decided it.
func (c *MachineAccessAdminClient) GetEffective(
	ctx context.Context, clientID string,
) (*MachineAccessEffective, ResponseMetadata, error) {
	var out MachineAccessEffective
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodGet, machineAccessClientPath(clientID, "/effective"), nil, &out,
	)
	return &out, metadata, err
}

func machineAccessClientPath(clientID, suffix string) string {
	return machineAccessPath + "/clients/" + url.PathEscape(clientID) + suffix
}
