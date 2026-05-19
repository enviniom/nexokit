# Delta for Middleware Auth

## MODIFIED Requirements

### Requirement: User lookup and context injection

The middleware MUST extract the `sub` claim from the token, lookup the corresponding user by `public_id`, and inject the user into the request context. The injected user MUST NOT contain `password` or `password_hash`. The injected user MUST include a `Permissions` field of type `[]string` populated by the `PermissionResolver`. If the user's role is `root`, the `Permissions` field MUST contain a designated root marker (e.g., `["*"]`) or be populated with all system permissions. If resolution fails, the middleware MUST proceed with an empty `Permissions` field and log a warning — it MUST NOT reject the request solely due to a permission resolution failure.
(Previously: Context injection did not include a Permissions field)

#### Scenario: User injected into context with permissions

- GIVEN a valid token for an active user with role `admin`
- WHEN the request reaches a protected handler
- THEN the handler can retrieve the user object from context without password fields
- AND `authctx.User.Permissions` contains the permission slugs for the admin role

#### Scenario: Root user gets full permissions in context

- GIVEN a valid token for a root user
- WHEN the request reaches a protected handler
- THEN `authctx.User.Permissions` contains all system permission slugs or the root marker `"*"`

#### Scenario: Permission resolution failure degrades gracefully

- GIVEN a valid token for an active user
- WHEN the PermissionResolver fails (e.g., cache and database both unavailable)
- THEN the handler can still retrieve the user object from context
- AND `authctx.User.Permissions` is an empty array
- AND a warning is logged