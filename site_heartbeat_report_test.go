package custd

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSubmitSiteHeartbeat(t *testing.T) {
	report := SiteHeartbeatReport{
		ScheduledFor: time.Date(2026, 11, 20, 6, 0, 0, 0, time.UTC),
		Outcome:      SiteHeartbeatFailed,
	}
	token := strings.Repeat("a", 64)
	path := heartbeatReportPrefix + "test/monitor"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != path || r.Header.Get("Authorization") != "Bearer "+token {
			t.Error("unexpected heartbeat request")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var submitted SiteHeartbeatReport
		if err := json.NewDecoder(r.Body).Decode(&submitted); err != nil || submitted != report {
			t.Errorf("report = %+v, %v", submitted, err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"report":{"scheduledFor":"2026-11-20T06:00:00Z","outcome":"failed"},"receivedAt":"2026-11-20T06:01:00Z"}`)
	}))
	defer server.Close()

	destination := SiteHeartbeatDestination{BaseURL: server.URL, ReportPath: path, WriteToken: token}
	receipt, err := SubmitSiteHeartbeat(context.Background(), destination, report)
	if err != nil || receipt.ReceivedAt != report.ScheduledFor.Add(time.Minute) {
		t.Fatalf("receipt = %+v, %v", receipt, err)
	}
	destination.ReportPath = privateUptimeReportPrefix + "test/monitor"
	if _, err := SubmitSiteHeartbeat(context.Background(), destination, report); err == nil {
		t.Fatal("heartbeat credential accepted for a private uptime path")
	}
}
