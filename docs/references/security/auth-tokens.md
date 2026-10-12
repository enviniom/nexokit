# Access and refresh tokens

IAM's internal authentication slice (`internal/modules/iam/slices/auth`) issues PASETO v4.local access tokens and database-backed opaque refresh tokens. [Request flow](../runtime/request-flow.md) owns route/middleware ordering.

## Access claims and key configuration

`platform/token.Manager` issues subject (user public ID), role (currently role name), optional **numeric** company ID, `token_type=access`, issued-at and expiration. Default access TTL is 15 minutes (`ACCESS_TTL_MINUTES`). Parsing requires a valid v4.local token and access type; middleware resolves the current user/role/permissions from IAM rather than relying only on claims.

`NewManager` SHA-256 hashes the supplied `PASETO_KEY` string into a 32-byte encryption key. Neither config load nor manager construction rejects an empty/weak secret. Provision a strong random secret as [deployment policy](../../deployment.md#configuration-and-secrets); do not describe that policy as boot-time enforcement. Rotating the key invalidates access tokens encrypted with the old key; no multi-key acceptance/rotation window is implemented.

## Refresh lifecycle

Refresh values are 32 random bytes encoded as hex; only their SHA-256 digests are persisted in `refresh_tokens`. Default expiration is seven days (`REFRESH_TTL_DAYS`).

| Operation | Current behavior |
|---|---|
| Login | Reject inactive user/invalid password; create refresh record; return token pair and user DTO |
| Refresh | Lookup digest; reject revoked/expired token or inactive user; issue pair, create new record, revoke old with replacement hash |
| Logout | Revoke the supplied refresh token |
| Authenticated request | Validate access token and resolve a currently active actor through IAM |

Revoked-token reuse is rejected, but there is no token-family revocation response or explicit compromise event. Refresh creation and old-token revocation are separate repository calls, not an atomic compare-and-swap; do not promise race-proof single-use rotation.

## Session safety limits

IAM password change currently updates the password hash only; it does **not** revoke every active refresh token. Logout revokes its refresh token, not stateless access tokens. Access tokens normally remain valid until expiry/key rotation, subject to middleware's current-user activity checks. Login/refresh do not universally check company active/deleted state.

Never log tokens or password hashes; HTTP error/logging policy belongs in [API validation and errors](../../api_validation_and_errors.md#6-response-safety-and-logging). Reassess revocation/concurrency requirements before treating this base as a complete session-security policy.

Sources: `internal/platform/token/token.go`; `authenticate_user`, `rotate_token`, `revoke_token` under `internal/modules/iam/slices/auth`; `internal/middleware/auth.go`; `internal/modules/iam/slices/users/change_user_password/repository.go`.
