// Command usage-report is a collector example for the Go SDK's tenant usage
// report.
//
// It shows the whole path a client needs to see what the platform attributes to
// it: read the trailing 30-day usage for the credential's own tenant, then read
// one meter over an explicit window, and decide whether each number is settled
// before acting on it.
//
// The tenant is derived from the credential, so this program never names a
// company slug and never sees another tenant's usage.
//
//	CUSTD_BASE_URL=https://custd.com \
//	CUSTD_TOKEN=<machine credential with tenant-admin usage read> \
//	go run ./examples/usage-report
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	custd "github.com/haakco/custd-sdk-go/v2"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "usage-report: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	baseURL := os.Getenv("CUSTD_BASE_URL")
	token := os.Getenv("CUSTD_TOKEN")
	if baseURL == "" || token == "" {
		return errors.New("set CUSTD_BASE_URL and CUSTD_TOKEN")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := custd.NewClient(&custd.ClientConfig{BaseURL: baseURL, APIKey: token})

	// No window means the service default: the trailing 30 days ending now.
	report, err := client.Admin.Usage.Get(ctx, custd.UsageQuery{})
	if err != nil {
		return fmt.Errorf("read usage: %w", err)
	}
	fmt.Println("trailing 30 days:")
	describe(report)

	// A caller that needs an authoritative reading for one meter over an explicit
	// window asks for exactly that window rather than trusting a default.
	end := time.Now().UTC()
	windowed, err := client.Admin.Usage.Get(ctx, custd.UsageQuery{
		MeterSlug: "events.ingested",
		Start:     end.AddDate(0, 0, -7),
		End:       end,
		Limit:     200,
	})
	if err != nil {
		return fmt.Errorf("read windowed usage: %w", err)
	}
	fmt.Println("events.ingested, last 7 days:")
	describe(windowed)

	return nil
}

func describe(report *custd.UsageReport) {
	fmt.Printf(
		"schema=%s tenant=%s window=[%s, %s) watermark=%d rows=%d\n",
		report.SchemaVersion, report.CompanySlug, report.Start, report.End, report.SourceWatermark, len(report.Rows),
	)
	for _, total := range report.Totals {
		fmt.Printf(
			"  total %s %s/%s = %d %s\n",
			total.MeterSlug, total.AccountCompanySlug, total.DataSpaceCompanySlug, total.Quantity, total.Unit,
		)
	}
	if report.ContainsProvisional || report.ContainsIncomplete {
		// A number that is not final is not a failure: it is the honest state
		// while the window settles or a source is behind. Report the state rather
		// than presenting the reading as final.
		fmt.Printf(
			"  containsProvisional=%t containsIncomplete=%t\n",
			report.ContainsProvisional, report.ContainsIncomplete,
		)
		return
	}
	for _, row := range report.Rows {
		fmt.Printf(
			"  row %s %s %s=%d state=%s\n",
			row.MeterSlug, row.WindowStart, row.Unit, row.Quantity, row.CompletenessState,
		)
	}
}
