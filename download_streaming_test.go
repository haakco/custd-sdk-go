package custd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// streamingTransport answers with a body whose second half is only written once the caller's
// writer has received the first, so a client that buffers the whole body before writing never
// finishes.
type streamingTransport struct {
	first, second string
	delivered     chan struct{}
	closed        atomic.Bool
}

func (s *streamingTransport) Do(req *http.Request) (*http.Response, error) {
	reader, writer := io.Pipe()
	go func() {
		_, _ = io.WriteString(writer, s.first)
		select {
		case <-s.delivered:
			_, _ = io.WriteString(writer, s.second)
			_ = writer.Close()
		case <-time.After(2 * time.Second):
			_ = writer.CloseWithError(io.ErrUnexpectedEOF)
		}
	}()
	sum := sha256.Sum256([]byte(s.first + s.second))
	header := http.Header{}
	header.Set("Content-Length", strconv.Itoa(len(s.first)+len(s.second)))
	header.Set("X-Checksum-SHA256", hex.EncodeToString(sum[:]))
	return &http.Response{StatusCode: http.StatusOK, Header: header, Body: closeRecorder{reader, &s.closed}, Request: req}, nil
}

type closeRecorder struct {
	io.Reader
	closed *atomic.Bool
}

func (c closeRecorder) Close() error {
	c.closed.Store(true)
	return nil
}

// signalWriter reports the first write, which is the proof the body reached it as a stream.
type signalWriter struct {
	written   []byte
	delivered chan struct{}
}

func (w *signalWriter) Write(p []byte) (int, error) {
	if len(w.written) == 0 {
		close(w.delivered)
	}
	w.written = append(w.written, p...)
	return len(p), nil
}

func TestDownloadsStreamThroughACustomHTTPClient(t *testing.T) {
	downloads := map[string]func(*CustdClient, io.Writer) error{
		"offboarding bundle": func(c *CustdClient, w io.Writer) error {
			_, err := c.Admin.Offboarding.Download(context.Background(), "request-1", w)
			return err
		},
		"audit export": func(c *CustdClient, w io.Writer) error {
			_, err := c.Admin.Audit.ExportEvents(context.Background(), AuditListOptions{}, AuditExportFormatCSV, w)
			return err
		},
	}
	for name, download := range downloads {
		t.Run(name, func(t *testing.T) {
			delivered := make(chan struct{})
			transport := &streamingTransport{first: "first half,", second: "second half", delivered: delivered}
			client := NewClient(&ClientConfig{BaseURL: "http://localhost:8080", APIKey: "token", HTTPClient: transport, FlushInterval: time.Hour})
			t.Cleanup(func() { _ = client.Close(context.Background()) })
			writer := &signalWriter{delivered: delivered}
			if err := download(client, writer); err != nil {
				t.Fatalf("download: %v", err)
			}
			if string(writer.written) != "first half,second half" || !transport.closed.Load() {
				t.Fatalf("written %q, body closed %v", writer.written, transport.closed.Load())
			}
		})
	}
}

// The built-in download path has no whole-body deadline: a body that keeps arriving outlasts the
// JSON client's timeout, and only a body that stops arriving fails. Both bounds are scaled down.
func TestBuiltInDownloadsFailOnlyWhenTheyStall(t *testing.T) {
	cases := map[string]struct {
		pause   time.Duration
		wantErr bool
	}{
		"slow but steady": {pause: 20 * time.Millisecond},
		"stalled":         {pause: time.Second, wantErr: true},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/csv")
				for range 10 {
					_, _ = io.WriteString(w, "row\n")
					w.(http.Flusher).Flush()
					select {
					case <-time.After(tc.pause):
					case <-r.Context().Done():
						return
					}
				}
			}))
			t.Cleanup(server.Close)
			client := NewClient(&ClientConfig{BaseURL: server.URL, APIKey: "token", FlushInterval: time.Hour})
			t.Cleanup(func() { _ = client.Close(context.Background()) })
			client.httpClient = &http.Client{Timeout: 50 * time.Millisecond}
			client.streamStall = 100 * time.Millisecond
			var written strings.Builder
			_, err := client.Admin.Audit.ExportEvents(context.Background(), AuditListOptions{}, AuditExportFormatCSV, &written)
			if tc.wantErr {
				if err == nil || !strings.Contains(err.Error(), "download stalled") {
					t.Fatalf("stalled download error = %v", err)
				}
				return
			}
			if err != nil || written.Len() != 40 {
				t.Fatalf("steady download wrote %d bytes, error %v", written.Len(), err)
			}
		})
	}
}
