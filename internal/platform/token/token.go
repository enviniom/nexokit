package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"aidanwoods.dev/go-paseto"
)

var (
	errInvalidToken = errors.New("invalid or expired token")
)

// AccessClaims holds the claims extracted from a PASETO v4.local access token.
type AccessClaims struct {
	Sub       string    `json:"sub"`
	Role      string    `json:"role"`
	CompanyID *uint     `json:"company_id,omitempty"`
	TokenType string    `json:"token_type"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Manager handles PASETO token operations.
type Manager struct {
	secretKey []byte
	accessTTL time.Duration
}

// NewManager creates a new token manager with the given secret key and access TTL.
// The secret key is hashed to a fixed 32-byte length suitable for v4.local.
func NewManager(secretKey string, accessTTL time.Duration) *Manager {
	h := sha256.Sum256([]byte(secretKey))
	return &Manager{
		secretKey: h[:],
		accessTTL: accessTTL,
	}
}

// IssueAccess creates a PASETO v4.local access token with the provided claims.
func (m *Manager) IssueAccess(sub, role string, companyID *uint) (string, error) {
	key, err := paseto.V4SymmetricKeyFromBytes(m.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to create paseto key: %w", err)
	}

	now := time.Now().UTC()
	claims := AccessClaims{
		Sub:       sub,
		Role:      role,
		CompanyID: companyID,
		TokenType: "access",
		IssuedAt:  now,
		ExpiresAt: now.Add(m.accessTTL),
	}

	token := paseto.NewToken()
	token.SetIssuedAt(claims.IssuedAt)
	token.SetExpiration(claims.ExpiresAt)
	token.SetSubject(claims.Sub)
	token.SetString("role", claims.Role)
	token.SetString("token_type", claims.TokenType)

	if claims.CompanyID != nil {
		token.Set("company_id", *claims.CompanyID)
	}

	return token.V4Encrypt(key, nil), nil
}

// ParseAccess validates and decrypts a PASETO v4.local access token.
func (m *Manager) ParseAccess(tokenStr string) (*AccessClaims, error) {
	key, err := paseto.V4SymmetricKeyFromBytes(m.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create paseto key: %w", err)
	}

	parser := paseto.NewParser()
	parsed, err := parser.ParseV4Local(key, tokenStr, nil)
	if err != nil {
		return nil, errInvalidToken
	}

	claims := &AccessClaims{}

	if v, err := parsed.GetSubject(); err == nil {
		claims.Sub = v
	}
	if v, err := parsed.GetString("role"); err == nil {
		claims.Role = v
	}
	if v, err := parsed.GetString("token_type"); err == nil {
		claims.TokenType = v
	}
	if v, err := parsed.GetIssuedAt(); err == nil {
		claims.IssuedAt = v
	}
	if v, err := parsed.GetExpiration(); err == nil {
		claims.ExpiresAt = v
	}

	var companyID uint
	if err := parsed.Get("company_id", &companyID); err == nil {
		claims.CompanyID = &companyID
	}

	if claims.TokenType != "access" {
		return nil, errInvalidToken
	}

	return claims, nil
}

// GenerateRefreshToken creates a new opaque refresh token using crypto/rand.
func (m *Manager) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// HashRefreshToken returns the SHA-256 hex digest of a refresh token.
func (m *Manager) HashRefreshToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
