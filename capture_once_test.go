package custd

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// failingDoer returns a transport-level failure so the SDK builds its error
// path before any decode happens. The SDK must never embed the request body in
// the wrapped error message; capture-once values must stay out of logs.
type failingDoer struct{}

func (failingDoer) Do(req *HTTPRequest) (*HTTPResponse, error) {
	return nil, errors.New("transport dial failure")
}

func TestAdminErrorsDoNotSurfaceCaptureOnceValues(t *testing.T) {
	plainExternal := "external-secret-DO-NOT-LOG"
	client := newAdminTestClient(t, &captureDoer{}, "http://localhost:8080/")
	client.config.HTTPClient = failingDoer{}

	_, err := client.Admin.Privacy.MapIdentifier(context.Background(), "acme", PrivacyIdentifierMapRequest{
		ExternalID: plainExternal,
	})
	if err == nil {
		t.Fatalf("expected transport error")
	}
	if strings.Contains(err.Error(), plainExternal) {
		t.Fatalf("transport error leaked ExternalID: %v", err)
	}
}
