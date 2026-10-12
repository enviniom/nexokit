package list_permissions

import (
	"github.com/enviniom/nexokit/internal/modules/iam/core/permission"
	"sort"
)

type Service interface {
	ListGrouped() ([]permission.PermissionGroupResponse, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) ListGrouped() ([]permission.PermissionGroupResponse, error) {
	items, err := s.repo.ListAllPermissions()
	if err != nil {
		return nil, err
	}
	return groupPermissions(items), nil
}

func groupPermissions(items []permission.IAMPermission) []permission.PermissionGroupResponse {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Module != items[j].Module {
			return items[i].Module < items[j].Module
		}
		if items[i].DisplayOrder != items[j].DisplayOrder {
			return items[i].DisplayOrder < items[j].DisplayOrder
		}
		return items[i].Slug < items[j].Slug
	})

	groups := make([]permission.PermissionGroupResponse, 0)
	moduleIndex := make(map[string]int)
	for i := range items {
		idx, ok := moduleIndex[items[i].Module]
		if !ok {
			groups = append(groups, permission.PermissionGroupResponse{Module: items[i].Module})
			idx = len(groups) - 1
			moduleIndex[items[i].Module] = idx
		}
		var response permission.PermissionResponse
		response.FromModel(items[i])
		groups[idx].Permissions = append(groups[idx].Permissions, response)
	}
	return groups
}
