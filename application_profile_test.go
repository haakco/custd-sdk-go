package custd

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestCurrentAppProfileUsesOnlyApplicationGrant(t *testing.T) {
	photoBytes := []byte{1, 2}
	var methodsMu sync.Mutex
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodsMu.Lock()
		methods = append(methods, r.Method+" "+r.URL.Path)
		methodsMu.Unlock()
		if r.Header.Get("Authorization") != "Bearer application-token" || r.Header.Get("Cookie") != "" ||
			r.Header.Get(authProjectOwningUserHeader) != "" {
			t.Error("profile must use the current application bearer only")
		}
		if r.Method == http.MethodGet && r.URL.Path == currentAppPhotoPath {
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(photoBytes)
			return
		}
		if r.URL.Path == currentAppPhotoPath && (r.Method == http.MethodPut || r.Method == http.MethodPost) {
			body, err := io.ReadAll(r.Body)
			if err != nil || !bytes.Equal(body, photoBytes) || r.Header.Get("Content-Type") != "image/png" {
				t.Error("upload must send the raw photo bytes")
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == currentAppPhotoPath && r.Method == http.MethodPost {
			_, _ = w.Write([]byte(`{"fields":[{"fieldKey":"picture_url","present":false,"initialized":true,"required":false,"editableBy":"user","editable":true}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"fields":[{"fieldKey":"display_name","value":"Canvas Person","present":true,"initialized":true,"required":false,"editableBy":"user","editable":true}],"managedPhoto":true}`))
	}))
	defer server.Close()
	client := NewClient(&ClientConfig{
		BaseURL: server.URL, APIKey: "platform-token-must-not-leak",
		ClientID: "machine-client", ClientSecret: "machine-secret",
		TokenURL: server.URL + "/oauth/token", FlushInterval: time.Hour,
	})
	t.Cleanup(func() { _ = client.Close(t.Context()) })
	ctx, token := t.Context(), "application-token"

	profile, err := client.GetCurrentAppProfile(ctx, token)
	if err != nil || !profile.ManagedPhoto || profile.Fields[0].Value != "Canvas Person" {
		t.Fatalf("read profile: %#v, %v", profile, err)
	}
	update := ApplicationProfileValuesRequest{Values: []ApplicationProfileValue{{FieldKey: "display_name", Value: "Canvas Person"}}}
	if _, err := client.UpdateCurrentAppProfile(ctx, token, update); err != nil {
		t.Fatal(err)
	}
	oldValues := ApplicationProfileValuesRequest{Values: []ApplicationProfileValue{{FieldKey: "display_name", Value: "Imported name"}}}
	initialized, err := client.InitializeCurrentAppProfile(ctx, token, oldValues)
	if err != nil || !initialized.Fields[0].Initialized || initialized.Fields[0].Value != "Canvas Person" {
		t.Fatalf("initialize profile must return current winning value: %#v, %v", initialized, err)
	}
	photo, err := client.GetCurrentAppProfilePhoto(ctx, token)
	if err != nil || photo.ContentType != "image/png" || !bytes.Equal(photo.Bytes, photoBytes) {
		t.Fatalf("read photo: %#v, %v", photo, err)
	}
	if _, err := client.UploadCurrentAppProfilePhoto(ctx, token, *photo); err != nil {
		t.Fatal(err)
	}
	withPhoto, err := client.InitializeCurrentAppProfilePhoto(ctx, token, *photo)
	if err != nil || !withPhoto.Fields[0].Initialized || withPhoto.Fields[0].Present || withPhoto.ManagedPhoto {
		t.Fatalf("initialize photo must preserve a clear: %#v, %v", withPhoto, err)
	}
	if _, err := client.RemoveCurrentAppProfilePhoto(ctx, token); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"GET " + currentAppProfilePath, "PATCH " + currentAppProfilePath, "POST " + currentAppProfilePath, "GET " + currentAppPhotoPath,
		"PUT " + currentAppPhotoPath, "POST " + currentAppPhotoPath, "DELETE " + currentAppPhotoPath,
	}
	methodsMu.Lock()
	defer methodsMu.Unlock()
	if !reflect.DeepEqual(methods, want) {
		t.Fatalf("profile requests = %v; want %v", methods, want)
	}
}

func TestCurrentAppProfileRefusesBearerRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != currentAppProfilePath {
			t.Error("unexpected redirected profile request")
		}
		http.Redirect(w, r, "/untrusted", http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	client := NewClient(&ClientConfig{BaseURL: server.URL, FlushInterval: time.Hour})
	t.Cleanup(func() { _ = client.Close(t.Context()) })
	_, err := client.GetCurrentAppProfile(t.Context(), "application-token")
	var response *RequestError
	if !errors.As(err, &response) || response.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected redirect refusal, got %v", err)
	}
}
