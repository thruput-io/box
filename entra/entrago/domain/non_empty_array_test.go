package domain_test

import (
	"errors"
	"reflect"
	"testing"

	"identity/domain"
)

func TestNewNonEmptyArray_EmptyReturnsError(t *testing.T) {
	t.Parallel()

	err, ok := domain.NewNonEmptyArray[int]().Left()
	if !ok {
		t.Fatal(expectedError)
	}

	if !errors.Is(err, domain.ErrNonEmptyArrayEmpty) {
		t.Fatalf(expectedFormat, domain.ErrNonEmptyArrayEmpty, err)
	}
}

func TestNewNonEmptyArray_StoresACopyOfInput(t *testing.T) {
	t.Parallel()

	input := []int{1, 2}
	arr := domain.NewNonEmptyArray(input...).MustRight()
	input[0] = 99

	if !reflect.DeepEqual(arr.Items(), []int{1, 2}) {
		t.Fatalf("items=%v", arr.Items())
	}
}

func TestNonEmptyArray_ItemsReturnsCopy(t *testing.T) {
	t.Parallel()

	arr := domain.NewNonEmptyArray(1, 2).MustRight()
	items := arr.Items()
	items[0] = 99

	if !reflect.DeepEqual(arr.Items(), []int{1, 2}) {
		t.Fatalf("items=%v", arr.Items())
	}
}
