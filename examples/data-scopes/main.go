// Command data-scopes is a tenant-administrator example for the Go SDK's
// data-scope management.
//
// It shows the whole self-service path a tenant administrator needs to publish
// sub-client scopes without HaakCo intervention: declare two scopes, add a
// resource under each, grant a machine principal read and write over each scope,
// then read the configuration back exactly as an auditor would.
//
// The tenant is derived from the credential, so this program never names a
// company slug, never copies a secret out of the platform, and never issues SQL.
// It reuses the SDK's existing client credential flow (a tenant-admin machine
// token, or the OAuth options a provisioned producer already carries).
//
//	CUSTD_BASE_URL=https://custd.example \
//	CUSTD_TOKEN=<machine credential holding data_scopes.manage> \
//	CUSTD_PRODUCER_CLIENT_ID=<the producer client the events arrive under> \
//	go run ./examples/data-scopes
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	custd "github.com/haakco/custd-sdk-go/v3"
)

// scopeDeclaration is one sub-client the tenant wants to publish. The code is
// the immutable selector event and query traffic use; the resource is the
// smallest addressable unit inside it.
type scopeDeclaration struct {
	scope    custd.DataScopeCreateRequest
	resource custd.DataScopeResourceCreateRequest
}

var declarations = []scopeDeclaration{
	{
		scope:    custd.DataScopeCreateRequest{Code: "acme-store", Name: "Acme Store"},
		resource: custd.DataScopeResourceCreateRequest{Code: "srv-01", Name: "Storefront server"},
	},
	{
		scope:    custd.DataScopeCreateRequest{Code: "acme-warehouse", Name: "Acme Warehouse"},
		resource: custd.DataScopeResourceCreateRequest{Code: "db-01", Name: "Warehouse database"},
	},
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "data-scopes: %v\n", err)
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
	scopes := client.Admin.DataScopes

	// A grant principal is the closed human/machine shape. This example grants a
	// producer machine: the client ID its events already arrive under, not an
	// email or a display name.
	principal := custd.DataScopePrincipalRef{
		Kind:     custd.DataScopePrincipalMachine,
		ClientID: envOr("CUSTD_PRODUCER_CLIENT_ID", "custd-example-producer"),
	}

	for _, declaration := range declarations {
		scope, err := scopes.CreateScope(ctx, declaration.scope)
		if err != nil {
			return fmt.Errorf("create scope %s: %w", declaration.scope.Code, err)
		}
		resource, err := scopes.CreateResource(ctx, scope.UUID, declaration.resource)
		if err != nil {
			return fmt.Errorf("create resource %s: %w", declaration.resource.Code, err)
		}

		// A sub-client grant covers every resource inside the scope, including the
		// one just created; a caller that needs exactly one resource uses Target
		// kind "resource" with both UUIDs instead.
		for _, action := range []custd.DataScopeAction{custd.DataScopeActionRead, custd.DataScopeActionWrite} {
			if _, err := scopes.CreateGrant(ctx, custd.DataScopeGrant{
				Principal: principal,
				Target:    custd.DataScopeGrantTarget{Kind: custd.DataScopeGrantTargetSubClient, ScopeUUID: scope.UUID},
				Action:    action,
			}); err != nil {
				return fmt.Errorf("grant %s %s: %w", action, scope.Code, err)
			}
		}

		fmt.Printf(
			"declared scope=%s uuid=%s status=%s resource=%s uuid=%s\n",
			scope.Code, scope.UUID, scope.Status, resource.Code, resource.UUID,
		)
	}

	// Read the configuration back so an operator can audit exactly what exists.
	if err := describeConfiguration(ctx, scopes, principal); err != nil {
		return err
	}

	// A scope or resource that should stop being selectable is disabled, not
	// deleted: already-written rows keep their scope, but no new grant or
	// selection resolves through a disabled identity.
	return nil
}

func describeConfiguration(
	ctx context.Context, scopes *custd.DataScopeAdminClient, principal custd.DataScopePrincipalRef,
) error {
	list, err := scopes.ListScopes(ctx, true)
	if err != nil {
		return fmt.Errorf("list scopes: %w", err)
	}
	fmt.Printf("scopes: %d\n", len(list.Scopes))
	for _, scope := range list.Scopes {
		resources, err := scopes.ListResources(ctx, scope.UUID, true)
		if err != nil {
			return fmt.Errorf("list resources for %s: %w", scope.Code, err)
		}
		fmt.Printf("  %s (%s) status=%s resources=%d\n", scope.Code, scope.UUID, scope.Status, len(resources.Resources))
	}

	grants, err := scopes.ListGrants(ctx, principal)
	if err != nil {
		return fmt.Errorf("list grants: %w", err)
	}
	fmt.Printf("grants for %s: %d\n", principal.Key(), len(grants.Grants))
	for _, grant := range grants.Grants {
		fmt.Printf("  %s %s target=%s\n", grant.Action, grant.Target.Kind, grant.Target.ScopeUUID)
	}

	return nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
