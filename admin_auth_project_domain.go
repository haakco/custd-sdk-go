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

// AuthProjectDomainBinding is the current issuer binding and DNS instructions.
type AuthProjectDomainBinding struct {
	EnvironmentID        string                    `json:"environmentId"`
	SiteDomainID         string                    `json:"siteDomainId,omitempty"`
	Host                 string                    `json:"host"`
	Issuer               string                    `json:"issuer"`
	Status               AuthProjectDomainStatus   `json:"status"`
	DNSTarget            string                    `json:"dnsTarget,omitempty"`
	DNSProof             AuthProjectDomainDNSProof `json:"dnsProof"`
	CreatedAt            string                    `json:"createdAt"`
	ObservedAt           string                    `json:"observedAt,omitempty"`
	CertificateExpiresAt string                    `json:"certificateExpiresAt,omitempty"`
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
