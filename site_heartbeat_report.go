package custd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SiteHeartbeatOutcome is the final result of a scheduled job occurrence.
type SiteHeartbeatOutcome string

const (
	SiteHeartbeatSucceeded SiteHeartbeatOutcome = "succeeded"
	SiteHeartbeatFailed    SiteHeartbeatOutcome = "failed"
)

// SiteHeartbeatDestination pairs the API origin with the job monitor's write credential.
type SiteHeartbeatDestination struct {
	BaseURL    string
	ReportPath string
	WriteToken string
}

const heartbeatReportPrefix = "/api/v1/site-heartbeat-reports/"

// SubmitSiteHeartbeat sends a job result and returns the first accepted receipt for that occurrence.
func SubmitSiteHeartbeat(ctx context.Context, destination SiteHeartbeatDestination, report SiteHeartbeatReport) (AdminSiteHeartbeatReceipt, error) {
	var receipt AdminSiteHeartbeatReceipt
	if report.ScheduledFor.IsZero() || (report.Outcome != SiteHeartbeatSucceeded && report.Outcome != SiteHeartbeatFailed) {
		return receipt, fmt.Errorf("custd: heartbeat requires a scheduled time and succeeded or failed outcome")
	}
	body, err := json.Marshal(report)
	if err != nil {
		return receipt, fmt.Errorf("custd: encode heartbeat report: %w", err)
	}
	response, err := submitSiteMonitorRequest(ctx, siteMonitorDestination(destination), heartbeatReportPrefix, http.MethodPost, body)
	if err != nil {
		return receipt, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return receipt, fmt.Errorf("custd: heartbeat report returned HTTP %d", response.StatusCode)
	}
	const receiptLimit = 1024
	raw, err := io.ReadAll(io.LimitReader(response.Body, receiptLimit+1))
	if err != nil {
		return receipt, fmt.Errorf("custd: read heartbeat receipt: %w", err)
	}
	if len(raw) > receiptLimit {
		return receipt, fmt.Errorf("custd: heartbeat receipt exceeds %d bytes", receiptLimit)
	}
	if err := json.Unmarshal(raw, &receipt); err != nil {
		return receipt, fmt.Errorf("custd: decode heartbeat receipt: %w", err)
	}
	if receipt.ReceivedAt.IsZero() || !receipt.Report.ScheduledFor.Equal(report.ScheduledFor) || receipt.Report.Outcome != report.Outcome {
		return receipt, fmt.Errorf("custd: heartbeat receipt does not match the submitted occurrence")
	}
	return receipt, nil
}
