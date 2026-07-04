package core

import "testing"

func TestAuthRole_TableName(t *testing.T) {
	m := AuthRole{}
	if got := m.TableName(); got != "roles" {
		t.Errorf("AuthRole.TableName() = %q, want %q", got, "roles")
	}
}

func TestAuthUser_TableName(t *testing.T) {
	m := AuthUser{}
	if got := m.TableName(); got != "users" {
		t.Errorf("AuthUser.TableName() = %q, want %q", got, "users")
	}
}

func TestRefreshToken_TableName(t *testing.T) {
	m := RefreshToken{}
	if got := m.TableName(); got != "refresh_tokens" {
		t.Errorf("RefreshToken.TableName() = %q, want %q", got, "refresh_tokens")
	}
}
