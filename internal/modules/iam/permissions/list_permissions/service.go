package list_permissions

import (
	"sort"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
)

type Service interface {
	ListGrouped() ([]core.PermissionGroupResponse, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) ListGrouped() ([]core.PermissionGroupResponse, error) {
	items, err := s.repo.ListAllPermissions()
	if err != nil {
		return nil, err
	}
	return groupPermissions(items), nil
}

func groupPermissions(items []core.IAMPermission) []core.PermissionGroupResponse {
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
		groups[idx].Permissions = append(groups[idx].Permissions, *toPermissionResponse(&items[i]))
	}
	return groups
}
