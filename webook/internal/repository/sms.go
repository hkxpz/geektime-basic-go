package repository

import (
	"context"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository/dao"
)

//go:generate mockgen -source=sms.go -package=mocks -destination=mocks/sms_mock_gen.go SMSRepository
type SMSRepository interface {
	Store(ctx context.Context, msg domain.SMS) error
	FindRetryWithMaxRetry(ctx context.Context, maxRetry int64, status int64) ([]dao.SMS, error)
	UpdateStatus(ctx context.Context, IDs []int64, status int64) error
	UpdateRetryCnt(ctx context.Context, IDs []int64) error
}

type repository struct {
	dao dao.SMSDao
}

func NewRepository(dao dao.SMSDao) SMSRepository {
	return &repository{dao: dao}
}

func (r *repository) Store(ctx context.Context, msg domain.SMS) error {
	return r.dao.Insert(ctx, dao.SMS{})
}

func (r *repository) FindRetryWithMaxRetry(ctx context.Context, maxRetry int64, status int64) ([]dao.SMS, error) {
	return r.dao.FindByMaxRetryAndStatus(ctx, maxRetry, status)
}

func (r *repository) UpdateStatus(ctx context.Context, IDs []int64, status int64) error {
	return r.dao.UpdateStatus(ctx, IDs)
}

func (r *repository) UpdateRetryCnt(ctx context.Context, IDs []int64) error {
	return r.dao.UpdateRetryCnt(ctx, IDs)
}
