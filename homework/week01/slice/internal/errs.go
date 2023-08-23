package internal

import (
	"errors"
	"fmt"
)

var ErrIndexOutOfRange = errors.New("slice: 下表超出范围")

func NewErrIndexOutOfRange(length, index int) error {
	return fmt.Errorf("%w，长度 %d，下标 %d", ErrIndexOutOfRange, length, index)
}
