package custd

import (
	"context"
	"net/http"
	"testing"
)

func TestAuthProjectInvitationByIDUsesTypedRoutes(t *testing.T) {
	tests := []struct {
		name         string
		verb         string
		responseBody string
		wantURL      string
		wantInvite   string
		wantReplayed bool
	}{
		{
			name:         "revoke",
			verb:         AuthProjectInvitationVerbRevoke,
			responseBody: `{"invitationId":"invite-1","email":"person@example.com","state":"revoked"}`,
			wantURL: "http://localhost:8080/api/v1/admin/auth-projects/project-1/" +
				"environments/environment-1/invitations/revoke",
		},
		{
			name: "redeliver",
			verb: AuthProjectInvitationVerbRedeliver,
			responseBody: `{"invitation":{"invitationId":"invite-1","email":"person@example.com",` +
				`"state":"open"},"inviteUrl":"https://example.com/invite/abc","replayed":true}`,
			wantURL: "http://localhost:8080/api/v1/admin/auth-projects/project-1/" +
				"environments/environment-1/invitations/redeliver",
			wantInvite:   "https://example.com/invite/abc",
			wantReplayed: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, test.responseBody)
			doer.headers = map[string]string{requestIDHeader: "req-1"}
			client := newAdminTestClient(t, doer, "http://localhost:8080")

			result, metadata, err := client.Admin.AuthProjects.InvitationByID(
				context.Background(),
				AuthProjectInvitationByIDTarget{
					Verb:          test.verb,
					ProjectID:     "project-1",
					EnvironmentID: "environment-1",
					InvitationID:  "invite-1",
				},
				AuthProjectRequestOptions{
					OwningUserUUID: authProjectOwningUser,
					IdempotencyKey: "idem-1",
				},
			)
			if err != nil {
				t.Fatalf("InvitationByID returned error: %v", err)
			}
			if metadata.RequestID != "req-1" {
				t.Fatalf("request id = %q", metadata.RequestID)
			}
			if result.Invitation.InvitationID != "invite-1" {
				t.Fatalf("invitation = %+v", result.Invitation)
			}
			if result.InviteURL != test.wantInvite || result.Replayed != test.wantReplayed {
				t.Fatalf("result = %+v", result)
			}

			if len(doer.requests) != 1 {
				t.Fatalf("requests = %d, want 1", len(doer.requests))
			}
			req := doer.requests[0]
			if req.Method != http.MethodPost {
				t.Fatalf("method = %s, want %s", req.Method, http.MethodPost)
			}
			if req.URL != test.wantURL {
				t.Fatalf("url = %s, want %s", req.URL, test.wantURL)
			}
			if req.Headers["Idempotency-Key"] != "idem-1" {
				t.Fatalf("idempotency key header = %q", req.Headers["Idempotency-Key"])
			}
			if req.Headers[authProjectOwningUserHeader] != authProjectOwningUser {
				t.Fatalf("owning user header = %q", req.Headers[authProjectOwningUserHeader])
			}
			body := decodeAuthProjectBody(t, req)
			if body["invitationId"] != "invite-1" {
				t.Fatalf("body = %+v", body)
			}
		})
	}
}

func TestAuthProjectInvitationByIDRefusesUnknownVerbAndMissingKey(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	target := AuthProjectInvitationByIDTarget{
		Verb:          "delete",
		ProjectID:     "project-1",
		EnvironmentID: "environment-1",
		InvitationID:  "invite-1",
	}

	if _, _, err := client.Admin.AuthProjects.InvitationByID(
		context.Background(), target,
		AuthProjectRequestOptions{IdempotencyKey: "idem-1"},
	); err == nil {
		t.Fatalf("InvitationByID with an unknown verb returned no error")
	}
	if _, _, err := client.Admin.AuthProjects.InvitationByID(
		context.Background(),
		AuthProjectInvitationByIDTarget{
			Verb:          AuthProjectInvitationVerbRevoke,
			ProjectID:     "project-1",
			EnvironmentID: "environment-1",
			InvitationID:  "invite-1",
		},
		AuthProjectRequestOptions{},
	); err == nil {
		t.Fatalf("InvitationByID without an idempotency key returned no error")
	}
	if len(doer.requests) != 0 {
		t.Fatalf("requests = %d, want 0", len(doer.requests))
	}
}
