package list_permissions

import (
	"sort"

	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/modules/permissions/core"
)

// Service defines the business logic contract for permissions.
type Service interface {
	ListGrouped() ([]core.PermissionGroupResponse, error)
}

// permissionService is the concrete implementation of Service.
type permissionService struct {
	repo  Repository
	cache cache.Cache
}

// ServiceOption configures optional permission service collaborators.
type ServiceOption func(*permissionService)

// NewService creates a new permissions service.
func NewService(repo Repository, opts ...ServiceOption) Service {
	s := &permissionService{repo: repo}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// WithCache configures cache-backed permission resolution.
func WithCache(c cache.Cache) ServiceOption {
	return func(s *permissionService) { s.cache = c }
}

// ListGrouped returns all permissions grouped by module and sorted for rendering.
func (s *permissionService) ListGrouped() ([]core.PermissionGroupResponse, error) {
	items, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}
	return groupPermissions(items), nil
}

func groupPermissions(items []core.Permission) []core.PermissionGroupResponse {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Module != items[j].Module {
			return items[i].Module < items[j].Module
		}
		if items[i].DisplayOrder != items[j].DisplayOrder {
			return items[i].DisplayOrder < items[j].DisplayOrder
		}
		return items[i].Slug < items[j].Slug
	})

	groups := make([]core.PermissionGroupResponse, 0)
	moduleIndex := make(map[string]int)
	for i := range items {
		idx, ok := moduleIndex[items[i].Module]
		if !ok {
			groups = append(groups, core.PermissionGroupResponse{Module: items[i].Module})
			idx = len(groups) - 1
			moduleIndex[items[i].Module] = idx
		}
		groups[idx].Permissions = append(groups[idx].Permissions, *core.ToResponse(&items[i]))
	}
	return groups
}
