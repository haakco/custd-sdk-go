# custd SDK (Go)

Ingestion client with retry, batching, and queueing.

## Compatibility

Version `1.0.0` targets the canonical ingest endpoint
`POST /api/v1/events`. The legacy `POST /v1/events` path is not supported.
This SDK was not released against the legacy path, so there is no compatibility
alias or deprecation window.

## Install

Consume the Go module from its dedicated mirror `github.com/haakco/custd-sdk-go/v3`
(tagged `vX.Y.Z`):

```bash
go get github.com/haakco/custd-sdk-go/v3@latest
```

```go
import custd "github.com/haakco/custd-sdk-go/v3"
```

> This module is developed in the [`custd-sdk`](https://github.com/haakco/custd-sdk)
> monorepo under `sdk-go/` and published to the read-only `custd-sdk-go` mirror on
> each release. Import the mirror path above, not the monorepo subdir.

## Migrating to `v3`

The module path is major-versioned. The current line is
`github.com/haakco/custd-sdk-go/v3`; the previous line served `.../v2`, and the original unversioned
`github.com/haakco/custd-sdk-go` path was frozen at v1 and receives no further releases.

To move from `v2` to `v3`:

```bash
# 1. Update import paths in your code
#    github.com/haakco/custd-sdk-go/v2  ->  github.com/haakco/custd-sdk-go/v3
# 2. Pull the new major
#    go get github.com/haakco/custd-sdk-go/v3@v3.0.0
# 3. Drop the now-unused old module
#    go mod tidy
```

The package name (`custd`) is unchanged, so the move is a mechanical find-and-replace. v3 also reshapes
the reporting-pack types to the canonical contract; the breaking list is in the release changelog under
`docs/changelog/`.

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

## Workflow timing imports

With a client configured for the Admin API and an authorized tenant credential:

```go
result, err := client.Admin.WorkflowTimings.AppendBatch(ctx, "acme", observations)
if err != nil {
    return err
}
return result.RequireAccepted()
```

Pass the complete import, including histories larger than 512 facts. The SDK
splits requests at 10,000 observations or 1 MiB of encoded JSON, whichever comes
first, and keeps result indexes relative to the original input. Each chunk
commits independently. If a request fails, replay with the same idempotency keys;
do not regenerate keys or discard already-observed history.

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
go run github.com/haakco/custd-sdk-go/v3/cmd/custd-sdk-setup@latest \
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

## Site Domain Admin Helpers

Use `CustdClient.Admin.SiteDomains` to claim a host for a Site (`Create`),
prove control of it (`Verify`), ask for an audit (`RequestAudit`), read its
findings (`ListFindings`), and remove it (`Delete`). `Create` returns the DNS
TXT record and well-known file instructions. Auditing an unverified domain
returns a `*RequestError` whose `Problem.Code` is `SiteDomainNotVerifiedCode`.
A system administrator names the tenant with `ClientConfig.ActingCompanySlug`.

`Ownership.VerifiedBy` says how control was proven: `dns`, `file`, `attested`,
`parent` (a verified parent domain, named by `VerifiedViaDomainUUID`) or
`operator`. A site-bound broker credential holding `site_domains.attest` that
has proven control itself calls `Attest` to verify the domain and its
subdomains, and `Unattest` to withdraw; `Ownership.Attestation` shows the
standing attestation. A finding's `Rule` is one of the `SiteRule*` constants;
key client-side fixes on those, never on the title.

Each domain also carries `Checks`: its uptime status, last check, certificate
expiry and warming measurements, and any open outage. `SetCheckMode` chooses
`AdminSiteCheckModeUptime` (the default), `AdminSiteCheckModeWarm` (adds one
home page GET a minute) or `AdminSiteCheckModePaused` (no requests at all; set
it when you stop a site, so nothing wakes it). `ListIncidents` returns the
domain's recent outages.

Search performance comes from the client's own Google Search Console and Bing
Webmaster Tools properties. `GetSearchAccount` says which providers this
deployment can import from and the Google service account email to add as a
(Restricted) user on the property. `ConnectSearch` connects a domain to
`AdminSiteSearchProviderGoogle` or `AdminSiteSearchProviderBing` (Bing also takes
the client's `APIKey`, which is never returned); a provider the deployment cannot
import from returns a 409 with `SiteSearchNotConfiguredCode`.
`GetSearchPerformance(ctx, domainUUID, days)` returns each connection's import
state, per-provider totals with the previous period (`Previous`) and each day
(`Daily`), the top pages and queries, and `StrikingDistance` queries averaging
position 5 to 20 (`days` 0 means 28).
`DisconnectSearch` removes a connection and its imported numbers.
`ListSearchRows` pages through one dimension (page, query, country or device);
`Unavailable` names providers that cannot report it, which is not zero traffic.
`RequestSearchInspection` queues a Search Console URL Inspection (202) and
`ListSearchInspections` returns the results and the tenant's daily use; refusals
carry the `SiteSearchInspection*Code` constants.
`GetIndexNow`, `EnableIndexNow` (returns the key to host), `VerifyIndexNowKey`,
`DisableIndexNow` and `SubmitIndexNow` (202) push changed URLs to IndexNow;
refusals carry the `SiteIndexNow*Code` constants.

Each finding carries a `Remedy`: what the fix changes (`Kind`, `Element`) and,
for some rules, `Generated` code to copy. The page tools work on one page:
`GetPageFacts` returns what the last audit read, `RequestPageRecheck` queues a
recheck (poll `GetPageRecheck`; `SitePageRecheckLimitCode` past the hourly
allowance, `SitePageNotOnDomainCode` for another host), `CheckPage` checks draft
HTML without storing it, and `TestRobots` answers whether a crawler may fetch a
URL under a pasted or the live robots.txt. `GetProjectContext` and
`SetProjectContext` read and replace a Site's brand names, competitors and
markets.

`GetReport`, `GetUptimeHistory` and `GetLighthouseHistory` read the client report
and its history; `ListReportShares`, `CreateReportShare` (the token is returned
once) and `RevokeReportShare` manage public report links. Measurements that were
not taken decode as `nil`, never 0.

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
- Hosted app releases: `CreateAppRelease` (manifest files and routing),
  `UploadAppReleaseFile` (one file's raw bytes, sent as `application/octet-stream`),
  `VerifyAppRelease`, `PromoteAppRelease` and `ListAppReleases`

An environment is served as a hosted app by the optional `Hosting` field of
`AuthProjectDesiredState`; a nil value keeps the stored section.
`AuthProjectStatus.Hosting` reports the derived host, origin and, with the
`sign_in` capability, the sign-in client.

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

## Native error capture

Error capture sends only exception types and bounded code filenames/line numbers
straight to Custd. Messages, arguments, absolute paths, users, request bodies and
headers are excluded before transport. Keep local operational logs; optional
capture is never the only copy of an operational failure.

```go
reporter, err := custd.NewErrorReporter(custd.ErrorCaptureConfig{
    BaseURL:     "https://custd.example.com",
    WriteKey:    os.Getenv("CUSTD_ERROR_WRITE_KEY"),
    Environment: "production",
    Release:     "vorrent@1.0.0",
})
if err == nil {
    // Admission is nonblocking. True means admitted, not delivered.
    reporter.Report(ctx, custd.CaptureExceptionRequest{
        Error: operationErr, Mode: custd.ErrorCaptureHandled,
    })
    // At shutdown use a bounded context; Close refuses new reports and drains sends.
    _ = reporter.Close(shutdownCtx)
}
// Invalid/missing configuration leaves capture off; record a safe warning locally.
```

Own one reporter per process and Site. It snapshots the calling stack before its
asynchronous handoff, admits at most four in-flight sends and has no pending queue,
persistence or retries. Each admitted action uses a fresh request-scoped capture
client, so a long-running process does not stop reporting after twenty errors.
`Stats()` returns accepted, rejected, unavailable, limited and in-flight counts;
these are delivery outcomes, not raw event data. Shutdown cancellation and a
two-second transport timeout bound sends. The supplied HTTP client is cloned,
with redirects and cookie jars disabled, so the caller's client is not changed.

For request-scoped synchronous capture use `NewErrorCaptureClient(config)` and
`CaptureException(ctx, request)`. Its typed result distinguishes `accepted`,
`rejected`, `unavailable` and `limited`; each client permits twenty captures and
four concurrent sends. Envelopes are capped at 16 KiB and 24 frames. Capture never
reads `Error.Error()` or panic string content. `SnapshotError(request)` freezes a
safe Go event at the call site when another component owns the handoff; set its
environment/release before `CaptureEvent`.

Mobile apps must send the SDK's `ErrorCaptureEvent` DTO to their own authenticated
application relay, never contain a server Site key. `CaptureEvent` and
`ReportEvent` revalidate that DTO and construct a new envelope with the relay's
Site key. Reject unknown JSON fields at the application HTTP boundary and enforce
its authentication and request-size limits. The DTO contains only optional
32-hex `eventId`, required `platform` (`javascript` or `go`), `exceptionType`,
`frames` (`filename`, `lineno`, optional `colno`), `mode` (`handled` or `unhandled`),
`environment`, optional `release` and optional application ULID `correlationId`.
The client generates an event ID when absent. Server and browser Sites use
different keys; never publish an originless server key in browser/mobile config.

For an authenticated application error relay, use
`reporter.CaptureEvent(ctx, event)` to return a **confirmed** collector result to
the mobile client. It shares the same four-send ceiling as asynchronous reports
and returns `limited` immediately when busy. Never acknowledge a queued/admitted
report as delivered. The separate diagnostic HTTP request may wait up to the
two-second send limit; the mobile UI operation must not wait for diagnostics.

## Confirmed event batches for application relays

An authenticated mobile analytics relay needs confirmed **per-event** acceptance,
not a successful enqueue or an HTTP 202 alone:

```go
result, err := client.SendEvents(ctx, custd.SendEventsRequest{Events: events})
if err != nil {
    // Acknowledge nothing: the outcome may be unknown.
    return err
}
acceptedIDs := result.AcceptedEventIDs
```

`SendEvents` sends once through the SDK's existing OAuth/authentication and gzip
transport, without the local queue or automatic retry. It accepts 1–1,000 events
with caller-supplied stable event UUIDs and the normal producer envelope fields.
The application relay supplies its authorized tenant and validates/allowlists
mobile data before sending. Do not regenerate a mobile queue's event UUIDs.

A partial batch response is returned with `Results` (typed `EventIngestResult`
values including each rejection's problem) and `AcceptedEventIDs`, in submission
order. Only explicitly successful 2xx results are acknowledged. Missing,
duplicate, unknown or inconsistent result IDs/statuses produce an error and no
acknowledgements. A transport failure can mean acceptance is unknown; keep the
original IDs for caller-controlled retry. A collector-supplied
`acceptedEventIds` field is never trusted. `Track` and `Flush` retain their
existing asynchronous queue contract and should not be used to acknowledge
mobile queue entries.

Caller deadlines also bound OAuth token acquisition and refresh. Valid producer
tokens are reused, and a slow refresh never holds another request behind a cache
mutex. Token endpoint redirects are refused.
