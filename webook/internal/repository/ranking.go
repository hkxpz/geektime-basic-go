package repository

import (
	"context"

	"github.com/ecodeclub/ekit/syncx/atomicx"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository/cache/memory"
	"geektime-basic-go/webook/internal/repository/cache/redis"
)

type RankingRepository interface {
	ReplaceTopN(ctx context.Context, arts []domain.Article) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type cacheRankingRepository struct {
	Cache      *redis.RankingCache
	localCache *memory.RankingCache
	topN       atomicx.Value[[]domain.Article]
}

func NewCacheRankingRepository(cache *redis.RankingCache, localCache *memory.RankingCache) RankingRepository {
	return &cacheRankingRepository{Cache: cache, localCache: localCache}
}

func (c *cacheRankingRepository) ReplaceTopN(ctx context.Context, arts []domain.Article) error {
	_ = c.localCache.Set(ctx, arts)
	return c.Cache.Set(ctx, arts)
}

func (c *cacheRankingRepository) GetTopN(ctx context.Context) ([]domain.Article, error) {
	arts, err := c.localCache.Get(ctx)
	if err != nil {
		return arts, nil
	}
	arts, err = c.Cache.Get(ctx)
	if err != nil {
		return c.localCache.ForceGet(ctx)
	}
	_ = c.localCache.Set(ctx, arts)
	return arts, err
}
