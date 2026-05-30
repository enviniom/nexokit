package queries

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"github.com/enviniom/nexokit/internal/shared"
)

func TestCheckDomainAvailable(t *testing.T) {
	db := newTestDB(t)

	if err := db.Create(&core.OnboardingCompanyDomain{
		BaseModel: shared.BaseModel{PublicID: "dom_01"},
		CompanyID: 1,
		Domain:    "acme.com",
		Status:    core.DomainStatusActive,
		Kind:      core.DomainKindPrimary,
	}).Error; err != nil {
		t.Fatalf("seed domain: %v", err)
	}

	tests := []struct {
		name         string
		domain       string
		duplicateErr error
		wantErr      error
	}{
		{name: "available", domain: "globex.com", duplicateErr: core.ErrDuplicateCompanyDomain, wantErr: nil},
		{name: "duplicate returns provided error", domain: "acme.com", duplicateErr: core.ErrDuplicateTechnicalDomain, wantErr: core.ErrDuplicateTechnicalDomain},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckDomainAvailable(db, tt.domain, tt.duplicateErr)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected err=%v, got %v", tt.wantErr, err)
			}
		})
	}
}
