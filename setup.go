package custd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var envNamePattern = regexp.MustCompile(`[^A-Z0-9]+`)

// ProducerSetupRequest describes the admin operation needed to provision a
// producer integration and the runtime URLs consumers need afterward.
type ProducerSetupRequest struct {
	BaseURL      string
	TokenURL     string
	Audience     string
	TenantSlug   string
	CompanyName  string
	ClientID     string
	Scopes       []string
	Environment  string
	EnsureTenant bool
	SchemaDir    string
}

// ProducerCredentials is the secret-bearing result returned once Custd creates
// a producer OAuth client.
type ProducerCredentials struct {
	BaseURL      string
	TokenURL     string
	Audience     string
	TenantSlug   string
	ClientID     string
	ClientSecret string
	Scopes       []string
	Environment  string
}

// SetupProducer ensures the tenant exists when requested, creates an OAuth2
// producer client, and returns the credential bundle consumers need.
func SetupProducer(
	ctx context.Context,
	admin *CustdClient,
	req ProducerSetupRequest,
) (*ProducerCredentials, error) {
	if err := ValidateProducerSetupRequest(req); err != nil {
		return nil, err
	}
	schemas, err := LoadSchemaRegistrationsFromDir(req.SchemaDir)
	if err != nil {
		return nil, err
	}
	if req.EnsureTenant {
		_, err := admin.Admin.Tenants.Create(ctx, AdminTenantCreate{
			Slug:        req.TenantSlug,
			CompanyName: req.CompanyName,
		})
		if err != nil && !isAlreadyExistsError(err) {
			return nil, fmt.Errorf("custd: create tenant: %w", err)
		}
	}
	if len(schemas) > 0 {
		if err := RegisterSchemas(ctx, admin, schemas); err != nil {
			return nil, err
		}
	}
	created, err := admin.Admin.OAuthClients.Create(ctx, AdminOAuthClientCreate{
		ClientID:    req.ClientID,
		CompanySlug: req.TenantSlug,
		Scopes:      normalizeScopes(req.Scopes),
	})
	if err != nil {
		return nil, fmt.Errorf("custd: create oauth client: %w", err)
	}
	return &ProducerCredentials{
		BaseURL:      strings.TrimRight(req.BaseURL, "/"),
		TokenURL:     req.TokenURL,
		Audience:     req.Audience,
		TenantSlug:   req.TenantSlug,
		ClientID:     created.ClientID,
		ClientSecret: created.ClientSecret,
		Scopes:       created.Scopes,
		Environment:  req.Environment,
	}, nil
}

func RegisterSchemasFromDir(ctx context.Context, admin *CustdClient, dir string) error {
	schemas, err := LoadSchemaRegistrationsFromDir(dir)
	if err != nil {
		return err
	}
	return RegisterSchemas(ctx, admin, schemas)
}

func LoadSchemaRegistrationsFromDir(dir string) ([]AdminSchemaRegister, error) {
	if dir == "" {
		return nil, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("custd: read schema directory: %w", err)
	}
	schemas := make([]AdminSchemaRegister, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("custd: read schema %s: %w", path, err)
		}
		var schema AdminSchemaRegister
		if err := json.Unmarshal(raw, &schema); err != nil {
			return nil, fmt.Errorf("custd: decode schema %s: %w", path, err)
		}
		schemas = append(schemas, schema)
	}
	return schemas, nil
}

func RegisterSchemas(ctx context.Context, admin *CustdClient, schemas []AdminSchemaRegister) error {
	for _, schema := range schemas {
		if _, err := admin.Admin.Schemas.Register(ctx, schema); err != nil {
			return fmt.Errorf("custd: register schema %s: %w", schema.EventTypeSlug, err)
		}
	}
	return nil
}

// ValidateProducerSetupRequest returns clear setup errors before network calls
// are made.
func ValidateProducerSetupRequest(req ProducerSetupRequest) error {
	missing := make([]string, 0)
	if strings.TrimSpace(req.BaseURL) == "" {
		missing = append(missing, "base URL")
	}
	if strings.TrimSpace(req.TokenURL) == "" {
		missing = append(missing, "token URL")
	}
	if strings.TrimSpace(req.TenantSlug) == "" {
		missing = append(missing, "tenant slug")
	}
	if strings.TrimSpace(req.ClientID) == "" {
		missing = append(missing, "client ID")
	}
	if len(missing) > 0 {
		return fmt.Errorf("custd: missing %s", strings.Join(missing, ", "))
	}
	if !isSecureOrLocalHTTP(req.BaseURL) {
		return fmt.Errorf("custd: base URL must use https unless it targets localhost")
	}
	if !isSecureOrLocalHTTP(req.TokenURL) {
		return fmt.Errorf("custd: token URL must use https unless it targets localhost")
	}
	return nil
}

// EnvSnippets renders copy-pasteable environment variable groups for the
// supported SDK consumers.
func EnvSnippets(prefix string, creds ProducerCredentials) string {
	name := envPrefix(prefix)
	scope := strings.Join(normalizeScopes(creds.Scopes), " ")
	blocks := []string{
		envBlock("Generic", name, creds, scope),
		envBlock("Go / TypeScript / Python / PHP", name, creds, scope),
		envBlock("Laravel", "CUSTD", creds, scope),
		envBlock("WordPress", "CUSTD_WP", creds, scope),
	}
	return strings.Join(blocks, "\n\n")
}

func envBlock(title string, prefix string, creds ProducerCredentials, scope string) string {
	lines := []string{
		"# " + title,
		fmt.Sprintf("%s_BASE_URL=%q", prefix, creds.BaseURL),
		fmt.Sprintf("%s_OAUTH_CLIENT_ID=%q", prefix, creds.ClientID),
		fmt.Sprintf("%s_OAUTH_CLIENT_SECRET=%q", prefix, creds.ClientSecret),
		fmt.Sprintf("%s_OAUTH_TOKEN_URL=%q", prefix, creds.TokenURL),
		fmt.Sprintf("%s_OAUTH_AUDIENCE=%q", prefix, creds.Audience),
		fmt.Sprintf("%s_OAUTH_SCOPES=%q", prefix, scope),
		fmt.Sprintf("%s_TENANT_SLUG=%q", prefix, creds.TenantSlug),
		fmt.Sprintf("%s_ENVIRONMENT=%q", prefix, creds.Environment),
	}
	return strings.Join(lines, "\n")
}

func envPrefix(prefix string) string {
	prefix = strings.ToUpper(strings.TrimSpace(prefix))
	prefix = envNamePattern.ReplaceAllString(prefix, "_")
	prefix = strings.Trim(prefix, "_")
	if prefix == "" {
		return "CUSTD"
	}
	return prefix
}

func normalizeScopes(scopes []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" || seen[scope] {
			continue
		}
		seen[scope] = true
		out = append(out, scope)
	}
	if len(out) == 0 {
		out = append(out, "events.write")
	}
	sort.Strings(out)
	return out
}

func isAlreadyExistsError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "409") || strings.Contains(msg, "already")
}
