package query

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

// PaginationParams carries normalized pagination inputs.
type PaginationParams struct {
	Page    int
	PerPage int
}

// Pagination is kept as a compatibility alias for existing module DTOs.
type Pagination = PaginationParams

// FilterParams carries generic list filters parsed from query parameters.
type FilterParams struct {
	Status      string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}

// SortParams carries sorting inputs parsed from query parameters.
type SortParams struct {
	Sort  string
	Order string
}

// SearchParams carries search inputs parsed from query parameters.
type SearchParams struct {
	Query string
}

// ListParams combines all generic list query parameters.
type ListParams struct {
	Pagination PaginationParams
	Filters    FilterParams
	Sort       SortParams
	Search     SearchParams
}

// ParsePagination normalizes page and per_page query values.
func ParsePagination(pageValue, perPageValue string) PaginationParams {
	page, _ := strconv.Atoi(pageValue)
	if page < 1 {
		page = DefaultPage
	}

	perPage, _ := strconv.Atoi(perPageValue)
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}

	return PaginationParams{Page: page, PerPage: perPage}
}

// PaginationFromGin reads and normalizes pagination query parameters from a Gin context.
func PaginationFromGin(c *gin.Context) PaginationParams {
	return ParsePagination(c.Query("page"), c.Query("per_page"))
}

// FiltersFromGin reads generic filters from a Gin context.
func FiltersFromGin(c *gin.Context) FilterParams {
	return FilterParams{
		Status:      c.Query("status"),
		CreatedFrom: parseDate(c.Query("created_from")),
		CreatedTo:   parseDate(c.Query("created_to")),
	}
}

// SortFromGin reads sorting query parameters from a Gin context.
func SortFromGin(c *gin.Context) SortParams {
	order := c.DefaultQuery("order", "desc")
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	return SortParams{
		Sort:  c.DefaultQuery("sort", "created_at"),
		Order: order,
	}
}

// SearchFromGin reads search query parameters from a Gin context.
func SearchFromGin(c *gin.Context) SearchParams {
	return SearchParams{Query: c.Query("search")}
}

// ListFromGin reads all generic list query parameters from a Gin context.
func ListFromGin(c *gin.Context) ListParams {
	return ListParams{
		Pagination: PaginationFromGin(c),
		Filters:    FiltersFromGin(c),
		Sort:       SortFromGin(c),
		Search:     SearchFromGin(c),
	}
}

func parseDate(value string) *time.Time {
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return nil
	}
	return &parsed
}
