package cache

import (
	"context"

	"github.com/gin-gonic/gin"

	"geektime-basic-go/webook/internal/domain"
)

// UserCache 用户服务缓存
type UserCache interface {
	Delete(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
}

// CodeCache 验证码服务缓存
type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type ArticleCache interface {
	DelFirstPage(ctx context.Context, author int64) error
	SetPub(ctx context.Context, article domain.Article) error
	GetPub(ctx *gin.Context, id int64) (domain.Article, error)
}

type InteractiveCache interface {
	Get(ctx *gin.Context, biz string, bizID int64) (domain.Interactive, error)
	Set(ctx *gin.Context, biz string, bizID int64, intr domain.Interactive) error
	IncrReadCntIfPresent(ctx context.Context, biz string, bizID int64) error
	DecrLikeCntIfPresent(ctx *gin.Context, biz string, bizID int64) error
	IncrLikeCntIfPresent(ctx *gin.Context, biz string, bizID int64) error
	IncrCollectCntIfPresent(ctx *gin.Context, biz string, bizID int64) error
}
