# Nexokit — Architecture Reference

This document defines the folder structure, layer responsibilities, import rules, and naming conventions for the Nexokit API. Every decision here has a reason; treat exceptions as explicit design decisions that must be documented.

---

## Folder overview

```
api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── domain/
│   ├── platform/
│   │   ├── contracts/
│   │   ├── response/
│   │   ├── apperror/
│   │   ├── database/
│   │   ├── tenant/
│   │   └── logger/
│   ├── modules/
│   │   ├── catalog/
│   │   │   ├── core/
│   │   │   ├── queries/
│   │   │   ├── slices/
│   │   │   ├── container.go
│   │   │   └── routes.go
│   │   ├── campaigns/
│   │   └── sales/
│   └── app/
│       └── container.go
└── migrations/
```

---

## Layer rules

### Import direction (non-negotiable)

```
cmd/  →  app/  →  modules/  →  platform/  →  domain/
                                    ↑
                               (contracts only,
                                no circular refs)
```

| Layer | Can import |
|---|---|
| `domain/` | Nothing internal |
| `platform/` | `domain/` |
| `modules/<m>/core/` | `domain/`, `platform/contracts/`, `platform/apperror/` |
| `modules/<m>/slices/` | `domain/`, `platform/`, `modules/<m>/core/`, `modules/<m>/queries/` |
| `app/container` | All modules, all platform |
| `cmd/` | `app/` only |

**Modules never import each other.** The only place where two modules coexist is `app/container.go`.

---

## `domain/`

**Purpose:** Canonical domain models. Plain Go structs with no GORM tags, no HTTP concerns, no behavior tied to any specific module. Every module that needs to read or write an entity uses these types.

**Rules:**
- No GORM tags — those belong in repository-level persistence records.
- No methods that import infrastructure packages.
- No DTOs, no response shapes, no validation tags.
- One file per aggregate or closely related group of entities.

**What goes here:** Structs that represent the core business entities of the system.

**What does NOT go here:** DTOs, response types, mappers, interfaces, GORM models, validation logic.

```go
// internal/domain/product.go
package domain

import "github.com/shopspring/decimal"

type Product struct {
    ID          uint
    Name        string
    Slug        string
    SKU         string
    Price       decimal.Decimal
    CategoryID  uint
    Description string
    ImageURL    string
    IsActive    bool
}

type Category struct {
    ID       uint
    Name     string
    Slug     string
    ParentID *uint
}
```

```go
// internal/domain/campaign.go
package domain

import (
    "time"
    "github.com/shopspring/decimal"
)

type Promotion struct {
    ID            uint
    Name          string
    DiscountType  string
    DiscountValue decimal.Decimal
    TargetType    string // "product" | "category" | "cart"
    TargetID      *uint
    StartsAt      time.Time
    EndsAt        time.Time
}

type Coupon struct {
    ID            uint
    Code          string
    DiscountType  string
    DiscountValue decimal.Decimal
    MinCartValue  decimal.Decimal
    UsageLimit    int
    UsedCount     int
    ExpiresAt     time.Time
}
```

---

## `platform/`

Platform contains shared infrastructure used by all modules. It can import `domain/`. Modules import from platform freely. Platform never imports from modules.

### `platform/apperror/`

**Purpose:** Sentinel errors, `AppError` wrapper, HTTP status resolution, and public message extraction. This is the single source of truth for error classification across the entire application.

**Rules:**
- Modules use `apperror.Wrap` in their `core/errors.go` to create domain-specific errors wrapping the appropriate sentinel.
- Handlers import `apperror` to call `HandleError`, `Status`, `PublicMessage`, and `Log`.
- Services and repositories must NOT import this package — they return domain errors from `core/`.
- Sentinel messages must be generic. Specific business messages belong in the `Wrap` call inside `core/`.

```go
// internal/platform/apperror/apperror.go
package apperror

import (
    "errors"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/your-org/nexokit/internal/platform/messages"
)

type AppError struct {
    Err     error  // sentinel (ErrNotFound, ErrConflict, etc.)
    Message string // public message shown to the client
    Cause   error  // original infrastructure error (for logs only)
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

// Sentinel errors — messages must be generic.
var (
    ErrNotFound         = &AppError{Message: messages.MsgNotFound}
    ErrForbidden        = &AppError{Message: messages.MsgForbidden}
    ErrUnauthorized     = &AppError{Message: messages.MsgUnauthorized}
    ErrConflict         = &AppError{Message: messages.MsgConflict}
    ErrBadRequest       = &AppError{Message: messages.MsgBadRequest}
    ErrTooManyRequests  = &AppError{Message: messages.MsgTooManyRequests}
    ErrValidation       = &AppError{Message: messages.MsgValidationError}
    ErrUnprocessable    = &AppError{Message: messages.MsgUnprocessable}
    ErrInternal         = &AppError{Message: messages.MsgInternalError}
)

// Wrap creates a domain-specific error wrapping a sentinel.
// Use this in modules/<m>/core/errors.go.
//
//   ErrCartNotFound = apperror.Wrap(apperror.ErrNotFound, "Carrito no encontrado")
//
// The optional cause is the original infrastructure error and is only used for logging.
func Wrap(sentinel error, message string, cause ...error) *AppError {
    var c error
    if len(cause) > 0 {
        c = cause[0]
    }
    return &AppError{Err: sentinel, Message: message, Cause: c}
}

// Status resolves the HTTP status code from any error.
// Returns 500 for unknown/unwrapped errors.
func Status(err error) int {
    var e *AppError
    if !errors.As(err, &e) {
        return http.StatusInternalServerError
    }
    switch {
    case errors.Is(e.Err, ErrNotFound):        return http.StatusNotFound
    case errors.Is(e.Err, ErrForbidden):       return http.StatusForbidden
    case errors.Is(e.Err, ErrUnauthorized):    return http.StatusUnauthorized
    case errors.Is(e.Err, ErrConflict):        return http.StatusConflict
    case errors.Is(e.Err, ErrBadRequest):      return http.StatusBadRequest
    case errors.Is(e.Err, ErrValidation):      return http.StatusUnprocessableEntity
    case errors.Is(e.Err, ErrUnprocessable):   return http.StatusUnprocessableEntity
    case errors.Is(e.Err, ErrTooManyRequests): return http.StatusTooManyRequests
    default:                                   return http.StatusInternalServerError
    }
}

// PublicMessage returns the safe message to send to the client.
// In release mode, unexpected errors (not AppError) return a generic message.
func PublicMessage(err error, mode string) string {
    var e *AppError
    if !errors.As(err, &e) {
        if mode == gin.ReleaseMode {
            return messages.MsgInternalError
        }
        return err.Error()
    }
    return e.Message
}

// Log extracts the underlying infrastructure cause for logging.
// Returns nil if no cause was recorded (expected domain errors need no logging).
func Log(err error) error {
    var e *AppError
    if errors.As(err, &e) {
        return e.Cause
    }
    return err
}
```

**Domain errors in `core/` wrap sentinels — never the reverse:**

```go
// internal/modules/sales/core/errors.go
package core

import "github.com/your-org/nexokit/internal/platform/apperror"

var (
    ErrCartNotFound            = apperror.Wrap(apperror.ErrNotFound,       "Carrito no encontrado")
    ErrOrderNotFound           = apperror.Wrap(apperror.ErrNotFound,       "Pedido no encontrado")
    ErrInsufficientStock       = apperror.Wrap(apperror.ErrConflict,       "Stock insuficiente en bodega")
    ErrVariantNotFound         = apperror.Wrap(apperror.ErrNotFound,       "Variante de producto no encontrada")
    ErrInvalidStatusTransition = apperror.Wrap(apperror.ErrBadRequest,     "Transición de estado de pedido inválida")
)
```

```go
// internal/modules/iam/core/errors.go
package core

import "github.com/your-org/nexokit/internal/platform/apperror"

var (
    ErrRoleNotFound        = apperror.Wrap(apperror.ErrNotFound,      "Rol no encontrado")
    ErrRoleHasAssignedUsers = apperror.Wrap(apperror.ErrUnprocessable, "El rol tiene usuarios asignados")
    ErrRoleProtected       = apperror.Wrap(apperror.ErrForbidden,      "El rol es del sistema y no puede modificarse")
)
```

### `platform/response/`

**Purpose:** `APIResponse[T]` envelope and HTTP writing helpers including `HandleError`.

```go
// internal/platform/response/response.go
package response

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/your-org/nexokit/internal/platform/apperror"
)

type APIResponse[T any] struct {
    Success bool   `json:"success"`
    Data    T      `json:"data,omitempty"`
    Error   *Error `json:"error,omitempty"`
}

type Error struct {
    Message string              `json:"message"`
    Fields  map[string][]string `json:"fields,omitempty"`
}

func HandleError(c *gin.Context, err error) {
    if err == nil {
        return
    }
    status := apperror.Status(err)
    message := apperror.PublicMessage(err, gin.Mode())
    c.JSON(status, APIResponse[any]{
        Success: false,
        Error:   &Error{Message: message},
    })
}

func Created[T any](data T) APIResponse[T] {
    return APIResponse[T]{Success: true, Data: data}
}

func OK[T any](data T) APIResponse[T] {
    return APIResponse[T]{Success: true, Data: data}
}
```

### `platform/contracts/`

**Purpose:** Cross-module interfaces. When module A needs a capability owned by module B, the interface is declared here. Neither module knows about the other — they only know about the contract.

**Rules:**
- Only interfaces and the value types those interfaces require in their signatures.
- No implementation, no business logic, no GORM, no HTTP.
- One file per capability domain.

```go
// internal/platform/contracts/discount.go
package contracts

import "github.com/shopspring/decimal"

// DiscountableItem is the neutral input type for the discount engine.
// Sales maps its CartItem to this type before calling the engine.
// Campaigns implements the engine without knowing anything about sales.
type DiscountableItem struct {
    ProductID  uint
    CategoryID uint
    Quantity   int
    UnitPrice  decimal.Decimal
}

type DiscountResult struct {
    ItemDiscounts map[uint]decimal.Decimal // productID → discount amount
    CartDiscount  decimal.Decimal
    CouponApplied string
}

// DiscountEngine is implemented by campaigns, consumed by sales.
type DiscountEngine interface {
    Apply(items []DiscountableItem, couponCode string) (DiscountResult, error)
}
```

### `platform/database/`, `platform/tenant/`, `platform/logger/`

Standard infrastructure packages. Modules receive these via constructor injection; they never instantiate connections themselves.

---

## `modules/<module>/`

Every module follows the same root shape:

```
modules/<module>/
  container.go
  routes.go
  core/
  queries/
  slices/
```

### `core/`

**Purpose:** The module's shared domain language — DTOs, domain errors, constants, and small pure helpers used by more than one slice within the module.

**Rules:**
- Can import `domain/` and `platform/apperror/` (for `Wrap`).
- No persistence, no GORM, no HTTP, no response helpers.
- DTOs are the module's own view of domain models — never imported from another module's `core/`.
- Mappers from `domain.X` to a local DTO are pure functions and live here next to the DTO they produce.

**What goes here:** DTOs, domain errors, constants, mappers `domain → DTO`, pure domain helpers.

**What does NOT go here:** Shared repositories, service logic, GORM queries, HTTP handlers.

```go
// internal/modules/catalog/core/product_dto.go
package core

import (
    "github.com/shopspring/decimal"
    "github.com/your-org/nexokit/internal/domain"
)

type ProductResponse struct {
    ID          uint             `json:"id"`
    Name        string           `json:"name"`
    Slug        string           `json:"slug"`
    Price       decimal.Decimal  `json:"price"`
    ImageURL    string           `json:"image_url"`
    Category    CategoryResponse `json:"category"`
    Description string           `json:"description"`
}

// ProductToResponse is a pure mapper. No business logic, no DB, no HTTP.
func ProductToResponse(p domain.Product, c domain.Category) ProductResponse {
    return ProductResponse{
        ID:          p.ID,
        Name:        p.Name,
        Slug:        p.Slug,
        Price:       p.Price,
        ImageURL:    p.ImageURL,
        Description: p.Description,
        Category:    CategoryToResponse(c),
    }
}
```

```go
// internal/modules/sales/core/cart_dto.go
package core

import "github.com/your-org/nexokit/internal/domain"

// CartProductSummary is sales' own projection of domain.Product.
// It does not import nor depend on catalog's ProductResponse.
type CartProductSummary struct {
    ID       uint   `json:"id"`
    Name     string `json:"name"`
    ImageURL string `json:"image_url"`
}

func ProductToCartSummary(p domain.Product) CartProductSummary {
    return CartProductSummary{ID: p.ID, Name: p.Name, ImageURL: p.ImageURL}
}
```

### DTO and mapper rules

> Each module declares its own DTOs. Two modules that need data from the same domain model each declare their own projection with only the fields they need. The overlap between projections is accidental — they are independent contracts.

| What | Where |
|---|---|
| Canonical domain model | `domain/` |
| Module HTTP response DTO | `modules/<m>/core/` |
| Mapper `domain.X → ModuleDTO` | `modules/<m>/core/` next to the DTO |
| Value type for cross-module contract | `platform/contracts/` |
| DTO reused between slices of the same module | `modules/<m>/core/` |
| DTO reused between different modules | Does not exist — each module declares its own |

**Why modules do not share DTOs:** `catalog.ProductResponse` and `sales.CartProductSummary` may share fields today. The moment `catalog` adds a field for its own API clients, it must not force a change in `sales`. They are independent representations with independent evolution paths.

### `queries/`

**Purpose:** Reusable persistence queries — one file per query. A query lives here only when it is used by more than one slice within the same module.

**Rules:**
- Only when reused by 2+ slices of the same module.
- One file per query, with dedicated tests.
- Contains only GORM/DB logic — no response mappers, no business policy.
- Single-use queries stay inside the slice repository.
- Returns `domain/` types or the module's `core/` types.

```go
// internal/modules/catalog/queries/find_product_by_slug.go
package queries

import (
    "context"
    "errors"
    "gorm.io/gorm"
    "github.com/your-org/nexokit/internal/domain"
    "github.com/your-org/nexokit/internal/modules/catalog/core"
)

func FindProductBySlug(ctx context.Context, db *gorm.DB, slug string) (*domain.Product, error) {
    var record productRecord
    err := db.WithContext(ctx).Where("slug = ? AND is_active = true", slug).First(&record).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, core.ErrProductNotFound
    }
    if err != nil {
        return nil, err
    }
    p := record.toDomain()
    return &p, nil
}
```

### `slices/`

**Purpose:** Business use-case slices. One folder per use case.

**Flat modules** (single entity or ≤3 use cases per entity):

```
slices/
  list_products/
  view_product/
  create_product/
```

**Multi-entity modules** (more than one entity AND more than 3 use cases per entity):

```
slices/
  users/
    list_users/
    create_user/
  roles/
    list_roles/
    create_role/
    assign_permission/
  permissions/
    list_permissions/
    sync_permissions/
```

#### Inside each slice

Every non-trivial slice owns exactly three files:

```
create_product/
  handler.go
  service.go
  repository.go
```

---

## Boundary responsibilities

### Handler

**Owns:** Request binding, request validation short-circuit, tenant/context extraction, response writing, domain error → HTTP mapping via `apperror` + `response.HandleError`.

**Must not own:** Business decisions, GORM calls, persistence error inspection, model construction.

```go
// internal/modules/catalog/slices/create_product/handler.go
package create_product

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/your-org/nexokit/internal/platform/apperror"
    "github.com/your-org/nexokit/internal/platform/logger"
    "github.com/your-org/nexokit/internal/platform/response"
    "go.uber.org/zap"
)

type Handler struct{ svc Service }

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Handle(c *gin.Context) {
    var req Request
    if err := c.ShouldBindJSON(&req); err != nil {
        response.HandleError(c, apperror.Wrap(apperror.ErrBadRequest, err.Error()))
        return
    }

    result, err := h.svc.Execute(c.Request.Context(), req)
    if err != nil {
        // log the infrastructure cause if present — expected domain errors have no cause
        if cause := apperror.Log(err); cause != nil {
            logger.Error(c, "create_product failed", zap.Error(cause))
        }
        response.HandleError(c, err)
        return
    }

    c.JSON(http.StatusCreated, response.Created(result))
}
```

### Service

**Owns:** Use-case orchestration, business rule validation, domain model construction from the request.

**Must not own:** GORM, SQL details, `platform/apperror` imports, HTTP status decisions, request binding.

The service is the only layer that builds `domain.X` structs. It applies business rules first, then constructs the model, then delegates persistence to the repository.

```go
// internal/modules/catalog/slices/create_product/service.go
package create_product

import (
    "context"
    "github.com/your-org/nexokit/internal/domain"
    "github.com/your-org/nexokit/internal/modules/catalog/core"
    "github.com/your-org/nexokit/internal/modules/catalog/slices/create_product/sku"
)

type Service interface {
    Execute(ctx context.Context, req Request) (core.ProductResponse, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) Execute(ctx context.Context, req Request) (core.ProductResponse, error) {
    // 1. validate business rules
    exists, err := s.repo.SlugExists(ctx, req.Slug)
    if err != nil {
        return core.ProductResponse{}, err
    }
    if exists {
        return core.ProductResponse{}, core.ErrProductSlugAlreadyExists
    }

    // 2. construct the domain model — service is responsible for this
    product := domain.Product{
        Name:        req.Name,
        Slug:        req.Slug,
        SKU:         sku.Generate(req.Name),   // business construction logic
        Price:       req.Price,
        CategoryID:  req.CategoryID,
        Description: req.Description,
        IsActive:    true,
    }

    // 3. delegate persistence
    return s.repo.Create(ctx, product)
}
```

### Repository

**Owns:** Slice persistence, GORM calls, persistence record ↔ domain model conversion, persistence-to-domain error translation.

**Must not own:** HTTP/API mapping, business policy, model construction from request data, `platform/apperror` imports.

The repository defines a local persistence record struct with GORM tags. `domain.X` structs have no GORM tags — the record struct is the adapter between GORM and the domain.

**Error mapping rule:** DB/GORM errors never leave the repository as GORM errors.

```
gorm.ErrRecordNotFound  →  core.ErrXxxNotFound
duplicate key violation  →  core.ErrXxxAlreadyExists
unexpected DB error      →  original wrapped error (becomes 500 at handler)
```

```go
// internal/modules/catalog/slices/create_product/repository.go
package create_product

import (
    "context"
    "errors"
    "time"
    "gorm.io/gorm"
    "github.com/shopspring/decimal"
    "github.com/your-org/nexokit/internal/domain"
    "github.com/your-org/nexokit/internal/modules/catalog/core"
)

type Repository interface {
    SlugExists(ctx context.Context, slug string) (bool, error)
    Create(ctx context.Context, product domain.Product) (core.ProductResponse, error)
}

// productRecord is the GORM persistence model. It never leaves this file.
type productRecord struct {
    ID          uint            `gorm:"primaryKey"`
    Name        string          `gorm:"column:name;not null"`
    Slug        string          `gorm:"column:slug;uniqueIndex"`
    SKU         string          `gorm:"column:sku;uniqueIndex"`
    Price       decimal.Decimal `gorm:"column:price;type:numeric(10,2)"`
    CategoryID  uint            `gorm:"column:category_id"`
    Description string          `gorm:"column:description"`
    ImageURL    string          `gorm:"column:image_url"`
    IsActive    bool            `gorm:"column:is_active;default:true"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

func (productRecord) TableName() string { return "products" }

func (r productRecord) toDomain() domain.Product {
    return domain.Product{
        ID:          r.ID,
        Name:        r.Name,
        Slug:        r.Slug,
        SKU:         r.SKU,
        Price:       r.Price,
        CategoryID:  r.CategoryID,
        Description: r.Description,
        ImageURL:    r.ImageURL,
        IsActive:    r.IsActive,
    }
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repository{db: db} }

func (r *repository) SlugExists(ctx context.Context, slug string) (bool, error) {
    var count int64
    err := r.db.WithContext(ctx).Model(&productRecord{}).
        Where("slug = ?", slug).Count(&count).Error
    return count > 0, err
}

func (r *repository) Create(ctx context.Context, p domain.Product) (core.ProductResponse, error) {
    record := productRecord{
        Name:        p.Name,
        Slug:        p.Slug,
        SKU:         p.SKU,
        Price:       p.Price,
        CategoryID:  p.CategoryID,
        Description: p.Description,
        IsActive:    p.IsActive,
    }
    if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
        if isDuplicateKeyError(err) {
            return core.ProductResponse{}, core.ErrProductSlugAlreadyExists
        }
        return core.ProductResponse{}, err
    }
    return core.ProductToResponse(record.toDomain(), domain.Category{}), nil
}
```

---

## Error flow

Errors travel inward as domain errors and outward as API responses:

```
DB/GORM error
    → repository translates to core domain error (core.ErrXxx)
        → service propagates or wraps with additional context
            → handler: apperror.Log() for logger + response.HandleError() for client
```

| Case | Repository returns | HTTP response |
|---|---|---|
| Row not found | `core.ErrXxxNotFound` (wraps `ErrNotFound`) | 404 |
| Duplicate field | `core.ErrXxxAlreadyExists` (wraps `ErrConflict`) | 409 |
| Business rule violation | `core.ErrXxxProtected` (wraps `ErrForbidden`) | 403 |
| Invalid state transition | `core.ErrInvalidXxx` (wraps `ErrBadRequest`) | 400 |
| Unprocessable entity | `core.ErrXxxHasRelations` (wraps `ErrUnprocessable`) | 422 |
| Unexpected DB failure | original wrapped error | 500 |

**Services never see GORM errors.** When a service needs to branch on existence, the repository exposes an existence contract instead of forcing error comparison:

```go
// preferred over returning gorm.ErrRecordNotFound to the service
FindBySlug(ctx context.Context, slug string) (*domain.Product, bool, error)
```

**Expected vs unexpected errors:**

- Expected domain errors (`core.ErrXxx` wrapping a sentinel) carry a public message and no `Cause`. The handler calls `response.HandleError` directly — no logging needed.
- Unexpected errors (raw DB errors, panics) carry no public message and have a `Cause`. The handler logs the cause via `apperror.Log(err)` before calling `response.HandleError`.

---

## `container.go`

**Purpose:** Module composition root. Constructs all slice dependencies and exposes cross-module interfaces.

**Rules:**
- No business logic, no conditionals beyond startup panics.
- Receives external dependencies (DB, cross-module contracts) as constructor parameters.
- Exposes implementations of `platform/contracts` interfaces for `app/container.go` to wire.

```go
// internal/modules/campaigns/container.go
package campaigns

import (
    "gorm.io/gorm"
    "github.com/your-org/nexokit/internal/platform/contracts"
    "github.com/your-org/nexokit/internal/modules/campaigns/slices/apply_discount"
)

type Container struct {
    ApplyDiscountH *apply_discount.Handler
    discountSvc    contracts.DiscountEngine
}

func NewContainer(db *gorm.DB) *Container {
    repo := apply_discount.NewRepository(db)
    svc  := apply_discount.NewService(repo)

    return &Container{
        ApplyDiscountH: apply_discount.NewHandler(svc),
        discountSvc:    svc,
    }
}

// DiscountEngine exposes the implementation as a platform contract.
// app/container.go calls this and injects it into sales.
func (c *Container) DiscountEngine() contracts.DiscountEngine {
    return c.discountSvc
}
```

---

## `routes.go`

**Purpose:** Route registration only. No middleware logic, no handler construction.

```go
// internal/modules/catalog/routes.go
package catalog

import "github.com/gin-gonic/gin"

func (c *Container) RegisterRoutes(r *gin.RouterGroup) {
    products := r.Group("/products")
    {
        products.GET("",       c.ListProductsH.Handle)
        products.GET("/:slug", c.ViewProductH.Handle)
        products.POST("",      c.CreateProductH.Handle)
    }
}
```

---

## `app/container.go`

**Purpose:** The single orchestrator of the full dependency graph. The only file in the codebase that knows about all modules simultaneously.

**Rules:**
- Constructs modules in dependency order.
- Wires cross-module contracts: calls `campaigns.DiscountEngine()` and passes it to `sales.NewContainer(...)`.
- No business logic, no HTTP handling.
- If this file grows beyond ~100 lines, Wire should be considered.

```go
// internal/app/container.go
package app

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "github.com/your-org/nexokit/internal/modules/campaigns"
    "github.com/your-org/nexokit/internal/modules/catalog"
    "github.com/your-org/nexokit/internal/modules/sales"
)

type Container struct {
    Catalog   *catalog.Container
    Campaigns *campaigns.Container
    Sales     *sales.Container
}

func NewContainer(db *gorm.DB) *Container {
    catalogC   := catalog.NewContainer(db)
    campaignsC := campaigns.NewContainer(db)

    // sales receives the discount engine as a contracts.DiscountEngine interface.
    // It never imports the campaigns package.
    salesC := sales.NewContainer(db, campaignsC.DiscountEngine())

    return &Container{
        Catalog:   catalogC,
        Campaigns: campaignsC,
        Sales:     salesC,
    }
}

func (c *Container) RegisterRoutes(r *gin.RouterGroup) {
    c.Catalog.RegisterRoutes(r)
    c.Campaigns.RegisterRoutes(r)
    c.Sales.RegisterRoutes(r)
}
```

---

## Cross-module capability rule

When module A needs a capability owned by module B:

1. Declare the interface and its input/output types in `platform/contracts/`.
2. Module B implements the interface internally; its `container.go` exposes it via a method returning the interface type.
3. Module A receives the interface as a constructor parameter; it never imports module B.
4. `app/container.go` calls B's exposure method and passes the result to A's constructor.

```
platform/contracts/  ←  declares DiscountEngine interface + DiscountableItem / DiscountResult
campaigns/           ←  implements DiscountEngine, exposes via container.DiscountEngine()
sales/               ←  receives contracts.DiscountEngine as constructor parameter
app/container        ←  wires: campaignsC.DiscountEngine() → sales.NewContainer(db, ...)
```

---

## GORM partial model rule

When a module defines a local struct that maps to a table it does not own, the struct must declare its table name explicitly.

```go
type IAMUser struct {
    ID    uint
    Email string
    Name  string
}

func (IAMUser) TableName() string { return "users" }
```

Checklist:
- Every partial model has `TableName()` when the struct name differs from the table name.
- Table names match the real Goose migration tables.
- A unit test covers the table name for every partial model.

---

## Multi-entity grouping heuristic

> If a module has more than one entity **and** each entity has more than 3 use-cases, group slices by entity under `slices/`. Otherwise, slices are flat under `slices/`.

---

## Quick reference

```
domain/                       canonical models — no GORM tags, no behavior
platform/apperror/            AppError, sentinels, Status(), PublicMessage(), Log(), Wrap()
platform/contracts/           cross-module interfaces + their input/output types
platform/response/            APIResponse[T], HandleError, HTTP helpers
platform/*/                   DB, tenant, logger, config

modules/<m>/core/             DTOs, domain errors (Wrap sentinels), pure mappers, constants
modules/<m>/queries/          reusable persistence queries (2+ slice consumers)
modules/<m>/slices/<s>/
  handler.go                  HTTP in/out, log cause, call response.HandleError
  service.go                  business rules, build domain.Model, no GORM, no apperror
  repository.go               persistence record (GORM tags), GORM calls,
                              record ↔ domain conversion, DB→domain error translation
modules/<m>/container.go      wiring only, exposes platform/contracts implementations
modules/<m>/routes.go         route registration only

app/container.go              full dependency graph, cross-module wiring
cmd/api/main.go               process entry point, server startup
```
