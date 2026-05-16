package identity

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
)

// Generate creates a new ULID public ID.
func Generate() (string, error) {
	id, err := ulid.New(ulid.Timestamp(time.Now().UTC()), rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate ulid: %w", err)
	}
	return id.String(), nil
}
