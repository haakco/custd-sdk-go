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
func (c *AdminUsersAdminClient) Get(
	ctx context.Context,
	userUUID string,
) (*AdminUser, ResponseMetadata, error) {
	var out AdminUser
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodGet, adminUserPath(userUUID), nil, &out)
	return &out, metadata, err
}

// Update changes a user's email or roles.
func (c *AdminUsersAdminClient) Update(
	ctx context.Context,
	userUUID string,
	req AdminUserUpdateRequest,
) (*AdminUser, ResponseMetadata, error) {
	var out AdminUser
	metadata, err := c.admin.requestWithMetadata(
		ctx, http.MethodPut, adminUserPath(userUUID), req, &out,
	)
	return &out, metadata, err
}

// Disable turns a user's global access off without removing them.
func (c *AdminUsersAdminClient) Disable(
	ctx context.Context,
	userUUID string,
) (*AdminUser, ResponseMetadata, error) {
	return c.userAction(ctx, userUUID, "/disable")
}

// Enable restores a disabled user's global access.
func (c *AdminUsersAdminClient) Enable(
	ctx context.Context,
	userUUID string,
) (*AdminUser, ResponseMetadata, error) {
	return c.userAction(ctx, userUUID, "/enable")
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
func (c *AdminUsersAdminClient) ResetPassword(
	ctx context.Context,
	userUUID string,
	req AdminUserPasswordReset,
) (ResponseMetadata, error) {
	return c.admin.requestWithMetadata(
		ctx, http.MethodPost, adminUserPath(userUUID)+"/reset-password", req, nil,
	)
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

func adminUserPath(userUUID string) string {
	return adminUsersPath + "/" + url.PathEscape(userUUID)
}
