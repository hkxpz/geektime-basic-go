package cache

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"github.com/redis/go-redis/v9"

	"geektime-basic-go/webook/interactive/domain"
)

type InteractiveCache interface {
	Get(ctx context.Context, biz string, bizID int64) (domain.Interactive, error)
	Set(ctx context.Context, biz string, bizID int64, intr domain.Interactive) error
	IncrReadCntIfPresent(ctx context.Context, biz string, bizID int64) error
	DecrLikeCntIfPresent(ctx context.Context, biz string, bizID int64) error
	IncrLikeCntIfPresent(ctx context.Context, biz string, bizID int64) error
	IncrCollectCntIfPresent(ctx context.Context, biz string, bizID int64) error
	BatchIncrLikeCntIfPresent(ctx context.Context, biz string, bizIDs []int64) error
	BatchDecrLikeCntIfPresent(ctx context.Context, biz string, bizIDs []int64) error
	BatchSetLikeCnt(ctx context.Context, biz string, bizIDs []int64, cnts []int64) ([]string, error)
}

const (
	fieldReadCnt    = "read_cnt"
	fieldCollectCnt = "collect_cnt"
	fieldLikeCnt    = "like_cnt"
)

const topLimit = 5000

var (
	//go:embed lua/interactive_incr_cnt.lua
	luaIncrCnt string

	//go:embed lua/interactive_remove_cnt.lua
	luaRemCnt string
)

// ErrKeyNotExist 因为我们目前还是只有一个实现，所以可以保持用别名
var ErrKeyNotExist = redis.Nil

type interactiveCache struct {
	client redis.Cmdable
}

func NewInteractiveCache(client redis.Cmdable) InteractiveCache {
	return &interactiveCache{client: client}
}

func (cache *interactiveCache) IncrReadCntIfPresent(ctx context.Context, biz string, bizID int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizID)}, fieldReadCnt, 1).Err()
}

func (cache *interactiveCache) key(biz string, bizID int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizID)
}

func (cache *interactiveCache) Get(ctx context.Context, biz string, bizID int64) (domain.Interactive, error) {
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
	return domain.Interactive{BizID: bizID, ReadCnt: readCnt, LikeCnt: lickCnt, CollectCnt: collectCnt}, err
}

func (cache *interactiveCache) Set(ctx context.Context, biz string, bizID int64, intr domain.Interactive) error {
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

func (cache *interactiveCache) DecrLikeCntIfPresent(ctx context.Context, biz string, bizID int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizID)}, fieldLikeCnt, -1).Err()
}

func (cache *interactiveCache) IncrLikeCntIfPresent(ctx context.Context, biz string, bizID int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizID)}, fieldLikeCnt, 1).Err()
}

func (cache *interactiveCache) IncrCollectCntIfPresent(ctx context.Context, biz string, bizID int64) error {
	return cache.client.Eval(ctx, luaIncrCnt, []string{cache.key(biz, bizID)}, fieldCollectCnt, 1).Err()
}

func (cache *interactiveCache) BatchIncrLikeCntIfPresent(ctx context.Context, biz string, bizIDs []int64) error {
	pipeClient := cache.client.Pipeline()
	for _, bizID := range bizIDs {
		pipeClient.ZIncrBy(ctx, "interactive", 1, cache.key(biz, bizID))
	}
	_, err := pipeClient.Exec(ctx)
	return err
}

func (cache *interactiveCache) BatchDecrLikeCntIfPresent(ctx context.Context, biz string, bizIDs []int64) error {
	pipeClient := cache.client.Pipeline()
	for _, bizID := range bizIDs {
		pipeClient.ZIncrBy(ctx, "interactive", -1, cache.key(biz, bizID))
	}
	_, err := pipeClient.Exec(ctx)
	return err
}

func (cache *interactiveCache) BatchSetLikeCnt(ctx context.Context, biz string, bizIDs []int64, cnts []int64) ([]string, error) {
	pipeClient := cache.client.Pipeline()
	for idx := range bizIDs {
		pipeClient.ZAdd(ctx, "interactive", redis.Z{Score: float64(cnts[idx]), Member: cache.key(biz, bizIDs[idx])})
	}
	_, err := pipeClient.Exec(ctx)
	res, err := cache.client.Eval(ctx, luaRemCnt, []string{"interactive"}, topLimit).Result()
	if err != nil {
		return nil, err
	}

	popIDs := slice.Map[any, string](res.([]any), func(idx int, src any) string {
		return src.(string)
	})
	return popIDs, err
}
