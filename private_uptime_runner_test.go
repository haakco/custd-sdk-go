package custd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPrivateUptimeRunnerRetriesAndPolls(t *testing.T) {
	requestedAt := time.Now().UTC()
	reported := make(chan PrivateUptimeReport, 3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(PrivateUptimeWork{RequestedAt: &requestedAt})
			return
		}
		var report PrivateUptimeReport
		if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		failing := len(reported) == 0
		reported <- report
		if failing {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)
	observations := 0
	runner := PrivateUptimeRunner{
		Destination: PrivateUptimeDestination{
			BaseURL: server.URL, ReportPath: "/api/v1/site-private-uptime-reports/acme/monitor",
			WriteToken: strings.Repeat("a", 64),
		},
		Interval: time.Minute,
		Observe: func(context.Context) (PrivateUptimeReport, error) {
			observations++
			id := "scheduled"
			if observations > 1 {
				id = "on-demand"
			}
			return PrivateUptimeReport{ReportUUID: id, ObservedAt: time.Now().UTC(), StatusCode: 200}, nil
		},
	}
	ctx := context.Background()
	var state privateUptimeRunnerState
	if err := runner.step(ctx, &state); err == nil {
		t.Fatal("upload failure must reach the runner's error handler")
	}
	if err := runner.step(ctx, &state); err != nil {
		t.Fatal(err)
	}
	if err := runner.step(ctx, &state); err != nil {
		t.Fatal(err)
	}
	if count := len(reported); count != 3 || observations != 2 {
		t.Fatalf("%d uploads, %d observations; want 3 uploads and 2 observations", count, observations)
	}
	first, retry, requested := <-reported, <-reported, <-reported
	if first != retry || requested.ReportUUID != "on-demand" {
		t.Fatal("failed upload was not replayed exactly before reporting the requested check")
	}
}
