package queries

import (
	"errors"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/onboarding/core"
	"github.com/enviniom/nexokit/internal/shared"
)

func TestCheckEmailAvailable(t *testing.T) {
	db := newTestDB(t)

	if err := db.Create(&core.OnboardingUser{
		BaseModel:    shared.BaseModel{PublicID: "usr_01"},
		Name:         "Admin",
		Email:        "admin@acme.com",
		PasswordHash: "hash",
		RoleID:       1,
		IsActive:     true,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	tests := []struct {
		name    string
		email   string
		wantErr error
	}{
		{name: "available", email: "other@acme.com", wantErr: nil},
		{name: "duplicate", email: "admin@acme.com", wantErr: core.ErrDuplicateAdminEmail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckEmailAvailable(db, tt.email)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected err=%v, got %v", tt.wantErr, err)
			}
		})
	}
}
