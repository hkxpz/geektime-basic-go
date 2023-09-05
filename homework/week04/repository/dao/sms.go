package dao

import (
	"context"
)

type SMSDao interface {
	Insert(ctx context.Context, sms SMS) error
	FindByID(ctx context.Context, id int64) (SMS, error)
}

type SMS struct {
	Id       int64 `gorm:"primaryKey,autoIncrement"`
	CreateAt int64 `gorm:"comment:创建时间"`
	UpdateAt int64 `gorm:"comment:更新时间"`
}
