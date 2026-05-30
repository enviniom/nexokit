package queries

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"github.com/enviniom/nexokit/internal/shared"
)

func TestCheckSlugAvailable(t *testing.T) {
	db := newTestDB(t)

	if err := db.Create(&core.OnboardingCompany{
		BaseModel: shared.BaseModel{PublicID: "cmp_01"},
		Name:      "Acme",
		Slug:      "acme",
		Status:    core.CompanyStatusActive,
	}).Error; err != nil {
		t.Fatalf("seed company: %v", err)
	}

	tests := []struct {
		name    string
		slug    string
		wantErr error
	}{
		{name: "available", slug: "globex", wantErr: nil},
		{name: "duplicate", slug: "acme", wantErr: core.ErrDuplicateCompanySlug},
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
