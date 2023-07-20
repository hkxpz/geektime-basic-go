package slice

func Delete[T any](src []T, index int) ([]T, error) {
	length := len(src)
	if index < 0 || index >= length {
		return nil, NewErrIndexOutOfRange(length, index)
	}

	copy(src[index:], src[index+1:])
	return src[:length-1], nil
}
