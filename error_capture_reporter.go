package custd

import (
	"context"
	"sync"
)

// ErrorReporterStats reports delivery outcomes without exposing errors, Site keys or event content.
type ErrorReporterStats struct {
	Accepted    uint64 `json:"accepted"`
	Rejected    uint64 `json:"rejected"`
	Unavailable uint64 `json:"unavailable"`
	Limited     uint64 `json:"limited"`
	Inflight    int    `json:"inflight"`
	Closed      bool   `json:"closed"`
}

// ErrorReporter owns four in-flight sends and no pending queue. Create one per process/Site, and Close at shutdown.
type ErrorReporter struct {
	config ErrorCaptureConfig
	mu     sync.Mutex
	stats  ErrorReporterStats
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewErrorReporter(config ErrorCaptureConfig) (*ErrorReporter, error) {
	if _, err := NewErrorCaptureClient(config); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &ErrorReporter{config: config, ctx: ctx, cancel: cancel}, nil
}

// Report snapshots the calling stack immediately. True means admitted, not delivered; inspect Stats for delivery.
func (r *ErrorReporter) Report(ctx context.Context, request CaptureExceptionRequest) bool {
	event := SnapshotError(request)
	event.Environment = r.config.Environment
	event.Release = r.config.Release
	return r.ReportEvent(ctx, event)
}

// ReportEvent copies and revalidates relay events. It never waits for network I/O.
func (r *ErrorReporter) ReportEvent(ctx context.Context, event ErrorCaptureEvent) bool {
	if outcome := r.admit(event); outcome != "" {
		return false
	}
	event.Frames = append([]ErrorCaptureFrame(nil), event.Frames...)
	// Request cancellation does not lose an admitted asynchronous report.
	go r.deliver(context.WithoutCancel(ctx), event)
	return true
}

// CaptureEvent returns confirmed collector acceptance for an optional relay request.
// It shares the four-send ceiling with Report and returns limited immediately when busy.
func (r *ErrorReporter) CaptureEvent(ctx context.Context, event ErrorCaptureEvent) ErrorCaptureResult {
	if outcome := r.admit(event); outcome != "" {
		return ErrorCaptureResult{Outcome: outcome}
	}
	return r.deliver(ctx, event)
}

func (r *ErrorReporter) admit(event ErrorCaptureEvent) ErrorCaptureOutcome {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.stats.Closed || r.stats.Inflight >= 4 {
		r.stats.Limited++
		return ErrorCaptureLimited
	}
	if !validErrorCaptureEvent(event) {
		r.stats.Rejected++
		return ErrorCaptureRejected
	}
	r.stats.Inflight++
	r.wg.Add(1)
	return ""
}

func (r *ErrorReporter) deliver(ctx context.Context, event ErrorCaptureEvent) ErrorCaptureResult {
	defer r.wg.Done()
	sendCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	stop := context.AfterFunc(r.ctx, cancel)
	defer stop()
	client, err := NewErrorCaptureClient(r.config)
	result := ErrorCaptureResult{Outcome: ErrorCaptureUnavailable}
	if err == nil {
		result = client.CaptureEvent(sendCtx, event)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.Inflight--
	switch result.Outcome {
	case ErrorCaptureAccepted:
		r.stats.Accepted++
	case ErrorCaptureRejected:
		r.stats.Rejected++
	case ErrorCaptureLimited:
		r.stats.Limited++
	default:
		r.stats.Unavailable++
	}
	return result
}

func (r *ErrorReporter) Stats() ErrorReporterStats {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.stats
}

// Close refuses new work, waits for admitted sends, and cancels them if the caller's deadline expires.
func (r *ErrorReporter) Close(ctx context.Context) error {
	r.mu.Lock()
	r.stats.Closed = true
	r.mu.Unlock()
	done := make(chan struct{})
	go func() { r.wg.Wait(); close(done) }()
	select {
	case <-done:
		r.cancel()
		return nil
	case <-ctx.Done():
		r.cancel()
		return ctx.Err()
	}
}
