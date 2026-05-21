package gormutil

import (
	"strings"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/platform/query"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testRecord struct {
	ID        uint
	Name      string
	Email     string
	Status    string
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func dryRunDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DryRun: true})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return db.Model(&testRecord{})
}

func renderedSQL(db *gorm.DB) string {
	stmt := db.Find(&[]testRecord{}).Statement
	return stmt.SQL.String()
}

func TestApplyPagination(t *testing.T) {
	sql := renderedSQL(ApplyPagination(dryRunDB(t), 2, 15))

	if !strings.Contains(sql, "LIMIT 15") {
		t.Fatalf("SQL = %q; want LIMIT 15", sql)
	}
	if !strings.Contains(sql, "OFFSET 15") {
		t.Fatalf("SQL = %q; want OFFSET 15", sql)
	}
}

func TestApplySorting(t *testing.T) {
	tests := []struct {
		name    string
		sort    query.SortParams
		allowed []string
		want    string
		deny    string
	}{
		{"allowed ascending sort", query.SortParams{Sort: "name", Order: "asc"}, []string{"name"}, "ORDER BY name asc", ""},
		{"invalid order defaults", query.SortParams{Sort: "created_at", Order: "sideways"}, []string{"created_at"}, "ORDER BY created_at desc", ""},
		{"empty sort is no-op", query.SortParams{}, nil, "", "ORDER BY"},
		{"disallowed sort is no-op", query.SortParams{Sort: "email", Order: "asc"}, []string{"name"}, "", "ORDER BY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql := renderedSQL(ApplySorting(dryRunDB(t), tt.sort, tt.allowed...))
			if tt.want != "" && !strings.Contains(sql, tt.want) {
				t.Fatalf("SQL = %q; want contains %q", sql, tt.want)
			}
			if tt.deny != "" && strings.Contains(sql, tt.deny) {
				t.Fatalf("SQL = %q; want no %q", sql, tt.deny)
			}
		})
	}
}

func TestApplySearch(t *testing.T) {
	t.Run("search across columns", func(t *testing.T) {
		db := ApplySearch(dryRunDB(t), query.SearchParams{Query: "john"}, "name", "email")
		sql := renderedSQL(db)
		if !strings.Contains(sql, "name LIKE ? OR email LIKE ?") {
			t.Fatalf("SQL = %q; want search conditions", sql)
		}
		if len(db.Statement.Vars) != 2 || db.Statement.Vars[0] != "%john%" || db.Statement.Vars[1] != "%john%" {
			t.Fatalf("vars = %#v; want two LIKE args", db.Statement.Vars)
		}
	})

	t.Run("empty query is no-op", func(t *testing.T) {
		sql := renderedSQL(ApplySearch(dryRunDB(t), query.SearchParams{}, "name"))
		if strings.Contains(sql, "LIKE") {
			t.Fatalf("SQL = %q; want no LIKE", sql)
		}
	})
}

func TestApplyDateRange(t *testing.T) {
	from := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		filters query.FilterParams
		want    []string
		deny    string
	}{
		{"both dates", query.FilterParams{CreatedFrom: &from, CreatedTo: &to}, []string{"created_at >= ?", "created_at <= ?"}, ""},
		{"from only", query.FilterParams{CreatedFrom: &from}, []string{"created_at >= ?"}, "created_at <= ?"},
		{"empty filters are no-op", query.FilterParams{}, nil, "created_at >= ?"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql := renderedSQL(ApplyDateRange(dryRunDB(t), tt.filters, "created_at"))
			for _, want := range tt.want {
				if !strings.Contains(sql, want) {
					t.Fatalf("SQL = %q; want contains %q", sql, want)
				}
			}
			if tt.deny != "" && strings.Contains(sql, tt.deny) {
				t.Fatalf("SQL = %q; want no %q", sql, tt.deny)
			}
		})
	}
}

func TestApplyStatusFilter(t *testing.T) {
	t.Run("status filter", func(t *testing.T) {
		db := ApplyStatusFilter(dryRunDB(t), query.FilterParams{Status: "active"}, "status")
		sql := renderedSQL(db)
		if !strings.Contains(sql, "status = ?") {
			t.Fatalf("SQL = %q; want status filter", sql)
		}
		if len(db.Statement.Vars) != 1 || db.Statement.Vars[0] != "active" {
			t.Fatalf("vars = %#v; want active", db.Statement.Vars)
		}
	})

	t.Run("empty status is no-op", func(t *testing.T) {
		sql := renderedSQL(ApplyStatusFilter(dryRunDB(t), query.FilterParams{}, "status"))
		if strings.Contains(sql, "status = ?") {
			t.Fatalf("SQL = %q; want no status filter", sql)
		}
	})
}

func TestHelpersPreserveSoftDeleteScope(t *testing.T) {
	db := dryRunDB(t)
	db = ApplyStatusFilter(db, query.FilterParams{Status: "active"}, "status")
	db = ApplySearch(db, query.SearchParams{Query: "john"}, "name")
	db = ApplySorting(db, query.SortParams{Sort: "created_at", Order: "desc"}, "created_at")
	db = ApplyPagination(db, 1, 10)

	sql := renderedSQL(db)
	if !strings.Contains(sql, "deleted_at") {
		t.Fatalf("SQL = %q; want GORM soft-delete scope to remain active", sql)
	}
}
