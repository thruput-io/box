package domain

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

// MarshalJSON implements json.Marshaler.

func (nes NonEmptyString) rawCallback(callback func(string) error) error {
	return callback(nes.value)
}
