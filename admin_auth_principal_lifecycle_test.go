package custd

import (
	"context"
	"net/http"
	"testing"
)

func TestPrincipalErasureRequestAndReceipt(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"principalId":"principal-1","status":"applied","countsRecorded":false,"replayed":true}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	target := ApplicationPrincipalTarget{ProjectID: "project-1", EnvironmentID: "environment-1", DirectoryID: "directory-1", ProviderSubject: "subject-1"}
	receipt, err := client.Admin.AuthProjects.ErasePrincipal(context.Background(), target, ApplicationPrincipalErasureRequest{Confirm: true}, AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: "erase-principal"})
	if err != nil {
		t.Fatal(err)
	}
	assertAuthProjectRequest(t, doer.requests[0], http.MethodPost, "http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/directories/directory-1/principals/subject-1/erase", authProjectOwningUser, "erase-principal")
	body := decodeAuthProjectBody(t, doer.requests[0])
	if len(body) != 1 || body["confirm"] != true {
		t.Fatalf("body = %+v", body)
	}
	if receipt.PrincipalID != "principal-1" || receipt.Status != "applied" || receipt.CountsRecorded || !receipt.Replayed {
		t.Fatalf("receipt = %+v", receipt)
	}
}

func TestPrincipalErasureRefusesMissingKey(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	_, err := client.Admin.AuthProjects.ErasePrincipal(context.Background(), ApplicationPrincipalTarget{}, ApplicationPrincipalErasureRequest{Confirm: true}, AuthProjectRequestOptions{})
	if err == nil || len(doer.requests) != 0 {
		t.Fatalf("error = %v; requests = %d", err, len(doer.requests))
	}
}

func TestPrincipalStateOperations(t *testing.T) {
	target := ApplicationPrincipalTarget{ProjectID: "project/1", EnvironmentID: "environment/1", DirectoryID: "directory/1", ProviderSubject: "subject/1"}
	for _, action := range []string{"suspend", "restore"} {
		t.Run(action, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, `{"principalId":"principal-1","enabled":false,"sessionsRevoked":2,"revision":8,"replayed":false}`)
			client := newAdminTestClient(t, doer, "http://localhost:8080")
			options := AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser, IdempotencyKey: "principal-state"}
			var receipt *ApplicationPrincipalSuspension
			var err error
			if action == "suspend" {
				receipt, err = client.Admin.AuthProjects.SuspendPrincipal(context.Background(), target, options)
			} else {
				receipt, err = client.Admin.AuthProjects.RestorePrincipal(context.Background(), target, options)
			}
			if err != nil {
				t.Fatal(err)
			}
			assertAuthProjectRequest(t, doer.requests[0], http.MethodPost, "http://localhost:8080/api/v1/admin/auth-projects/project%2F1/environments/environment%2F1/directories/directory%2F1/principals/subject%2F1/"+action, authProjectOwningUser, "principal-state")
			if receipt.PrincipalID != "principal-1" || receipt.SessionsRevoked != 2 || receipt.Revision != 8 {
				t.Fatalf("receipt = %+v", receipt)
			}
		})
	}
}

func TestPrincipalStateRefusesMissingKeyBeforeTransport(t *testing.T) {
	for _, action := range []string{"suspend", "restore"} {
		t.Run(action, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, `{}`)
			client := newAdminTestClient(t, doer, "http://localhost:8080")
			var err error
			if action == "suspend" {
				_, err = client.Admin.AuthProjects.SuspendPrincipal(context.Background(), ApplicationPrincipalTarget{}, AuthProjectRequestOptions{})
			} else {
				_, err = client.Admin.AuthProjects.RestorePrincipal(context.Background(), ApplicationPrincipalTarget{}, AuthProjectRequestOptions{})
			}
			if err == nil || len(doer.requests) != 0 {
				t.Fatalf("error = %v; requests = %d", err, len(doer.requests))
			}
		})
	}
}
