package auth

import (
	"time"

	"github.com/enviniom/nexokit/internal/modules/auth/core"
	"github.com/enviniom/nexokit/internal/modules/auth/slices/authenticate_user"
	"github.com/enviniom/nexokit/internal/modules/auth/slices/revoke_token"
	"github.com/enviniom/nexokit/internal/modules/auth/slices/rotate_token"
	"github.com/enviniom/nexokit/internal/modules/auth/slices/view_session"
	"gorm.io/gorm"
)

type Container struct {
	AuthenticateUser *authenticate_user.Handler
	RotateToken      *rotate_token.Handler
	RevokeToken      *revoke_token.Handler
	ViewSession      *view_session.Handler
}

func NewContainer(db *gorm.DB, verifier core.PasswordVerifier, tokenManager core.TokenManager, refreshTTL time.Duration) *Container {
	authenticateRepository := authenticate_user.NewRepository(db)
	authenticateService := authenticate_user.NewService(authenticateRepository, verifier, tokenManager, refreshTTL)
	rotateRepository := rotate_token.NewRepository(db)
	rotateService := rotate_token.NewService(rotateRepository, tokenManager, refreshTTL)
	revokeRepository := revoke_token.NewRepository(db)
	revokeService := revoke_token.NewService(revokeRepository, tokenManager)
	viewSessionRepository := view_session.NewRepository()
	viewSessionService := view_session.NewService(viewSessionRepository)

	return &Container{
		AuthenticateUser: authenticate_user.NewHandler(authenticateService),
		RotateToken:      rotate_token.NewHandler(rotateService),
		RevokeToken:      revoke_token.NewHandler(revokeService),
		ViewSession:      view_session.NewHandler(viewSessionService),
	}
}
