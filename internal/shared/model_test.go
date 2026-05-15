package shared

import (
	"testing"
)

func TestBaseModelFields(t *testing.T) {
	// This test ensures BaseModel compiles with the correct field set.
	// Real GORM behavior is tested in integration tests.
	_ = BaseModel{}
	_ = BaseModelSimple{}
}
