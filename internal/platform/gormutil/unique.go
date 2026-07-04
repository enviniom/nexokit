package gormutil

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// IsUniqueConstraintError returns true when err indicates a unique-constraint
// violation across the supported GORM drivers. It detects gorm.ErrDuplicatedKey
// and Postgres/SQLite-style duplicate or unique-constraint messages.
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}

	msg := strings.ToLower(err.Error())
	// Keep matching specific to unique-constraint violations. The bare
	// "constraint failed" substring is intentionally excluded because it also
	// matches foreign-key, check, and not-null failures on SQLite.
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "unique failed")
}
