package custd

import (
	"context"
	"net/http"
	"testing"
)

func TestAuthProjectsListPrincipalsSendsRouteLimitCursorAndOwningUser(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{
		"principals": [{
			"principalId": "01957abc-0000-7000-8000-0000000000c1",
			"directoryId": "01957abc-0000-7000-8000-0000000000d1",
			"providerSubject": "subject-1",
			"enabled": true
		}],
		"nextCursor": "cursor-2"
	}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	page, err := client.Admin.AuthProjects.ListPrincipals(
		context.Background(),
		"project-1",
		"environment-1",
		ApplicationPrincipalListQuery{Cursor: "cursor-1", Limit: 50},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err != nil {
		t.Fatalf("ListPrincipals returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodGet,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1"+
			"/principals?cursor=cursor-1&limit=50",
		authProjectOwningUser, "")
	if len(page.Principals) != 1 {
		t.Fatalf("principals = %+v", page.Principals)
	}
	item := page.Principals[0]
	if item.PrincipalID != "01957abc-0000-7000-8000-0000000000c1" ||
		item.DirectoryID != "01957abc-0000-7000-8000-0000000000d1" ||
		item.ProviderSubject != "subject-1" || !item.Enabled {
		t.Fatalf("principal = %+v", item)
	}
	if page.NextCursor != "cursor-2" {
		t.Fatalf("nextCursor = %q", page.NextCursor)
	}
}

func TestAuthProjectsListPrincipalsOmitsEmptyQueryAndDecodesEmptyPage(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"principals": []}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")

	page, err := client.Admin.AuthProjects.ListPrincipals(
		context.Background(),
		"project-1",
		"environment-1",
		ApplicationPrincipalListQuery{},
		AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
	)
	if err != nil {
		t.Fatalf("ListPrincipals returned error: %v", err)
	}

	assertAuthProjectRequest(t, doer.requests[0], http.MethodGet,
		"http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/principals",
		authProjectOwningUser, "")
	if page.Principals == nil || len(page.Principals) != 0 {
		t.Fatalf("principals = %+v, want empty non-nil slice", page.Principals)
	}
	if page.NextCursor != "" {
		t.Fatalf("nextCursor = %q, want empty", page.NextCursor)
	}
}

func TestAuthProjectsListPrincipalsRoundTripsOpaqueCursorUnchanged(t *testing.T) {
	base := "http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/principals"
	// The cursor is opaque: only an absent or genuinely empty value is dropped, and
	// every other string is url-encoded byte-for-byte, including surrounding and
	// whitespace-only values the client must not normalize.
	cases := []struct {
		cursor string
		want   string
	}{
		{" cursor/2?x=+ ", base + "?cursor=+cursor%2F2%3Fx%3D%2B+"},
		{"   ", base + "?cursor=+++"},
	}
	for _, tc := range cases {
		doer := newCaptureDoer(http.StatusOK, `{"principals": []}`)
		client := newAdminTestClient(t, doer, "http://localhost:8080")

		if _, err := client.Admin.AuthProjects.ListPrincipals(
			context.Background(),
			"project-1",
			"environment-1",
			ApplicationPrincipalListQuery{Cursor: tc.cursor},
			AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
		); err != nil {
			t.Fatalf("ListPrincipals(%q) returned error: %v", tc.cursor, err)
		}

		assertAuthProjectRequest(t, doer.requests[0], http.MethodGet, tc.want, authProjectOwningUser, "")
	}
}
