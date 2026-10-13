package assign_permissions_to_role

import (
	"context"
	"errors"
	"fmt"
	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"

	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
	"github.com/enviniom/nexokit/internal/modules/iam/queries"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	GetByPublicID(tc tenant.TenantContext, publicID string) (*iamrole.IAMRole, error)
	ListAllPermissions() ([]permission.IAMPermission, error)
	ReplacePermissions(tc tenant.TenantContext, roleID uint, permissionIDs []uint) error
	InvalidateRoleMemberPermissionCache(tc tenant.TenantContext, roleID uint, c cache.Cache) error
	ResolvePermissionSelection(catalog []permission.IAMPermission, slugs []string) ([]string, map[string]bool, []uint, error)
	BuildRolePermissionCatalog(catalog []permission.IAMPermission, granted map[string]bool) []iamrole.RolePermissionGroupResponse
}

type GormRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &GormRepository{db: db} }

func (r *GormRepository) GetByPublicID(tc tenant.TenantContext, publicID string) (*iamrole.IAMRole, error) {
	role, err := queries.GetRoleByPublicID(r.db, tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, core.ErrNotFound
		}
		return nil, err
	}
	return role, nil
}

func (r *GormRepository) ListAllPermissions() ([]permission.IAMPermission, error) {
	var items []permission.IAMPermission
	if err := r.db.Order("module ASC, display_order ASC, slug ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *GormRepository) ReplacePermissions(tc tenant.TenantContext, roleID uint, permissionIDs []uint) error {
	if !tc.IsRootScope && tc.CompanyID == 0 {
		return core.ErrNotFound
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var role iamrole.IAMRole
		if err := tenant.ApplyTenantScope(tx, tc).Clauses(clause.Locking{Strength: "UPDATE"}).First(&role, roleID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return core.ErrNotFound
			}
			return err
		}
		if err := tx.Exec("DELETE FROM role_permissions WHERE role_id = ?", roleID).Error; err != nil {
			return err
		}
		for _, permissionID := range permissionIDs {
			if err := tx.Table("role_permissions").Create(map[string]any{"role_id": roleID, "permission_id": permissionID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GormRepository) InvalidateRoleMemberPermissionCache(tc tenant.TenantContext, roleID uint, c cache.Cache) error {
	if c == nil {
		return nil
	}
	ids, err := r.ListRoleMemberPublicIDs(tc, roleID)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if err := c.Delete(context.Background(), fmt.Sprintf("rbac:permissions:%s", id)); err != nil {
			return err
		}
	}
	return nil
}

func (r *GormRepository) ListRoleMemberPublicIDs(tc tenant.TenantContext, roleID uint) ([]string, error) {
	if !tc.IsRootScope && tc.CompanyID == 0 {
		return nil, core.ErrNotFound
	}
	var ids []string
	if err := tenant.ApplyTenantScope(r.db.Model(&iamuser.IAMUser{}), tc).Where("role_id = ?", roleID).Pluck("public_id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *GormRepository) ResolvePermissionSelection(catalog []permission.IAMPermission, slugs []string) ([]string, map[string]bool, []uint, error) {
	bySlug := make(map[string]permission.IAMPermission, len(catalog))
	for _, p := range catalog {
		bySlug[p.Slug] = p
	}
	normalized := queries.NormalizeSlugs(slugs)
	selected, ids := map[string]bool{}, []uint{}
	for _, slug := range normalized {
		p, ok := bySlug[slug]
		if !ok {
			return nil, nil, nil, core.ErrInvalidPermissionSlug
		}
		if selected[slug] {
			continue
		}
		selected[slug] = true
		ids = append(ids, p.ID)
	}
	return normalized, selected, ids, nil
}

func (r *GormRepository) BuildRolePermissionCatalog(catalog []permission.IAMPermission, granted map[string]bool) []iamrole.RolePermissionGroupResponse {
	return iamrole.BuildPermissionCatalog(catalog, granted)
}
