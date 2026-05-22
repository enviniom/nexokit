package roles

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/permissions"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/identity"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"gorm.io/gorm"
)

// Service defines the business logic contract for roles.
type Service interface {
	List(tc tenant.TenantContext, page, perPage int) ([]RoleResponse, int64, error)
	GetByPublicID(tc tenant.TenantContext, publicID string) (*RoleResponse, error)
	Create(tc tenant.TenantContext, req CreateRoleRequest) (*RoleResponse, error)
	Update(tc tenant.TenantContext, publicID string, req UpdateRoleRequest) (*RoleResponse, error)
	Delete(tc tenant.TenantContext, publicID string) error
	GetPermissionCatalog(tc tenant.TenantContext, publicID string) ([]RolePermissionGroupResponse, error)
	AssignPermissions(tc tenant.TenantContext, publicID string, req AssignRolePermissionsRequest, actorPermissions []string) (*RolePermissionAssignmentResponse, error)
}

type permissionCatalogRepository interface {
	ListAll() ([]permissions.Permission, error)
}

type roleMemberRepository interface {
	ListPublicIDsByRoleID(roleID uint) ([]string, error)
	CountByRoleID(roleID uint) (int64, error)
}

// ServiceOption configures optional role service collaborators.
type ServiceOption func(*roleService)

// roleService is the concrete implementation of Service.
type roleService struct {
	repo              Repository
	permissionCatalog permissionCatalogRepository
	roleMembers       roleMemberRepository
	cache             cache.Cache
}

// NewService creates a new roles service.
func NewService(repo Repository, opts ...ServiceOption) Service {
	s := &roleService{repo: repo}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// WithPermissionCatalog configures the permission catalog dependency.
func WithPermissionCatalog(repo permissionCatalogRepository) ServiceOption {
	return func(s *roleService) { s.permissionCatalog = repo }
}

// WithRoleMembers configures the role-member lookup dependency.
func WithRoleMembers(repo roleMemberRepository) ServiceOption {
	return func(s *roleService) { s.roleMembers = repo }
}

// WithCache configures cache invalidation for role permission assignments.
func WithCache(c cache.Cache) ServiceOption {
	return func(s *roleService) { s.cache = c }
}

// List returns paginated roles as DTOs.
func (s *roleService) List(tc tenant.TenantContext, page, perPage int) ([]RoleResponse, int64, error) {
	roles, err := s.repo.List(tc, page, perPage)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.Count(tc)
	if err != nil {
		return nil, 0, err
	}

	result := make([]RoleResponse, len(roles))
	for i, r := range roles {
		result[i] = *toResponse(&r)
	}

	return result, total, nil
}

// GetByPublicID returns a single role by public ID.
func (s *roleService) GetByPublicID(tc tenant.TenantContext, publicID string) (*RoleResponse, error) {
	role, err := s.repo.GetByPublicID(tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	return toResponse(role), nil
}

// Create creates a new role after checking name uniqueness.
func (s *roleService) Create(tc tenant.TenantContext, req CreateRoleRequest) (*RoleResponse, error) {
	if isReservedIdentity(req.Name, req.Slug) {
		return nil, apperror.ErrConflict
	}

	if exists, err := s.repo.ExistsByName(tc, req.Name); err != nil {
		return nil, err
	} else if exists {
		return nil, apperror.ErrConflict
	}

	if exists, err := s.repo.ExistsBySlug(tc, req.Slug); err != nil {
		return nil, err
	} else if exists {
		return nil, apperror.ErrConflict
	}

	publicID, err := identity.Generate()
	if err != nil {
		return nil, err
	}

	role := &Role{
		BaseModel: shared.BaseModel{
			PublicID: publicID,
		},
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		IsSystem:    false,
	}

	if !tc.IsRootScope {
		companyID := tc.CompanyID
		role.CompanyID = &companyID
	}

	if err := s.repo.Create(role); err != nil {
		if isUniqueConstraintError(err) {
			return nil, apperror.ErrConflict
		}
		return nil, err
	}

	return toResponse(role), nil
}

// Update updates a role if it is not a system role and the new name is unique.
func (s *roleService) Update(tc tenant.TenantContext, publicID string, req UpdateRoleRequest) (*RoleResponse, error) {
	role, err := s.repo.GetByPublicID(tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}

	if role.IsSystem {
		return nil, apperror.ErrForbidden
	}
	if isReservedIdentity(role.Name, role.Slug) || isReservedIdentity(req.Name, req.Slug) {
		return nil, apperror.ErrForbidden
	}

	if role.Name != req.Name {
		exists, err := s.repo.ExistsByName(tc, req.Name)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.ErrConflict
		}
	}
	if role.Slug != req.Slug {
		exists, err := s.repo.ExistsBySlug(tc, req.Slug)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.ErrConflict
		}
	}

	role.Name = req.Name
	role.Slug = req.Slug
	role.Description = req.Description

	if err := s.repo.Update(role); err != nil {
		if isUniqueConstraintError(err) {
			return nil, apperror.ErrConflict
		}
		return nil, err
	}

	return toResponse(role), nil
}

func isReservedIdentity(name, slug string) bool {
	normalizedName := strings.TrimSpace(name)
	normalizedSlug := strings.TrimSpace(slug)
	for _, reserved := range ReservedSlugs {
		if strings.EqualFold(normalizedName, reserved) || strings.EqualFold(normalizedSlug, reserved) {
			return true
		}
	}
	return false
}

// Delete deletes a role if it is not a system role.
func (s *roleService) Delete(tc tenant.TenantContext, publicID string) error {
	role, err := s.repo.GetByPublicID(tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.ErrNotFound
		}
		return err
	}

	if role.IsSystem {
		return apperror.ErrForbidden
	}
	if isReservedIdentity(role.Name, role.Slug) {
		return apperror.ErrForbidden
	}

	if s.roleMembers != nil {
		count, err := s.roleMembers.CountByRoleID(role.ID)
		if err != nil {
			return err
		}
		if count > 0 {
			return apperror.ErrUnprocessable
		}
	}

	return s.repo.Delete(tc, publicID)
}

// GetPermissionCatalog returns the full permission catalog annotated with role grants.
func (s *roleService) GetPermissionCatalog(tc tenant.TenantContext, publicID string) ([]RolePermissionGroupResponse, error) {
	role, err := s.repo.GetByPublicID(tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	if s.permissionCatalog == nil {
		return nil, fmt.Errorf("permission catalog repository is not configured")
	}
	items, err := s.permissionCatalog.ListAll()
	if err != nil {
		return nil, err
	}
	return buildRolePermissionCatalog(items, grantedSlugSet(role.Permissions)), nil
}

// AssignPermissions replaces role permissions by slug, protects system roles, and invalidates member caches.
func (s *roleService) AssignPermissions(tc tenant.TenantContext, publicID string, req AssignRolePermissionsRequest, actorPermissions []string) (*RolePermissionAssignmentResponse, error) {
	if !hasPermission(actorPermissions, "roles.assign_permissions") {
		return nil, apperror.ErrForbidden
	}
	role, err := s.repo.GetByPublicID(tc, publicID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, err
	}
	if s.permissionCatalog == nil {
		return nil, fmt.Errorf("permission catalog repository is not configured")
	}
	items, err := s.permissionCatalog.ListAll()
	if err != nil {
		return nil, err
	}
	selected, ids, err := resolvePermissionSelection(items, req.Permissions)
	if err != nil {
		return nil, err
	}
	if role.IsSystem && removesSystemPermission(role.Permissions, selected) {
		return nil, apperror.ErrForbidden
	}
	if err := s.repo.ReplacePermissions(role.ID, ids); err != nil {
		return nil, err
	}
	if err := s.invalidateRoleMemberCaches(role.ID); err != nil {
		return nil, err
	}
	catalog := buildRolePermissionCatalog(items, selected)
	return &RolePermissionAssignmentResponse{RoleID: role.PublicID, Permissions: normalizedSlugs(req.Permissions), Catalog: catalog}, nil
}

func isUniqueConstraintError(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint")
}

func toResponse(r *Role) *RoleResponse {
	permissionSlugs := make([]string, 0, len(r.Permissions))
	for _, permission := range r.Permissions {
		permissionSlugs = append(permissionSlugs, permission.Slug)
	}
	sort.Strings(permissionSlugs)
	return &RoleResponse{
		PublicID:    r.PublicID,
		CompanyID:   r.CompanyID,
		Name:        r.Name,
		Slug:        r.Slug,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		Permissions: permissionSlugs,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
		CreatedBy:   r.CreatedBy,
		UpdatedBy:   r.UpdatedBy,
	}
}

func grantedSlugSet(items []permissions.Permission) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item.Slug] = true
	}
	return set
}

func buildRolePermissionCatalog(items []permissions.Permission, granted map[string]bool) []RolePermissionGroupResponse {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Module != items[j].Module {
			return items[i].Module < items[j].Module
		}
		if items[i].DisplayOrder != items[j].DisplayOrder {
			return items[i].DisplayOrder < items[j].DisplayOrder
		}
		return items[i].Slug < items[j].Slug
	})

	groups := make([]RolePermissionGroupResponse, 0)
	moduleIndex := make(map[string]int)
	for i := range items {
		idx, ok := moduleIndex[items[i].Module]
		if !ok {
			groups = append(groups, RolePermissionGroupResponse{Module: items[i].Module})
			idx = len(groups) - 1
			moduleIndex[items[i].Module] = idx
		}
		groups[idx].Permissions = append(groups[idx].Permissions, RolePermissionResponse{
			PublicID:     items[i].PublicID,
			Slug:         items[i].Slug,
			Name:         items[i].Name,
			Module:       items[i].Module,
			Action:       items[i].Action,
			Description:  items[i].Description,
			IsSystem:     items[i].IsSystem,
			DisplayOrder: items[i].DisplayOrder,
			Granted:      granted[items[i].Slug],
		})
	}
	return groups
}

func hasPermission(items []string, slug string) bool {
	for _, item := range items {
		if item == slug {
			return true
		}
	}
	return false
}

func resolvePermissionSelection(items []permissions.Permission, slugs []string) (map[string]bool, []uint, error) {
	bySlug := make(map[string]permissions.Permission, len(items))
	for _, item := range items {
		bySlug[item.Slug] = item
	}
	selected := make(map[string]bool, len(slugs))
	ids := make([]uint, 0, len(slugs))
	for _, slug := range normalizedSlugs(slugs) {
		item, ok := bySlug[slug]
		if !ok {
			return nil, nil, apperror.ErrBadRequest
		}
		if selected[slug] {
			continue
		}
		selected[slug] = true
		ids = append(ids, item.ID)
	}
	return selected, ids, nil
}

func normalizedSlugs(slugs []string) []string {
	normalized := make([]string, 0, len(slugs))
	seen := make(map[string]bool, len(slugs))
	for _, slug := range slugs {
		slug = strings.TrimSpace(slug)
		if slug == "" || seen[slug] {
			continue
		}
		seen[slug] = true
		normalized = append(normalized, slug)
	}
	return normalized
}

func removesSystemPermission(existing []permissions.Permission, selected map[string]bool) bool {
	for _, item := range existing {
		if item.IsSystem && !selected[item.Slug] {
			return true
		}
	}
	return false
}

func (s *roleService) invalidateRoleMemberCaches(roleID uint) error {
	if s.roleMembers == nil || s.cache == nil {
		return nil
	}
	publicIDs, err := s.roleMembers.ListPublicIDsByRoleID(roleID)
	if err != nil {
		return err
	}
	for _, publicID := range publicIDs {
		if err := s.cache.Delete(context.Background(), fmt.Sprintf("rbac:permissions:%s", publicID)); err != nil {
			return err
		}
	}
	return nil
}
