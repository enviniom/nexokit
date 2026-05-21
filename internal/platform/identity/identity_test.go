package identity

import (
	"slices"
	"testing"

	"github.com/oklog/ulid/v2"
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
		const count = 64
		ids := make([]string, 0, count)
		for range count {
			id, err := Generate()
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}
			ids = append(ids, id)
		}

		sorted := append([]string(nil), ids...)
		slices.Sort(sorted)

		var previous uint64
		for i, raw := range sorted {
			parsed, err := ulid.Parse(raw)
			if err != nil {
				t.Fatalf("generated id is not a valid ULID: %v", err)
			}
			ts := parsed.Time()
			if i > 0 && ts < previous {
				t.Fatalf("lexicographic order broke timestamp monotonicity: prev=%d curr=%d (id=%s)", previous, ts, raw)
			}
			previous = ts
		}
	})
}
