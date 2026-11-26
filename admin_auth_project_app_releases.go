package custd

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// A hosted environment's releases. A release is created from its manifest, each
// listed file is uploaded as raw bytes, the release is verified once every file
// reads back intact, and a verified release is promoted to the environment's
// host. The environment must have hosting enabled in its desired state.

// AppReleaseState is a release's lifecycle state.
type AppReleaseState string

const (
	// AppReleaseStatePending accepts uploads until every file reads back intact.
	AppReleaseStatePending AppReleaseState = "pending"
	// AppReleaseStateVerified accepts no further uploads and can be promoted.
	AppReleaseStateVerified AppReleaseState = "verified"
)

// AppReleaseCacheMode is the caching a cache rule applies to the files it matches.
type AppReleaseCacheMode string

const (
	AppReleaseCacheImmutable  AppReleaseCacheMode = "immutable"
	AppReleaseCacheRevalidate AppReleaseCacheMode = "revalidate"
	AppReleaseCacheShort      AppReleaseCacheMode = "short"
	AppReleaseCacheNoStore    AppReleaseCacheMode = "no-store"
)

// AppReleaseFramingMode is who may frame the routes a header rule matches.
type AppReleaseFramingMode string

const (
	AppReleaseFramingDeny       AppReleaseFramingMode = "deny"
	AppReleaseFramingSameOrigin AppReleaseFramingMode = "same-origin"
	AppReleaseFramingAny        AppReleaseFramingMode = "any"
)

// AppReleaseFile is one manifest file. SHA256 is the lowercase hex digest of the
// bytes the upload must carry.
type AppReleaseFile struct {
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	SHA256      string `json:"sha256"`
	ContentType string `json:"contentType"`
}

// AppReleaseRouting is how the host serves the release. IndexFile defaults to
// index.html on the server.
type AppReleaseRouting struct {
	IndexFile       string                     `json:"indexFile,omitempty"`
	NotFoundPage    string                     `json:"notFoundPage,omitempty"`
	SPAFallbacks    []AppReleaseSPAFallback    `json:"spaFallbacks,omitempty"`
	Headers         []AppReleaseHeaderRule     `json:"headers,omitempty"`
	Cache           []AppReleaseCacheRule      `json:"cache,omitempty"`
	ShareableRoutes []AppReleaseShareableRoute `json:"shareableRoutes,omitempty"`
	SiteCard        *AppReleaseSiteCard        `json:"siteCard,omitempty"`
}

// AppReleaseSiteCard is the link preview a shared link to the site shows.
type AppReleaseSiteCard struct {
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	CardTemplate string `json:"cardTemplate"`
}

// AppReleaseSPAFallback serves File for any path under Prefix that names no file.
type AppReleaseSPAFallback struct {
	Prefix string `json:"prefix"`
	File   string `json:"file"`
}

// AppReleaseHeaderRule sets framing and extra response headers on a route.
type AppReleaseHeaderRule struct {
	Route   string                     `json:"route"`
	Framing AppReleaseFramingMode      `json:"framing,omitempty"`
	Headers []AppReleaseResponseHeader `json:"headers,omitempty"`
}

// AppReleaseResponseHeader is one extra response header.
type AppReleaseResponseHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// AppReleaseCacheRule applies a cache mode to the files Match selects.
type AppReleaseCacheRule struct {
	Match string              `json:"match"`
	Mode  AppReleaseCacheMode `json:"mode"`
}

// AppReleaseShareableRoute is a route the host may serve with share previews.
type AppReleaseShareableRoute struct {
	Route        string `json:"route"`
	Kind         string `json:"kind"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	CardTemplate string `json:"cardTemplate"`
}

// AppReleaseCreateRequest is a new release's files and routing. The server
// assigns the release identifier and creation time.
type AppReleaseCreateRequest struct {
	Files   []AppReleaseFile  `json:"files"`
	Routing AppReleaseRouting `json:"routing"`
}

// AppReleaseManifest is the stored manifest of one release.
type AppReleaseManifest struct {
	Version   int               `json:"version"`
	ReleaseID string            `json:"releaseId"`
	CreatedAt string            `json:"createdAt"`
	Files     []AppReleaseFile  `json:"files"`
	Routing   AppReleaseRouting `json:"routing"`
}

// AppRelease is one stored release with its manifest.
type AppRelease struct {
	ReleaseID  string             `json:"releaseId"`
	State      AppReleaseState    `json:"state"`
	Active     bool               `json:"active"`
	Label      string             `json:"label,omitempty"`
	FileCount  int                `json:"fileCount"`
	TotalBytes int64              `json:"totalBytes"`
	CreatedAt  string             `json:"createdAt"`
	VerifiedAt string             `json:"verifiedAt,omitempty"`
	Manifest   AppReleaseManifest `json:"manifest"`
}

// AppReleaseSummary is one release in a list, without its file manifest.
type AppReleaseSummary struct {
	ReleaseID  string          `json:"releaseId"`
	State      AppReleaseState `json:"state"`
	Active     bool            `json:"active"`
	Label      string          `json:"label,omitempty"`
	FileCount  int             `json:"fileCount"`
	TotalBytes int64           `json:"totalBytes"`
	CreatedAt  string          `json:"createdAt"`
	VerifiedAt string          `json:"verifiedAt,omitempty"`
}

// AppReleaseList is an environment's most recent releases, newest first.
type AppReleaseList struct {
	Releases []AppReleaseSummary `json:"releases"`
}

// AppReleaseUploadedFile is the receipt for one stored release file.
type AppReleaseUploadedFile struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

// AppReleasePromotion is the receipt for one promotion. PreviousReleaseID is
// empty when the environment served nothing before.
type AppReleasePromotion struct {
	ReleaseID         string `json:"releaseId"`
	Host              string `json:"host"`
	PreviousReleaseID string `json:"previousReleaseId,omitempty"`
	PromotedAt        string `json:"promotedAt"`
}

// ListAppReleases returns the environment's most recent releases. A limit of 0
// uses the server default (20); the server caps it at 100.
func (c *AuthProjectAdminClient) ListAppReleases(
	ctx context.Context, projectID, environmentID string, limit int, options AuthProjectRequestOptions,
) (*AppReleaseList, error) {
	path := appReleasesPath(projectID, environmentID)
	if limit > 0 {
		path += "?limit=" + strconv.Itoa(limit)
	}
	var out AppReleaseList
	if err := c.admin.requestWithHeaders(ctx, http.MethodGet, path, nil, &out, authProjectHeaders(options)); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateAppRelease records a new, pending release from its manifest.
func (c *AuthProjectAdminClient) CreateAppRelease(
	ctx context.Context, projectID, environmentID string, req AppReleaseCreateRequest, options AuthProjectRequestOptions,
) (*AppRelease, error) {
	var out AppRelease
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, appReleasesPath(projectID, environmentID), req, &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// UploadAppReleaseFile stores one manifest file of a pending release. The body
// is the file's bytes; the server refuses bytes whose size or sha256 differ
// from the manifest.
func (c *AuthProjectAdminClient) UploadAppReleaseFile(
	ctx context.Context, projectID, environmentID, releaseID, filePath string, data []byte,
	options AuthProjectRequestOptions,
) (*AppReleaseUploadedFile, error) {
	var out AppReleaseUploadedFile
	path := appReleasePath(projectID, environmentID, releaseID) + "/files?path=" + url.QueryEscape(filePath)
	if _, err := sendAuthProjectRawRequest(
		ctx, c.admin, http.MethodPost, path, data, "application/octet-stream", &out, authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// VerifyAppRelease reads every file back and marks the release verified when
// all of them match the manifest.
func (c *AuthProjectAdminClient) VerifyAppRelease(
	ctx context.Context, projectID, environmentID, releaseID string, options AuthProjectRequestOptions,
) (*AppRelease, error) {
	var out AppRelease
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, appReleasePath(projectID, environmentID, releaseID)+"/verify", nil, &out,
		authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

// PromoteAppRelease makes a verified release the one the environment's host serves.
func (c *AuthProjectAdminClient) PromoteAppRelease(
	ctx context.Context, projectID, environmentID, releaseID string, options AuthProjectRequestOptions,
) (*AppReleasePromotion, error) {
	var out AppReleasePromotion
	if err := c.admin.requestWithHeaders(
		ctx, http.MethodPost, appReleasePath(projectID, environmentID, releaseID)+"/promote", nil, &out,
		authProjectHeaders(options),
	); err != nil {
		return nil, err
	}
	return &out, nil
}

func appReleasesPath(projectID, environmentID string) string {
	return authProjectEnvironmentPath(projectID, environmentID) + "/releases"
}

func appReleasePath(projectID, environmentID, releaseID string) string {
	return appReleasesPath(projectID, environmentID) + "/" + url.PathEscape(releaseID)
}
