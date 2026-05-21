# Query Parsing Specification

## Purpose
Define structs and Gin context parsers for pagination, filters, sorting, and search query parameters consumed by handlers and GORM helpers.

## Requirements

### Requirement: PaginationParams struct and parser

The system MUST provide a `PaginationParams` struct with `Page` and `PerPage`, parsed from `page` and `per_page` using the existing defaults and maximum page size.

#### Scenario: Pagination params parsed

- GIVEN `?page=2&per_page=10`
- WHEN `PaginationFromGin(c)` is called
- THEN `Page` is 2 and `PerPage` is 10

#### Scenario: Invalid pagination defaults

- GIVEN `?page=0&per_page=999`
- WHEN `PaginationFromGin(c)` is called
- THEN page/per_page are normalized to configured bounds

### Requirement: FilterParams struct and parser

The system MUST provide a `FilterParams` struct with fields `Status`, `CreatedFrom`, `CreatedTo` (time.Time pointers) and a `FiltersFromGin(c *gin.Context)` function that reads `status`, `created_from`, `created_to` query params.

#### Scenario: Default empty filters

- GIVEN a request with no filter query params
- WHEN `FiltersFromGin(c)` is called
- THEN `FilterParams.Status` is empty and date fields are nil

#### Scenario: Status and date filters parsed

- GIVEN a request with `?status=active&created_from=2025-01-01&created_to=2025-12-31`
- WHEN `FiltersFromGin(c)` is called
- THEN `Status` equals `"active"`, `CreatedFrom` is Jan 1, `CreatedTo` is Dec 31

#### Scenario: Invalid date format ignored

- GIVEN a request with `?created_from=not-a-date`
- WHEN `FiltersFromGin(c)` is called
- THEN `CreatedFrom` is nil and other fields are unaffected

### Requirement: SortParams struct and parser

The system MUST provide a `SortParams` struct with `Sort` (column name, default `created_at`) and `Order` (`asc` or `desc`, default `desc`), parsed via `SortFromGin(c *gin.Context)`.

#### Scenario: Default sort order

- GIVEN a request with no sort params
- WHEN `SortFromGin(c)` is called
- THEN `Sort` is `"created_at"` and `Order` is `"desc"`

#### Scenario: Explicit sort

- GIVEN `?sort=name&order=asc`
- WHEN `SortFromGin(c)` is called
- THEN `Sort` is `"name"` and `Order` is `"asc"`

#### Scenario: Invalid order defaults to desc

- GIVEN `?order=sideways`
- WHEN `SortFromGin(c)` is called
- THEN `Order` is `"desc"`

### Requirement: SearchParams struct and parser

The system MUST provide a `SearchParams` struct with `Query` (string, default empty), parsed via `SearchFromGin(c *gin.Context)`.

#### Scenario: Search query present

- GIVEN `?search=jhon`
- WHEN `SearchFromGin(c)` is called
- THEN `Query` is `"jhon"`

#### Scenario: No search query

- GIVEN a request with no `search` param
- WHEN `SearchFromGin(c)` is called
- THEN `Query` is `""`

### Requirement: Combined ListParams struct

The system MUST provide a `ListParams` struct composing `PaginationParams`, `FilterParams`, `SortParams`, and `SearchParams` with a `ListFromGin(c *gin.Context)` parser.

#### Scenario: Full params parsed together

- GIVEN `?page=2&per_page=10&status=active&sort=name&order=asc&search=test`
- WHEN `ListFromGin(c)` is called
- THEN `Pagination.Page` is 2, `Filters.Status` is `"active"`, `Sort.Sort` is `"name"`, `Search.Query` is `"test"`
