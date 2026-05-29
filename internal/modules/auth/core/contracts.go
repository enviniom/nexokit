package core

// PasswordVerifier verifies user passwords.
type PasswordVerifier interface {
	VerifyPassword(password, hash string) error
}

// TokenManager generates and hashes opaque refresh tokens.
type TokenManager interface {
	IssueAccess(sub, role string, companyID *uint) (string, error)
	GenerateRefreshToken() (string, error)
	HashRefreshToken(refreshToken string) string
}
