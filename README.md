# custd SDK (Go)

Ingestion client with retry, batching, and queueing.

## Compatibility

Version `1.0.0` targets the canonical ingest endpoint
`POST /api/v1/events`. The legacy `POST /v1/events` path is not supported.
This SDK was not released against the legacy path, so there is no compatibility
alias or deprecation window.

## Install

Consume the Go module from its dedicated mirror `github.com/haakco/custd-sdk-go`
(tagged `vX.Y.Z`):

```bash
go get github.com/haakco/custd-sdk-go@latest
```

```go
import custd "github.com/haakco/custd-sdk-go"
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
go run github.com/haakco/custd-sdk-go/cmd/custd-sdk-setup@latest \
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
