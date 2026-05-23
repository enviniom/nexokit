package permissions

import (
	"sort"
	"sync"
	"testing"
)

func TestFormat(t *testing.T) {
	cases := []struct {
		module string
		action string
		want   string
	}{
		{module: ModuleUsers, action: ActionList, want: "users.list"},
		{module: ModuleRoles, action: ActionSelect, want: "roles.select"},
		{module: ModulePermissions, action: ActionManage, want: "permissions.manage"},
	}

	for _, tc := range cases {
		if got := Format(tc.module, tc.action); got != tc.want {
			t.Errorf("Format(%q, %q) = %q, want %q", tc.module, tc.action, got, tc.want)
		}
	}
}

func TestRegistry(t *testing.T) {
	// Reset the registered map for clean testing
	mu.Lock()
	registered = make(map[string]bool)
	mu.Unlock()

	Register("") // Ignore empty
	Register("users.list")
	Register("users.list") // Duplicate
	Register("roles.view")

	got := ListRegistered()
	if len(got) != 2 {
		t.Fatalf("expected 2 unique registered permissions, got %d", len(got))
	}

	sort.Strings(got)
	if got[0] != "roles.view" || got[1] != "users.list" {
		t.Errorf("unexpected registered list: %v", got)
	}
}

func TestRegistryConcurrentSafety(t *testing.T) {
	// Reset
	mu.Lock()
	registered = make(map[string]bool)
	mu.Unlock()

	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(val int) {
			defer wg.Done()
			Register("users.list")
			Register("roles.view")
		}(i)
	}
	wg.Wait()

	got := ListRegistered()
	if len(got) != 2 {
		t.Errorf("expected 2 unique registered permissions under concurrency, got %d", len(got))
	}
}

func TestHumanizeAndOrder(t *testing.T) {
	if name := HumanizeName("users", ActionList); name != "List users" {
		t.Errorf("unexpected humanized name for list users: %q", name)
	}
	if name := HumanizeName("roles", ActionAssignPermissions); name != "Assign permissions roles" {
		t.Errorf("unexpected humanized name for assign_permissions roles: %q", name)
	}

	if desc := HumanizeDescription("users", ActionCreate); desc != "Allows creating users" {
		t.Errorf("unexpected humanized description for create users: %q", desc)
	}

	if order := DefaultDisplayOrder(ActionList); order != 10 {
		t.Errorf("unexpected display order for list action: %d", order)
	}
	if order := DefaultDisplayOrder("unknown"); order != 100 {
		t.Errorf("unexpected display order for unknown action: %d", order)
	}
}
