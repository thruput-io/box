package domain

import (
	"github.com/samber/mo"
)

type NonEmptyArray[T any] struct {
	items []T
}

func NewNonEmptyArray[T any](elements ...T) mo.Either[Error, NonEmptyArray[T]] {
	if len(elements) == emptyLen {
		return mo.Left[Error, NonEmptyArray[T]](ErrNonEmptyArrayEmpty)
	}

	copied := append([]T(nil), elements...)

	return mo.Right[Error](NonEmptyArray[T]{items: copied})
}

func (n NonEmptyArray[T]) Items() []T {
	return append([]T(nil), n.items...)
}
