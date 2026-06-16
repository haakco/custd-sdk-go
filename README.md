# custd SDK (Go)

Ingestion client with retry, batching, and queueing.

## Compatibility

Version `1.0.0` targets the canonical ingest endpoint
`POST /api/v1/events`. The legacy `POST /v1/events` path is not supported.
This SDK was not released against the legacy path, so there is no compatibility
alias or deprecation window.

## Install

Consume the Go module from the public monorepo subdir (its `go.mod` module path
is `github.com/haakco/custd-sdk/sdk-go`, tagged `sdk-go/vX.Y.Z`):

```bash
go get github.com/haakco/custd-sdk/sdk-go@latest
```

```go
import custd "github.com/haakco/custd-sdk/sdk-go"
```

> The `haakco/custd-sdk-go` mirror exists but is not yet `go get`-able under that
> name — its `go.mod` still declares the monorepo subdir path. Use the import
> above until the mirror's module path is renamed.

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
go run github.com/haakco/custd-sdk/sdk-go/cmd/custd-sdk-setup@latest \
  --base-url=https://custd.k8.haak.co \
  --admin-url=https://custd.k8.haak.co \
  --admin-token="$CUSTD_ADMIN_TOKEN" \
  --token-url=https://auth.k8.haak.co/oauth2/token \
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

```go
_, err := client.Admin.Schemas.Register(ctx, custd.AdminSchemaRegister{
    EventTypeSlug: "courib.delivery.created",
    Version:       "1.0.0",
    JSONSchema:    map[string]any{"type": "object"},
})
```
