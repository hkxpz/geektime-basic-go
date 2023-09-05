package repository

import (
	"context"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository/dao"
)

type Repository interface {
	Store(ctx context.Context, msg domain.SMS) error
	Load(ctx context.Context, Id int64) (dao.SMS, error)
}

type repository struct {
	dao dao.SMSDao
}

func NewRepository(dao dao.SMSDao) Repository {
	return &repository{dao: dao}
}

func (r *repository) Store(ctx context.Context, msg domain.SMS) error {
	return r.dao.Insert(ctx, dao.SMS{})
}

func (r *repository) Load(ctx context.Context, Id int64) (dao.SMS, error) {
	return r.dao.FindByID(ctx, Id)
}
