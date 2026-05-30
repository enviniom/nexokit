package resolve_permissions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/enviniom/nexokit/internal/infra/cache"
)

type Service interface {
	Resolve(publicID string) ([]string, error)
}

type service struct {
	repo  Repository
	cache cache.Cache
}

func NewService(repo Repository, c cache.Cache) Service {
	return &service{repo: repo, cache: c}
}

func (s *service) Resolve(publicID string) ([]string, error) {
	key := fmt.Sprintf("rbac:permissions:%s", publicID)
	if s.cache != nil {
		cached, err := s.cache.Get(context.Background(), key)
		if err != nil {
			return nil, err
		}
		if len(cached) > 0 {
			var slugs []string
			if err := json.Unmarshal(cached, &slugs); err != nil {
				return nil, err
			}
			return slugs, nil
		}
	}

	slugs, err := s.repo.ListSlugsByUserPublicID(publicID)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		payload, err := json.Marshal(slugs)
		if err != nil {
			return nil, err
		}
		if err := s.cache.Set(context.Background(), key, payload, 5*time.Minute); err != nil {
			return nil, err
		}
	}

	return slugs, nil
}
