package repository

import (
	"context"

	"geektime-basic-go/webook/internal/repository/cache"
)

var (
	ErrCodeVerifyTooManyTimes = cache.ErrCodeVerifyTooManyTimes
	ErrCodeSendTooMany        = cache.ErrCodeSendTooMany
)

type CodeRepository interface {
	Store(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type cachedCodeRepository struct {
	cache cache.CodeCache
}

func NewCachedCodeRepository(c cache.CodeCache) CodeRepository {
	return &cachedCodeRepository{cache: c}
}

func (repo *cachedCodeRepository) Store(ctx context.Context, biz, phone, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}

func (repo *cachedCodeRepository) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	return repo.cache.Verify(ctx, biz, phone, code)
}
