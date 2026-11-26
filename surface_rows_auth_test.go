package custd

import (
	"context"
)

var authProjectSurfaceRows = []surfaceRow{
	{
		id:   "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/domain/admit",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/domain/admit",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.AdmitEnvironmentAuthDomain(context.Background(), "a1", "a2", AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api PUT /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/domain",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/domain",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.BindEnvironmentAuthDomain(context.Background(), "a1", "a2", AuthProjectDomainBindRequest{}, AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/auth-projects/{projectId}/company-binding",
		path: "/api/v1/admin/auth-projects/a1/company-binding",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.BindProjectCompany(context.Background(), "a1", AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/brand/logo",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/brand/logo",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.ClearAuthProjectBrandLogo(context.Background(), "a1", "a2", AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/apps/{slug}/deploy-credentials",
		path: "/api/v1/admin/apps/a1/deploy-credentials",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.CreateAppDeployCredential(context.Background(), "a1", AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/releases",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/releases",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.CreateAppRelease(context.Background(), "a1", "a2", AppReleaseCreateRequest{}, AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments",
		path: "/api/v1/admin/auth-projects/a1/environments",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.CreateEnvironment(context.Background(), "a1", AuthProjectEnvironmentCreateRequest{}, AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api DELETE /api/v1/admin/apps/{slug}/deploy-credentials/{clientId}",
		path: "/api/v1/admin/apps/a1/deploy-credentials/a2",
		call: func(e *surfaceEnv) error {
			return e.client.Admin.AuthProjects.DeleteAppDeployCredential(context.Background(), "a1", "a2", AuthProjectRequestOptions{})
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/delete",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/delete",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.DeleteEnvironment(context.Background(), "a1", "a2", AuthProjectEnvironmentDeletionRequest{}, AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/auth-projects/{projectId}/delete",
		path: "/api/v1/admin/auth-projects/a1/delete",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.DeleteProject(context.Background(), "a1", AuthProjectDeletionRequest{}, AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/domain/disable",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/domain/disable",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.DisableEnvironmentAuthDomain(context.Background(), "a1", "a2", AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/apps/{slug}",
		path: "/api/v1/admin/apps/a1",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.GetApp(context.Background(), "a1", AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/domain",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/domain",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.GetEnvironmentAuthDomain(context.Background(), "a1", "a2", AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/capability",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/capability",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.GetEnvironmentCapability(context.Background(), "a1", "a2", AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/status",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/status",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.GetEnvironmentStatus(context.Background(), "a1", "a2", AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/releases",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/releases",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.ListAppReleases(context.Background(), "a1", "a2", 3, AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/apps",
		path: "/api/v1/admin/apps",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.ListApps(context.Background(), "a1", AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/directories",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/directories",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.ListDirectories(context.Background(), "a1", "a2", AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/organisations",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/organisations",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.ListEnvironmentOrganisations(context.Background(), "a1", "a2", AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/invitations",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/invitations",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.ListInvitations(context.Background(), "a1", "a2", AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/directories/{directoryId}/principals/{providerSubject}/sessions",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/directories/a3/principals/a4/sessions",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.ListPrincipalSessions(context.Background(), "a1", "a2", "a3", "a4", AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/principals",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/principals",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.ListPrincipals(context.Background(), "a1", "a2", ApplicationPrincipalListQuery{Cursor: "c1", Limit: 5}, AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api GET /api/v1/admin/auth-projects",
		path: "/api/v1/admin/auth-projects",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.ListProjects(context.Background(), "a1", AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/offboard",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/offboard",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.OffboardEnvironment(context.Background(), "a1", "a2", AuthProjectEnvironmentOffboardRequest{}, AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/preview",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/preview",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.PreviewEnvironmentDesiredState(context.Background(), "a1", "a2", AuthProjectDesiredStateRequest{}, AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/releases/{releaseId}/promote",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/releases/a3/promote",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.PromoteAppRelease(context.Background(), "a1", "a2", "a3", AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/domain/release",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/domain/release",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.ReleaseEnvironmentAuthDomain(context.Background(), "a1", "a2", AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/domain/retry",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/domain/retry",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.RetryEnvironmentAuthDomain(context.Background(), "a1", "a2", AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/directories/{directoryId}/principals/{providerSubject}/memberships/revoke",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/directories/a3/principals/a4/memberships/revoke",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.RevokePrincipalMembership(context.Background(), "a1", "a2", "a3", "a4", ApplicationMembershipRevokeRequest{}, AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/brand/logo",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/brand/logo",
		call: func(e *surfaceEnv) error {
			_, _, err := e.client.Admin.AuthProjects.SetAuthProjectBrandLogo(context.Background(), "a1", "a2", "a3", nil, AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/releases/{releaseId}/files",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/releases/a3/files",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.UploadAppReleaseFile(context.Background(), "a1", "a2", "a3", "a4", nil, AuthProjectRequestOptions{})
			return err
		},
	},
	{
		id:   "admin-api POST /api/v1/admin/auth-projects/{projectId}/environments/{environmentId}/releases/{releaseId}/verify",
		path: "/api/v1/admin/auth-projects/a1/environments/a2/releases/a3/verify",
		call: func(e *surfaceEnv) error {
			_, err := e.client.Admin.AuthProjects.VerifyAppRelease(context.Background(), "a1", "a2", "a3", AuthProjectRequestOptions{})
			return err
		},
	},
}
