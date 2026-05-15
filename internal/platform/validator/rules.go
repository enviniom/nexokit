package validator

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode"

	"github.com/enviniom/nexokit/internal/platform/messages"
)

// MinLength returns a rule that validates the minimum number of runes.
func MinLength(n int) Rule {
	return func(v string) string {
		if len([]rune(v)) < n {
			return fmt.Sprintf(messages.MsgMinLength, n)
		}
		return ""
	}
}

// MaxLength returns a rule that validates the maximum number of runes.
func MaxLength(n int) Rule {
	return func(v string) string {
		if len([]rune(v)) > n {
			return fmt.Sprintf(messages.MsgMaxLength, n)
		}
		return ""
	}
}

// ValidEmail returns a rule that validates an email address.
func ValidEmail() Rule {
	return func(v string) string {
		_, err := mail.ParseAddress(v)
		if err != nil {
			return messages.MsgValidEmail
		}
		return ""
	}
}

// HasUppercase returns a rule that requires at least one uppercase letter.
func HasUppercase() Rule {
	return func(v string) string {
		for _, ch := range v {
			if unicode.IsUpper(ch) {
				return ""
			}
		}
		return messages.MsgHasUppercase
	}
}

// HasDigit returns a rule that requires at least one digit.
func HasDigit() Rule {
	return func(v string) string {
		for _, ch := range v {
			if unicode.IsDigit(ch) {
				return ""
			}
		}
		return messages.MsgHasDigit
	}
}

// HasSpecialChar returns a rule that requires at least one special character.
func HasSpecialChar() Rule {
	return func(v string) string {
		for _, ch := range v {
			if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
				return ""
			}
		}
		return messages.MsgHasSpecialChar
	}
}

// MinWords returns a rule that validates a minimum number of words.
func MinWords(n int) Rule {
	return func(v string) string {
		if len(strings.Fields(v)) < n {
			return fmt.Sprintf(messages.MsgMinWords, n)
		}
		return ""
	}
}

// NoNumbers returns a rule that rejects any digit in the value.
func NoNumbers() Rule {
	return func(v string) string {
		for _, ch := range v {
			if unicode.IsDigit(ch) {
				return messages.MsgNoNumbers
			}
		}
		return ""
	}
}

// Matches returns a rule that validates a value against a regular expression.
func Matches(pattern string) Rule {
	re := regexp.MustCompile(pattern)
	return func(v string) string {
		if !re.MatchString(v) {
			return messages.MsgInvalidFormat
		}
		return ""
	}
}
