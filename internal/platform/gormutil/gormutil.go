package gormutil

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/enviniom/nexokit/internal/platform/query"
	"gorm.io/gorm"
)

var safeColumnPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_\.]*$`)

// ApplyPagination applies a one-indexed pagination window.
func ApplyPagination(db *gorm.DB, page, perPage int) *gorm.DB {
	if page < 1 {
		page = query.DefaultPage
	}
	if perPage < 1 {
		perPage = query.DefaultPerPage
	}
	return db.Offset((page - 1) * perPage).Limit(perPage)
}

// ApplySorting applies an ORDER BY clause when the requested sort is safe and allowed.
func ApplySorting(db *gorm.DB, sort query.SortParams, allowedColumns ...string) *gorm.DB {
	if sort.Sort == "" || !isAllowedColumn(sort.Sort, allowedColumns) {
		return db
	}
	order := strings.ToLower(sort.Order)
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	return db.Order(fmt.Sprintf("%s %s", sort.Sort, order))
}

// ApplySearch applies a LIKE search across the provided columns.
func ApplySearch(db *gorm.DB, search query.SearchParams, columns ...string) *gorm.DB {
	if search.Query == "" || len(columns) == 0 {
		return db
	}

	conditions := make([]string, 0, len(columns))
	args := make([]any, 0, len(columns))
	for _, column := range columns {
		if !safeColumnPattern.MatchString(column) {
			continue
		}
		conditions = append(conditions, column+" LIKE ?")
		args = append(args, "%"+search.Query+"%")
	}
	if len(conditions) == 0 {
		return db
	}

	return db.Where(strings.Join(conditions, " OR "), args...)
}

// ApplyDateRange applies inclusive date range filters to the provided column.
func ApplyDateRange(db *gorm.DB, filters query.FilterParams, column string) *gorm.DB {
	if !safeColumnPattern.MatchString(column) {
		return db
	}
	if filters.CreatedFrom != nil {
		db = db.Where(column+" >= ?", *filters.CreatedFrom)
	}
	if filters.CreatedTo != nil {
		db = db.Where(column+" <= ?", *filters.CreatedTo)
	}
	return db
}

// ApplyStatusFilter applies an equality filter to the provided status column.
func ApplyStatusFilter(db *gorm.DB, filters query.FilterParams, column string) *gorm.DB {
	if filters.Status == "" || !safeColumnPattern.MatchString(column) {
		return db
	}
	return db.Where(column+" = ?", filters.Status)
}

func isAllowedColumn(column string, allowedColumns []string) bool {
	if !safeColumnPattern.MatchString(column) {
		return false
	}
	if len(allowedColumns) == 0 {
		return true
	}
	for _, allowed := range allowedColumns {
		if column == allowed {
			return true
		}
	}
	return false
}
