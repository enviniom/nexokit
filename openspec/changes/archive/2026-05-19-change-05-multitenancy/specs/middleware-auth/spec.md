# Delta for Middleware Auth

## MODIFIED Requirements

### Requirement: User lookup and context injection

The middleware MUST extract the `sub` claim from the token, lookup the corresponding user by `public_id`, and inject the user into the request context. The injected user MUST NOT contain `password` or `password_hash`. The injected user MUST include a `Permissions` field of type `[]string` populated by the `PermissionResolver`. If the user's role is `root`, the `Permissions` field MUST contain a designated root marker (e.g., `["*"]`) or be populated with all system permissions. If resolution fails, the middleware MUST proceed with an empty `Permissions` field and log a warning — it MUST NOT reject the request solely due to a permission resolution failure. The middleware MUST also set `TenantContext` into the Gin context, derived from the authenticated user: if the user's `company_id` is non-null, `TenantContext.CompanyID` and `TenantContext.CompanySlug` are populated with `IsRootScope = false`; if the user is root with null `company_id`, `TenantContext.IsRootScope` is `true`. For root users, if a valid `X-Company-ID` header is present, `TenantContext` MUST be scoped to that company with `IsRootScope = false`.
(Previously: Context injection only set user and permissions; no TenantContext was set.)

#### Scenario: User injected into context with permissions and tenant

- GIVEN a valid token for an active admin user with `company_id = 5`
- WHEN the request reaches a protected handler
- THEN the handler can retrieve the user object from context without password fields
- AND `authctx.User.Permissions` contains the permission slugs for the admin role
- AND Gin context contains `TenantContext{CompanyID: 5, IsRootScope: false}`

#### Scenario: Root user gets full permissions and global tenant scope

- GIVEN a valid token for a root user with null `company_id`
- WHEN the request reaches a protected handler
- THEN `authctx.User.Permissions` contains all system permission slugs or the root marker `"*"`
- AND Gin context contains `TenantContext{IsRootScope: true, CompanyID: 0}`

#### Scenario: Root user with X-Company-ID header gets scoped tenant

- GIVEN a valid token for a root user with null `company_id` and header `X-Company-ID: 7`
- WHEN the request reaches a protected handler
- THEN Gin context contains `TenantContext{CompanyID: 7, IsRootScope: false}`
- AND `authctx.User.Permissions` still contains all system permissions

#### Scenario: Permission resolution failure degrades gracefully

- GIVEN a valid token for an active user
- WHEN the PermissionResolver fails (e.g., cache and database both unavailable)
- THEN the handler can still retrieve the user object from context
- AND `authctx.User.Permissions` is an empty array
- AND a warning is logged

#### Scenario: Admin user with company_id gets scoped tenant context

- GIVEN a valid token for an active admin user with `company_id = 3`
- WHEN the request reaches a protected handler
- THEN Gin context contains `TenantContext{CompanyID: 3, IsRootScope: false}`
- AND `TenantContext.CompanySlug` matches the company's slug value