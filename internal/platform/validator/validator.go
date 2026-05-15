package validator

import (
	"github.com/enviniom/nexokit/internal/platform/messages"
	"github.com/enviniom/nexokit/internal/platform/response"
)

// Rule is a validation function that receives a field value and returns
// an error message or an empty string if the value is valid.
type Rule func(value string) string

// FieldValidator chains rules for a single field.
type FieldValidator struct {
	field string
	value string
	skip  bool
	errs  response.ValidationErrors
}

// Field initiates validation for a field.
func Field(errs response.ValidationErrors, field, value string) *FieldValidator {
	return &FieldValidator{
		field: field,
		value: value,
		errs:  errs,
	}
}

// Required fails if the field value is empty.
func (fv *FieldValidator) Required() *FieldValidator {
	if fv.value == "" {
		fv.errs.Add(fv.field, messages.MsgRequired)
		fv.skip = true
	}
	return fv
}

// Optional skips subsequent rules if the field value is empty.
func (fv *FieldValidator) Optional() *FieldValidator {
	if fv.value == "" {
		fv.skip = true
	}
	return fv
}

// Apply executes a rule unless skip is active.
func (fv *FieldValidator) Apply(rule Rule) *FieldValidator {
	if fv.skip {
		return fv
	}
	if msg := rule(fv.value); msg != "" {
		fv.errs.Add(fv.field, msg)
	}
	return fv
}
