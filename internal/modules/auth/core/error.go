package core

import "errors"

var (
	// ErrInvalidCredentials represents a generic auth failure.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInvalidRefreshToken represents an invalid or unusable refresh token.
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)
