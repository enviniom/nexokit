package password

import (
	"testing"
)

func TestHashAndVerify(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"simple password", "password123"},
		{"complex password", "My$uperStr0ng!P@ss"},
		{"unicode password", "пароль123!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := Hash(tt.password)
			if err != nil {
				t.Fatalf("Hash failed: %v", err)
			}

			if hash == "" {
				t.Error("expected non-empty hash")
			}

			if hash == tt.password {
				t.Error("hash must not equal plain password")
			}

			if err := Verify(tt.password, hash); err != nil {
				t.Errorf("Verify failed for correct password: %v", err)
			}

			if err := Verify("wrong-password", hash); err == nil {
				t.Error("Verify should fail for wrong password")
			}
		})
	}
}

func TestHash_DifferentHashesForSamePassword(t *testing.T) {
	password := "same-password"

	hash1, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}

	hash2, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}

	if hash1 == hash2 {
		t.Error("same password should produce different hashes due to salt")
	}
}
