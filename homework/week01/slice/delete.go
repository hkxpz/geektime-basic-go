package slice

import (
	"geektime-basic-go/homework/week01/slice/internal"
)

func Delete[T any](src []T, index int) ([]T, error) {
	res, _, err := internal.Delete(src, index)
	return res, err
}

func DeleteWithReduceCapacity[T any](src []T, index int) ([]T, error) {
	res, _, err := internal.Delete(src, index)
	if err != nil {
		return src, err
	}

	length, capacity := len(src), cap(src)
	if capacity >= 10 && length <= capacity/4 {
		return src[: length : capacity/2], nil
	}
	return res, nil
}
