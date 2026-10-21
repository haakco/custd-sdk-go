package custd

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestAuthProjectBrandLogoRoutesUseTypedPaths(t *testing.T) {
	const logoPath = "http://localhost:8080/api/v1/admin/auth-projects/project-1/environments/environment-1/brand/logo"
	tests := []struct {
		name         string
		responseBody string
		call         func(t *testing.T, client *CustdClient) (ResponseMetadata, error)
		wantMethod   string
		wantURL      string
		wantBody     bool
		check        func(t *testing.T, req *HTTPRequest)
	}{
		{
			name:         "upload",
			responseBody: `{"contentType":"image/png","updatedAt":"2026-08-12T12:00:00Z","byteSize":4,"width":2,"height":2}`,
			wantMethod:   http.MethodPost,
			wantURL:      logoPath,
			wantBody:     true,
			call: func(t *testing.T, client *CustdClient) (ResponseMetadata, error) {
				t.Helper()
				logo, metadata, err := client.Admin.AuthProjects.SetAuthProjectBrandLogo(
					context.Background(), "project-1", "environment-1", "logo.png",
					[]byte{0x89, 'P', 'N', 'G'},
					AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
				)
				if err != nil {
					t.Fatalf("SetAuthProjectBrandLogo returned error: %v", err)
				}
				if logo.ContentType != "image/png" || logo.ByteSize != 4 || logo.Width != 2 || logo.Height != 2 {
					t.Fatalf("logo = %+v", logo)
				}
				return metadata, nil
			},
			check: func(t *testing.T, req *HTTPRequest) {
				t.Helper()
				if req.Headers[authProjectOwningUserHeader] != authProjectOwningUser {
					t.Fatalf("owning user header = %q", req.Headers[authProjectOwningUserHeader])
				}
				contentType := req.Headers["Content-Type"]
				if !strings.HasPrefix(contentType, "multipart/form-data; boundary=") {
					t.Fatalf("content type = %q", contentType)
				}
				if !bytes.Contains(req.Body, []byte("logo.png")) {
					t.Fatalf("upload body does not carry the file name: %q", req.Body)
				}
				if !bytes.Contains(req.Body, []byte{0x89, 'P', 'N', 'G'}) {
					t.Fatalf("upload body does not carry the bytes: %q", req.Body)
				}
			},
		},
		{
			name:         "clear",
			responseBody: `{}`,
			wantMethod:   http.MethodDelete,
			wantURL:      logoPath,
			wantBody:     false,
			call: func(t *testing.T, client *CustdClient) (ResponseMetadata, error) {
				t.Helper()
				return client.Admin.AuthProjects.ClearAuthProjectBrandLogo(
					context.Background(), "project-1", "environment-1",
					AuthProjectRequestOptions{OwningUserUUID: authProjectOwningUser},
				)
			},
			check: func(t *testing.T, req *HTTPRequest) {
				t.Helper()
				if req.Headers[authProjectOwningUserHeader] != authProjectOwningUser {
					t.Fatalf("owning user header = %q", req.Headers[authProjectOwningUserHeader])
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			doer := newCaptureDoer(http.StatusOK, test.responseBody)
			doer.headers = map[string]string{requestIDHeader: "req-1"}
			client := newAdminTestClient(t, doer, "http://localhost:8080")

			metadata, err := test.call(t, client)
			if err != nil {
				t.Fatalf("%s returned error: %v", test.name, err)
			}
			if metadata.RequestID != "req-1" {
				t.Fatalf("request id = %q", metadata.RequestID)
			}
			if len(doer.requests) != 1 {
				t.Fatalf("requests = %d, want 1", len(doer.requests))
			}
			req := doer.requests[0]
			if req.Method != test.wantMethod {
				t.Fatalf("method = %s, want %s", req.Method, test.wantMethod)
			}
			if req.URL != test.wantURL {
				t.Fatalf("url = %s, want %s", req.URL, test.wantURL)
			}
			if test.wantBody && len(req.Body) == 0 {
				t.Fatalf("body is empty")
			}
			if !test.wantBody && len(req.Body) != 0 {
				t.Fatalf("body = %q, want empty", req.Body)
			}
			test.check(t, req)
		})
	}
}
