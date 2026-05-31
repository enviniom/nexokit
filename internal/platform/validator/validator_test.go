package validator

import (
	"testing"

	"github.com/enviniom/nexokit/internal/platform/messages"
)

func TestRequired(t *testing.T) {
	errs := make(ValidationErrors)
	Field(errs, "email", "").Required()

	if !errs.HasErrors() {
		t.Fatal("expected error for empty required field")
	}
	if errs["email"][0] != messages.MsgRequired {
		t.Errorf("unexpected message: %s", errs["email"][0])
	}
}

func TestRequired_NonEmpty(t *testing.T) {
	errs := make(ValidationErrors)
	Field(errs, "email", "test@example.com").Required()

	if errs.HasErrors() {
		t.Fatal("expected no errors for non-empty required field")
	}
}

func TestOptional_SkipsRules(t *testing.T) {
	errs := make(ValidationErrors)
	Field(errs, "bio", "").Optional().Apply(MinLength(10))

	if errs.HasErrors() {
		t.Fatal("expected no errors for empty optional field")
	}
}

func TestOptional_AppliesWhenPresent(t *testing.T) {
	errs := make(ValidationErrors)
	Field(errs, "bio", "hi").Optional().Apply(MinLength(10))

	if !errs.HasErrors() {
		t.Fatal("expected error for short optional field")
	}
}

func TestOptional_SkipsInList(t *testing.T) {
	errs := make(ValidationErrors)
	Field(errs, "status", "").Optional().Apply(InList("active", "inactive"))

	if errs.HasErrors() {
		t.Fatalf("expected no errors for empty optional InList, got %#v", errs)
	}
}

func TestApply_SkipAfterRequiredFailure(t *testing.T) {
	errs := make(ValidationErrors)
	Field(errs, "name", "").Required().Apply(MinLength(3))

	if len(errs["name"]) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs["name"]))
	}
	if errs["name"][0] != messages.MsgRequired {
		t.Errorf("unexpected message: %s", errs["name"][0])
	}
}

func TestValidationErrors_Add(t *testing.T) {
	errs := make(ValidationErrors)
	errs.Add("field1", "error one")
	errs.Add("field1", "error two")
	errs.Add("field2", "error three")

	if len(errs["field1"]) != 2 {
		t.Errorf("expected 2 errors for field1, got %d", len(errs["field1"]))
	}
	if len(errs["field2"]) != 1 {
		t.Errorf("expected 1 error for field2, got %d", len(errs["field2"]))
	}
}

func TestValidationErrors_HasErrors(t *testing.T) {
	errs := make(ValidationErrors)
	if errs.HasErrors() {
		t.Error("expected no errors on empty map")
	}
	errs.Add("x", "y")
	if !errs.HasErrors() {
		t.Error("expected errors after Add")
	}
}
