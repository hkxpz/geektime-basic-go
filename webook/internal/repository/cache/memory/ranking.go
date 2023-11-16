package memory

import (
	"context"
	"errors"
	"time"

	"github.com/ecodeclub/ekit/syncx/atomicx"

	"geektime-basic-go/webook/internal/domain"
)

type RankingCache struct {
	topN       *atomicx.Value[[]domain.Article]
	ddl        *atomicx.Value[time.Time]
	expiration time.Duration
}

func NewRankingCache() *RankingCache {
	return &RankingCache{
		topN:       atomicx.NewValue[[]domain.Article](),
		ddl:        atomicx.NewValueOf(time.Now()),
		expiration: 3 * time.Minute,
	}
}

func (r *RankingCache) Set(ctx context.Context, arts []domain.Article) error {
	r.ddl.Store(time.Now().Add(3 * time.Minute))
	r.topN.Store(arts)
	return nil
}

func (r *RankingCache) Get(context.Context) ([]domain.Article, error) {
	arts := r.topN.Load()
	if len(arts) == 0 || r.ddl.Load().Before(time.Now()) {
		return nil, errors.New("本地缓存失效了")
	}
	return arts, nil
}

func (r *RankingCache) ForceGet(context.Context) ([]domain.Article, error) {
	return r.topN.Load(), nil
}
