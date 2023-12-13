package gorm

import (
	"gorm.io/gorm"

	"geektime-basic-go/webook/migrator"
	"geektime-basic-go/webook/migrator/events"
	"geektime-basic-go/webook/pkg/logger"
)

type DynamicValidator[T migrator.Entity] struct {
	srcFirst *Validator[T]
	dstFirst *Validator[T]
}

func NewDynamicValidator[T migrator.Entity](src, dst *gorm.DB, l logger.Logger, producer events.Producer) *DynamicValidator[T] {
	return &DynamicValidator[T]{
		srcFirst: NewValidator[T](src, dst, "src", l, producer),
		dstFirst: NewValidator[T](dst, src, "dst", l, producer),
	}
}
