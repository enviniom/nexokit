package token

import (
	"testing"
	"time"
)

func TestIssueAndParseAccess(t *testing.T) {
	m := NewManager("test-key-must-be-32-bytes-long!!", 15*time.Minute)

	t.Run("happy path with company", func(t *testing.T) {
		companyID := uint(42)
		tokenStr, err := m.IssueAccess("user-123", "admin", &companyID)
		if err != nil {
			t.Fatalf("IssueAccess failed: %v", err)
		}
		if tokenStr == "" {
			t.Fatal("expected non-empty token")
		}

		claims, err := m.ParseAccess(tokenStr)
		if err != nil {
			t.Fatalf("ParseAccess failed: %v", err)
		}

		if claims.Sub != "user-123" {
			t.Errorf("expected sub 'user-123', got %s", claims.Sub)
		}
		if claims.Role != "admin" {
			t.Errorf("expected role 'admin', got %s", claims.Role)
		}
		if claims.TokenType != "access" {
			t.Errorf("expected token_type 'access', got %s", claims.TokenType)
		}
		if claims.CompanyID == nil || *claims.CompanyID != 42 {
			t.Errorf("expected company_id 42, got %v", claims.CompanyID)
		}
		if claims.ExpiresAt.Before(claims.IssuedAt) {
			t.Error("expires_at must be after issued_at")
		}
	})

	t.Run("happy path without company", func(t *testing.T) {
		tokenStr, err := m.IssueAccess("user-456", "user", nil)
		if err != nil {
			t.Fatalf("IssueAccess failed: %v", err)
		}

		claims, err := m.ParseAccess(tokenStr)
		if err != nil {
			t.Fatalf("ParseAccess failed: %v", err)
		}

		if claims.CompanyID != nil {
			t.Errorf("expected nil company_id, got %v", claims.CompanyID)
		}
	})

	t.Run("expired token rejected", func(t *testing.T) {
		shortMgr := NewManager("test-key-must-be-32-bytes-long!!", -1*time.Minute)
		tokenStr, err := shortMgr.IssueAccess("user-789", "user", nil)
		if err != nil {
			t.Fatalf("IssueAccess failed: %v", err)
		}

		_, err = shortMgr.ParseAccess(tokenStr)
		if err == nil {
			t.Error("expected error for expired token")
		}
	})

	t.Run("invalid token rejected", func(t *testing.T) {
		_, err := m.ParseAccess("not-a-valid-token")
		if err == nil {
			t.Error("expected error for invalid token")
		}
	})

	t.Run("tampered token rejected", func(t *testing.T) {
		tokenStr, err := m.IssueAccess("user-000", "user", nil)
		if err != nil {
			t.Fatalf("IssueAccess failed: %v", err)
		}

		tampered := tokenStr + "tampered"
		_, err = m.ParseAccess(tampered)
		if err == nil {
			t.Error("expected error for tampered token")
		}
	})
}

func TestManager_GenerateRefreshToken(t *testing.T) {
	m := NewManager("test-key-must-be-32-bytes-long!!", 15*time.Minute)

	t.Run("generates non-empty token", func(t *testing.T) {
		token, err := m.GenerateRefreshToken()
		if err != nil {
			t.Fatalf("GenerateRefreshToken failed: %v", err)
		}
		if token == "" {
			t.Error("expected non-empty refresh token")
		}
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		t1, err := m.GenerateRefreshToken()
		if err != nil {
			t.Fatalf("GenerateRefreshToken failed: %v", err)
		}
		t2, err := m.GenerateRefreshToken()
		if err != nil {
			t.Fatalf("GenerateRefreshToken failed: %v", err)
		}
		if t1 == t2 {
			t.Error("expected unique refresh tokens")
		}
	})
}

func TestManager_HashRefreshToken(t *testing.T) {
	m := NewManager("test-key-must-be-32-bytes-long!!", 15*time.Minute)

	t.Run("produces consistent hash", func(t *testing.T) {
		token := "test-refresh-token"
		hash1 := m.HashRefreshToken(token)
		hash2 := m.HashRefreshToken(token)
		if hash1 != hash2 {
			t.Error("expected consistent hash for same token")
		}
	})

	t.Run("different tokens produce different hashes", func(t *testing.T) {
		hash1 := m.HashRefreshToken("token-a")
		hash2 := m.HashRefreshToken("token-b")
		if hash1 == hash2 {
			t.Error("expected different hashes for different tokens")
		}
	})

	t.Run("hash is hex encoded", func(t *testing.T) {
		hash := m.HashRefreshToken("any-token")
		if len(hash) != 64 {
			t.Errorf("expected SHA-256 hex length 64, got %d", len(hash))
		}
	})
}

func TestManager_DifferentKeyFails(t *testing.T) {
	m1 := NewManager("test-key-must-be-32-bytes-long!!", 15*time.Minute)
	m2 := NewManager("different-key-32-bytes-long!", 15*time.Minute)

	tokenStr, err := m1.IssueAccess("user-111", "user", nil)
	if err != nil {
		t.Fatalf("IssueAccess failed: %v", err)
	}

	_, err = m2.ParseAccess(tokenStr)
	if err == nil {
		t.Error("expected error when parsing with different key")
	}
}
