package custd

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestErrorCaptureSanitizesAndValidatesRelay(t *testing.T) {
	var body string
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.URL.Path != "/api/v1/errors/api/1/envelope/" {
			t.Error(r.URL.Path)
		}
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		body = string(data)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()
	client, err := NewErrorCaptureClient(ErrorCaptureConfig{BaseURL: server.URL, WriteKey: "site_pk_" + strings.Repeat("a", 48), Environment: "test", Release: "vorrent@1.0.0"})
	if err != nil {
		t.Fatal(err)
	}
	result := client.CaptureException(context.Background(), CaptureExceptionRequest{Error: errors.New("password=private"), Mode: ErrorCaptureHandled})
	if result.Outcome != ErrorCaptureAccepted {
		t.Fatalf("%+v", result)
	}
	for _, forbidden := range []string{"password", "private", "/home/", `"value"`, `"request"`} {
		if strings.Contains(body, forbidden) {
			t.Errorf("private field %q in envelope", forbidden)
		}
	}
	if !strings.Contains(body, "error_capture_test.go") || !strings.Contains(body, `"platform":"go"`) {
		t.Error("missing origin stack/platform")
	}
	fixture, err := os.ReadFile("../contract-fixtures/error-capture-relay.json")
	if err != nil {
		t.Fatal(err)
	}
	var event ErrorCaptureEvent
	if err := json.Unmarshal(fixture, &event); err != nil {
		t.Fatal(err)
	}
	if result = client.CaptureEvent(context.Background(), event); result.Outcome != ErrorCaptureAccepted {
		t.Fatalf("%+v", result)
	}
	event.Frames[0].Filename = "/private/user/screen.js"
	if result = client.CaptureEvent(context.Background(), event); result.Outcome != ErrorCaptureRejected {
		t.Fatalf("unsafe relay: %+v", result)
	}
	if requests.Load() != 2 {
		t.Fatal("unsafe relay reached collector")
	}
}

func TestErrorCaptureBoundsAndRedirects(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Location", "http://example.invalid/private")
		w.WriteHeader(http.StatusFound)
	}))
	defer server.Close()
	transport := &http.Client{}
	client, err := NewErrorCaptureClient(ErrorCaptureConfig{BaseURL: server.URL, WriteKey: "site_pk_" + strings.Repeat("b", 48), Environment: "test", HTTPClient: transport})
	if err != nil {
		t.Fatal(err)
	}
	for range 20 {
		if got := client.CaptureException(context.Background(), CaptureExceptionRequest{Error: errors.New("secret")}); got.Outcome != ErrorCaptureRejected || got.Status != 302 {
			t.Fatalf("%+v", got)
		}
	}
	if got := client.CaptureException(context.Background(), CaptureExceptionRequest{}); got.Outcome != ErrorCaptureLimited {
		t.Fatal(got)
	}
	if requests.Load() != 20 || transport.CheckRedirect != nil || transport.Timeout != 0 {
		t.Fatal("mutated caller transport or retried")
	}
	if _, err := NewErrorCaptureClient(ErrorCaptureConfig{BaseURL: "http://remote.invalid", WriteKey: "secret", Environment: "test"}); err == nil {
		t.Fatal("invalid config accepted")
	}
}

func TestErrorReporterBoundedOriginAndClose(t *testing.T) {
	arrived := make(chan struct{}, 4)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { arrived <- struct{}{}; <-release; w.WriteHeader(202) }))
	defer server.Close()
	reporter, err := NewErrorReporter(ErrorCaptureConfig{BaseURL: server.URL, WriteKey: "site_pk_" + strings.Repeat("c", 48), Environment: "test"})
	if err != nil {
		t.Fatal(err)
	}
	for range 4 {
		if !reporter.Report(context.Background(), CaptureExceptionRequest{Error: errors.New("secret")}) {
			t.Fatal("not admitted")
		}
	}
	for range 4 {
		<-arrived
	}
	if reporter.Report(context.Background(), CaptureExceptionRequest{}) {
		t.Fatal("unbounded admission")
	}
	event := ErrorCaptureEvent{Platform: ErrorPlatformJavaScript, ExceptionType: "Error", Frames: []ErrorCaptureFrame{}, Mode: ErrorCaptureHandled, Environment: "test"}
	if result := reporter.CaptureEvent(context.Background(), event); result.Outcome != ErrorCaptureLimited {
		t.Fatal(result)
	}
	stats := reporter.Stats()
	if stats.Inflight != 4 || stats.Limited != 2 {
		t.Fatal(stats)
	}
	close(release)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := reporter.Close(ctx); err != nil {
		t.Fatal(err)
	}
	stats = reporter.Stats()
	if stats.Accepted != 4 || stats.Inflight != 0 {
		t.Fatal(stats)
	}
	if reporter.Report(context.Background(), CaptureExceptionRequest{}) {
		t.Fatal("closed reporter admitted")
	}
}

func TestErrorReporterSurvivesTwentyActionsAndCancelsClose(t *testing.T) {
	var requests atomic.Int32
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) > 20 {
			select {
			case <-r.Context().Done():
			case <-release:
			}
			return
		}
		w.WriteHeader(202)
	}))
	defer server.Close()
	defer close(release)
	reporter, err := NewErrorReporter(ErrorCaptureConfig{BaseURL: server.URL, WriteKey: "site_pk_" + strings.Repeat("d", 48), Environment: "test"})
	if err != nil {
		t.Fatal(err)
	}
	for i := uint64(1); i <= 20; i++ {
		if !reporter.Report(context.Background(), CaptureExceptionRequest{}) {
			t.Fatal("not admitted")
		}
		deadline := time.Now().Add(time.Second)
		for reporter.Stats().Accepted < i {
			if time.Now().After(deadline) {
				t.Fatal("delivery did not finish")
			}
			time.Sleep(time.Millisecond)
		}
	}
	if !reporter.Report(context.Background(), CaptureExceptionRequest{}) {
		t.Fatal("per-process lifetime cutoff")
	}
	deadline := time.Now().Add(time.Second)
	for requests.Load() < 21 {
		if time.Now().After(deadline) {
			t.Fatal("send did not start")
		}
		time.Sleep(time.Millisecond)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := reporter.Close(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("close: %v", err)
	}
	deadline = time.Now().Add(time.Second)
	for reporter.Stats().Inflight != 0 {
		if time.Now().After(deadline) {
			t.Fatal("send not canceled")
		}
		time.Sleep(time.Millisecond)
	}
	if reporter.Stats().Unavailable != 1 {
		t.Fatal(reporter.Stats())
	}
}

func TestErrorReporterRelayReturnsConfirmedOutcome(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusUnauthorized) }))
	defer server.Close()
	reporter, err := NewErrorReporter(ErrorCaptureConfig{BaseURL: server.URL, WriteKey: "site_pk_" + strings.Repeat("e", 48), Environment: "test"})
	if err != nil {
		t.Fatal(err)
	}
	event := ErrorCaptureEvent{Platform: ErrorPlatformJavaScript, ExceptionType: "Error", Frames: []ErrorCaptureFrame{}, Mode: ErrorCaptureHandled, Environment: "test"}
	result := reporter.CaptureEvent(context.Background(), event)
	if result.Outcome != ErrorCaptureRejected || result.Status != http.StatusUnauthorized || reporter.Stats().Rejected != 1 {
		t.Fatal(result, reporter.Stats())
	}
	if err := reporter.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
}
