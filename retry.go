package custd

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	mrand "math/rand"
	"sync"
	"time"
)

// retryableStatusSet returns a lookup set from a slice of status codes.
func retryableStatusSet(codes []int) map[int]bool {
	m := make(map[int]bool, len(codes))
	for _, c := range codes {
		m[c] = true
	}
	return m
}

// isRetryableStatus checks whether an HTTP status code is retryable.
func isRetryableStatus(code int, retrySet map[int]bool) bool {
	return retrySet[code]
}

// sendError represents an HTTP error with status code context. When the server
// returned an RFC 9457 problem+json body, Problem carries the parsed detail so
// callers can branch on the typed status/code instead of string-matching.
type sendError struct {
	StatusCode int
	Message    string
	Retryable  bool
	Problem    *Problem
}

func (e *sendError) Error() string {
	return e.Message
}

// newProblemError builds a send error from a parsed RFC 9457 problem, carrying
// the problem through so callers can branch on its status and code.
func newProblemError(statusCode int, retryable bool, problem *Problem) *sendError {
	return &sendError{
		StatusCode: statusCode,
		Message:    problem.Error(),
		Retryable:  retryable,
		Problem:    problem,
	}
}

// newRetryableError creates a retryable send error.
func newRetryableError(statusCode int) *sendError {
	return &sendError{
		StatusCode: statusCode,
		Message:    fmt.Sprintf("custd: retryable status %d", statusCode),
		Retryable:  true,
	}
}

// newNonRetryableError creates a non-retryable send error.
func newNonRetryableError(statusCode int) *sendError {
	return &sendError{
		StatusCode: statusCode,
		Message:    fmt.Sprintf("custd: request failed with status %d", statusCode),
		Retryable:  false,
	}
}

// withRetry executes op with exponential backoff, respecting context cancellation.
func withRetry(ctx context.Context, cfg RetryConfig, rng *mrand.Rand, rngMu *sync.Mutex, op func() error) error {
	maxAttempts := cfg.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}

	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = op()
		if err == nil {
			return nil
		}
		if !isRetryableErr(err) || attempt == maxAttempts {
			return err
		}
		if waitErr := sleepWithContext(ctx, cfg, attempt, rng, rngMu); waitErr != nil {
			return waitErr
		}
	}
	return err
}

// isRetryableErr checks if an error should trigger a retry.
func isRetryableErr(err error) bool {
	var se *sendError
	if errors.As(err, &se) {
		return se.Retryable
	}
	return false
}

// sleepWithContext pauses with backoff, returning early on context cancellation.
func sleepWithContext(ctx context.Context, cfg RetryConfig, attempt int, rng *mrand.Rand, rngMu *sync.Mutex) error {
	delay := backoffDelay(cfg, attempt, rng, rngMu)
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// newSecureRand creates a math/rand source seeded from crypto/rand.
func newSecureRand() *mrand.Rand {
	var seed int64
	// nolint:errcheck // jitter-only PRNG seed; a crypto/rand read error falls back to the zero seed
	_ = binary.Read(rand.Reader, binary.LittleEndian, &seed)
	return mrand.New(mrand.NewSource(seed)) //nolint:gosec // seeded from crypto/rand
}

// backoffDelay calculates exponential backoff with jitter.
// The rng and rngMu parameters allow callers to share a single
// crypto/rand-seeded source across the client's lifetime.
func backoffDelay(cfg RetryConfig, attempt int, rng *mrand.Rand, rngMu *sync.Mutex) time.Duration {
	base := float64(cfg.BaseDelay)
	exp := base * math.Pow(2, float64(attempt-1))

	maxDelay := float64(cfg.MaxDelay)
	capped := math.Min(exp, maxDelay)

	rngMu.Lock()
	jitterRange := capped * cfg.Jitter
	jittered := capped + jitterRange*(rng.Float64()*2-1)
	rngMu.Unlock()

	if jittered < 0 {
		return 0
	}
	return time.Duration(jittered)
}
