package custd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Problem is the RFC 9457 problem detail the ingest API returns for error
// responses (and for failed per-event results inside a batch). Optional fields
// are omitempty on the server side and may be absent.
//
// Lifecycle endpoints (tenant-storage, subject-exports, privacy/erasures,
// retention, offboarding) use a flatter {error, message, safeNextAction,
// requestId} envelope. parseProblem normalizes that shape into Problem so
// callers can branch on Code and SafeNextAction the same way they do for
// ingest problems.
type Problem struct {
	Type     string            `json:"type,omitempty"`
	Title    string            `json:"title,omitempty"`
	Status   int               `json:"status,omitempty"`
	Detail   string            `json:"detail,omitempty"`
	Code     string            `json:"code,omitempty"`
	Instance string            `json:"instance,omitempty"`
	TraceID  string            `json:"traceId,omitempty"`
	Fields   map[string]string `json:"fields,omitempty"`
	// SafeNextAction is the server's recommended next step for retryable
	// destructive flows (e.g. retry_erasure). Empty when the server does
	// not provide one.
	SafeNextAction string `json:"safeNextAction,omitempty"`
	// RequestID is the server-side request identifier the flat error
	// envelope carries. Useful for operator lookup; never echo into
	// per-subject logs.
	RequestID string `json:"requestId,omitempty"`
}

// Error renders the problem as a human-readable message. It leads with the
// detail (or title) and appends the status, code, and field errors when present
// so a logged error is diagnosable without re-fetching the body.
func (p *Problem) Error() string {
	msg := p.Detail
	if msg == "" {
		msg = p.Title
	}
	if msg == "" {
		msg = "request failed"
	}
	parts := []string{msg}
	if p.Status != 0 {
		parts = append(parts, fmt.Sprintf("status %d", p.Status))
	}
	if p.Code != "" {
		parts = append(parts, "code "+p.Code)
	}
	if p.SafeNextAction != "" {
		parts = append(parts, "safeNextAction="+p.SafeNextAction)
	}
	if len(p.Fields) > 0 {
		parts = append(parts, "fields: "+formatFields(p.Fields))
	}
	return "custd: " + strings.Join(parts, "; ")
}

func formatFields(fields map[string]string) string {
	parts := make([]string, 0, len(fields))
	for name, msg := range fields {
		parts = append(parts, name+"="+msg)
	}
	// Stable ordering keeps error strings deterministic for tests and logs.
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

// flatErrorEnvelope is the lifecycle endpoints' non-RFC error shape. It is
// parsed leniently: only the {error, message} pair is required.
type flatErrorEnvelope struct {
	Error          string `json:"error"`
	Message        string `json:"message"`
	SafeNextAction string `json:"safeNextAction"`
	RequestID      string `json:"requestId"`
}

// parseProblem decodes an RFC 9457 problem+json body, or a flat
// {error, message, safeNextAction} envelope. It returns nil when the body
// is empty or cannot be decoded as either, so callers can fall back to a
// status-only error.
func parseProblem(body []byte) *Problem {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return nil
	}
	var p Problem
	if err := json.Unmarshal([]byte(trimmed), &p); err == nil {
		if p.Type != "" || p.Title != "" || p.Detail != "" || p.Status != 0 || p.Code != "" {
			return &p
		}
	}
	var flat flatErrorEnvelope
	if err := json.Unmarshal([]byte(trimmed), &flat); err != nil {
		return nil
	}
	if flat.Error == "" && flat.Message == "" {
		return nil
	}
	return &Problem{
		Title:          flat.Message,
		Code:           flat.Error,
		SafeNextAction: flat.SafeNextAction,
		RequestID:      flat.RequestID,
	}
}
