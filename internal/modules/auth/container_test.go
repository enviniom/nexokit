package auth

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type fakeContainerPasswordVerifier struct{}

func (fakeContainerPasswordVerifier) VerifyPassword(password, hash string) error { return nil }

type fakeContainerTokenManager struct{}

func (fakeContainerTokenManager) IssueAccess(sub, role string, companyID *uint) (string, error) {
	return "access", nil
}

func (fakeContainerTokenManager) GenerateRefreshToken() (string, error) { return "refresh", nil }
func (fakeContainerTokenManager) HashRefreshToken(refreshToken string) string {
	return "hash:" + refreshToken
}

func TestNewContainer_WiresAllSlices(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	c := NewContainer(db, fakeContainerPasswordVerifier{}, fakeContainerTokenManager{}, time.Hour)
	if c == nil || c.AuthenticateUser == nil || c.RotateToken == nil || c.RevokeToken == nil || c.ViewSession == nil {
		t.Fatalf("expected auth slices wired, got %#v", c)
	}
}
