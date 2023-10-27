package redis

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository/cache"
)

const (
	fieldReadCnt    = "read_cnt"
	fieldCollectCnt = "collect_cnt"
	fieldLikeCnt    = "like_cnt"
)

var (
	//go:embed lua/interactive_incr_read_cnt.lua
	luaIncrCnt string
)

// ErrKeyNotExist 因为我们目前还是只有一个实现，所以可以保持用别名
var ErrKeyNotExist = redis.Nil

type interactiveCache struct {
	client redis.Cmdable
}

func NewInteractiveCache(client redis.Cmdable) cache.InteractiveCache {
	return &interactiveCache{client: client}
}

func (cache *interactiveCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizID int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizID)}, fieldReadCnt, 1).Err()
}

func (cache *interactiveCache) key(biz string, bizID int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizID)
}

func (cache *interactiveCache) Get(ctx *gin.Context, biz string, bizID int64) (domain.Interactive, error) {
	data, err := cache.client.HGetAll(ctx, cache.key(biz, bizID)).Result()
	if err != nil {
		return domain.Interactive{}, err
	}

	if len(data) == 0 {
		return domain.Interactive{}, ErrKeyNotExist
	}

	collectCnt, _ := strconv.ParseInt(data[fieldCollectCnt], 10, 64)
	lickCnt, _ := strconv.ParseInt(data[fieldLikeCnt], 10, 64)
	readCnt, _ := strconv.ParseInt(data[fieldReadCnt], 10, 64)
	return domain.Interactive{ReadCnt: readCnt, LikeCnt: lickCnt, CollectCnt: collectCnt}, err
}

func (cache *interactiveCache) Set(ctx *gin.Context, biz string, bizID int64, intr domain.Interactive) error {
	key := cache.key(biz, bizID)
	err := cache.client.HMSet(ctx, key,
		fieldLikeCnt, intr.LikeCnt,
		fieldCollectCnt, intr.CollectCnt,
		fieldReadCnt, intr.ReadCnt,
	).Err()
	if err != nil {
		return err
	}
	return cache.client.Expire(ctx, key, 15*time.Minute).Err()
}

func (cache *interactiveCache) DecrLikeCntIfPresent(ctx *gin.Context, biz string, bizID int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizID)}, fieldLikeCnt, -1).Err()
}

func (cache *interactiveCache) IncrLikeCntIfPresent(ctx *gin.Context, biz string, bizID int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizID)}, fieldLikeCnt, 1).Err()
}

func (cache *interactiveCache) IncrCollectCntIfPresent(ctx *gin.Context, biz string, bizID int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizID)}, fieldCollectCnt, 1).Err()
}
