# Lifecycle contract fixtures

Shared request/response/error fixtures consumed by every supported SDK
(`sdk-go`, `sdk-js`, `sdk-python`, `sdk-php`) and the `clean-consumer`
gate. The fixtures describe the **current** server contract under
`custd` `feat/client-data-lifecycle-d3-d5` @ `35e6a3e1`; the SDK plan
forbids inventing endpoints ahead of the server contract.

## Namespaces

- `tenant-storage/` — location create, list, get, and delete endpoints.
- `subject-exports/` — request create, list, get, cancel, download, and force
  endpoints.
- `privacy-erasures/` — erasure create, list, get, and force endpoints.
- `retention/` — policy list, get, upsert, delete, preview, apply, and run-list
  endpoints.
- `offboarding/` — request lifecycle, preview, export, download, receipt,
  execution, retry, and schedule endpoints.

## Filename convention

- `valid-*.json` — happy-path request or response the SDK must accept.
- `invalid-*.json` — error response the SDK must surface without leaking
  internal details.
- `isolation-*.json` — proves cross-tenant boundary (e.g. other tenant
  returns an empty list or a mismatch error).
- `expired-*.json` — terminal-state response the SDK must not retry.
- `legal-hold-*.json` — proves legal-hold rows are preserved, not deleted.
- `partial-*.json` — partial completion the SDK must surface verbatim.

## Cross-language matrix

Every SDK loads the same fixtures and asserts the same set of
fields/values before any network implementation. See
`tests/matrix/lifecycle_matrix.json` for the per-namespace assertion
list shared by Go, JS, Python, and PHP test suites.

## Forward-only rule

If the server removes an endpoint, the SDK removes the matching
method in the same release. If the server adds an endpoint, the SDK
adds the matching method in the same release. No aliases, no
dual-write fallbacks, no compatibility wrappers. Per CLAUDE.md
pre-live rule.
