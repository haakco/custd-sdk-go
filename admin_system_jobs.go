package custd

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// JobStatusResponse is the current worker queue and outbox snapshot.
type JobStatusResponse struct {
	QueueCounts    []JobQueueCount   `json:"queueCounts"`
	RecentFailures []FailedJob       `json:"recentFailures"`
	Outboxes       []JobOutboxStatus `json:"outboxes"`
	RealtimeEvents int64             `json:"realtimeEvents"`
}

// JobQueueCount counts one state for a job kind in a queue.
type JobQueueCount struct {
	Queue           string     `json:"queue"`
	Kind            string     `json:"kind"`
	State           string     `json:"state"`
	Count           int64      `json:"count"`
	OldestWaitingAt *time.Time `json:"oldestWaitingAt,omitempty"`
}

// FailedJob identifies a discarded job without its payload or raw error.
type FailedJob struct {
	JobID       int64     `json:"jobId"`
	Queue       string    `json:"queue"`
	Kind        string    `json:"kind"`
	Attempts    int       `json:"attempts"`
	MaxAttempts int       `json:"maxAttempts"`
	FailedAt    time.Time `json:"failedAt"`
}

// JobOutboxRetryReason is a sanitized notification delivery retry category.
type JobOutboxRetryReason string

const (
	JobOutboxRetryProviderTransient      JobOutboxRetryReason = "provider_transient"
	JobOutboxRetryAuthenticationRejected JobOutboxRetryReason = "authentication_rejected"
	JobOutboxRetryTLSFailure             JobOutboxRetryReason = "tls_failure"
	JobOutboxRetryConnectionFailure      JobOutboxRetryReason = "connection_failure"
	JobOutboxRetrySenderRejected         JobOutboxRetryReason = "sender_rejected"
	JobOutboxRetryRecipientRejected      JobOutboxRetryReason = "recipient_rejected"
	JobOutboxRetryUnclassified           JobOutboxRetryReason = "unclassified_retry"
)

// JobOutboxStatus gives the backlog reported by an available outbox.
// The notification outbox may include its oldest pending delivery and latest retry category.
type JobOutboxStatus struct {
	Name            string               `json:"name"`
	Pending         int64                `json:"pending"`
	Failed          int64                `json:"failed"`
	OldestPendingAt *time.Time           `json:"oldestPendingAt,omitempty"`
	LastRetryReason JobOutboxRetryReason `json:"lastRetryReason,omitempty"`
}

// JobRetryReceipt identifies the job accepted for retry.
type JobRetryReceipt struct {
	JobID int64 `json:"jobId"`
}

const jobsPath = "/jobs"

// ListJobs returns worker queue counts, recent failures, and available outbox counts.
// This operation requires system administrator access.
func (c *SystemAdminClient) ListJobs(ctx context.Context) (*JobStatusResponse, ResponseMetadata, error) {
	var out JobStatusResponse
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodGet, jobsPath, nil, &out)
	return &out, metadata, err
}

// RetryJob asks the server to requeue a discarded job. The server audits acceptance.
// This operation requires system administrator access.
func (c *SystemAdminClient) RetryJob(ctx context.Context, jobID int64) (*JobRetryReceipt, ResponseMetadata, error) {
	var out JobRetryReceipt
	path := jobsPath + "/" + strconv.FormatInt(jobID, 10) + "/retry"
	metadata, err := c.admin.requestWithMetadata(ctx, http.MethodPost, path, nil, &out)
	return &out, metadata, err
}
