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
type Problem struct {
	Type     string            `json:"type,omitempty"`
	Title    string            `json:"title,omitempty"`
	Status   int               `json:"status,omitempty"`
	Detail   string            `json:"detail,omitempty"`
	Code     string            `json:"code,omitempty"`
	Instance string            `json:"instance,omitempty"`
	TraceID  string            `json:"traceId,omitempty"`
	Fields   map[string]string `json:"fields,omitempty"`
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

// parseProblem decodes an RFC 9457 problem+json body. It returns nil when the
// body is empty or cannot be decoded as a problem, so callers can fall back to
// a status-only error.
func parseProblem(body []byte) *Problem {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return nil
	}
	var p Problem
	if err := json.Unmarshal([]byte(trimmed), &p); err != nil {
		return nil
	}
	if p.Type == "" && p.Title == "" && p.Detail == "" && p.Status == 0 && p.Code == "" {
		return nil
	}
	return &p
}
