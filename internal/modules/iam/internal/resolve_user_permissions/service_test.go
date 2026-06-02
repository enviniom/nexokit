package resolve_user_permissions

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

type fakeRepo struct{ slugs []string }

func (f fakeRepo) ListSlugsByUserPublicID(string) ([]string, error) { return f.slugs, nil }

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
