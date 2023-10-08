package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository/cache"
)

type articleCache struct {
	client redis.Cmdable
}

func NewArticleCache(client redis.Cmdable) cache.ArticleCache {
	return &articleCache{client: client}
}

func (r *articleCache) DelFirstPage(ctx context.Context, author int64) error {
	return r.client.Del(ctx, r.firstPageKey(author)).Err()
}

func (r *articleCache) SetPub(ctx context.Context, art domain.Article) error {
	data, err := json.Marshal(art)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, r.readerArtKey(art.ID), data, time.Minute*30).Err()
}

func (r *articleCache) firstPageKey(author int64) string {
	return fmt.Sprintf("article:first_page:%d", author)
}

func (r *articleCache) readerArtKey(author int64) string {
	return fmt.Sprintf("article:first_page:%d", author)
}
