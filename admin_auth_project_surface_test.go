package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
)

func TestAuthProjectCapabilityPinsRequestAndDecodesModes(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"identityModes": ["isolated", "shared"],
		"registrationPolicies": ["invite_only", "open"],
		"profileFieldKeys": ["contact_email"],
		"audienceConstraints": "one audience per environment",
		"subjectUniqueness": "per environment"
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	capability, err := client.Admin.AuthProjects.GetEnvironmentCapability(
		context.Background(), "project-1", "environment-1",
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err != nil {
		t.Fatalf("GetEnvironmentCapability returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodGet,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/capability",
		authProjectOwningUser, "")
	if len(capability.IdentityModes) != 2 || capability.IdentityModes[0] != "isolated" {
		t.Fatalf("identity modes = %+v", capability.IdentityModes)
	}
	if capability.SubjectUniqueness != "per environment" {
		t.Fatalf("subject uniqueness = %q", capability.SubjectUniqueness)
	}
}

func TestAuthProjectListDirectoriesDecodesCapacityAndStates(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"projectId": "project-1",
		"environmentId": "environment-1",
		"revision": 4,
		"directories": [{
			"poolId": "pool-1",
			"poolSlug": "primary",
			"state": "ready",
			"slotRef": "store-1",
			"runtimePublicUrl": "https://auth.example.com",
			"revision": 4,
			"attemptCount": 1,
			"readyAt": "2026-10-16T00:00:00Z"
		}],
		"ready": true,
		"capacity": {"declared": 3, "inUse": 1, "free": 2}
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	status, err := client.Admin.AuthProjects.ListDirectories(
		context.Background(), "project-1", "environment-1",
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err != nil {
		t.Fatalf("ListDirectories returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodGet,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/directories",
		authProjectOwningUser, "")
	if !status.Ready || status.Capacity.Free != 2 || status.Capacity.Declared != 3 {
		t.Fatalf("status = %+v", status)
	}
	if len(status.Directories) != 1 || status.Directories[0].SlotRef != "store-1" ||
		status.Directories[0].State != "ready" {
		t.Fatalf("directories = %+v", status.Directories)
	}
}

func TestAuthProjectActOnDirectoryPinsActionAndEscapesPool(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"id": "op-1", "kind": "directory_enable", "status": "applied", "revision": 5,
		"idempotencyKey": "idem-dir", "appliedAt": "2026-10-16T00:00:00Z", "replayed": false
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	operation, err := client.Admin.AuthProjects.ActOnDirectory(
		context.Background(), "project-1", "environment-1", AuthProjectDirectoryEnable,
		AuthProjectDirectoryRequest{PoolID: "pool/1", ExpectedRevision: 4},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: "idem-dir"},
	)
	if err != nil {
		t.Fatalf("ActOnDirectory returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodPost,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/directories/enable",
		authProjectOwningUser, "idem-dir")
	body := decodeAuthProjectBody(t, doer.requests[0])
	if body["poolId"] != "pool/1" || body["expectedRevision"] != float64(4) {
		t.Fatalf("body = %+v", body)
	}
	if operation.Kind != "directory_enable" || operation.Revision != 5 {
		t.Fatalf("operation = %+v", operation)
	}
}

func TestAuthProjectActOnDirectoryRejectsUnknownActionBeforeSending(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	_, err := client.Admin.AuthProjects.ActOnDirectory(
		context.Background(), "project-1", "environment-1", "destroy",
		AuthProjectDirectoryRequest{PoolID: "pool-1"},
		AuthProjectRequestOptions{IdempotencyKey: "idem-dir"},
	)
	if err == nil {
		t.Fatal("ActOnDirectory accepted an unknown action")
	}
	if len(doer.requests) != 0 {
		t.Fatalf("a request was sent for an unknown action: %d", len(doer.requests))
	}
}

func TestAuthProjectListEnvironmentsOrganisationsDecodesList(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"projectId": "project-1",
		"environmentId": "environment-1",
		"organisations": [{
			"organisationId": "org-1", "slug": "acme", "name": "Acme",
			"identityPoolId": "pool-1"
		}]
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	list, err := client.Admin.AuthProjects.ListEnvironmentOrganisations(
		context.Background(), "project-1", "environment-1",
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err != nil {
		t.Fatalf("ListEnvironmentOrganisations returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodGet,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/organisations",
		authProjectOwningUser, "")
	if len(list.Organisations) != 1 || list.Organisations[0].OrganisationID != "org-1" ||
		list.Organisations[0].IdentityPoolID != "pool-1" {
		t.Fatalf("organisations = %+v", list.Organisations)
	}
}

func TestAuthProjectReconcileEnvironmentClientsPinsRequest(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"environmentId": "environment-1",
		"revision": 6,
		"checkedAt": "2026-10-16T00:00:00Z",
		"superseded": false,
		"clients": [{
			"audience": "hosting-edge", "clientId": "custd-app-environment-1-hosting-edge",
			"action": "created"
		}]
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	report, err := client.Admin.AuthProjects.ReconcileEnvironmentClients(
		context.Background(), "project-1", "environment-1",
		AuthProjectClientSyncRequest{ExpectedRevision: 5},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: "idem-reconcile"},
	)
	if err != nil {
		t.Fatalf("ReconcileEnvironmentClients returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodPost,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/clients/reconcile",
		authProjectOwningUser, "idem-reconcile")
	if body := decodeAuthProjectBody(t, doer.requests[0]); body["expectedRevision"] != float64(5) {
		t.Fatalf("body = %+v", body)
	}
	if report.Superseded || len(report.Clients) != 1 || report.Clients[0].Action != "created" {
		t.Fatalf("report = %+v", report)
	}
}

func TestAuthProjectListInvitationsDecodesViews(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"invitations": [{
			"invitationId": "inv-1", "email": "person@example.com", "membershipRole": "member",
			"poolId": "pool-1", "organisationId": "org-1",
			"expiresAt": "2026-10-20T00:00:00Z", "createdAt": "2026-10-16T00:00:00Z",
			"state": "open", "acceptedPrincipalId": null
		}, {
			"invitationId": "inv-accepted", "email": "person@example.com", "membershipRole": "member",
			"poolId": "pool-1", "organisationId": "org-1",
			"expiresAt": "2026-10-20T00:00:00Z", "createdAt": "2026-10-16T00:00:00Z",
			"state": "accepted", "acceptedPrincipalId": "017f22e2-79b0-7f4a-8f90-ec7c7c502334"
		}]
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	list, err := client.Admin.AuthProjects.ListInvitations(
		context.Background(), "project-1", "environment-1",
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err != nil {
		t.Fatalf("ListInvitations returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodGet,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/invitations",
		authProjectOwningUser, "")
	if len(list.Invitations) != 2 || list.Invitations[0].State != "open" ||
		list.Invitations[0].InvitationID != "inv-1" || list.Invitations[0].AcceptedPrincipalID != nil {
		t.Fatalf("invitations = %+v", list.Invitations)
	}
	if principal := list.Invitations[1].AcceptedPrincipalID; principal == nil || *principal != "017f22e2-79b0-7f4a-8f90-ec7c7c502334" {
		t.Fatal("accepted invitation must retain its exact server-recorded principal")
	}
}

func TestAuthProjectIssueInvitationReturnsOneTimeLink(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"invitation": {
			"invitationId": "inv-1", "email": "person@example.com", "membershipRole": "member",
			"poolId": "pool-1", "organisationId": "org-1",
			"expiresAt": "2026-10-20T00:00:00Z", "createdAt": "2026-10-16T00:00:00Z",
			"state": "open", "delivered": true
		},
		"inviteUrl": "https://auth.example.com/invite/abc",
		"replayed": false
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	issued, err := client.Admin.AuthProjects.IssueInvitation(
		context.Background(), "project-1", "environment-1",
		AuthProjectInvitationIssueRequest{
			PoolID: "pool-1", OrganisationID: "org-1", Email: "person@example.com",
			MembershipRole: "member", ExpiresInSeconds: 86400,
		},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: "idem-invite"},
	)
	if err != nil {
		t.Fatalf("IssueInvitation returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodPost,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/invitations",
		authProjectOwningUser, "idem-invite")
	if issued.InviteURL != "https://auth.example.com/invite/abc" || !issued.Invitation.Delivered {
		t.Fatalf("issued = %+v", issued)
	}
}

func TestAuthProjectRedeliverInvitationReplacesLink(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"invitation": {
			"invitationId": "inv-1", "email": "person@example.com", "membershipRole": "member",
			"poolId": "pool-1", "organisationId": "org-1",
			"expiresAt": "2026-10-20T00:00:00Z", "createdAt": "2026-10-16T00:00:00Z",
			"state": "open"
		},
		"inviteUrl": "https://auth.example.com/invite/def",
		"replayed": false
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	issued, err := client.Admin.AuthProjects.RedeliverInvitation(
		context.Background(), "project-1", "environment-1",
		AuthProjectInvitationRedeliveryRequest{InvitationID: "inv-1"},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: "idem-redeliver"},
	)
	if err != nil {
		t.Fatalf("RedeliverInvitation returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodPost,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/invitations/redeliver",
		authProjectOwningUser, "idem-redeliver")
	if issued.InviteURL != "https://auth.example.com/invite/def" {
		t.Fatalf("issued = %+v", issued)
	}
}

func TestAuthProjectRevokeInvitationDecodesView(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"invitationId": "inv-1", "email": "person@example.com", "membershipRole": "member",
		"poolId": "pool-1", "organisationId": "org-1",
		"expiresAt": "2026-10-20T00:00:00Z", "createdAt": "2026-10-16T00:00:00Z",
		"state": "revoked"
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	view, err := client.Admin.AuthProjects.RevokeInvitation(
		context.Background(), "project-1", "environment-1",
		AuthProjectInvitationRevocationRequest{InvitationID: "inv-1"},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: "idem-revoke"},
	)
	if err != nil {
		t.Fatalf("RevokeInvitation returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodPost,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/invitations/revoke",
		authProjectOwningUser, "idem-revoke")
	if view.State != "revoked" {
		t.Fatalf("view = %+v", view)
	}
}

func TestAuthProjectInvitationRoutesRejectMissingIdempotencyKey(t *testing.T) {
	for name, call := range map[string]func(client *CustdClient) error{
		"issue": func(client *CustdClient) error {
			_, err := client.Admin.AuthProjects.IssueInvitation(
				context.Background(), "project-1", "environment-1",
				AuthProjectInvitationIssueRequest{PoolID: "pool-1"}, AuthProjectRequestOptions{})
			return err
		},
		"redeliver": func(client *CustdClient) error {
			_, err := client.Admin.AuthProjects.RedeliverInvitation(
				context.Background(), "project-1", "environment-1",
				AuthProjectInvitationRedeliveryRequest{InvitationID: "inv-1"}, AuthProjectRequestOptions{})
			return err
		},
		"revoke": func(client *CustdClient) error {
			_, err := client.Admin.AuthProjects.RevokeInvitation(
				context.Background(), "project-1", "environment-1",
				AuthProjectInvitationRevocationRequest{InvitationID: "inv-1"}, AuthProjectRequestOptions{})
			return err
		},
	} {
		t.Run(name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, `{}`)
			client := newAdminTestClient(t, doer, "http://localhost:8080")
			if err := call(client); err == nil {
				t.Fatal("call returned nil error without an idempotency key")
			}
			if len(doer.requests) != 0 {
				t.Fatalf("a request was sent without an idempotency key: %d", len(doer.requests))
			}
		})
	}
}

func TestAuthProjectGetPrincipalProfileReportsMissingRequired(t *testing.T) {
	principal := ApplicationPrincipalTarget{
		ProjectID: "project-1", EnvironmentID: "environment-1",
		DirectoryID: "directory-1", ProviderSubject: "subject-1",
	}
	doer := newCaptureDoer(http.StatusOK, `{
		"projectId": "project-1", "environmentId": "environment-1",
		"directoryId": "directory-1", "principalId": "principal-1",
		"fields": [{
			"fieldKey": "contact_email", "value": "person@example.com",
			"present": true, "required": true, "editableBy": "user", "editable": false
		}],
		"missingRequired": ["display_name"]
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	profile, err := client.Admin.AuthProjects.GetPrincipalProfile(
		context.Background(), principal, AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err != nil {
		t.Fatalf("GetPrincipalProfile returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodGet,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1"+
			"/directories/directory-1/principals/subject-1/profile",
		authProjectOwningUser, "")
	if len(profile.Fields) != 1 || !profile.Fields[0].Present || profile.Fields[0].Editable {
		t.Fatalf("fields = %+v", profile.Fields)
	}
	if !reflect.DeepEqual(profile.MissingRequired, []string{"display_name"}) {
		t.Fatalf("missingRequired = %+v", profile.MissingRequired)
	}
}

func TestAuthProjectSetPrincipalProfilePinsValuesAndKey(t *testing.T) {
	principal := ApplicationPrincipalTarget{
		ProjectID: "project-1", EnvironmentID: "environment-1",
		DirectoryID: "directory-1", ProviderSubject: "subject-1",
	}
	doer := newCaptureDoer(http.StatusOK, `{
		"projectId": "project-1", "environmentId": "environment-1",
		"directoryId": "directory-1", "principalId": "principal-1",
		"fields": [], "missingRequired": []
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	_, err := client.Admin.AuthProjects.SetPrincipalProfile(
		context.Background(), principal,
		ApplicationProfileValuesRequest{Values: []ApplicationProfileValue{
			{FieldKey: "display_name", Value: "Ada"},
		}},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: "idem-profile"},
	)
	if err != nil {
		t.Fatalf("SetPrincipalProfile returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodPost,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1"+
			"/directories/directory-1/principals/subject-1/profile",
		authProjectOwningUser, "idem-profile")
	var sent ApplicationProfileValuesRequest
	if err := json.Unmarshal(doer.requests[0].Body, &sent); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(sent.Values) != 1 || sent.Values[0].FieldKey != "display_name" || sent.Values[0].Value != "Ada" {
		t.Fatalf("body = %+v", sent)
	}
}

func TestAuthProjectListPrincipalMembershipsDecodesChoices(t *testing.T) {
	principal := ApplicationPrincipalTarget{
		ProjectID: "project-1", EnvironmentID: "environment-1",
		DirectoryID: "directory-1", ProviderSubject: "subject-1",
	}
	doer := newCaptureDoer(http.StatusOK, `{
		"projectId": "project-1", "environmentId": "environment-1",
		"directoryId": "directory-1", "principalId": "principal-1",
		"memberships": [{
			"organisationId": "org-1", "organisationSlug": "acme",
			"organisationName": "Acme", "role": "member"
		}]
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	contextResult, err := client.Admin.AuthProjects.ListPrincipalMemberships(
		context.Background(), principal, AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err != nil {
		t.Fatalf("ListPrincipalMemberships returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodGet,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1"+
			"/directories/directory-1/principals/subject-1/memberships",
		authProjectOwningUser, "")
	if len(contextResult.Memberships) != 1 || contextResult.Memberships[0].Role != "member" {
		t.Fatalf("memberships = %+v", contextResult.Memberships)
	}
}

func TestAuthProjectDeleteProjectDecodesReleasedCounts(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"projectId": "project-1", "slug": "hosting-eu", "name": "Hosting EU",
		"released": {
			"environments": 2, "directories": 2, "organisations": 2, "memberships": 1,
			"invitations": 3, "audienceBindings": 1, "operations": 4
		}
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	deletion, err := client.Admin.AuthProjects.DeleteProject(
		context.Background(), "project-1",
		AuthProjectDeletionRequest{ConfirmSlug: "hosting-eu"},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err != nil {
		t.Fatalf("DeleteProject returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodPost,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/delete",
		authProjectOwningUser, "")
	if body := decodeAuthProjectBody(t, doer.requests[0]); body["confirmSlug"] != "hosting-eu" {
		t.Fatalf("body = %+v", body)
	}
	if deletion.Slug != "hosting-eu" || deletion.Released.Environments != 2 ||
		deletion.Released.AudienceBindings != 1 {
		t.Fatalf("deletion = %+v", deletion)
	}
}

func TestAuthProjectDeleteEnvironmentDecodesReceipt(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"projectId": "project-1", "projectSlug": "hosting-eu",
		"environmentId": "environment-1", "slug": "staging", "name": "Staging",
		"released": {
			"directories": 1, "organisations": 1, "memberships": 0, "invitations": 1,
			"audienceBindings": 1, "operations": 2, "profileValues": 3
		}
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	deletion, err := client.Admin.AuthProjects.DeleteEnvironment(
		context.Background(), "project-1", "environment-1",
		AuthProjectEnvironmentDeletionRequest{ConfirmSlug: "staging"},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err != nil {
		t.Fatalf("DeleteEnvironment returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodPost,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/delete",
		authProjectOwningUser, "")
	if deletion.ProjectSlug != "hosting-eu" || deletion.Slug != "staging" ||
		deletion.Released.ProfileValues != 3 {
		t.Fatalf("deletion = %+v", deletion)
	}
}

func TestAuthProjectOffboardEnvironmentDecodesRetainedReceipt(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"projectId": "project-1", "projectSlug": "hosting-eu",
		"environmentId": "environment-1", "slug": "staging", "name": "Staging",
		"authorityWithdrawn": true, "loginPaused": true, "revision": 8,
		"clientCleanup": {"status": "complete", "clientsRemoved": 2},
		"retained": {
			"directories": 1, "organisations": 1, "memberships": 3,
			"invitations": 4, "profileValues": 5, "reservedSlots": 1
		}
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	offboarding, err := client.Admin.AuthProjects.OffboardEnvironment(
		context.Background(), "project-1", "environment-1",
		AuthProjectEnvironmentOffboardRequest{ConfirmSlug: "staging"},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err != nil {
		t.Fatalf("OffboardEnvironment returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodPost,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/offboard",
		authProjectOwningUser, "")
	if body := decodeAuthProjectBody(t, doer.requests[0]); body["confirmSlug"] != "staging" {
		t.Fatalf("body = %+v", body)
	}
	if !offboarding.AuthorityWithdrawn || !offboarding.LoginPaused || offboarding.Revision != 8 {
		t.Fatalf("offboarding = %+v", offboarding)
	}
	if offboarding.ClientCleanup.Status != "complete" || offboarding.ClientCleanup.ClientsRemoved != 2 {
		t.Fatalf("client cleanup = %+v", offboarding.ClientCleanup)
	}
	if offboarding.Retained.ReservedSlots != 1 || offboarding.Retained.ProfileValues != 5 ||
		offboarding.Retained.Memberships != 3 {
		t.Fatalf("retained = %+v", offboarding.Retained)
	}
}
