package query

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func queryContext(target string) *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", target, nil)
	return c
}

func TestPaginationFromGin(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		wantPage    int
		wantPerPage int
	}{
		{"explicit values", "/?page=2&per_page=10", 2, 10},
		{"invalid values use bounds", "/?page=0&per_page=999", DefaultPage, MaxPerPage},
		{"missing values use defaults", "/", DefaultPage, DefaultPerPage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PaginationFromGin(queryContext(tt.target))
			if got.Page != tt.wantPage || got.PerPage != tt.wantPerPage {
				t.Fatalf("PaginationFromGin() = %+v; want page=%d per_page=%d", got, tt.wantPage, tt.wantPerPage)
			}
		})
	}
}

func TestFiltersFromGin(t *testing.T) {
	t.Run("default empty filters", func(t *testing.T) {
		got := FiltersFromGin(queryContext("/"))
		if got.Status != "" || got.CreatedFrom != nil || got.CreatedTo != nil {
			t.Fatalf("FiltersFromGin() = %+v; want empty filters", got)
		}
	})

	t.Run("status and dates", func(t *testing.T) {
		got := FiltersFromGin(queryContext("/?status=active&created_from=2025-01-01&created_to=2025-12-31"))
		from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
		if got.Status != "active" {
			t.Fatalf("Status = %q; want active", got.Status)
		}
		if got.CreatedFrom == nil || !got.CreatedFrom.Equal(from) {
			t.Fatalf("CreatedFrom = %v; want %v", got.CreatedFrom, from)
		}
		if got.CreatedTo == nil || !got.CreatedTo.Equal(to) {
			t.Fatalf("CreatedTo = %v; want %v", got.CreatedTo, to)
		}
	})

	t.Run("invalid dates are ignored", func(t *testing.T) {
		got := FiltersFromGin(queryContext("/?created_from=not-a-date"))
		if got.CreatedFrom != nil {
			t.Fatalf("CreatedFrom = %v; want nil", got.CreatedFrom)
		}
	})
}

func TestSortFromGin(t *testing.T) {
	tests := []struct {
		name      string
		target    string
		wantSort  string
		wantOrder string
	}{
		{"default sort", "/", "created_at", "desc"},
		{"explicit sort", "/?sort=name&order=asc", "name", "asc"},
		{"invalid order defaults", "/?sort=name&order=sideways", "name", "desc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SortFromGin(queryContext(tt.target))
			if got.Sort != tt.wantSort || got.Order != tt.wantOrder {
				t.Fatalf("SortFromGin() = %+v; want sort=%q order=%q", got, tt.wantSort, tt.wantOrder)
			}
		})
	}
}

func TestSearchFromGin(t *testing.T) {
	if got := SearchFromGin(queryContext("/?search=jhon")); got.Query != "jhon" {
		t.Fatalf("Query = %q; want jhon", got.Query)
	}
	if got := SearchFromGin(queryContext("/")); got.Query != "" {
		t.Fatalf("Query = %q; want empty", got.Query)
	}
}

func TestListFromGin(t *testing.T) {
	got := ListFromGin(queryContext("/?page=2&per_page=10&status=active&sort=name&order=asc&search=test"))
	if got.Pagination.Page != 2 || got.Pagination.PerPage != 10 {
		t.Fatalf("Pagination = %+v; want page=2 per_page=10", got.Pagination)
	}
	if got.Filters.Status != "active" {
		t.Fatalf("Status = %q; want active", got.Filters.Status)
	}
	if got.Sort.Sort != "name" || got.Sort.Order != "asc" {
		t.Fatalf("Sort = %+v; want name asc", got.Sort)
	}
	if got.Search.Query != "test" {
		t.Fatalf("Query = %q; want test", got.Search.Query)
	}
}
