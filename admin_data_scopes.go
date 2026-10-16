package custd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// dataScopeEndpoint is the parent-owned data-scope management collection. It is
// an admin surface addressed by external UUIDs; the immutable scope/resource
// codes are the selectors event and query traffic use.
const dataScopeEndpoint = "/data-scopes"

// dataScopeGrantEndpoint is the explicit principal-grant collection.
const dataScopeGrantEndpoint = "/data-scope-grants"

// DataScopeCodePattern is the shape of an immutable external scope or resource
// code. The codes are stable identifiers (for example "acme-store",
// "srv-01"), not free text.
var DataScopeCodePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]*$`)

// dataScopeUUIDPattern matches a canonical RFC 4122 UUID, the external identity
// configuration surfaces address scopes, resources and human principals by.
var dataScopeUUIDPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// DataScopeStatus is the closed lifecycle state of a scope or resource. It maps
// to the persisted enabled flag; a disabled row stays readable but cannot be
// granted or selected.
type DataScopeStatus string

const (
	DataScopeStatusActive   DataScopeStatus = "active"
	DataScopeStatusDisabled DataScopeStatus = "disabled"
)

// Valid reports whether the status is one of the closed active/disabled values.
func (s DataScopeStatus) Valid() bool {
	return s == DataScopeStatusActive || s == DataScopeStatusDisabled
}

// DataScopeCreateRequest creates one parent-company data scope. Code is the
// immutable external selector event and query traffic use; Name is the mutable
// display label.
type DataScopeCreateRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Validate mirrors the server's request validation: a bounded immutable code and
// a bounded display name. It rejects locally what the server would reject anyway.
func (r DataScopeCreateRequest) Validate() error {
	if !ValidDataScopeCode(r.Code) {
		return fmt.Errorf("custd: data scope code must be 1-128 characters matching %s", DataScopeCodePattern.String())
	}
	if !ValidDataScopeName(r.Name) {
		return fmt.Errorf("custd: data scope name must be 1-100 characters")
	}
	return nil
}

// DataScopeUpdateRequest changes the editable fields of a scope. Both fields are
// optional, but at least one must be present. The scope identity (uuid, code)
// is immutable.
type DataScopeUpdateRequest struct {
	Name   *string          `json:"name,omitempty"`
	Status *DataScopeStatus `json:"status,omitempty"`
}

// Validate mirrors the server: at least one field, and every supplied field valid.
func (r DataScopeUpdateRequest) Validate() error {
	return validateDataScopeUpdate(r.Name, r.Status)
}

// DataScope is the management view of one parent-company data scope.
type DataScope struct {
	UUID      string          `json:"uuid"`
	Code      string          `json:"code"`
	Name      string          `json:"name"`
	Status    DataScopeStatus `json:"status"`
	CreatedAt string          `json:"createdAt"`
	UpdatedAt string          `json:"updatedAt"`
}

// DataScopeListResponse is the named collection boundary for scopes.
type DataScopeListResponse struct {
	Scopes []DataScope `json:"scopes"`
}

// DataScopeResourceCreateRequest creates one resource inside a data scope. The
// owning scope is taken from the route, never the body.
type DataScopeResourceCreateRequest struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Validate mirrors the server's request validation.
func (r DataScopeResourceCreateRequest) Validate() error {
	if !ValidDataScopeCode(r.Code) {
		return fmt.Errorf("custd: data scope resource code must be 1-128 characters matching %s", DataScopeCodePattern.String())
	}
	if !ValidDataScopeName(r.Name) {
		return fmt.Errorf("custd: data scope resource name must be 1-100 characters")
	}
	return nil
}

// DataScopeResourceUpdateRequest changes the editable fields of a resource.
type DataScopeResourceUpdateRequest struct {
	Name   *string          `json:"name,omitempty"`
	Status *DataScopeStatus `json:"status,omitempty"`
}

// Validate mirrors the server: at least one field, and every supplied field valid.
func (r DataScopeResourceUpdateRequest) Validate() error {
	return validateDataScopeUpdate(r.Name, r.Status)
}

// DataScopeResource is the management view of one resource inside a scope.
type DataScopeResource struct {
	UUID      string          `json:"uuid"`
	ScopeUUID string          `json:"scopeUuid"`
	Code      string          `json:"code"`
	Name      string          `json:"name"`
	Status    DataScopeStatus `json:"status"`
	CreatedAt string          `json:"createdAt"`
	UpdatedAt string          `json:"updatedAt"`
}

// DataScopeResourceListResponse is the named collection boundary for resources.
type DataScopeResourceListResponse struct {
	Resources []DataScopeResource `json:"resources"`
}

// DataScopePrincipalKind is the closed discriminator between the two product
// principals that can hold a data grant.
type DataScopePrincipalKind string

const (
	// DataScopePrincipalHuman is a verified human, identified by the stable
	// membership user UUID.
	DataScopePrincipalHuman DataScopePrincipalKind = "human"
	// DataScopePrincipalMachine is a machine, identified by its producer client ID.
	DataScopePrincipalMachine DataScopePrincipalKind = "machine"
)

// DataScopePrincipalRef names the verified principal a data grant belongs to. A
// human is identified by the stable membership user UUID; a machine by its
// producer client ID. It never carries an email, a display name, or free actor
// text.
type DataScopePrincipalRef struct {
	Kind     DataScopePrincipalKind `json:"kind"`
	UserUUID string                 `json:"userUuid,omitempty"`
	ClientID string                 `json:"clientId,omitempty"`
}

// Validate enforces the closed human/machine shape: exactly one identity, in the
// field that identity belongs to. It is a shape check only; the server is the
// authority on which principals exist and what they may hold.
func (p DataScopePrincipalRef) Validate() error {
	switch p.Kind {
	case DataScopePrincipalHuman:
		if !isDataScopeUUID(p.UserUUID) || p.ClientID != "" {
			return fmt.Errorf("custd: a human principal requires userUuid and no clientId")
		}
	case DataScopePrincipalMachine:
		if !ValidDataScopeCode(p.ClientID) || p.UserUUID != "" {
			return fmt.Errorf("custd: a machine principal requires a bounded clientId and no userUuid")
		}
	default:
		return fmt.Errorf("custd: data scope principal kind must be %q or %q",
			DataScopePrincipalHuman, DataScopePrincipalMachine)
	}
	return nil
}

// Key returns a stable, non-secret identity string for the closed principal.
func (p DataScopePrincipalRef) Key() string {
	if p.Kind == DataScopePrincipalMachine {
		return "machine:" + p.ClientID
	}
	return "human:" + p.UserUUID
}

// DataScopeGrantTargetKind is the closed grant-target discriminator.
type DataScopeGrantTargetKind string

const (
	// DataScopeGrantTargetParent grants every scope and resource in the company,
	// including rows written before any scope existed. It is explicit; a missing
	// scope never means parent.
	DataScopeGrantTargetParent DataScopeGrantTargetKind = "parent"
	// DataScopeGrantTargetSubClient grants one scope and every resource inside it.
	DataScopeGrantTargetSubClient DataScopeGrantTargetKind = "sub_client"
	// DataScopeGrantTargetResource grants exactly one resource within one scope.
	DataScopeGrantTargetResource DataScopeGrantTargetKind = "resource"
)

// DataScopeGrantTarget names what a grant covers. Configuration surfaces address
// the target by UUID; the immutable codes are for event/query selection.
type DataScopeGrantTarget struct {
	Kind         DataScopeGrantTargetKind `json:"kind"`
	ScopeUUID    string                   `json:"scopeUuid,omitempty"`
	ResourceUUID string                   `json:"resourceUuid,omitempty"`
}

// Validate enforces the explicit target mapping: parent has neither reference,
// sub-client has a scope only, resource has both.
func (t DataScopeGrantTarget) Validate() error {
	switch t.Kind {
	case DataScopeGrantTargetParent:
		if t.ScopeUUID != "" || t.ResourceUUID != "" {
			return fmt.Errorf("custd: a parent target carries no scope or resource")
		}
	case DataScopeGrantTargetSubClient:
		if !isDataScopeUUID(t.ScopeUUID) || t.ResourceUUID != "" {
			return fmt.Errorf("custd: a sub_client target requires scopeUuid and no resourceUuid")
		}
	case DataScopeGrantTargetResource:
		if !isDataScopeUUID(t.ScopeUUID) || !isDataScopeUUID(t.ResourceUUID) {
			return fmt.Errorf("custd: a resource target requires scopeUuid and resourceUuid")
		}
	default:
		return fmt.Errorf("custd: data scope grant target kind must be %q, %q or %q",
			DataScopeGrantTargetParent, DataScopeGrantTargetSubClient, DataScopeGrantTargetResource)
	}
	return nil
}

// DataScopeAction is a closed data action.
type DataScopeAction string

const (
	DataScopeActionRead  DataScopeAction = "read"
	DataScopeActionWrite DataScopeAction = "write"
)

// Valid reports whether the action is one of the closed read/write actions.
func (a DataScopeAction) Valid() bool {
	return a == DataScopeActionRead || a == DataScopeActionWrite
}

// DataScopeGrant is one principal's grant over one explicit target and action.
// The wire shape is identical for the grant create/revoke request and the grant
// list response, so one named DTO serves both.
type DataScopeGrant struct {
	Principal DataScopePrincipalRef `json:"principal"`
	Target    DataScopeGrantTarget  `json:"target"`
	Action    DataScopeAction       `json:"action"`
}

// Validate checks the principal, target and action together, mirroring the
// server's grant contract. It makes no authorization decision; the server
// decides what a principal may hold.
func (g DataScopeGrant) Validate() error {
	if err := g.Principal.Validate(); err != nil {
		return err
	}
	if err := g.Target.Validate(); err != nil {
		return err
	}
	if !g.Action.Valid() {
		return fmt.Errorf("custd: data scope action must be %q or %q", DataScopeActionRead, DataScopeActionWrite)
	}
	return nil
}

// DataScopeGrantListResponse is the named list boundary for a principal's grants.
type DataScopeGrantListResponse struct {
	Grants []DataScopeGrant `json:"grants"`
}

// DataScopeAdminClient owns parent-owned data-scope management: scopes,
// resources and explicit principal grants. It is reached through
// CustdClient.Admin.DataScopes and reuses the admin transport.
type DataScopeAdminClient struct {
	admin *AdminClient
}

func newDataScopeAdminClient(admin *AdminClient) *DataScopeAdminClient {
	return &DataScopeAdminClient{admin: admin}
}

// CreateScope creates a parent-company data scope.
func (c *DataScopeAdminClient) CreateScope(
	ctx context.Context, req DataScopeCreateRequest,
) (*DataScope, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var scope DataScope
	if err := c.admin.request(ctx, http.MethodPost, dataScopeEndpoint, req, &scope); err != nil {
		return nil, err
	}
	return &scope, nil
}

// ListScopes lists the parent company's data scopes. includeDisabled also
// returns disabled scopes.
func (c *DataScopeAdminClient) ListScopes(
	ctx context.Context, includeDisabled bool,
) (*DataScopeListResponse, error) {
	var out DataScopeListResponse
	path := dataScopeEndpoint + dataScopeIncludeDisabledParam(includeDisabled)
	if err := c.admin.request(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	if out.Scopes == nil {
		out.Scopes = []DataScope{}
	}
	return &out, nil
}

// GetScope returns one data scope by its external UUID.
func (c *DataScopeAdminClient) GetScope(ctx context.Context, scopeUUID string) (*DataScope, error) {
	var scope DataScope
	path := dataScopeEndpoint + "/" + url.PathEscape(scopeUUID)
	if err := c.admin.request(ctx, http.MethodGet, path, nil, &scope); err != nil {
		return nil, err
	}
	return &scope, nil
}

// UpdateScope changes the editable fields of a scope.
func (c *DataScopeAdminClient) UpdateScope(
	ctx context.Context, scopeUUID string, req DataScopeUpdateRequest,
) (*DataScope, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var scope DataScope
	path := dataScopeEndpoint + "/" + url.PathEscape(scopeUUID)
	if err := c.admin.request(ctx, http.MethodPatch, path, req, &scope); err != nil {
		return nil, err
	}
	return &scope, nil
}

// DisableScope disables a data scope. The route returns 204 No Content.
func (c *DataScopeAdminClient) DisableScope(ctx context.Context, scopeUUID string) error {
	path := dataScopeEndpoint + "/" + url.PathEscape(scopeUUID) + "/disable"
	return c.admin.request(ctx, http.MethodPost, path, nil, nil)
}

// CreateResource creates a resource inside a scope. The scope is addressed by
// route, never by the body.
func (c *DataScopeAdminClient) CreateResource(
	ctx context.Context, scopeUUID string, req DataScopeResourceCreateRequest,
) (*DataScopeResource, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var resource DataScopeResource
	path := dataScopeEndpoint + "/" + url.PathEscape(scopeUUID) + "/resources"
	if err := c.admin.request(ctx, http.MethodPost, path, req, &resource); err != nil {
		return nil, err
	}
	return &resource, nil
}

// ListResources lists the resources inside a scope. includeDisabled also returns
// disabled resources.
func (c *DataScopeAdminClient) ListResources(
	ctx context.Context, scopeUUID string, includeDisabled bool,
) (*DataScopeResourceListResponse, error) {
	var out DataScopeResourceListResponse
	path := dataScopeEndpoint + "/" + url.PathEscape(scopeUUID) + "/resources" +
		dataScopeIncludeDisabledParam(includeDisabled)
	if err := c.admin.request(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	if out.Resources == nil {
		out.Resources = []DataScopeResource{}
	}
	return &out, nil
}

// GetResource returns one resource inside a scope by its external UUID.
func (c *DataScopeAdminClient) GetResource(
	ctx context.Context, scopeUUID, resourceUUID string,
) (*DataScopeResource, error) {
	var resource DataScopeResource
	path := dataScopeResourcePath(scopeUUID, resourceUUID)
	if err := c.admin.request(ctx, http.MethodGet, path, nil, &resource); err != nil {
		return nil, err
	}
	return &resource, nil
}

// UpdateResource changes the editable fields of a resource.
func (c *DataScopeAdminClient) UpdateResource(
	ctx context.Context, scopeUUID, resourceUUID string, req DataScopeResourceUpdateRequest,
) (*DataScopeResource, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	var resource DataScopeResource
	path := dataScopeResourcePath(scopeUUID, resourceUUID)
	if err := c.admin.request(ctx, http.MethodPatch, path, req, &resource); err != nil {
		return nil, err
	}
	return &resource, nil
}

// DisableResource disables a resource. The route returns 204 No Content.
func (c *DataScopeAdminClient) DisableResource(ctx context.Context, scopeUUID, resourceUUID string) error {
	path := dataScopeResourcePath(scopeUUID, resourceUUID) + "/disable"
	return c.admin.request(ctx, http.MethodPost, path, nil, nil)
}

// ListGrants lists one principal's explicit grants. The principal is the closed
// human/machine shape: a human membership user UUID or a producer client ID.
func (c *DataScopeAdminClient) ListGrants(
	ctx context.Context, principal DataScopePrincipalRef,
) (*DataScopeGrantListResponse, error) {
	if err := principal.Validate(); err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Set("kind", string(principal.Kind))
	if principal.UserUUID != "" {
		params.Set("userUuid", principal.UserUUID)
	}
	if principal.ClientID != "" {
		params.Set("clientId", principal.ClientID)
	}
	var out DataScopeGrantListResponse
	path := dataScopeGrantEndpoint + "?" + params.Encode()
	if err := c.admin.request(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	if out.Grants == nil {
		out.Grants = []DataScopeGrant{}
	}
	return &out, nil
}

// CreateGrant creates one explicit principal grant.
func (c *DataScopeAdminClient) CreateGrant(
	ctx context.Context, grant DataScopeGrant,
) (*DataScopeGrant, error) {
	if err := grant.Validate(); err != nil {
		return nil, err
	}
	var created DataScopeGrant
	if err := c.admin.request(ctx, http.MethodPost, dataScopeGrantEndpoint, grant, &created); err != nil {
		return nil, err
	}
	return &created, nil
}

// RevokeGrant revokes one explicit principal grant. The route returns 204 No
// Content.
func (c *DataScopeAdminClient) RevokeGrant(ctx context.Context, grant DataScopeGrant) error {
	if err := grant.Validate(); err != nil {
		return err
	}
	return c.admin.request(ctx, http.MethodPost, dataScopeGrantEndpoint+"/revoke", grant, nil)
}

func dataScopeResourcePath(scopeUUID, resourceUUID string) string {
	return dataScopeEndpoint + "/" + url.PathEscape(scopeUUID) + "/resources/" + url.PathEscape(resourceUUID)
}

func dataScopeIncludeDisabledParam(includeDisabled bool) string {
	if !includeDisabled {
		return ""
	}
	return "?includeDisabled=" + strconv.FormatBool(includeDisabled)
}

func validateDataScopeUpdate(name *string, status *DataScopeStatus) error {
	if name == nil && status == nil {
		return fmt.Errorf("custd: data scope update requires a name or a status")
	}
	if name != nil && !ValidDataScopeName(*name) {
		return fmt.Errorf("custd: data scope name must be 1-100 characters")
	}
	if status != nil && !status.Valid() {
		return fmt.Errorf("custd: data scope status must be %q or %q",
			DataScopeStatusActive, DataScopeStatusDisabled)
	}
	return nil
}

// ValidDataScopeCode reports whether value is a valid immutable external scope
// or resource code.
func ValidDataScopeCode(value string) bool {
	return len(value) >= 1 && len(value) <= 128 && DataScopeCodePattern.MatchString(value)
}

// ValidDataScopeName reports whether value is a valid mutable display name.
func ValidDataScopeName(value string) bool {
	trimmed := strings.TrimSpace(value)
	return len(trimmed) >= 1 && len(trimmed) <= 100
}

func isDataScopeUUID(value string) bool {
	return dataScopeUUIDPattern.MatchString(value)
}
