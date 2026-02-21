package utilities

import "errors"

var (
	ErrOutOfBounds = errors.New("DeleteFromSlice: out of bounds")
)

func FindInSlice[T comparable](slice []T, searchedValue T) (int, bool) {
	for i, v := range slice {
		if v == searchedValue {
			return i, true
		}
	}
	return 0, false
}

func DeleteFromSlice[T any](slice []T, pos int) (*[]T, error) {
	if pos >= len(slice) || pos < 0 {
		return nil, ErrOutOfBounds
	}

	retSlice := make([]T, len(slice) - 1)

	for i, v := range slice {
		if i < pos {
			retSlice[i] = v
		} else if i > pos {
			retSlice[i-1] = v
		}
	}

	return &retSlice, nil
}
