package commands

import (
	"errors"

	"github.com/enviniom/nexokit/internal/cli/root"
	iamcore "github.com/enviniom/nexokit/internal/modules/iam/core"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
	"github.com/enviniom/nexokit/internal/platform/identity"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

// rootStorage implements root.RootStorage using the application's database.
type rootStorage struct {
	db *gorm.DB
}

// newRootStorage creates a new root storage backed by the given database.
func newRootStorage(db *gorm.DB) root.RootStorage {
	return &rootStorage{db: db}
}

// RootExists returns true if a user with the root role already exists.
func (s *rootStorage) RootExists() (bool, error) {
	var rootRole iamrole.IAMRole
	if err := s.db.Where("slug = ?", iamcore.RootRoleSlug).First(&rootRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	var count int64
	if err := s.db.Model(&iamuser.IAMUser{}).Where("role_id = ?", rootRole.ID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateRoot persists a new root user linked to the root role.
func (s *rootStorage) CreateRoot(name, email, passwordHash string) error {
	var rootRole iamrole.IAMRole
	if err := s.db.Where("slug = ?", iamcore.RootRoleSlug).First(&rootRole).Error; err != nil {
		return err
	}

	publicID, err := identity.Generate()
	if err != nil {
		return err
	}

	user := &iamuser.IAMUser{
		BaseModel: shared.BaseModel{
			PublicID: publicID,
		},
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		RoleID:       rootRole.ID,
		IsActive:     true,
	}
	return s.db.Create(user).Error
}
