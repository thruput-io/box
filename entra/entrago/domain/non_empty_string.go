package domain

import (
	"github.com/samber/mo"
)

// NonEmptyString is an opaque string value object that guarantees the underlying value is not empty.
//
// Note: Like most Go value types, the zero value is presentable but invalid; always construct via
// NewNonEmptyString (or domain-specific constructors that use it).
type NonEmptyString struct {
	value string
}

// NewNonEmptyString constructs a NonEmptyString, returning an error if the input is empty.
func NewNonEmptyString(raw string) mo.Either[Error, NonEmptyString] {
	if raw == emptyString {
		return mo.Left[Error, NonEmptyString](errNonEmptyStringEmpty)
	}

	return mo.Right[Error](NonEmptyString{value: raw})
}

// Value returns the underlying string value.
func (nes NonEmptyString) Value() string {
	return nes.value
}
