package custd

import (
	"context"
	"net/http"
)

// PreparedDataClient owns the tenant's prepared-data bootstrap projection in the
// admin namespace. The prepared-data receipt, output catalogue and output query
// are served by the reporting API and live on ReportingClient.
//
// The client is not wired onto AdminClient, because that would change
// admin.go. Construct it explicitly:
//
//	prepared := custd.NewPreparedDataClient(client.Admin)
type PreparedDataClient struct {
	admin *AdminClient
}

// NewPreparedDataClient builds the prepared-data client over an admin client.
func NewPreparedDataClient(admin *AdminClient) *PreparedDataClient {
	return &PreparedDataClient{admin: admin}
}

// PreparedDataBootstrapProofKind names the channel a client presents proof on.
type PreparedDataBootstrapProofKind string

const (
	// PreparedDataBootstrapProofKindDNSTXT is a DNS TXT record the client controls.
	PreparedDataBootstrapProofKindDNSTXT PreparedDataBootstrapProofKind = "dns_txt"
	// PreparedDataBootstrapProofKindEmailLink is an emailed link the client followed.
	PreparedDataBootstrapProofKindEmailLink PreparedDataBootstrapProofKind = "email_link"
)

// PreparedDataBootstrapProof is one opaque proof reference. The server never
// dereferences the value, so the client keeps it verbatim.
type PreparedDataBootstrapProof struct {
	Kind       PreparedDataBootstrapProofKind `json:"kind"`
	Value      string                         `json:"value"`
	ObservedAt string                         `json:"observedAt"`
}

// PreparedDataBootstrapSubmitRequest is the bounded body of the bootstrap
// submit route. It carries no tenant or authority field: the server derives the
// tenant from the caller and the state from what the client submitted.
type PreparedDataBootstrapSubmitRequest struct {
	Note         string                       `json:"note,omitempty"`
	ContactEmail string                       `json:"contactEmail,omitempty"`
	Proofs       []PreparedDataBootstrapProof `json:"proofs,omitempty"`
}

// ClientBootstrapState is whether a client's prepared-data bootstrap can be used.
type ClientBootstrapState string

const (
	ClientBootstrapStateReady             ClientBootstrapState = "ready"
	ClientBootstrapStateAttentionRequired ClientBootstrapState = "attention_required"
	ClientBootstrapStateSuspended         ClientBootstrapState = "suspended"
)

// PreparedDataBootstrapStatus is the bootstrap projection of one tenant: where
// the tenant stands, what Custd produced it from, and the next safe action.
type PreparedDataBootstrapStatus struct {
	TenantSlug   string                 `json:"tenantSlug"`
	State        ClientBootstrapState   `json:"state"`
	ObservedAt   string                 `json:"observedAt"`
	Provenance   PreparedDataProvenance `json:"provenance"`
	Retryability Retryability           `json:"retryability"`
	Warnings     []PreparedDataWarning  `json:"warnings,omitempty"`
	NextAction   PreparedDataNextAction `json:"nextAction"`
}

const (
	// preparedDataBootstrapPath is the admin-namespace-relative bootstrap path.
	preparedDataBootstrapPath = "/bootstrap"
)

// GetClientBootstrap reads the tenant's bootstrap projection.
func (c *PreparedDataClient) GetClientBootstrap(
	ctx context.Context,
) (*PreparedDataBootstrapStatus, ResponseMetadata, error) {
	var out PreparedDataBootstrapStatus
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodGet, preparedDataBootstrapPath, nil, &out,
	)
	return &out, metadata, err
}

// SubmitClientBootstrap submits the tenant's bootstrap evidence and returns the
// resulting projection.
func (c *PreparedDataClient) SubmitClientBootstrap(
	ctx context.Context, req PreparedDataBootstrapSubmitRequest,
) (*PreparedDataBootstrapStatus, ResponseMetadata, error) {
	var out PreparedDataBootstrapStatus
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodPost, preparedDataBootstrapPath, req, &out,
	)
	return &out, metadata, err
}
