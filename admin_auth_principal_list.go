package custd

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// ApplicationPrincipalListQuery selects one page of an environment's application
// principals. Cursor is the opaque nextCursor the previous page returned; it is
// dropped only when absent or genuinely empty and is otherwise sent unchanged. Limit
// bounds the page; Custd applies its own default of 100 and maximum of 100, so a
// positive limit is passed through unchanged and the server's bound stays the single
// source of truth.
type ApplicationPrincipalListQuery struct {
	Cursor string
	Limit  int
}

// ApplicationPrincipalListItem addresses one application principal admitted to one
// directory. ProviderSubject is the opaque management address the principal routes
// take; PrincipalID is the application identity UUID that address resolves to.
type ApplicationPrincipalListItem struct {
	PrincipalID     string `json:"principalId"`
	DirectoryID     string `json:"directoryId"`
	ProviderSubject string `json:"providerSubject"`
	Enabled         bool   `json:"enabled"`
}

// ApplicationPrincipalListResponse is one page of an environment's application
// principals. Principals is empty, never null, when the environment holds none.
// NextCursor is empty on the last page.
type ApplicationPrincipalListResponse struct {
	Principals []ApplicationPrincipalListItem `json:"principals"`
	NextCursor string                         `json:"nextCursor,omitempty"`
}

// ListPrincipals lists the application principals every directory in the
// environment holds, ordered so the cursor selects a stable key. It carries the
// owning-user header from options like the other project-auth reads and pages with
// the opaque cursor the previous page returned.
func (c *AuthProjectAdminClient) ListPrincipals(
	ctx context.Context,
	projectID string,
	environmentID string,
	query ApplicationPrincipalListQuery,
	options AuthProjectRequestOptions,
) (*ApplicationPrincipalListResponse, error) {
	var out ApplicationPrincipalListResponse
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodGet,
		authProjectEnvironmentPath(projectID, environmentID)+"/principals"+query.encode(),
		nil, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// encode renders the query parameters the route accepts. The cursor is opaque, so
// only an absent or genuinely empty value is left off; every other value is
// url-encoded byte-for-byte and never normalized. The limit is omitted when unset
// so the service applies its own bound.
func (query ApplicationPrincipalListQuery) encode() string {
	params := url.Values{}
	if query.Cursor != "" {
		params.Set("cursor", query.Cursor)
	}
	if query.Limit > 0 {
		params.Set("limit", strconv.Itoa(query.Limit))
	}
	if len(params) == 0 {
		return ""
	}
	return "?" + params.Encode()
}
