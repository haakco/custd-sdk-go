# custd SDK (Go)

Ingestion client with retry, batching, and queueing.

## Compatibility

Version `1.0.0` targets the canonical ingest endpoint
`POST /api/v1/events`. The legacy `POST /v1/events` path is not supported.
This SDK was not released against the legacy path, so there is no compatibility
alias or deprecation window.

## Install

Consume the Go module from its dedicated mirror `github.com/haakco/custd-sdk-go/v2`
(tagged `vX.Y.Z`):

```bash
go get github.com/haakco/custd-sdk-go/v2@latest
```

```go
import custd "github.com/haakco/custd-sdk-go/v2"
```

> This module is developed in the [`custd-sdk`](https://github.com/haakco/custd-sdk)
> monorepo under `sdk-go/` and published to the read-only `custd-sdk-go` mirror on
> each release. Import the mirror path above, not the monorepo subdir.

## Migrating from `github.com/haakco/custd-sdk/sdk-go`

The module path **changed in `v1.3.2`** (breaking). Versions up to and including
`sdk-go/v1.3.1` were imported from the monorepo subdir path
`github.com/haakco/custd-sdk/sdk-go`; that path is now frozen and will not receive
further releases.

To move to `v1.3.2`+:

```bash
# 1. Update import paths in your code
#    github.com/haakco/custd-sdk/sdk-go  ->  github.com/haakco/custd-sdk-go
# 2. Pull the renamed module
go get github.com/haakco/custd-sdk-go@v1.3.2
# 3. Drop the now-unused old module
go mod tidy
```

The package name (`custd`) and every exported symbol are unchanged — only the
module/import path moves, so it is a mechanical find-and-replace.

## Usage

Static token:

```go
client := custd.NewClient(&custd.ClientConfig{
    BaseURL: "http://localhost:8087",
    APIKey:  "<token>",
})
defer client.Close(context.Background())
```

OAuth2 producer client:

```go
client := custd.NewClient(&custd.ClientConfig{
    BaseURL:      "https://ingest.custd.example",
    ClientID:     "producer-client",
    ClientSecret: os.Getenv("CUSTD_CLIENT_SECRET"),
    TokenURL:     "https://hydra.example/oauth2/token",
    Audience:     "custd",
    Scopes:       []string{"events.write"},
    BatchSize:    50,
    MaxQueueSize: 1000,
})
defer client.Close(context.Background())
```

Provisioned producer bundle (no manual OAuth mapping):

```go
client, err := custd.NewClientFromProvisionedProducer(creds)
if err != nil {
    return err
}
defer client.Close(context.Background())
_ = client.Track(context.Background(), &custd.EventEnvelope{
    EventTypeSlug: "order.completed",
    SchemaVersion: "1.0.0",
    CompanySlug:   creds.CompanySlug,
    Context:       custd.EventContext{Device: &custd.DeviceContext{Type: "server"}},
    Payload:       map[string]any{"orderTotal": 42},
})
```

Use `custd.RedactedProvisionedProducer(creds)` to show the bundle on a dashboard
without exposing the client secret.

The client rejects plaintext non-local Custd and token URLs. Localhost HTTP is
allowed for development.

Dogfood producers can use `NewDogfoodEvent` to build the canonical event shape
with `sourceSystem`, `sourceCompany`, `environment`, `schemaVersion`, and
`correlationId` in the payload while keeping `companySlug` on the envelope.

## Declaring an environment

One credential serves every environment, so sending from `dev`, a preview build,
or a local machine needs no extra provisioning:

```go
client := custd.NewClient(&custd.ClientConfig{BaseURL: baseURL, APIKey: token, Environment: "dev"})

// Per event, when this one came from somewhere else.
_ = client.Track(ctx, &custd.EventEnvelope{ ..., Environment: "preview-pr-9" })
```

The value is sent as the reserved `custd.environment` label (writing that label
through `Labels` is rejected). It must be lowercase letters, digits and hyphens,
at most 32 characters; `unclassified` is reserved. A tenant may restrict a
credential to a set of environments, in which case a declaration outside that set
is rejected by the API.

## Dev smoke test (Hydra)

Requires the dev stack running with Hydra using JWT access tokens and ingest-api configured with `AUTH_JWKS_URL`.

```bash
cd sdk-go
go run ./cmd/smoke-dev
```

To run all SDK checks, use this from the repository root:

```bash
mise exec -- just check
```

## Producer setup CLI

Create the tenant-bound OAuth2 producer client and print env snippets:

```bash
go run github.com/haakco/custd-sdk-go/v2/cmd/custd-sdk-setup@latest \
  --base-url=https://custd.com \
  --admin-url=https://custd.com \
  --admin-token="$CUSTD_ADMIN_TOKEN" \
  --token-url=https://auth.custd.com/oauth2/token \
  --tenant=tracklab \
  --company-name="TrackLab" \
  --client-id=tl-custd-bridge \
  --scope=events.write \
  --environment=production \
  --env-prefix=TL_CUSTD_BRIDGE
```

The helper uses `CustdClient.Admin.Tenants` and
`CustdClient.Admin.OAuthClients`, so key/bootstrap behavior stays in the SDK.
Pass `--register-schemas ./schemas` to register every `.json` schema file in a
directory after producer credential creation.

## Browser Site Admin Helpers

Use `CustdClient.Admin.Sites` to create, list, get, delete, and rotate browser
tracker Sites. `Create` returns the public write key once. `List` and `Get`
return `AdminSite` metadata without the write key. `RotateWriteKey` returns the
replacement write key once; update browser tracker config and stop using the old
key after rotation.

## Schema Admin Helpers

Use `CustdClient.Admin.Schemas` from setup code:

Supported feature parity and intentionally missing helpers are documented in the SDK
root README.

```go
_, err := client.Admin.Schemas.Register(ctx, custd.AdminSchemaRegister{
    EventTypeSlug: "courib.delivery.created",
    Version:       "1.0.0",
    JSONSchema:    map[string]any{"type": "object"},
})
```

## Lifecycle administration

The Go SDK exposes typed admin clients for the five lifecycle
namespaces. Every method takes `context.Context` and returns server
responses verbatim. Forward-only: no deprecated aliases.

```go
admin := client.Admin()

// Tenant storage: list/create/get/revoke.
loc, err := admin.TenantStorage.Create(ctx, custd.TenantStorageCreateRequest{
    TenantSlug:    "acme",
    ClientLocation: "s3://acme-prod-warehouse/events/",
})

// Subject exports: full request lifecycle.
exp, err := admin.SubjectExports.Create(ctx, custd.SubjectExportCreateRequest{
    TenantSlug:     "acme",
    Subject:        custd.Subject{Type: "userUuid", Value: "01J5..."},
    Scope:          "portability",
    IdempotencyKey: "acme-2026-07-31",
})

// Physical erasures: NO cancel/retry — server has none.
force, err := admin.PrivacyErasures.Force(ctx, "pe_01J5...")

// Retention policies: list/get/upsert/delete + preview/apply/listRuns.
runs, err := admin.Retention.ListRuns(ctx, "acme")

// Offboarding: full request lifecycle + schedules.
sched, err := admin.Offboarding.Schedule(ctx, custd.OffboardingScheduleRequest{
    TenantSlug: "acme",
    ExecuteAt:  "2026-12-31T00:00:00Z",
    Reason:     "contract_end",
})
```

## Time-plan administration

`client.Admin.TimePlans` exposes typed `TimePlan*` request and response
structs for plan drafts/revisions, previews, publication and retirement, runs,
commands, history, and annotation correction/redaction. Calls are tenant-scoped
through the required company slug; allocation and command results come from the
server.

```go
definition := custd.TimePlanDefinition{
    HorizonMS: 60_000,
    Blocks: []custd.TimePlanBlock{{
        UUID: "block-1", SemanticKey: "focus", Title: "Focus", Basis: "absolute",
    }},
}

preview, err := client.Admin.TimePlans.Preview(ctx, "acme", definition)
plan, err := client.Admin.TimePlans.Create(ctx, "acme", custd.TimePlanDraftRequest{
    PlanKey: "focus", Name: "Focus", Definition: definition,
})
```

The typed clients are available in `v1.8.25` and later.

## Usage reporting

`client.Admin.Usage` reads the attributed usage for the authenticated tenant
through `GET /api/v1/admin/usage/me`. The tenant comes from the credential, so
the call never names a company slug. The system-admin `/usage` and `/usage/export`
surfaces are deliberately not exposed.

```go
report, err := client.Admin.Usage.Get(ctx, custd.UsageQuery{
    MeterSlug: "events.ingested",
    Start:     time.Now().UTC().AddDate(0, 0, -30),
    End:       time.Now().UTC(),
    Limit:     200,
})
```

`UsageReport.Totals` carries the per-meter totals and `UsageReport.Rows` the
per-window detail. `ContainsProvisional` and `ContainsIncomplete` are the
server's own assessment, so a caller deciding whether a number is settled reads
them rather than assuming every row is final. An omitted `Start`/`End` uses the
service default (the trailing 30 days) and an omitted `Limit` uses
`custd.UsageDefaultLimit`; a limit outside `1..custd.UsageMaxLimit` is rejected
before a request is sent.

A runnable example lives at `sdk-go/examples/usage-report`.

## Auth-project admin helpers

`client.Admin.AuthProjects` manages Custd projects, their environments, and the
application principals a directory holds, through `/api/v1/admin/auth-projects`:

- `CreateProject` / `ListProjects` (pass the previous `NextAfter` as the `after` cursor)
- `CreateEnvironment`
- `ListPrincipalSessions`, `RevokePrincipalSession`, `RevokePrincipalSessions`
- `RevokePrincipalMembership`

Every call takes a `custd.AuthProjectRequestOptions`. A machine credential must
name the platform user it acts for in `OwningUserUUID`, sent as
`X-Custd-Owning-User-UUID`; Custd validates the named user as a live member of
the machine's own company. A human administrator's own token subject is the actor
and leaves it empty. The operations Custd makes retry-safe (`CreateProject`,
`RevokePrincipalSession`, `RevokePrincipalSessions`) require `IdempotencyKey` and
reject an empty value before a request is sent.

```go
creation, err := client.Admin.AuthProjects.CreateProject(ctx, custd.AuthProjectCreateRequest{
    Slug:            "hosting-eu",
    Name:            "Hosting EU",
    EnvironmentSlug: "production",
    IdentityMode:    custd.AuthProjectIdentityIsolated,
}, custd.AuthProjectRequestOptions{
    OwningUserUUID: owningUserUUID,
    IdempotencyKey: "create-hosting-eu",
})
```

Responses are typed: `AuthProjectCreation`, `AuthProjectListResponse`,
`ApplicationSessionInventory`, `ApplicationSessionRevocation` and
`ApplicationMembershipRevocation`.

## Analytics range query

`client.Analytics.QueryRange` reads a tenant's own events across an inclusive
date range of at most `custd.AnalyticsMaxRangeDays` days, with
`groupBy: "day"`. `Buckets` carries each day's `count`/`source`/`complete` and
`Rows` the capped detail. The per-day completeness is the server's assessment
and is surfaced unchanged.

```go
response, err := client.Analytics.QueryRange(ctx, custd.AnalyticsEventRangeQueryRequest{
    From:      "2026-02-23",
    To:        "2026-05-23",
    EventType: "page-view",
    Limit:     10000,
    Source:    custd.AnalyticsRangeSourceAuto,
    GroupBy:   custd.AnalyticsRangeGroupByDay,
})
```

SDKs never log signed URLs, raw personal data, export bytes, or
subject identifiers outside opaque IDs.
