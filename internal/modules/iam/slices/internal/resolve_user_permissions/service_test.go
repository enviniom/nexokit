package resolve_user_permissions

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

type fakeRepo struct {
	slugs []string
	err   error
}

func (f fakeRepo) ListSlugsByUserPublicID(string) ([]string, error) { return f.slugs, f.err }

type fakeCache struct {
	values map[string][]byte
	sets   map[string][]byte
	ttl    map[string]time.Duration
}

func (f *fakeCache) Get(context.Context, string) ([]byte, error) {
	return f.values["rbac:permissions:u1"], nil
}
func (f *fakeCache) Set(_ context.Context, key string, value []byte, d time.Duration) error {
	if f.sets == nil {
		f.sets, f.ttl = map[string][]byte{}, map[string]time.Duration{}
	}
	f.sets[key] = value
	f.ttl[key] = d
	return nil
}
func (f *fakeCache) Delete(context.Context, string) error         { return nil }
func (f *fakeCache) Exists(context.Context, string) (bool, error) { return false, nil }
func (f *fakeCache) Close() error                                 { return nil }

func TestResolveUsesCacheAndStoresTTL(t *testing.T) {
	c := &fakeCache{}
	svc := NewService(fakeRepo{slugs: []string{"a", "b"}}, c)
	r, err := svc.Resolve("u1")
	if err != nil || len(r) != 2 || c.ttl["rbac:permissions:u1"] != 5*time.Minute {
		t.Fatalf("unexpected: %v %v", r, err)
	}
	var stored []string
	_ = json.Unmarshal(c.sets["rbac:permissions:u1"], &stored)
	if len(stored) != 2 {
		t.Fatalf("unexpected payload: %v", stored)
	}
}

func TestResolvePropagatesRepositoryError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := NewService(fakeRepo{err: repoErr}, &fakeCache{})
	_, err := svc.Resolve("u1")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}

func TestResolveReturnsCachedSlugsWithoutCallingRepo(t *testing.T) {
	cached := []string{"cached.read"}
	payload, _ := json.Marshal(cached)
	cache := &fakeCache{values: map[string][]byte{"rbac:permissions:u1": payload}}
	svc := NewService(fakeRepo{slugs: []string{"repo.read"}}, cache)
	got, err := svc.Resolve("u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != "cached.read" {
		t.Fatalf("expected cached slugs, got %v", got)
	}
	if cache.sets != nil {
		t.Fatalf("expected cache Set to be skipped on hit, got %v", cache.sets)
	}
}
