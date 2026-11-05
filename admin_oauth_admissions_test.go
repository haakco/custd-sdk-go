package custd

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"
)

// TestAdminOAuthClientAdmissionsUseDocumentedContract covers the four tenant
// admission operations in one place: the route, the query, the identifier
// escaping and the exact camelCase body fields. A wrong field name here would be
// a silent 400 at runtime, because the server validates the body it is given.
func TestAdminOAuthClientAdmissionsUseDocumentedContract(t *testing.T) {
	const (
		admissionID      = "adm/1"
		securityRevision = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	)
	doer := newCaptureDoer(http.StatusOK, `{"admissions":[{"admissionUuid":"adm/1",`+
		`"companyUuid":"company-1","clientId":"client-1","requestedSecurityRevision":"`+securityRevision+`",`+
		`"requestedGrants":[{"resourceUri":"https://api.example.com","scopes":["orders.read"]}],`+
		`"approvedGrants":[],"revision":2}]}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	list, err := client.Admin.OAuthClientAdmissions.List(context.Background(), "acme")
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(list.Admissions) != 1 || list.Admissions[0].ClientID != "client-1" ||
		list.Admissions[0].Revision != 2 || len(list.Admissions[0].RequestedGrants) != 1 ||
		list.Admissions[0].RequestedGrants[0].ResourceURI != "https://api.example.com" {
		t.Fatalf("admissions = %+v", list.Admissions)
	}
	assertAdmissionRequest(t, doer.requests[0], http.MethodGet,
		"http://localhost:8080/api/v1/admin/oauth-client-admissions?companySlug=acme", nil)

	doer.body = `{"admission":{"admissionUuid":"adm/1","companyUuid":"company-1","clientId":"client-1",` +
		`"requestedSecurityRevision":"` + securityRevision + `","requestedGrants":[],"approvedGrants":[],"revision":3},` +
		`"metadata":{"clientName":"External App","redirectUris":["https://app.example.com/cb"],` +
		`"securityRevision":"` + securityRevision + `","expiresAt":"2026-12-01T00:00:00Z"}}`
	review, err := client.Admin.OAuthClientAdmissions.Review(context.Background(), admissionID,
		AdminOAuthClientAdmissionReviewRequest{CompanySlug: "acme"})
	if err != nil {
		t.Fatalf("Review returned error: %v", err)
	}
	if review.Admission.Revision != 3 || review.Metadata.ClientName != "External App" ||
		len(review.Metadata.RedirectURIs) != 1 || review.Metadata.ExpiresAt == "" {
		t.Fatalf("review = %+v", review)
	}
	assertAdmissionRequest(t, doer.requests[1], http.MethodPost,
		"http://localhost:8080/api/v1/admin/oauth-client-admissions/adm%2F1/review",
		map[string]any{"companySlug": "acme"})

	doer.body = `{"admission":{"admissionUuid":"adm/1","companyUuid":"company-1","clientId":"client-1",` +
		`"requestedSecurityRevision":"` + securityRevision + `","approvedSecurityRevision":"` + securityRevision + `",` +
		`"requestedGrants":[],"approvedGrants":[{"resourceUri":"https://api.example.com","scopes":["orders.read"]}],` +
		`"revision":4,"approvedAt":"2026-11-01T00:00:00Z"}}`
	approved, err := client.Admin.OAuthClientAdmissions.Approve(context.Background(), admissionID,
		AdminOAuthClientAdmissionApproveRequest{
			CompanySlug:      "acme",
			ClientID:         "client-1",
			ExpectedRevision: 3,
			SecurityRevision: securityRevision,
			Grants:           []AdminOAuthClientAdmissionGrant{{ResourceURI: "https://api.example.com", Scopes: []string{"orders.read"}}},
		})
	if err != nil {
		t.Fatalf("Approve returned error: %v", err)
	}
	if approved.Admission.Revision != 4 || approved.Admission.ApprovedAt == nil ||
		len(approved.Admission.ApprovedGrants) != 1 {
		t.Fatalf("approved = %+v", approved.Admission)
	}
	assertAdmissionRequest(t, doer.requests[2], http.MethodPost,
		"http://localhost:8080/api/v1/admin/oauth-client-admissions/adm%2F1/approve",
		map[string]any{
			"companySlug":      "acme",
			"clientId":         "client-1",
			"expectedRevision": float64(3),
			"securityRevision": securityRevision,
			"grants": []any{map[string]any{
				"resourceUri": "https://api.example.com",
				"scopes":      []any{"orders.read"},
			}},
		})

	doer.body = `{"admission":{"admissionUuid":"adm/1","companyUuid":"company-1","clientId":"client-1",` +
		`"requestedSecurityRevision":"` + securityRevision + `","requestedGrants":[],"approvedGrants":[],` +
		`"revision":5,"revokedAt":"2026-11-02T00:00:00Z"}}`
	revoked, err := client.Admin.OAuthClientAdmissions.Revoke(context.Background(), admissionID,
		AdminOAuthClientAdmissionRevokeRequest{CompanySlug: "acme", ExpectedRevision: 4})
	if err != nil {
		t.Fatalf("Revoke returned error: %v", err)
	}
	if revoked.Admission.Revision != 5 || revoked.Admission.RevokedAt == nil {
		t.Fatalf("revoked = %+v", revoked.Admission)
	}
	assertAdmissionRequest(t, doer.requests[3], http.MethodPost,
		"http://localhost:8080/api/v1/admin/oauth-client-admissions/adm%2F1/revoke",
		map[string]any{"companySlug": "acme", "expectedRevision": float64(4)})

	if len(doer.requests) != 4 {
		t.Fatalf("requests = %d, want 4", len(doer.requests))
	}
}

// TestAdminOAuthClientAdmissionConflictStaysVisible proves a 409 stays a typed
// SDK error and is not retried or swallowed: a changed admission requires a new
// review, so replaying the same approval would repeat a rejected decision.
func TestAdminOAuthClientAdmissionConflictStaysVisible(t *testing.T) {
	doer := newCaptureDoer(http.StatusConflict,
		`{"title":"revision_conflict","status":409,"detail":"admission changed","code":"revision_conflict"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	_, err := client.Admin.OAuthClientAdmissions.Approve(context.Background(), "adm-1",
		AdminOAuthClientAdmissionApproveRequest{
			CompanySlug:      "acme",
			ClientID:         "client-1",
			ExpectedRevision: 3,
			SecurityRevision: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Grants:           []AdminOAuthClientAdmissionGrant{{ResourceURI: "https://api.example.com", Scopes: []string{"orders.read"}}},
		})
	if err == nil {
		t.Fatal("Approve accepted a 409")
	}
	var requestErr *RequestError
	if !errors.As(err, &requestErr) {
		t.Fatalf("error type = %T, want *RequestError", err)
	}
	if requestErr.StatusCode != http.StatusConflict || requestErr.Problem == nil ||
		requestErr.Problem.Code != "revision_conflict" {
		t.Fatalf("error = %+v problem = %+v", requestErr, requestErr.Problem)
	}
	if len(doer.requests) != 1 {
		t.Fatalf("requests = %d, want 1: a conflict must not be retried", len(doer.requests))
	}
}

func assertAdmissionRequest(t *testing.T, req *HTTPRequest, method, url string, body map[string]any) {
	t.Helper()
	if req.Method != method {
		t.Fatalf("method = %s, want %s", req.Method, method)
	}
	if req.URL != url {
		t.Fatalf("url = %s, want %s", req.URL, url)
	}
	if req.Headers["Idempotency-Key"] != "" {
		t.Fatalf("unexpected Idempotency-Key = %q", req.Headers["Idempotency-Key"])
	}
	if body == nil {
		if len(req.Body) != 0 {
			t.Fatalf("body = %s, want empty", req.Body)
		}
		return
	}
	var sent map[string]any
	if err := json.Unmarshal(req.Body, &sent); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !reflect.DeepEqual(sent, body) {
		t.Fatalf("body = %#v, want %#v", sent, body)
	}
}
