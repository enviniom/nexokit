package identity

import (
	"testing"
)

func TestGenerate(t *testing.T) {
	t.Run("generates 26 character string", func(t *testing.T) {
		id, err := Generate()
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		if len(id) != 26 {
			t.Errorf("expected length 26, got %d", len(id))
		}
	})

	t.Run("generates unique ids", func(t *testing.T) {
		id1, err := Generate()
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		id2, err := Generate()
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		if id1 == id2 {
			t.Error("expected unique ids")
		}
	})

	t.Run("generates sortable ids", func(t *testing.T) {
		// ULIDs are lexicographically sortable by time.
		// Two consecutive generations should produce ids where id2 > id1.
		id1, err := Generate()
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		id2, err := Generate()
		if err != nil {
			t.Fatalf("Generate failed: %v", err)
		}
		if id2 <= id1 {
			t.Errorf("expected id2 > id1 for sortability, got id1=%s id2=%s", id1, id2)
		}
	})
}
