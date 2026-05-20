package query

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

// Pagination carries normalized pagination inputs.
type Pagination struct {
	Page    int
	PerPage int
}

// ParsePagination normalizes page and per_page query values.
func ParsePagination(pageValue, perPageValue string) Pagination {
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

	return Pagination{Page: page, PerPage: perPage}
}

// PaginationFromGin reads and normalizes pagination query parameters from a Gin context.
func PaginationFromGin(c *gin.Context) Pagination {
	return ParsePagination(c.Query("page"), c.Query("per_page"))
}
