package custd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PrivateUptimeFailure is why a private uptime check failed.
type PrivateUptimeFailure string

const (
	PrivateUptimeFailureDNS        PrivateUptimeFailure = "dns"
	PrivateUptimeFailureConnection PrivateUptimeFailure = "connection"
	PrivateUptimeFailureTLS        PrivateUptimeFailure = "tls"
	PrivateUptimeFailureTimeout    PrivateUptimeFailure = "timeout"
)

// PrivateUptimeReport is an observation from a probe inside the monitored network.
type PrivateUptimeReport struct {
	ReportUUID string               `json:"reportUuid"`
	ObservedAt time.Time            `json:"observedAt"`
	StatusCode int                  `json:"statusCode,omitempty"`
	ResponseMs int                  `json:"responseMs"`
	Failure    PrivateUptimeFailure `json:"failure,omitempty"`
}

// PrivateUptimeDestination holds the API origin and the monitor's report path and write token.
type PrivateUptimeDestination struct {
	BaseURL    string
	ReportPath string
	WriteToken string
}

const privateUptimeReportPrefix = "/api/v1/site-private-uptime-reports/"

// SubmitPrivateUptime sends a report using only the monitor's write credential.
func SubmitPrivateUptime(ctx context.Context, destination PrivateUptimeDestination, report PrivateUptimeReport) error {
	body, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("custd: encode uptime report: %w", err)
	}
	response, err := submitSiteMonitorRequest(ctx, siteMonitorDestination(destination), privateUptimeReportPrefix, http.MethodPost, body)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("custd: uptime report returned HTTP %d", response.StatusCode)
	}
	return nil
}

// PrivateUptimeWork is the pending check for the configured monitor, if any.
type PrivateUptimeWork struct {
	RequestedAt *time.Time `json:"requestedAt,omitempty"`
}

// GetPrivateUptimeWork reads pending check work using the monitor credential.
func GetPrivateUptimeWork(ctx context.Context, destination PrivateUptimeDestination) (PrivateUptimeWork, error) {
	var work PrivateUptimeWork
	response, err := submitSiteMonitorRequest(ctx, siteMonitorDestination(destination), privateUptimeReportPrefix, http.MethodGet, nil)
	if err != nil {
		return work, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return work, fmt.Errorf("custd: uptime work returned HTTP %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 1025))
	if err != nil {
		return work, fmt.Errorf("custd: read uptime work: %w", err)
	}
	if len(body) > 1024 {
		return work, fmt.Errorf("custd: uptime work exceeds 1024 bytes")
	}
	if err := json.Unmarshal(body, &work); err != nil {
		return work, fmt.Errorf("custd: decode uptime work: %w", err)
	}
	return work, nil
}
