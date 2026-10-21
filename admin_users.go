package custd

import (
	"context"
	"errors"
	"net/http"
	"net/url"
)

// AdminUsersAdminClient owns tenant administrator users and their roles.
//
// A user has both a global state and a per-company membership state; the two
// are reported separately because disabling a user globally and removing their
// membership are different decisions with different consequences.
type AdminUsersAdminClient struct {
	admin *AdminClient
}

// UserRemovalAttribution records who removed a user, when, and why. It travels
// with the user representation rather than in a separate audit read.
type UserRemovalAttribution struct {
	DeletedAt         *string `json:"deletedAt"`
	DeletedByUserUUID *string `json:"deletedByUserUuid"`
	DeleteReason      *string `json:"deleteReason"`
}

// AdminUser is one administrator user.
type AdminUser struct {
	UserUUID                  string   `json:"userUuid"`
	Email                     string   `json:"email"`
	CompanySlug               string   `json:"companySlug"`
	Roles                     []string `json:"roles"`
	GlobalUserState           string   `json:"globalUserState"`
	MembershipState           string   `json:"membershipState"`
	CreatedAt                 string   `json:"createdAt"`
	WelcomeNotificationStatus string   `json:"welcomeNotificationStatus"`
	UserRemovalAttribution
}

// AdminUserCreateRequest creates a user. The company is server-derived unless a
// system administrator names one.
type AdminUserCreateRequest struct {
	Email       string   `json:"email"`
	CompanySlug string   `json:"companySlug,omitempty"`
	Roles       []string `json:"roles,omitempty"`
}

// AdminUserUpdateRequest changes a user. Omitted fields are left untouched.
type AdminUserUpdateRequest struct {
	Email string   `json:"email,omitempty"`
	Roles []string `json:"roles,omitempty"`
}

// AdminUserMembershipCreateRequest adds a user to a company.
type AdminUserMembershipCreateRequest struct {
	CompanySlug string   `json:"companySlug"`
	Roles       []string `json:"roles,omitempty"`
}

// AdminUserPasswordReset carries a replacement password. The value is sensitive
// and must not be logged.
type AdminUserPasswordReset struct {
	Password string `json:"password"`
}

// AdminUserListResponse is the body for GET /users. The list is a named
// collection, not a bare array, so the wrapper field is part of the contract.
type AdminUserListResponse struct {
	Users []AdminUser `json:"users"`
}

// AdminUserRemoval is the removal action applied to a user. Soft-delete removes the
// user's access while keeping the record; restore reverses it. They are different
// operations, not a boolean, because the server records which one happened.
type AdminUserRemoval string

const (
	AdminUserRemovalSoftDelete AdminUserRemoval = "soft-delete"
	AdminUserRemovalRestore    AdminUserRemoval = "restore"
)

// AdminUserRemovalRequest carries the reason the operator typed. It lands in the
// tenant's audit stream, so it is part of the contract rather than decoration.
type AdminUserRemovalRequest struct {
	Reason string `json:"reason"`
}

// AdminUserRole is one assignable role.
type AdminUserRole struct {
	Role        string `json:"role"`
	Description string `json:"description,omitempty"`
}

// AdminUserRoleListResponse lists assignable roles.
type AdminUserRoleListResponse struct {
	Roles []AdminUserRole `json:"roles"`
}

const (
	adminUsersPath     = "/users"
	adminUserRolesPath = "/user-roles"
)

// List returns the users visible to the caller. CompanySlug narrows the answer to one
// tenant; IncludeEnded keeps removed memberships in the result so an operator can audit
// them.
func (c *AdminUsersAdminClient) List(
	ctx context.Context,
	companySlug string,
	includeEnded bool,
) (*AdminUserListResponse, error) {
	query := url.Values{}
	if companySlug != "" {
		query.Set(companySlugQueryParameter, companySlug)
	}
	if includeEnded {
		query.Set("includeEnded", "true")
	}
	path := adminUsersPath
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	var out AdminUserListResponse
	if err := c.admin.request(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Create adds a user.
func (c *AdminUsersAdminClient) Create(
	ctx context.Context,
	req AdminUserCreateRequest,
) (*AdminUser, ResponseMetadata, error) {
	var out AdminUser
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodPost, adminUsersPath, req, &out)
	return &out, metadata, err
}

// Get reads one user.
// Get reads one user. IncludeEnded keeps removed memberships in the answer so an
// operator can audit them.
func (c *AdminUsersAdminClient) Get(
	ctx context.Context,
	userUUID string,
	includeEnded bool,
) (*AdminUser, ResponseMetadata, error) {
	var out AdminUser
	path := adminUserPath(userUUID)
	if includeEnded {
		path += "?includeEnded=true"
	}
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodGet, path, nil, &out)
	return &out, metadata, err
}

// Update changes a user's email or roles.
func (c *AdminUsersAdminClient) Update(
	ctx context.Context,
	userUUID string,
	roles []string,
) (*AdminUser, ResponseMetadata, error) {
	if roles == nil {
		roles = []string{}
	}
	req := AdminUserUpdateRequest{Roles: roles}
	var out AdminUser
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodPut, adminUserPath(userUUID), req, &out,
	)
	return &out, metadata, err
}

// Disable turns a user's global access off without removing them.
// Disable turns a user's access off. Reason is optional and lands in the tenant's audit
// stream; GlobalUser disables the identity rather than one membership.
func (c *AdminUsersAdminClient) Disable(
	ctx context.Context,
	userUUID string,
	reason string,
	globalUser bool,
) (ResponseMetadata, error) {
	var body any
	if reason != "" {
		body = AdminUserRemovalRequest{Reason: reason}
	}
	return c.admin.requestWithMetadata(
		ctx, http.MethodPost, adminUserActionPath(userUUID, "/disable", globalUser), body, nil,
	)
}

// Enable restores a disabled user's global access.
// Enable re-establishes a user's access. Roles are required to reopen a membership: it
// cannot be reopened without saying what authority it carries.
func (c *AdminUsersAdminClient) Enable(
	ctx context.Context,
	userUUID string,
	roles []string,
	globalUser bool,
) (ResponseMetadata, error) {
	if roles == nil {
		roles = []string{}
	}
	body := AdminUserEnableRequest{Roles: roles}
	return c.admin.requestWithMetadata(
		ctx, http.MethodPost, adminUserActionPath(userUUID, "/enable", globalUser), body, nil,
	)
}

// Remove applies a removal action to one user and records the reason on the
// representation. CompanySlug is required because the route is tenant-scoped.
func (c *AdminUsersAdminClient) Remove(
	ctx context.Context,
	userUUID string,
	removal AdminUserRemoval,
	companySlug string,
	reason string,
) (*AdminUser, ResponseMetadata, error) {
	if removal != AdminUserRemovalSoftDelete && removal != AdminUserRemovalRestore {
		return nil, ResponseMetadata{}, errors.New("custd: admin user removal must be soft-delete or restore")
	}
	var out AdminUser
	query := url.Values{companySlugQueryParameter: {companySlug}}
	path := adminUserPath(userUUID) + "/" + string(removal) + "?" + query.Encode()
	body := AdminUserRemovalRequest{Reason: reason}
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodPost, path, body, &out)
	return &out, metadata, err
}

// AddMembership adds an existing user to a company with the given roles.
func (c *AdminUsersAdminClient) AddMembership(
	ctx context.Context,
	userUUID string,
	req AdminUserMembershipCreateRequest,
) (*AdminUser, ResponseMetadata, error) {
	var out AdminUser
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodPost, adminUserPath(userUUID)+"/memberships", req, &out,
	)
	return &out, metadata, err
}

// ResetPassword sets a new password. There is no body to carry the request
// identifier, so both are returned.
// ResetPassword queues a reset notification and returns the notification it queued, not
// a password.
func (c *AdminUsersAdminClient) ResetPassword(
	ctx context.Context,
	userUUID string,
) (*AdminUserPasswordReset, ResponseMetadata, error) {
	var out AdminUserPasswordReset
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodPost, adminUserPath(userUUID)+"/reset-password", nil, &out,
	)
	return &out, metadata, err
}

// RetryWelcome re-sends the welcome notification for one company.
func (c *AdminUsersAdminClient) RetryWelcome(
	ctx context.Context,
	userUUID string,
	companySlug string,
) (ResponseMetadata, error) {
	query := url.Values{}
	if companySlug != "" {
		query.Set(companySlugQueryParameter, companySlug)
	}
	path := adminUserPath(userUUID) + "/retry-welcome"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}
	return c.admin.requestWithMetadata(ctx, http.MethodPost, path, nil, nil)
}

// ListRoles returns the roles a caller may assign.
func (c *AdminUsersAdminClient) ListRoles(
	ctx context.Context,
) (*AdminUserRoleListResponse, ResponseMetadata, error) {
	var out AdminUserRoleListResponse
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodGet, adminUserRolesPath, nil, &out)
	return &out, metadata, err
}

func (c *AdminUsersAdminClient) userAction(
	ctx context.Context,
	userUUID string,
	action string,
) (*AdminUser, ResponseMetadata, error) {
	var out AdminUser
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodPost, adminUserPath(userUUID)+action, nil, &out,
	)
	return &out, metadata, err
}

// AdminUserEnableRequest is the body for an enable: the roles the membership reopens
// with.
type AdminUserEnableRequest struct {
	Roles []string `json:"roles"`
}

// adminUserActionPath builds an action route, adding globalUser when the action is on the
// identity rather than one membership.
func adminUserActionPath(userUUID, action string, globalUser bool) string {
	path := adminUserPath(userUUID) + action
	if globalUser {
		path += "?globalUser=true"
	}
	return path
}

func adminUserPath(userUUID string) string {
	return adminUsersPath + "/" + url.PathEscape(userUUID)
}
