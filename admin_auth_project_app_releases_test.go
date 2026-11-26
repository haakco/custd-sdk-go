package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

const appReleasesURL = "http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/releases"

func TestAppReleaseOperationsSendTheirRoutes(t *testing.T) {
	options := AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser}
	tests := []struct {
		name    string
		body    string
		call    func(*AuthProjectAdminClient) error
		method  string
		url     string
		content string
	}{
		{
			name: "list", body: `{"releases":[{"releaseId":"r1","state":"verified","active":true,"fileCount":2,"totalBytes":10,"createdAt":"2026-11-17T00:00:00Z"}]}`,
			call: func(c *AuthProjectAdminClient) error {
				list, err := c.ListAppReleases(context.Background(), "project-1", "environment-1", 5, options)
				if err == nil && (len(list.Releases) != 1 || !list.Releases[0].Active) {
					t.Fatalf("list = %+v", list)
				}
				return err
			},
			method: http.MethodGet, url: appReleasesURL + "?limit=5",
		},
		{
			name: "create", body: `{"releaseId":"r1","state":"pending","manifest":{"version":1,"files":[]}}`,
			call: func(c *AuthProjectAdminClient) error {
				_, err := c.CreateAppRelease(context.Background(), "project-1", "environment-1", AppReleaseCreateRequest{
					Files:   []AppReleaseFile{{Path: "index.html", Size: 4, SHA256: "ab", ContentType: "text/html"}},
					Routing: AppReleaseRouting{SPAFallbacks: []AppReleaseSPAFallback{{Prefix: "/", File: "index.html"}}},
				}, options)
				return err
			},
			method: http.MethodPost, url: appReleasesURL, content: "application/json",
		},
		{
			name: "upload", body: `{"path":"assets/a b.js","size":3,"sha256":"cd"}`,
			call: func(c *AuthProjectAdminClient) error {
				uploaded, err := c.UploadAppReleaseFile(context.Background(), "project-1", "environment-1", "r1",
					"assets/a b.js", []byte("abc"), options)
				if err == nil && uploaded.SHA256 != "cd" {
					t.Fatalf("uploaded = %+v", uploaded)
				}
				return err
			},
			method: http.MethodPost, url: appReleasesURL + "/r1/files?path=assets%2Fa+b.js", content: "application/octet-stream",
		},
		{
			name: "verify", body: `{"releaseId":"r1","state":"verified"}`,
			call: func(c *AuthProjectAdminClient) error {
				_, err := c.VerifyAppRelease(context.Background(), "project-1", "environment-1", "r1", options)
				return err
			},
			method: http.MethodPost, url: appReleasesURL + "/r1/verify",
		},
		{
			name: "promote", body: `{"releaseId":"r1","host":"app.example","promotedAt":"2026-11-17T00:00:00Z"}`,
			call: func(c *AuthProjectAdminClient) error {
				promotion, err := c.PromoteAppRelease(context.Background(), "project-1", "environment-1", "r1", options)
				if err == nil && promotion.Host != "app.example" {
					t.Fatalf("promotion = %+v", promotion)
				}
				return err
			},
			method: http.MethodPost, url: appReleasesURL + "/r1/promote",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, test.body)
			client := newAdminTestClient(t, doer, "http://localhost:8080")
			if err := test.call(client.Admin.AuthProjects); err != nil {
				t.Fatalf("%s returned error: %v", test.name, err)
			}
			req := doer.requests[0]
			if req.Method != test.method || req.URL != test.url {
				t.Fatalf("request = %s %s, want %s %s", req.Method, req.URL, test.method, test.url)
			}
			if req.Header(authProjectOwningUserHeader) != authProjectOwningUser {
				t.Fatalf("owning user header = %q", req.Header(authProjectOwningUserHeader))
			}
			if test.content != "" && req.Header("Content-Type") != test.content {
				t.Fatalf("content type = %q, want %q", req.Header("Content-Type"), test.content)
			}
		})
	}
}

func TestAppReleaseUploadSendsTheRawBytes(t *testing.T) {
	doer := newCaptureDoer(http.StatusOK, `{"path":"index.html","size":4,"sha256":"ab"}`)
	client := newAdminTestClient(t, doer, "http://localhost:8080")
	if _, err := client.Admin.AuthProjects.UploadAppReleaseFile(context.Background(), "project-1", "environment-1",
		"r1", "index.html", []byte{0x00, 0x01, 'h', 'i'}, AuthProjectRequestOptions{}); err != nil {
		t.Fatalf("upload: %v", err)
	}
	if got := doer.requests[0].Body; string(got) != "\x00\x01hi" {
		t.Fatalf("body = %q, want the file bytes unchanged", got)
	}
}

func TestAppReleaseCreateRequestUsesTheManifestJSONNames(t *testing.T) {
	encoded, err := json.Marshal(AppReleaseCreateRequest{
		Files: []AppReleaseFile{{Path: "index.html", Size: 4, SHA256: "ab", ContentType: "text/html"}},
		Routing: AppReleaseRouting{
			Headers: []AppReleaseHeaderRule{{Route: "/", Framing: AppReleaseFramingDeny}},
			Cache:   []AppReleaseCacheRule{{Match: "assets/*", Mode: AppReleaseCacheImmutable}},
		},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := `{"files":[{"path":"index.html","size":4,"sha256":"ab","contentType":"text/html"}],` +
		`"routing":{"headers":[{"route":"/","framing":"deny"}],"cache":[{"match":"assets/*","mode":"immutable"}]}}`
	if string(encoded) != want {
		t.Fatalf("body = %s\nwant  %s", encoded, want)
	}
}
