# API Conventions Specification

## Purpose
Document the patterns and conventions for creating new modules, DTOs, validation, pagination, filters, soft deletes, tenant scope, permissions, and route registration in NexoKit.

## Requirements

### Requirement: Module creation guide

The system MUST provide documentation explaining how to create a new module: files to create (`handler.go`, `service.go`, `repository.go`, `dto.go`, `model.go`, `routes.go`, `validation.go`), where to place them (`internal/modules/{name}/`), and how to register routes in `server/router.go`.

#### Scenario: Developer follows guide to create module

- GIVEN a developer reads the conventions doc
- WHEN they create a new `products` module
- THEN they produce all seven files and a `Register` function matching the documented pattern

### Requirement: DTO and validation guide

The system MUST document the DTO pattern (request/response structs), the `Validate()` method convention returning `response.ValidationErrors`, and how to use `Field().Required().Apply(Rule)` chains.

The documentation MUST name the base response DTOs explicitly: `APIResponse`, `ErrorResponse`, `ValidationErrorResponse`, `PaginatedResponse`, and `PaginationMeta`.

#### Scenario: Developer adds DTO with validation

- GIVEN the conventions doc
- WHEN a developer creates a `CreateProductRequest` with validated fields
- THEN the DTO includes `Validate()` and the handler calls `RespondIfInvalid`

### Requirement: Pagination and filters guide

The system MUST document how to use `ListFromGin`, `ApplyPagination`, `ApplySorting`, `ApplySearch`, date/status filters, and `PaginatedWithFilters` in a handler.

#### Scenario: Developer implements paginated list endpoint

- GIVEN the conventions doc
- WHEN a developer implements `GET /api/v1/products` with pagination, filters, and search
- THEN the handler uses `ListFromGin`, GORM helpers, and `PaginatedWithFilters`

### Requirement: Error handling guide

The system MUST document the `HandleError(c, err)` pattern for centralized error→response mapping, replacing manual `switch apperror.Status(err)` blocks.

#### Scenario: Developer replaces manual error mapping

- GIVEN a handler with manual status switch
- WHEN the developer reads the guide and refactors
- THEN the handler uses `HandleError(c, err)` in place of the switch block

### Requirement: Tenant scope and permissions guide

The system MUST document how to apply `ApplyTenantScope(db, ctx)` and protect routes with `RequirePermission(slug)`.

#### Scenario: Developer adds tenant-scoped endpoint

- GIVEN the conventions doc
- WHEN a developer adds a tenant-scoped list endpoint
- THEN the handler calls `tenant.FromGin(c)` and the repository applies tenant filtering

### Requirement: Soft delete conventions guide

The system MUST document that models embedding `BaseModel` or `BaseModelSimple` use `gorm.DeletedAt`, normal GORM queries exclude soft-deleted rows, delete endpoints SHOULD soft-delete, and hard delete/unscoped reads require an explicit documented exception.

#### Scenario: Developer implements delete endpoint

- GIVEN the conventions doc
- WHEN a developer adds `DELETE /api/v1/products/:id`
- THEN the repository uses GORM soft delete behavior and does not expose `DeletedAt` in API DTOs
