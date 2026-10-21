package custd

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

// PreparedDataClient owns the client-facing prepared-data surfaces: the tenant
// bootstrap projection and the prepared-data receipt, output catalogue and
// output query.
//
// The six routes live under three different roots. /api/v1/admin is the admin
// namespace, so the admin request helper addresses it with a path relative to
// that namespace. /api/v1/processing and /api/v1/reporting are tenant-scoped
// roots outside the admin namespace, so their paths are built in full here and
// requested against the configured base URL with the same credentials the admin
// client uses.
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

// PreparedDataBootstrapStatus is the bootstrap projection of one tenant: where
// the tenant stands, what Custd produced it from, and the next safe action.
type PreparedDataBootstrapStatus struct {
	TenantSlug   string                 `json:"tenantSlug"`
	State        string                 `json:"state"`
	ObservedAt   string                 `json:"observedAt"`
	Provenance   PreparedDataProvenance `json:"provenance"`
	Retryability string                 `json:"retryability"`
	Warnings     []PreparedDataWarning  `json:"warnings,omitempty"`
	NextAction   PreparedDataNextAction `json:"nextAction"`
}

const (
	// preparedDataBootstrapPath is the admin-namespace-relative bootstrap path.
	preparedDataBootstrapPath = "/bootstrap"
	// preparedDataProcessingRoot is the receipt root. It is not under the admin
	// namespace, so the path is built in full.
	preparedDataProcessingRoot = "/api/v1/processing"
	// preparedDataReportingRoot is the prepared-output catalogue root. It is not
	// under the admin namespace, so the path is built in full.
	preparedDataReportingRoot = "/api/v1/reporting/outputs"
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

// GetPreparedDataReceipt reads the status of one accepted submission.
//
// The receipt identifier is escaped into the path. The server owns identifier
// validation: an identifier it does not know answers exactly like one that
// does not exist.
func (c *PreparedDataClient) GetPreparedDataReceipt(
	ctx context.Context, receiptUUID string,
) (*PreparedDataReceiptStatus, ResponseMetadata, error) {
	var out PreparedDataReceiptStatus
	metadata, err := c.requestAtRoot(
		ctx, http.MethodGet, preparedDataProcessingRoot+"/"+url.PathEscape(receiptUUID), nil, &out,
	)
	return &out, metadata, err
}

// ListPreparedDataOutputs lists the tenant's prepared outputs.
func (c *PreparedDataClient) ListPreparedDataOutputs(
	ctx context.Context,
) (*PreparedDataOutputList, ResponseMetadata, error) {
	var out PreparedDataOutputList
	metadata, err := c.requestAtRoot(ctx, http.MethodGet, preparedDataReportingRoot, nil, &out)
	return &out, metadata, err
}

// GetPreparedDataOutput reads the status of one prepared output.
func (c *PreparedDataClient) GetPreparedDataOutput(
	ctx context.Context, outputUUID string,
) (*PreparedDataOutputStatus, ResponseMetadata, error) {
	var out PreparedDataOutputStatus
	metadata, err := c.requestAtRoot(
		ctx, http.MethodGet, preparedDataReportingRoot+"/"+url.PathEscape(outputUUID), nil, &out,
	)
	return &out, metadata, err
}

// QueryPreparedDataOutput runs one query against a ready prepared output and
// returns the output status together with the rendered data.
//
// The query body is the same validated reporting request the reporting
// namespace accepts, so one request type serves both surfaces.
func (c *PreparedDataClient) QueryPreparedDataOutput(
	ctx context.Context, outputUUID string, req ReportingQueryRequest,
) (*PreparedDataQueryEnvelope, ResponseMetadata, error) {
	var out PreparedDataQueryEnvelope
	metadata, err := c.requestAtRoot(
		ctx,
		http.MethodPost,
		preparedDataReportingRoot+"/"+url.PathEscape(outputUUID)+"/query",
		req,
		&out,
	)
	return &out, metadata, err
}

// requestAtRoot issues one request against a prepared-data root that is not
// the admin namespace.
//
// The admin request helper always prefixes /api/v1/admin and rewrites the
// acting-company query parameter, so it cannot address these roots. The path
// is therefore built in full and requested against the configured base URL
// with the same credential and audit labels the admin client sends.
func (c *PreparedDataClient) requestAtRoot(
	ctx context.Context, method, path string, payload, out any,
) (ResponseMetadata, error) {
	endpoint := strings.TrimRight(c.admin.client.config.BaseURL, "/") + path
	headers := c.admin.client.headers(false)
	c.admin.applyActingCompany(headers)
	metadata, body, err := c.admin.requestOutsideAdmin(ctx, method, endpoint, headers, payload)
	if err != nil {
		return metadata, err
	}
	if err := decodeAdminResponse(body, out); err != nil {
		return metadata, err
	}
	return metadata, nil
}
