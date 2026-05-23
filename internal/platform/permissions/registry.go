package permissions

import "sync"

var (
	mu         sync.RWMutex
	registered = make(map[string]bool)
)

// Register records a permission slug thread-safely in the in-memory registry.
func Register(slug string) {
	if slug == "" {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	registered[slug] = true
}

// ListRegistered returns all unique registered permission slugs as a slice.
func ListRegistered() []string {
	mu.RLock()
	defer mu.RUnlock()
	slugs := make([]string, 0, len(registered))
	for slug := range registered {
		slugs = append(slugs, slug)
	}
	return slugs
}
