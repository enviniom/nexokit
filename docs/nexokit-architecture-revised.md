# Nexokit — Guía de Arquitectura de Referencia (Versión Optimizada)

Esta guía establece el diseño de arquitectura, la estructura de directorios, las responsabilidades de cada capa y las reglas de importación para la API de Nexokit. El objetivo primordial es garantizar un acoplamiento bajo, alta cohesión y una clara separación de conceptos (Separation of Concerns).

---

## Estructura de Directorios

```
api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── domain/              # Modelos de dominio puros (sin etiquetas GORM ni HTTP)
│   ├── platform/            # Código compartido transversal e infraestructura básica
│   │   ├── contracts/       # Interfaces para comunicación inter-módulos
│   │   ├── response/        # Envolturas de respuesta API uniformes
│   │   ├── apperror/        # Definición y manejo de errores de aplicación
│   │   ├── database/        # Inicialización y pool de base de datos
│   │   ├── tenant/          # Contexto y utilidades multi-tenant
│   │   └── logger/          # Utilidad de logging estructurado
│   ├── modules/             # Monolito modular por capacidades de negocio
│   │   ├── catalog/
│   │   │   ├── core/        # DTOs del módulo, errores locales y mapeadores
│   │   │   ├── queries/     # Consultas de lectura reutilizables (compartidas por slices)
│   │   │   ├── slices/      # Casos de uso implementados como Vertical Slices
│   │   │   ├── container.go # Composition Root del módulo
│   │   │   └── routes.go    # Registro de rutas del módulo
│   │   ├── campaigns/
│   │   └── sales/
│   └── app/
│       └── container.go     # Wiring global y Composition Root principal
└── migrations/              # Migraciones físicas de base de datos (Goose)
```

---

## Reglas de Importación e Integridad de Capas

El flujo de dependencias es estrictamente unidireccional y hacia adentro. Las capas externas conocen a las internas, pero las internas jamás conocen a las externas.

```
cmd/  →  app/  →  modules/  →  platform/  →  domain/
                                    ↑
                               (contracts únicamente,
                                sin referencias circulares)
```

### Matriz de Dependencias Permitidas

| Capa | Dependencias Permitidas | Justificación |
|---|---|---|
| `domain/` | Ninguna interna del proyecto | Debe ser Go puro, representando reglas y datos de negocio sin acoplamiento tecnológico. |
| `platform/` | `domain/` | Provee herramientas genéricas e infraestructura que opera sobre tipos de dominio. |
| `modules/<m>/core/` | `domain/`, `platform/contracts/`, `platform/apperror/` | Contiene los tipos y errores propios del módulo. Jamás importa otros módulos. |
| `modules/<m>/slices/` | `domain/`, `platform/`, `modules/<m>/core/`, `modules/<m>/queries/` | Implementa la lógica de los casos de uso. |
| `app/` | Todos los módulos, todo `platform/` | Es el ensamblador global (Composition Root). |
| `cmd/` | Únicamente `app/` | Punto de entrada del binario. |

> [!IMPORTANT]
> **LOS MÓDULOS NUNCA SE IMPORTAN ENTRE SÍ.** La comunicación entre ellos se realiza única y exclusivamente mediante interfaces definidas en `platform/contracts/`, las cuales son inyectadas en `app/container.go`.

---

## 1. Capa de Dominio (`domain/`)

Representa la verdad del negocio. Son estructuras de Go puras (Plain Old Go Objects).

### Reglas Críticas
1. **SIN etiquetas GORM o JSON:** El dominio no debe saber cómo se guarda en base de datos ni cómo se serializa en la red. Las etiquetas GORM pertenecen a la capa de persistencia (`repository.go`) y las etiquetas JSON a los DTOs (`core/`).
2. **Métodos puros únicamente:** Se permiten métodos de comportamiento (ej. reglas de cálculo), siempre y cuando no importen paquetes de infraestructura o base de datos.

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

---

## 2. Capa Compartida (`platform/`)

### 2.1. Gestión de Errores (`platform/apperror/`)

Nexokit utiliza errores semánticos basados en sentinelas para desacoplar el transporte (HTTP/gRPC) de la lógica de negocio.

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
	Err     error  // Error semántico base (ErrNotFound, ErrConflict...)
	Message string // Mensaje legible para el usuario/cliente
	Cause   error  // Error técnico original (ej. error de SQL), NO expuesto al cliente
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

// Sentinelas globales
var (
	ErrNotFound        = &AppError{Message: messages.MsgNotFound}
	ErrForbidden       = &AppError{Message: messages.MsgForbidden}
	ErrUnauthorized    = &AppError{Message: messages.MsgUnauthorized}
	ErrConflict        = &AppError{Message: messages.MsgConflict}
	ErrBadRequest      = &AppError{Message: messages.MsgBadRequest}
	ErrTooManyRequests = &AppError{Message: messages.MsgTooManyRequests}
	ErrValidation      = &AppError{Message: messages.MsgValidationError}
	ErrUnprocessable   = &AppError{Message: messages.MsgUnprocessable}
	ErrInternal        = &AppError{Message: messages.MsgInternalError}
)

// Wrap asocia un error técnico o contexto a un sentinela semántico
func Wrap(sentinel error, message string, cause ...error) *AppError {
	var c error
	if len(cause) > 0 {
		c = cause[0]
	}
	return &AppError{Err: sentinel, Message: message, Cause: c}
}

// Status mapea el error semántico al código HTTP correspondiente
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

func Log(err error) error {
	var e *AppError
	if errors.As(err, &e) {
		return e.Cause
	}
	return err
}
```

### 2.2. Contratos Inter-Módulos (`platform/contracts/`)

Cuando el Módulo A necesita invocar una funcionalidad del Módulo B, **no debe importarlo**. En su lugar, el Módulo A importa una interfaz genérica ubicada aquí.

```go
// internal/platform/contracts/discount.go
package contracts

import "github.com/shopspring/decimal"

type DiscountableItem struct {
	ProductID  uint
	CategoryID uint
	Quantity   int
	UnitPrice  decimal.Decimal
}

type DiscountResult struct {
	ItemDiscounts map[uint]decimal.Decimal
	CartDiscount  decimal.Decimal
	CouponApplied string
}

type DiscountEngine interface {
	Apply(items []DiscountableItem, couponCode string) (DiscountResult, error)
}
```

---

## 3. Estructura de Módulos y Vertical Slices

Cada módulo encapsula su propia lógica y se subdivide en **Vertical Slices** (un subdirectorio por caso de uso dentro de `slices/`).

```
modules/<modulo>/
  container.go   # Composition root del módulo
  routes.go      # Rutas HTTP asociadas al módulo
  core/          # Lenguaje común: DTOs, mappers puros y errores del módulo
  queries/       # Queries de base de datos compartidas (solo si las usan 2+ slices)
  slices/
    create_product/
      handler.go
      service.go
      repository.go
```

### 3.1. Responsabilidades de la Frontera (Flujo de Datos)

Un error muy común es acoplar las capas internas a las firmas externas. La arquitectura de Nexokit exige el siguiente flujo estricto de tipos de datos:

```
[Cliente]  ── (JSON) ──>  [Handler]  ── (Request DTO) ──>  [Service]  ── (domain.Entity) ──>  [Repository]
                                                                                                 │
[Cliente]  <── (JSON) ──  [Handler]  <── (Response DTO) ◄── [Service]  ◄── (domain.Entity) ◄─────┘
```

> [!WARNING]
> **EL REPOSITORIO NUNCA DEBE CONOCER NI RETORNAR DTOs.** Su única responsabilidad es leer/escribir registros de persistencia y mapearlos a entidades puras de dominio (`domain.Entity`). Devolver un DTO (`core.ProductResponse`) desde el repositorio rompe la independencia de la persistencia y acopla la base de datos a la presentación.

---

## 4. Ejemplo Práctico de Vertical Slice: `create_product`

A continuación se presenta el código de referencia que ejemplifica la correcta separación de responsabilidades y flujo de datos.

### 4.1. Definición de Errores y DTOs en `core/`

```go
// internal/modules/catalog/core/errors.go
package core

import "github.com/your-org/nexokit/internal/platform/apperror"

var (
	ErrProductNotFound          = apperror.Wrap(apperror.ErrNotFound, "Producto no encontrado")
	ErrProductSlugAlreadyExists = apperror.Wrap(apperror.ErrConflict, "El slug del producto ya existe")
)
```

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
	Description string           `json:"description"`
}

// ProductToResponse es un mapeador puro. Sin lógica de BD, sin HTTP.
func ProductToResponse(p domain.Product) ProductResponse {
	return ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Slug:        p.Slug,
		Price:       p.Price,
		ImageURL:    p.ImageURL,
		Description: p.Description,
	}
}
```

### 4.2. Capa de Transporte: Handler (`handler.go`)

- **Responsabilidades:** Binding de JSON, validación sintáctica/estructural rápida, extracción de tenant desde el contexto, llamada al servicio y delegación del mapeo de errores HTTP a `response.HandleError`.
- **Prohibido:** Contener lógica de negocio o queries directas a base de datos.

```go
// internal/modules/catalog/slices/create_product/handler.go
package create_product

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/your-org/nexokit/internal/platform/apperror"
	"github.com/your-org/nexokit/internal/platform/logger"
	"github.com/your-org/nexokit/internal/platform/response"
	"github.com/your-org/nexokit/internal/modules/catalog/core"
	"go.uber.org/zap"
)

type Request struct {
	Name        string              `json:"name" binding:"required"`
	Slug        string              `json:"slug" binding:"required"`
	Price       string              `json:"price" binding:"required"`
	CategoryID  uint                `json:"category_id" binding:"required"`
	Description string              `json:"description"`
}

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Handle(c *gin.Context) {
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleError(c, apperror.Wrap(apperror.ErrBadRequest, err.Error()))
		return
	}

	res, err := h.svc.Execute(c.Request.Context(), req)
	if err != nil {
		if cause := apperror.Log(err); cause != nil {
			logger.Error(c, "create_product failed", zap.Error(cause))
		}
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response.Created(res))
}
```

### 4.3. Capa de Negocio: Service (`service.go`)

- **Responsabilidades:** Orquestar el caso de uso, validar reglas de negocio utilizando el repositorio, construir la entidad de dominio pura (`domain.Product`) y mapear el resultado final a un DTO de salida (`core.ProductResponse`).
- **Simplificación:** Evitamos el exceso de abstracciones (Interface Pollution). Los servicios internos de los slices no necesitan una interfaz Go a menos que deban ser inyectados en otros módulos mediante `platform/contracts`. Un struct concreto es más que suficiente y reduce el ruido cognitivo.

```go
// internal/modules/catalog/slices/create_product/service.go
package create_product

import (
	"context"
	"github.com/shopspring/decimal"
	"github.com/your-org/nexokit/internal/domain"
	"github.com/your-org/nexokit/internal/modules/catalog/core"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Execute(ctx context.Context, req Request) (core.ProductResponse, error) {
	// 1. Validar reglas de negocio
	exists, err := s.repo.SlugExists(ctx, req.Slug)
	if err != nil {
		return core.ProductResponse{}, err
	}
	if exists {
		return core.ProductResponse{}, core.ErrProductSlugAlreadyExists
	}

	price, err := decimal.NewFromString(req.Price)
	if err != nil {
		return core.ProductResponse{}, core.Wrap(core.ErrProductNotFound.Err, "Precio inválido", err)
	}

	// 2. Construir la entidad de dominio (responsabilidad exclusiva del Service)
	product := domain.Product{
		Name:        req.Name,
		Slug:        req.Slug,
		Price:       price,
		CategoryID:  req.CategoryID,
		Description: req.Description,
		IsActive:    true,
	}

	// 3. Persistir usando el repositorio (que retorna una entidad limpia de dominio)
	savedProduct, err := s.repo.Create(ctx, product)
	if err != nil {
		return core.ProductResponse{}, err
	}

	// 4. Mapear dominio a DTO para consumo externo
	return core.ProductToResponse(*savedProduct), nil
}
```

### 4.4. Capa de Persistencia: Repository (`repository.go`)

- **Responsabilidades:** Traducir consultas SQL/GORM, mapear base de datos (`productRecord`) a dominio (`domain.Product`) y transformar errores de persistencia a errores semánticos del negocio (`core.ErrXxx`).
- **Interfaces:** Mantener interfaces en el repositorio sigue siendo recomendado para poder probar la lógica del servicio en aislamiento usando mocks rápidos.

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
	"github.com/your-org/nexokit/internal/platform/tenant"
)

type Repository interface {
	SlugExists(ctx context.Context, slug string) (bool, error)
	Create(ctx context.Context, p domain.Product) (*domain.Product, error)
}

// productRecord es el modelo privado de persistencia de GORM. Jamás se expone fuera del repositorio.
type productRecord struct {
	ID          uint            `gorm:"primaryKey"`
	TenantID    string          `gorm:"column:tenant_id;not null"`
	Name        string          `gorm:"column:name;not null"`
	Slug        string          `gorm:"column:slug;uniqueIndex"`
	Price       decimal.Decimal `gorm:"column:price;type:numeric(10,2)"`
	CategoryID  uint            `gorm:"column:category_id"`
	Description string          `gorm:"column:description"`
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
		Price:       r.Price,
		CategoryID:  r.CategoryID,
		Description: r.Description,
		IsActive:    r.IsActive,
	}
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) SlugExists(ctx context.Context, slug string) (bool, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return false, err
	}

	var count int64
	err = r.db.WithContext(ctx).Model(&productRecord{}).
		Where("tenant_id = ? AND slug = ?", tenantID, slug).
		Count(&count).Error
	return count > 0, err
}

func (r *repository) Create(ctx context.Context, p domain.Product) (*domain.Product, error) {
	tenantID, err := tenant.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	record := productRecord{
		TenantID:    tenantID,
		Name:        p.Name,
		Slug:        p.Slug,
		Price:       p.Price,
		CategoryID:  p.CategoryID,
		Description: p.Description,
		IsActive:    p.IsActive,
	}

	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		// Mapeo de errores de infraestructura a semántica de negocio
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, core.ErrProductSlugAlreadyExists
		}
		return nil, err // Errores inesperados se elevan directamente
	}

	saved := record.toDomain()
	return &saved, nil
}
```

---

## 5. Regla Multi-Tenant y Seguridad de Aislamiento

Para prevenir la fuga de información entre clientes de la plataforma, el aislamiento debe ser estricto:

1. **Propagación del Contexto:** El Tenant ID es inyectado en la request por un middleware y viaja obligatoriamente en el `context.Context`.
2. **Aplicación Obligatoria en Repositorio:** Todo método del repositorio que realice consultas SQL debe extraer el `tenant_id` y filtrarlo explícitamente en la consulta:
   ```go
   tenantID, err := tenant.FromContext(ctx)
   // ...
   db.Where("tenant_id = ?", tenantID)
   ```
3. **Validación:** Se debe acompañar de pruebas unitarias que verifiquen que ninguna consulta carece de filtro de inquilino.

---

## 6. Modelos GORM Parciales (GORM Partial Models)

Cuando un módulo necesita consultar información de un modelo ajeno (ej. `sales` necesita el nombre de un producto), **no importa el modelo del módulo `catalog`**.

En su lugar, declara un modelo parcial específico para lectura local en el repositorio de su slice:

```go
// En internal/modules/sales/slices/add_item/repository.go
type catalogProductRecord struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"column:name"`
}

func (catalogProductRecord) TableName() string { return "products" }
```

### Reglas para Modelos Parciales
- Debe declarar explícitamente `TableName()` devolviendo el nombre real de la tabla en base de datos.
- Solo debe contener los campos necesarios para satisfacer la necesidad de lectura del caso de uso.
- Debe ir acompañado de un test unitario rápido que verifique que el nombre de la tabla mapeado coincide con las migraciones globales del sistema.

---

## 7. Referencia Rápida de Responsabilidades

```
domain/                       Modelos canónicos de negocio (estructuras puras sin tags).
platform/apperror/            Gestión de errores del sistema (sentinelas, mapeador HTTP).
platform/contracts/           Definición de interfaces para integraciones inter-módulos.
platform/response/            Sobres uniformes de respuesta HTTP genéricos.

modules/<m>/core/             Contratos de red de módulo (DTOs) y mapeadores puramente funcionales.
modules/<m>/queries/          Consultas read-only compartidas internamente en el módulo.
modules/<m>/slices/<s>/
  handler.go                  Capa HTTP/Transporte. Bindea datos y maneja respuestas de error.
  service.go                  Capa de Negocio. Valida reglas y orquesta la operación (struct concreto).
  repository.go               Capa de Persistencia. Escribe/lee base de datos y mapea a domain (GORM).
```
