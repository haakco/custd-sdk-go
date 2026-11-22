package custd

import (
	"context"
	"net/http"
)

// AuthProjectDomainStatus describes the observed lifecycle of a custom issuer.
type AuthProjectDomainStatus string

const (
	AuthProjectDomainPending        AuthProjectDomainStatus = "pending"
	AuthProjectDomainActive         AuthProjectDomainStatus = "active"
	AuthProjectDomainProofWithdrawn AuthProjectDomainStatus = "proof-withdrawn"
	AuthProjectDomainDisabled       AuthProjectDomainStatus = "disabled"
	AuthProjectDomainReleasing      AuthProjectDomainStatus = "releasing"
)

// AuthProjectDomainLifecycle is the realm-level state an operator acts on. It is derived from the
// stored reservation, so a pause is a durable hold rather than a reading of provisioning detail.
// A realm is unclaimed while it has a hostname but has not claimed installed capacity.
type AuthProjectDomainLifecycle string

const (
	AuthProjectDomainUnclaimed AuthProjectDomainLifecycle = "unclaimed"
	AuthProjectDomainAdmitted  AuthProjectDomainLifecycle = "admitted"
	AuthProjectDomainPaused    AuthProjectDomainLifecycle = "paused"
	AuthProjectDomainRetiring  AuthProjectDomainLifecycle = "retiring"
	AuthProjectDomainRetired   AuthProjectDomainLifecycle = "retired"
)

// AuthProjectDomainConvergence answers whether the realm runs the released image identity. It never
// infers a running identity: unverified is a distinct answer from converged.
type AuthProjectDomainConvergence string

const (
	AuthProjectDomainConverged  AuthProjectDomainConvergence = "converged"
	AuthProjectDomainDrifted    AuthProjectDomainConvergence = "drifted"
	AuthProjectDomainNotRunning AuthProjectDomainConvergence = "not-running"
	AuthProjectDomainUnverified AuthProjectDomainConvergence = "unverified"
)

// AuthProjectDomainImageSet is one image identity pair: the gateway and its Hydra server/init
// lineage. Both are empty when the deployment declares no slot images.
type AuthProjectDomainImageSet struct {
	Gateway string `json:"gateway,omitempty"`
	Hydra   string `json:"hydra,omitempty"`
}

// AuthProjectDomainImages reports the identity a realm should run against the identity it was
// observed running. A paused, retiring or retired realm runs no container, so it answers
// not-running with no running identity rather than reading as drift.
type AuthProjectDomainImages struct {
	Expected    AuthProjectDomainImageSet    `json:"expected"`
	Running     AuthProjectDomainImageSet    `json:"running"`
	Convergence AuthProjectDomainConvergence `json:"convergence"`
}

// The problem codes the admission and lifecycle routes answer with. Each is stable, so a caller
// branches on the parsed problem instead of the detail text: errors.As(err, &requestErr) then
// requestErr.Problem.Code. A code this SDK does not name is still readable the same way.
const (
	// AuthProjectDomainCapacityExhaustedCode is the 409 an admit carries when every installed
	// realm slot is claimed by a live realm or quarantined by a retirement. Admitting again is a
	// no-op rather than a second attempt, so the operator frees capacity or raises it instead.
	AuthProjectDomainCapacityExhaustedCode = "auth_domain_capacity_exhausted"
	// AuthProjectDomainPausedCode is the 409 an admit or bind carries for a realm the operator
	// paused. Admission never clears the hold.
	AuthProjectDomainPausedCode = "auth_domain_paused"
	// AuthProjectDomainRetiringCode is the 409 a mutation carries while the realm is being retired.
	AuthProjectDomainRetiringCode = "auth_domain_retiring"
	// AuthProjectDomainRetiredCode is the terminal 409: a retired realm never binds or admits again.
	AuthProjectDomainRetiredCode = "auth_domain_retired"
	// AuthProjectDomainSlotUninstalledCode is the 409 an admit carries when the stored claim names
	// a realm slot this deployment no longer installs. The claim is not silently moved.
	AuthProjectDomainSlotUninstalledCode = "auth_domain_slot_uninstalled"
	// AuthProjectDomainNotBoundCode is the 404 a read or mutation carries when the environment
	// holds no reservation.
	AuthProjectDomainNotBoundCode = "auth_domain_not_bound"
	// AuthProjectDomainConflictCode is the 409 a bind or admit carries when the reservation
	// conflicts with another binding, or when the stored claim disagrees with the reservation row.
	AuthProjectDomainConflictCode = "auth_domain_conflict"
	// AuthProjectDomainBusyCode is the 409 an operation carries while reconciliation is already
	// running. Read the domain again once it finishes.
	AuthProjectDomainBusyCode = "auth_domain_busy"
	// AuthProjectDomainUnverifiedCode is the 422 a bind carries for a Site domain whose control the
	// tenant has not yet proven.
	AuthProjectDomainUnverifiedCode = "auth_domain_unverified"
	// AuthProjectDomainUnavailableCode is the 503 a deployment answers when it cannot provision
	// authentication domains at all.
	AuthProjectDomainUnavailableCode = "auth_domain_unavailable"
)

// AuthProjectDomainBindRequest reserves a hostname using the environment's
// tenant-owned Site domain as the source of authority.
type AuthProjectDomainBindRequest struct {
	SiteDomainID string `json:"siteDomainId"`
	Host         string `json:"host"`
}

// AuthProjectDomainDNSProof is the TXT record the domain administrator publishes.
type AuthProjectDomainDNSProof struct {
	RecordName  string `json:"recordName"`
	RecordValue string `json:"recordValue"`
}

// AuthProjectDomainBinding is the current issuer binding and DNS instructions, plus the realm
// admission it holds. SlotRef is the stored claim and IntentRevision the domain intent revision the
// claim was written at, so a remediation compares the revision instead of the presentation one.
type AuthProjectDomainBinding struct {
	EnvironmentID        string                     `json:"environmentId"`
	SlotRef              string                     `json:"slotRef,omitempty"`
	SiteDomainID         string                     `json:"siteDomainId,omitempty"`
	Host                 string                     `json:"host"`
	Issuer               string                     `json:"issuer"`
	Status               AuthProjectDomainStatus    `json:"status"`
	Lifecycle            AuthProjectDomainLifecycle `json:"lifecycle"`
	IntentRevision       int64                      `json:"intentRevision"`
	Images               AuthProjectDomainImages    `json:"images"`
	DNSTarget            string                     `json:"dnsTarget,omitempty"`
	DNSProof             AuthProjectDomainDNSProof  `json:"dnsProof"`
	CreatedAt            string                     `json:"createdAt"`
	ObservedAt           string                     `json:"observedAt,omitempty"`
	CertificateExpiresAt string                     `json:"certificateExpiresAt,omitempty"`
}

// GetEnvironmentAuthDomain reads an environment's authentication hostname and status.
func (c *AuthProjectAdminClient) GetEnvironmentAuthDomain(
	ctx context.Context, projectID, environmentID string, options AuthProjectRequestOptions,
) (*AuthProjectDomainBinding, error) {
	return c.authDomainRequest(ctx, http.MethodGet, authProjectEnvironmentPath(projectID, environmentID)+"/domain", nil, options)
}

// BindEnvironmentAuthDomain reserves a hostname and returns the DNS records needed for activation.
func (c *AuthProjectAdminClient) BindEnvironmentAuthDomain(
	ctx context.Context, projectID, environmentID string, request AuthProjectDomainBindRequest, options AuthProjectRequestOptions,
) (*AuthProjectDomainBinding, error) {
	return c.authDomainRequest(ctx, http.MethodPut, authProjectEnvironmentPath(projectID, environmentID)+"/domain", &request, options)
}

// AdmitEnvironmentAuthDomain claims installed capacity for an already reserved realm and returns the
// stored claim. It changes no hostname, so it is the operation an operator repeats: a replay returns
// the same claim without advancing the intent revision. A deployment with no free slot is refused
// with AuthProjectDomainCapacityExhaustedCode, and a paused or retiring realm is refused rather than
// re-admitted.
func (c *AuthProjectAdminClient) AdmitEnvironmentAuthDomain(
	ctx context.Context, projectID, environmentID string, options AuthProjectRequestOptions,
) (*AuthProjectDomainBinding, error) {
	return c.authDomainRequest(ctx, http.MethodPost, authProjectEnvironmentPath(projectID, environmentID)+"/domain/admit", nil, options)
}

// RetryEnvironmentAuthDomain requests reconciliation of the environment's authentication hostname.
func (c *AuthProjectAdminClient) RetryEnvironmentAuthDomain(
	ctx context.Context, projectID, environmentID string, options AuthProjectRequestOptions,
) (*AuthProjectDomainBinding, error) {
	return c.authDomainRequest(ctx, http.MethodPost, authProjectEnvironmentPath(projectID, environmentID)+"/domain/retry", nil, options)
}

// DisableEnvironmentAuthDomain disables the issuer while retaining the hostname reservation.
func (c *AuthProjectAdminClient) DisableEnvironmentAuthDomain(
	ctx context.Context, projectID, environmentID string, options AuthProjectRequestOptions,
) (*AuthProjectDomainBinding, error) {
	return c.authDomainRequest(ctx, http.MethodPost, authProjectEnvironmentPath(projectID, environmentID)+"/domain/disable", nil, options)
}

// ReleaseEnvironmentAuthDomain starts cleanup before the hostname becomes available for a new binding.
func (c *AuthProjectAdminClient) ReleaseEnvironmentAuthDomain(
	ctx context.Context, projectID, environmentID string, options AuthProjectRequestOptions,
) (*AuthProjectDomainBinding, error) {
	return c.authDomainRequest(ctx, http.MethodPost, authProjectEnvironmentPath(projectID, environmentID)+"/domain/release", nil, options)
}

func (c *AuthProjectAdminClient) authDomainRequest(
	ctx context.Context, method, path string, request any, options AuthProjectRequestOptions,
) (*AuthProjectDomainBinding, error) {
	var out AuthProjectDomainBinding
	if err := c.admin.requestWithHeaders(ctx, method, path, request, &out, authProjectHeaders(options)); err != nil {
		return nil, err
	}
	return &out, nil
}
