package custd

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const appArchiveTestUUID = "0195f0a0-0000-4000-8000-000000000001"

func TestAppMessageArchiveMedia(t *testing.T) {
	image := []byte{0xff, 0xd8, 0xff, 0xd9}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tenant-human" {
			t.Error("archive media request did not use the tenant human credential")
		}
		switch r.Method {
		case http.MethodPut:
			if r.URL.Path != appMessagePath+"/"+appArchiveTestUUID+"/media" ||
				r.Header.Get("Content-Type") != "application/octet-stream" {
				t.Error("archive image upload did not send octet-stream to its message")
			}
			read := new(bytes.Buffer)
			if _, err := read.ReadFrom(r.Body); err != nil || !bytes.Equal(read.Bytes(), image) {
				t.Error("archive image upload changed the original bytes")
			}
			w.Header().Set("Content-Type", "application/json")
			if _, err := fmt.Fprintf(w, `{"mediaUuid":%q,"contentType":"image/jpeg","byteCount":4,"ocrStatus":"pending"}`, appArchiveTestUUID); err != nil {
				t.Error(err)
			}
		case http.MethodGet:
			if r.URL.Path != appMessagePath+"/"+appArchiveTestUUID+"/media/"+appArchiveTestUUID {
				t.Error("archive image download used the wrong path")
			}
			w.Header().Set("Content-Type", "image/jpeg")
			if _, err := w.Write(image); err != nil {
				t.Error(err)
			}
		default:
			t.Errorf("unexpected HTTP method %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	t.Cleanup(server.Close)
	client := NewClient(&ClientConfig{BaseURL: server.URL, APIKey: "tenant-human", HTTPClient: server.Client()})
	uploaded, err := client.AppMessages.UploadMedia(context.Background(), &AppMessageMediaUpload{
		MessageUUID: appArchiveTestUUID, Image: image,
	})
	if err != nil || uploaded.ByteCount != 4 {
		t.Fatalf("upload metadata = %+v, error = %v", uploaded, err)
	}
	original, err := client.AppMessages.DownloadMedia(context.Background(), AppMessageMediaRef{
		MessageUUID: appArchiveTestUUID, MediaUUID: uploaded.MediaUUID,
	})
	if err != nil || original.ContentType != "image/jpeg" || !bytes.Equal(original.Image, image) {
		t.Fatalf("download original: %v", err)
	}
}
