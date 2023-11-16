package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"geektime-basic-go/webook/internal/domain"
)

type RankingCache struct {
	client     redis.Cmdable
	key        string
	expiration time.Duration
}

func NewRankingCache(client redis.Cmdable) *RankingCache {
	return &RankingCache{client: client, key: "ranking:article", expiration: 3 * time.Minute}
}

func (r *RankingCache) Set(ctx context.Context, arts []domain.Article) error {
	for i := 0; i < len(arts); i++ {
		arts[i].Content = arts[i].Abstract()
	}
	val, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key, val, r.expiration).Err()
}

func (r *RankingCache) Get(ctx context.Context) ([]domain.Article, error) {
	val, err := r.client.Get(ctx, r.key).Bytes()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal(val, &res)
	return res, err
}
