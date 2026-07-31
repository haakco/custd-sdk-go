# Lifecycle contract fixtures

Shared request/response/error fixtures consumed by every supported SDK
(`sdk-go`, `sdk-js`, `sdk-python`, `sdk-php`) and the `clean-consumer`
gate. The fixtures describe the **current** server contract under
`custd` `feat/client-data-lifecycle-d3-d5` @ `35e6a3e1`; the SDK plan
forbids inventing endpoints ahead of the server contract.

## Namespaces

| Namespace              | Endpoints (server source of truth)                                                                                                                                                                                                                                                                |
| ---------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `tenant-storage/`      | `GET /api/v1/tenant-storage-locations`, `POST /api/v1/tenant-storage-locations`, `GET /api/v1/tenant-storage-locations/{id}`, `DELETE /api/v1/tenant-storage-locations/{id}`                                                                                                                          |
| `subject-exports/`     | `POST /api/v1/admin/subject-exports`, `GET /api/v1/admin/subject-exports`, `GET /api/v1/admin/subject-exports/{requestId}`, `POST /api/v1/admin/subject-exports/{requestId}/cancel`, `GET /api/v1/admin/subject-exports/{requestId}/download`, `POST /api/v1/admin/subject-exports/{requestId}/force` |
| `privacy-erasures/`    | `POST /api/v1/admin/privacy/erasures`, `GET /api/v1/admin/privacy/erasures`, `GET /api/v1/admin/privacy/erasures/{requestUuid}`, `POST /api/v1/admin/privacy/erasures/{requestUuid}/force`                                                                                                          |
| `retention/`           | `GET /api/v1/admin/retention/policies`, `GET /api/v1/admin/retention/policies/{tenantSlug}`, `PUT /api/v1/admin/retention/policies/{tenantSlug}`, `DELETE /api/v1/admin/retention/policies/{tenantSlug}`, `POST /api/v1/admin/retention/policies/{tenantSlug}/preview`, `POST /api/v1/admin/retention/policies/{tenantSlug}/apply`, `GET /api/v1/admin/retention/policies/{tenantSlug}/runs` |
| `offboarding/`         | `POST /api/v1/admin/offboarding`, `GET /api/v1/admin/offboarding/{requestUuid}`, `POST /api/v1/admin/offboarding/{requestUuid}/cancel`, `POST /api/v1/admin/offboarding/{requestUuid}/confirm`, `POST /api/v1/admin/offboarding/requests/{requestUuid}/preview`, `POST /api/v1/admin/offboarding/requests/{requestUuid}/export`, `GET /api/v1/admin/offboarding/requests/{requestUuid}/download`, `POST /api/v1/admin/offboarding/requests/{requestUuid}/acknowledge`, `POST /api/v1/admin/offboarding/requests/{requestUuid}/execute`, `POST /api/v1/admin/offboarding/requests/{requestUuid}/retry`, `GET /api/v1/admin/offboarding/requests/{requestUuid}/receipt`, `POST /api/v1/admin/offboarding/schedules`, `GET /api/v1/admin/offboarding/schedules/{tenantSlug}`, `POST /api/v1/admin/offboarding/schedules/{tenantSlug}/cancel`, `GET /api/v1/admin/offboarding/schedules` |

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