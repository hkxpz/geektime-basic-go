package internal

func Delete[T any](src []T, index int) ([]T, T, error) {
	length := len(src)
	if index < 0 || index >= length {
		var zero T
		return src, zero, NewErrIndexOutOfRange(length, index)
	}

	res := src[index]
	copy(src[index:], src[index+1:])
	return src[:length-1], res, nil
}
