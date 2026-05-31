# custd SDK (Go)

Ingestion client with retry, batching, and queueing.

## Compatibility

Version `1.0.0` targets the canonical ingest endpoint
`POST /api/v1/events`. The legacy `POST /v1/events` path is not supported.
This SDK was not released against the legacy path, so there is no compatibility
alias or deprecation window.

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
