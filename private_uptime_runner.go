package custd

import (
	"context"
	"fmt"
	"time"
)

const privateUptimePollInterval = 15 * time.Second

// PrivateUptimeRunner runs one explicitly configured probe. Observe makes the local
// check. OnError reports a failed poll, observation or upload while the runner retries.
type PrivateUptimeRunner struct {
	Destination PrivateUptimeDestination
	Interval    time.Duration
	Observe     func(context.Context) (PrivateUptimeReport, error)
	OnError     func(error)
}

// Run starts with a check and continues until the context is canceled. Run one
// instance per monitor. Check intervals must be at least 15 seconds.
func (r PrivateUptimeRunner) Run(ctx context.Context) error {
	if r.Observe == nil || r.OnError == nil || r.Interval < privateUptimePollInterval {
		return fmt.Errorf("custd: private uptime runner requires an observer, an error handler and at least a 15-second interval")
	}
	if _, err := siteMonitorDestination(r.Destination).endpoint(privateUptimeReportPrefix); err != nil {
		return err
	}
	return r.run(ctx, privateUptimePollInterval)
}

func (r PrivateUptimeRunner) run(ctx context.Context, pollInterval time.Duration) error {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	var state privateUptimeRunnerState
	for ctx.Err() == nil {
		if err := r.step(ctx, &state); err != nil && ctx.Err() == nil {
			r.OnError(err)
		}
		select {
		case <-ctx.Done():
		case <-ticker.C:
		}
	}
	return ctx.Err()
}

type privateUptimeRunnerState struct {
	nextCheck time.Time
	retry     *PrivateUptimeReport
}

func (r PrivateUptimeRunner) step(ctx context.Context, state *privateUptimeRunnerState) error {
	if state.retry != nil {
		report := *state.retry
		state.retry = nil
		return SubmitPrivateUptime(ctx, r.Destination, report)
	}
	scheduled := !time.Now().Before(state.nextCheck)
	if !scheduled {
		work, err := GetPrivateUptimeWork(ctx, r.Destination)
		if err != nil {
			return err
		}
		if work.RequestedAt == nil {
			return nil
		}
	}
	report, err := r.Observe(ctx)
	if err != nil {
		return fmt.Errorf("custd: observe private uptime: %w", err)
	}
	if scheduled {
		state.nextCheck = time.Now().Add(r.Interval)
	}
	if err := SubmitPrivateUptime(ctx, r.Destination, report); err != nil {
		state.retry = &report
		return err
	}
	return nil
}
