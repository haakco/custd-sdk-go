package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestPrivateUptimeWorkAndSubmit(t *testing.T) {
	token := strings.Repeat("a", 64)
	path := "/api/v1/site-private-uptime-reports/acme/monitor"
	report := PrivateUptimeReport{ReportUUID: "fixture", ObservedAt: time.Now().UTC(), StatusCode: 200, ResponseMs: 12}
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path || r.Header.Get("Authorization") != "Bearer "+token {
			t.Error("report did not use the bound path and credential")
		}
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(PrivateUptimeWork{RequestedAt: &report.ObservedAt})
			return
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		var got PrivateUptimeReport
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil || got != report {
			t.Errorf("unexpected report: %+v, error %v", got, err)
		}
		if requests.Add(1) > 1 {
			http.Redirect(w, r, "/other", http.StatusTemporaryRedirect)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)
	destination := PrivateUptimeDestination{BaseURL: server.URL, ReportPath: path, WriteToken: token}
	work, err := GetPrivateUptimeWork(context.Background(), destination)
	if err != nil || work.RequestedAt == nil || !work.RequestedAt.Equal(report.ObservedAt) {
		t.Fatalf("unexpected work: %+v, error %v", work, err)
	}
	if err := SubmitPrivateUptime(context.Background(), destination, report); err != nil {
		t.Fatal(err)
	}
	if err := SubmitPrivateUptime(context.Background(), destination, report); err == nil {
		t.Fatal("redirect was accepted as a receipt")
	}
	if got := requests.Load(); got != 2 {
		t.Fatalf("made %d requests, want two", got)
	}
}
