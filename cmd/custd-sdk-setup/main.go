package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	custd "github.com/haakco/custd-sdk/sdk-go"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	var scopes csvFlag
	fs := flag.NewFlagSet("custd-sdk-setup", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	baseURL := fs.String("base-url", env("CUSTD_BASE_URL", ""), "Custd ingest/API base URL")
	adminURL := fs.String("admin-url", env("CUSTD_ADMIN_URL", ""), "Custd admin API base URL; defaults to base-url")
	adminToken := fs.String("admin-token", env("CUSTD_ADMIN_TOKEN", ""), "Custd admin bearer token")
	tokenURL := fs.String("token-url", env("CUSTD_OAUTH_TOKEN_URL", ""), "OAuth2 token URL")
	audience := fs.String("audience", env("CUSTD_OAUTH_AUDIENCE", "custd"), "OAuth2 audience")
	tenant := fs.String("tenant", env("CUSTD_TENANT_SLUG", ""), "Tenant/company slug")
	companyName := fs.String("company-name", env("CUSTD_COMPANY_NAME", ""), "Company display name")
	clientID := fs.String("client-id", env("CUSTD_OAUTH_CLIENT_ID", ""), "Producer OAuth client ID")
	environment := fs.String("environment", env("CUSTD_ENVIRONMENT", "production"), "Producer environment label")
	prefix := fs.String("env-prefix", env("CUSTD_ENV_PREFIX", "CUSTD"), "Environment variable prefix for generic output")
	schemaDir := fs.String("register-schemas", env("CUSTD_SCHEMA_DIR", ""), "Directory of schema JSON files to register after producer setup")
	ensureTenant := fs.Bool("ensure-tenant", true, "Create the tenant before creating the OAuth client")
	fs.Var(&scopes, "scope", "OAuth scope; may be repeated or comma-separated")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *adminURL == "" {
		*adminURL = *baseURL
	}
	if *adminToken == "" {
		return fmt.Errorf("custd-sdk-setup: --admin-token or CUSTD_ADMIN_TOKEN is required")
	}
	if *companyName == "" {
		*companyName = *tenant
	}

	admin := custd.NewClient(&custd.ClientConfig{
		BaseURL: *adminURL,
		APIKey:  *adminToken,
	})
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = admin.Close(shutdownCtx)
	}()

	creds, err := custd.SetupProducer(ctx, admin, custd.ProducerSetupRequest{
		BaseURL:      *baseURL,
		TokenURL:     *tokenURL,
		Audience:     *audience,
		TenantSlug:   *tenant,
		CompanyName:  *companyName,
		ClientID:     *clientID,
		Scopes:       scopes.values(),
		Environment:  *environment,
		EnsureTenant: *ensureTenant,
		SchemaDir:    *schemaDir,
	})
	if err != nil {
		return err
	}
	fmt.Println(custd.EnvSnippets(*prefix, *creds))
	return nil
}

func env(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

type csvFlag []string

func (f *csvFlag) String() string {
	return strings.Join(*f, ",")
}

func (f *csvFlag) Set(value string) error {
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			*f = append(*f, part)
		}
	}
	return nil
}

func (f *csvFlag) values() []string {
	return append([]string(nil), *f...)
}
