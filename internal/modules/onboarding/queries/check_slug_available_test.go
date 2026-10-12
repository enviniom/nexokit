package queries

import (
	"errors"
	companyentity "github.com/enviniom/nexokit/internal/modules/onboarding/core/company"
	"gorm.io/gorm"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"github.com/enviniom/nexokit/internal/shared"
)

func TestCheckSlugAvailable(t *testing.T) {
	db := newTestDB(t)

	if err := db.Create(&companyentity.OnboardingCompany{
		BaseModel: shared.BaseModel{PublicID: "cmp_01"},
		Name:      "Acme",
		Slug:      "acme",
		Status:    core.CompanyStatusActive,
	}).Error; err != nil {
		t.Fatalf("seed company: %v", err)
	}

	if err := db.Create(&companyentity.OnboardingCompany{
		BaseModel: shared.BaseModel{PublicID: "deleted_slug", DeletedAt: gorm.DeletedAt{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true}},
		Name:      "Deleted", Slug: "reserved", Status: core.CompanyStatusActive,
	}).Error; err != nil {
		t.Fatalf("seed soft-deleted record: %v", err)
	}

	tests := []struct {
		name    string
		slug    string
		wantErr error
	}{
		{name: "available", slug: "globex", wantErr: nil},
		{name: "duplicate", slug: "acme", wantErr: core.ErrDuplicateCompanySlug},
		{name: "soft-deleted slug remains reserved", slug: "reserved", wantErr: core.ErrDuplicateCompanySlug},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckSlugAvailable(db, tt.slug)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected err=%v, got %v", tt.wantErr, err)
			}
		})
	}
}
