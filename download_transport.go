package custd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

// streamStallTimeout bounds a streamed download's wait for its response headers and for each
// read. The whole body has no deadline, so a large archive is not cut off while it still arrives.
const streamStallTimeout = 30 * time.Second

// buildStreamTransport is the caller's HTTPClient when one is configured, else a client like the
// built one without its whole-request timeout: connecting keeps the default dial and TLS bounds
// and the response headers must arrive within streamStallTimeout.
func buildStreamTransport(cfg *ClientConfig) HTTPDoer {
	if cfg.HTTPClient != nil {
		return cfg.HTTPClient
	}
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return buildHTTPClient(cfg)
	}
	transport := base.Clone()
	transport.ResponseHeaderTimeout = streamStallTimeout
	return &http.Client{Transport: buildRoundTripper(cfg, transport)}
}

// sendStream issues a download and returns its unread answer. The body fails once no bytes arrive
// for the client's stall bound; closing it releases the request.
func (c *CustdClient) sendStream(ctx context.Context, request *bufferedRequest) (*http.Response, error) {
	ctx, cancel := context.WithCancel(ctx)
	response, err := send(ctx, c.streamClient, request)
	if err != nil {
		cancel()
		return nil, err
	}
	response.Body = newStallGuard(response.Body, c.streamStall, cancel)
	return response, nil
}

// stallGuard gives each Read of a streamed body the stall bound to return, and cancels the request
// when one does not. Time the caller spends between reads is not counted.
type stallGuard struct {
	body    io.ReadCloser
	stall   time.Duration
	timer   *time.Timer
	stalled atomic.Bool
	cancel  context.CancelFunc
}

func newStallGuard(body io.ReadCloser, stall time.Duration, cancel context.CancelFunc) *stallGuard {
	guard := &stallGuard{body: body, stall: stall, cancel: cancel}
	guard.timer = time.AfterFunc(stall, func() {
		guard.stalled.Store(true)
		cancel()
	})
	guard.timer.Stop()
	return guard
}

func (g *stallGuard) Read(p []byte) (int, error) {
	g.timer.Reset(g.stall)
	n, err := g.body.Read(p)
	g.timer.Stop()
	if err != nil && err != io.EOF && g.stalled.Load() {
		return n, fmt.Errorf("custd: download stalled: no bytes for %s: %w", g.stall, err)
	}
	return n, err
}

func (g *stallGuard) Close() error {
	g.timer.Stop()
	err := g.body.Close()
	g.cancel()
	return err
}
