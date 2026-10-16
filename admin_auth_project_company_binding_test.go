package custd

import (
	"context"
	"net/http"
	"testing"
)

func TestAuthProjectsBindCompanyIsAHumanOwnerActionWithNoBody(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"projectId": "01957abc-0000-7000-8000-000000000001",
		"companyUuid": "01957abc-0000-7000-8000-0000000000bb"
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	binding, err := client.Admin.AuthProjects.BindProjectCompany(
		context.Background(),
		"01957abc-0000-7000-8000-000000000001",
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err != nil {
		t.Fatalf("BindProjectCompany returned error: %v", err)
	}
	if binding.CompanyUUID != "01957abc-0000-7000-8000-0000000000bb" ||
		binding.ProjectID != "01957abc-0000-7000-8000-000000000001" {
		t.Fatalf("binding = %+v", binding)
	}

	request := doer.requests[0]
	if request.Method != http.MethodPost ||
		request.URL != "http://localhost:8080/api/v1/admin/auth-projects/01957abc-0000-7000-8000-000000000001/company-binding" {
		t.Fatalf("request = %s %s", request.Method, request.URL)
	}
	// The service refuses a non-empty body, so the binding must send none.
	if len(request.Body) != 0 {
		t.Fatalf("body = %q", string(request.Body))
	}
	if request.Headers[authProjectOwningUserHeader] != authProjectOwningUser {
		t.Fatalf("owning user header = %q", request.Headers[authProjectOwningUserHeader])
	}
	if _, present := request.Headers["Idempotency-Key"]; present {
		t.Fatal("company binding is safe to repeat and must not require an idempotency key")
	}
}
