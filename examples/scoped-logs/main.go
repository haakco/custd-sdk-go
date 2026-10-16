// Command scoped-logs is a tenant-administrator walkthrough of the reusable-log
// setup on the Go SDK.
//
// It shows the whole self-service path with no operator involved: authenticate
// with the tenant's own credential, declare two scopes with a resource each,
// grant read and write to the administrator and the producer, upload a raw log
// file scoped to one resource, then read that resource's logs back.
//
// There is no company slug, no SQL, and no secret copy: the tenant comes from
// the credential and the target from registered immutable codes.
//
//	CUSTD_BASE_URL=https://custd.example \
//	CUSTD_TOKEN=<tenant credential with data_scopes.manage and logs ingest/read> \
//	CUSTD_PARSER_UUID=<uuid of the enabled parser to execute> \
//	CUSTD_PRODUCER_CLIENT_ID=<the producer client the events arrive under> \
//	CUSTD_ADMIN_USER_UUID=<the administrator's membership user UUID> \
//	CUSTD_LOG_FILE=./access.log \
//	go run ./examples/scoped-logs
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	custd "github.com/haakco/custd-sdk-go/v3"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "scoped-logs: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	baseURL := requireEnv("CUSTD_BASE_URL")
	token := requireEnv("CUSTD_TOKEN")
	parserUUID := requireEnv("CUSTD_PARSER_UUID")
	logPath := requireEnv("CUSTD_LOG_FILE")

	// 1. Authenticate. The credential already names the tenant and the scopes it
	// may use; a real deployment uses a provisioned producer bundle or OAuth
	// rather than a copied token. No secret is read from a file here.
	client := custd.NewClient(&custd.ClientConfig{BaseURL: baseURL, APIKey: token})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 2. Declare the sub-client scope and one resource inside it. The codes are
	// the immutable selectors events and queries use.
	scope, err := client.Admin.DataScopes.CreateScope(ctx, custd.DataScopeCreateRequest{
		Code: "acme-store", Name: "Acme Store",
	})
	if err != nil {
		return fmt.Errorf("create scope: %w", err)
	}
	resource, err := client.Admin.DataScopes.CreateResource(ctx, scope.UUID, custd.DataScopeResourceCreateRequest{
		Code: "srv-01", Name: "Storefront server",
	})
	if err != nil {
		return fmt.Errorf("create resource: %w", err)
	}

	// 3. Grant read+write explicitly. The administrator is a human (membership
	// user UUID); the producer that will upload is a machine (its client ID). A
	// parser binding is not a data grant: without one of these the upload is
	// refused.
	principals := []custd.DataScopePrincipalRef{
		{Kind: custd.DataScopePrincipalHuman, UserUUID: requireEnv("CUSTD_ADMIN_USER_UUID")},
		{Kind: custd.DataScopePrincipalMachine, ClientID: requireEnv("CUSTD_PRODUCER_CLIENT_ID")},
	}
	for _, principal := range principals {
		for _, action := range []custd.DataScopeAction{custd.DataScopeActionRead, custd.DataScopeActionWrite} {
			if _, err := client.Admin.DataScopes.CreateGrant(ctx, custd.DataScopeGrant{
				Principal: principal,
				Target: custd.DataScopeGrantTarget{
					Kind:         custd.DataScopeGrantTargetResource,
					ScopeUUID:    scope.UUID,
					ResourceUUID: resource.UUID,
				},
				Action: action,
			}); err != nil {
				return fmt.Errorf("grant %s %s: %w", action, principal.Key(), err)
			}
		}
	}

	// 4. Upload the raw file scoped to the resource. Apache/Nginx/PHP files are
	// parsed as-is: the producer names the registered target once and the backend
	// fills the context and checks the grant before anything is durably written.
	content, err := os.ReadFile(logPath)
	if err != nil {
		return fmt.Errorf("read log file: %w", err)
	}
	result, err := client.IngestRawLines(ctx, custd.RawIngestRequest{
		Content:    string(content),
		ParserUUID: parserUUID,
		Target:     &custd.RawSourceTarget{SubClient: scope.Code, Resource: resource.Code},
	})
	if err != nil {
		return fmt.Errorf("upload raw file: %w", err)
	}
	// The route answers 202 Accepted: the status alone is not per-record success.
	fmt.Printf("uploaded: processed=%d failed=%d\n", result.Processed, result.Failed)
	if result.Failed > 0 {
		fmt.Println("some lines were refused; see the ingest errors for the missing grant or selection")
	}

	// 5. Read the scoped logs back. The same immutable codes narrow the query, so
	// only the granted resource's rows are returned. Scoping a reporting query
	// uses the same codes on the reporting surface.
	today := time.Now().UTC().Format("2006-01-02")
	logs, err := client.Analytics.Query(ctx, custd.AnalyticsEventQueryRequest{
		Date:      today,
		SubClient: scope.Code,
		Resource:  resource.Code,
		Limit:     50,
	})
	if err != nil {
		return fmt.Errorf("query scoped logs: %w", err)
	}
	fmt.Printf("scoped rows: %d\n", logs.Count)

	return nil
}

func requireEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		fmt.Fprintf(os.Stderr, "scoped-logs: set %s\n", key)
		os.Exit(2)
	}
	return value
}
