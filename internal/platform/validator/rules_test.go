package validator

import (
	"fmt"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/messages"
)

func TestMinLength_Rune(t *testing.T) {
	rule := MinLength(3)
	// "añ" is 2 runes, not 2 bytes
	if msg := rule("añ"); msg == "" {
		t.Error("expected error for string shorter than 3 runes")
	}
	if msg := rule("año"); msg != "" {
		t.Errorf("unexpected error: %s", msg)
	}
	if msg := rule("abcd"); msg != "" {
		t.Errorf("unexpected error: %s", msg)
	}
}

func TestMaxLength_Rune(t *testing.T) {
	rule := MaxLength(3)
	if msg := rule("abc"); msg != "" {
		t.Errorf("unexpected error: %s", msg)
	}
	if msg := rule("año"); msg != "" {
		t.Errorf("unexpected error for 3 runes: %s", msg)
	}
	if msg := rule("abcd"); msg == "" {
		t.Error("expected error for string longer than 3 runes")
	}
}

func TestValidEmail(t *testing.T) {
	rule := ValidEmail()
	if msg := rule("valid@example.com"); msg != "" {
		t.Errorf("unexpected error: %s", msg)
	}
	if msg := rule("not-an-email"); msg == "" {
		t.Error("expected error for invalid email")
	}
	if msg := rule(""); msg == "" {
		t.Error("expected error for empty email")
	}
}

func TestHasUppercase(t *testing.T) {
	rule := HasUppercase()
	if msg := rule("Hello"); msg != "" {
		t.Errorf("unexpected error: %s", msg)
	}
	if msg := rule("hello"); msg == "" {
		t.Error("expected error for missing uppercase")
	}
	if msg := rule("1234"); msg == "" {
		t.Error("expected error for missing uppercase")
	}
}

func TestHasDigit(t *testing.T) {
	rule := HasDigit()
	if msg := rule("hello1"); msg != "" {
		t.Errorf("unexpected error: %s", msg)
	}
	if msg := rule("hello"); msg == "" {
		t.Error("expected error for missing digit")
	}
}

func TestHasSpecialChar(t *testing.T) {
	rule := HasSpecialChar()
	if msg := rule("hello!"); msg != "" {
		t.Errorf("unexpected error: %s", msg)
	}
	if msg := rule("hello"); msg == "" {
		t.Error("expected error for missing special character")
	}
	if msg := rule("hello1"); msg == "" {
		t.Error("expected error for missing special character")
	}
}

func TestMinWords(t *testing.T) {
	rule := MinWords(2)
	if msg := rule("John Doe"); msg != "" {
		t.Errorf("unexpected error: %s", msg)
	}
	if msg := rule("John"); msg == "" {
		t.Error("expected error for single word")
	}
	if msg := rule("One Two Three"); msg != "" {
		t.Errorf("unexpected error: %s", msg)
	}
}

func TestNoNumbers(t *testing.T) {
	rule := NoNumbers()
	if msg := rule("hello"); msg != "" {
		t.Errorf("unexpected error: %s", msg)
	}
	if msg := rule("hello1"); msg == "" {
		t.Error("expected error for string containing numbers")
	}
}

func TestMatches(t *testing.T) {
	rule := Matches(`^[a-z]+$`)
	if msg := rule("abc"); msg != "" {
		t.Errorf("unexpected error: %s", msg)
	}
	if msg := rule("ABC"); msg == "" {
		t.Error("expected error for non-matching pattern")
	}
	if msg := rule("ab1"); msg == "" {
		t.Error("expected error for non-matching pattern")
	}
}

func TestRuleMessages(t *testing.T) {
	if msg := MinLength(5)("abc"); msg != fmt.Sprintf(messages.MsgMinLength, 5) {
		t.Errorf("unexpected MinLength message: %s", msg)
	}
	if msg := MaxLength(3)("abcd"); msg != fmt.Sprintf(messages.MsgMaxLength, 3) {
		t.Errorf("unexpected MaxLength message: %s", msg)
	}
	if msg := ValidEmail()("bad"); msg != messages.MsgValidEmail {
		t.Errorf("unexpected ValidEmail message: %s", msg)
	}
	if msg := HasUppercase()("abc"); msg != messages.MsgHasUppercase {
		t.Errorf("unexpected HasUppercase message: %s", msg)
	}
	if msg := HasDigit()("abc"); msg != messages.MsgHasDigit {
		t.Errorf("unexpected HasDigit message: %s", msg)
	}
	if msg := HasSpecialChar()("abc"); msg != messages.MsgHasSpecialChar {
		t.Errorf("unexpected HasSpecialChar message: %s", msg)
	}
	if msg := MinWords(2)("one"); msg != fmt.Sprintf(messages.MsgMinWords, 2) {
		t.Errorf("unexpected MinWords message: %s", msg)
	}
	if msg := NoNumbers()("a1"); msg != messages.MsgNoNumbers {
		t.Errorf("unexpected NoNumbers message: %s", msg)
	}
	if msg := Matches(`^[a-z]+$`)("A"); msg != messages.MsgInvalidFormat {
		t.Errorf("unexpected Matches message: %s", msg)
	}
}
