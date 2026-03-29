package domain

import (
	"encoding/json"
)

// NonEmptyString is an opaque string value object that guarantees the underlying value is not empty.
//
// Note: Like most Go value types, the zero value is presentable but invalid; always construct via
// NewNonEmptyString (or domain-specific constructors that use it).
type NonEmptyString struct {
	value string
}

// NewNonEmptyString constructs a NonEmptyString, returning an error if the input is empty.
func NewNonEmptyString(raw string) (NonEmptyString, error) {
	if raw == emptyString {
		return NonEmptyString{}, errNonEmptyStringEmpty
	}

	return NonEmptyString{value: raw}, nil
}

// MustNonEmptyString constructs a NonEmptyString, panicking if invalid. For use in tests and constants only.
func MustNonEmptyString(raw string) NonEmptyString {
	v, err := NewNonEmptyString(raw)
	if err != nil {
		panic(err)
	}

	return v
}

func (nes NonEmptyString) MarshalJSON() ([]byte, error) {
	return json.Marshal(nes.value)
}

func (nes NonEmptyString) RawString() string {
	return nes.value
}
