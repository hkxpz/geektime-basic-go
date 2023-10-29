package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
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

func (cache *articleCache) DelFirstPage(ctx context.Context, author int64) error {
	return cache.client.Del(ctx, cache.firstPageKey(author)).Err()
}

func (cache *articleCache) SetPub(ctx context.Context, art domain.Article) error {
	data, err := json.Marshal(art)
	if err != nil {
		return err
	}

	return cache.client.Set(ctx, cache.readerArtKey(art.ID), data, time.Minute*30).Err()
}

func (cache *articleCache) GetPub(ctx *gin.Context, id int64) (domain.Article, error) {
	data, err := cache.client.Get(ctx, cache.readerArtKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(data, &res)
	return res, err
}

func (cache *articleCache) GetFirstPage(ctx *gin.Context, author int64) ([]domain.Article, error) {
	bs, err := cache.client.Get(ctx, cache.firstPageKey(author)).Bytes()
	if err != nil {
		return nil, err
	}
	var arts []domain.Article
	err = json.Unmarshal(bs, &arts)
	return arts, err
}

func (cache *articleCache) SetFirstPage(ctx *gin.Context, author int64, arts []domain.Article) error {
	for i := range arts {
		arts[i].Content = arts[i].Abstract()
	}
	bs, err := json.Marshal(arts)
	if err != nil {
		return err
	}
	return cache.client.Set(ctx, cache.firstPageKey(author), bs, 10*time.Minute).Err()
}

func (cache *articleCache) Get(ctx *gin.Context, id int64) (domain.Article, error) {
	data, err := cache.client.Get(ctx, cache.authorArtKey(id)).Bytes()
	if err != nil {
		return domain.Article{}, err
	}
	var res domain.Article
	err = json.Unmarshal(data, &res)
	return res, err
}

func (cache *articleCache) Set(ctx context.Context, art domain.Article) error {
	data, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return cache.client.Set(ctx, cache.authorArtKey(art.ID), data, time.Minute).Err()
}

func (cache *articleCache) firstPageKey(author int64) string {
	return fmt.Sprintf("article:first_page:%d", author)
}

// 读者端的缓存设置
func (cache *articleCache) readerArtKey(author int64) string {
	return fmt.Sprintf("article:first_page:%d", author)
}

// 创作端的缓存设置
func (cache *articleCache) authorArtKey(id int64) string {
	return fmt.Sprintf("article:author:%d", id)
}
