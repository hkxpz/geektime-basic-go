package validator

import (
	"context"
	"errors"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"

	"geektime-basic-go/webook/pkg/logger"
	"geektime-basic-go/webook/pkg/migrator"
	"geektime-basic-go/webook/pkg/migrator/events"
)

type GORMValidator[T migrator.Entity] struct {
	base   *gorm.DB
	target *gorm.DB

	// 这边需要告知是以 src 为准 还是以 dst 为准
	direction string

	batchSize int

	l        logger.Logger
	producer events.Producer

	updateAt int64
	// 默认是全量校验，如果没有数据了，就睡眠
	// 如果不是正数，那么就说明直接返回，结束这一次的循环
	sleepInterval time.Duration
}

func NewValidator[T migrator.Entity](base *gorm.DB, target *gorm.DB, direction string, l logger.Logger, producer events.Producer) *GORMValidator[T] {
	return &GORMValidator[T]{
		base:          base,
		target:        target,
		direction:     direction,
		l:             l,
		producer:      producer,
		batchSize:     100,
		sleepInterval: 0,
	}
}

func (g *GORMValidator[T]) UpdateAt(updateAt int64) *GORMValidator[T] {
	g.updateAt = updateAt
	return g
}

func (g *GORMValidator[T]) SleepInterval(i time.Duration) *GORMValidator[T] {
	g.sleepInterval = i
	return g
}

func (g *GORMValidator[T]) Validate(ctx context.Context) error {
	var wg errgroup.Group
	wg.Go(func() error { return g.baseToTarget(ctx) })
	wg.Go(func() error { return g.targetToBase(ctx) })
	return wg.Wait()
}

// baseToTarget 从 base 到 target 的验证
// 找出 dst 中错误的数据
func (g *GORMValidator[T]) baseToTarget(ctx context.Context) error {
	var offset int
	for {
		var src T
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		err := g.base.WithContext(dbCtx).Where("update_at > ?", g.updateAt).Order("id").Offset(offset).First(&src).Error
		cancel()
		switch {
		case err == nil:
			g.dstDiff(ctx, src)
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return nil
		case errors.Is(err, gorm.ErrRecordNotFound):
			if g.sleepInterval < 1 {
				return nil
			}
			time.Sleep(g.sleepInterval)
			continue
		default:
			g.l.Error("src to dst 查询源表失败", logger.Error(err))
		}
		offset++
	}
}

func (g *GORMValidator[T]) dstDiff(ctx context.Context, src T) {
	var dst T
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err := g.target.WithContext(dbCtx).Where("id = ?", src.Id()).First(&dst).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		g.notify(src.Id(), events.InconsistentEventTypeTargetMissing)
	case err == nil:
		if equal := src.CompareTo(dst); !equal {
			g.notify(src.Id(), events.InconsistentEventTypeNotEqual)
		}
	default:
		g.l.Error("src to dst 查询目标表失败", logger.Error(err))
	}
}

// targetToBase 从 target 到 base 的验证
// 找出 dst 中多余的数据
func (g *GORMValidator[T]) targetToBase(ctx context.Context) error {
	var offset int
	for {
		ts := make([]T, 0, g.batchSize)
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		err := g.target.WithContext(dbCtx).Model(new(T)).Select("id").Offset(offset).Limit(g.batchSize).Find(&ts).Error
		cancel()

		switch {
		case err == nil && len(ts) > 1:
			g.srcMissingRecords(ctx, ts)
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return nil
		case len(ts) < 1, errors.Is(err, gorm.ErrRecordNotFound):
			if g.sleepInterval < 1 {
				return nil
			}
			time.Sleep(g.sleepInterval)
			continue
		default:
			g.l.Error("dst to src 查询目标表失败", logger.Error(err))
		}
		offset += g.batchSize
	}
}

func (g *GORMValidator[T]) srcMissingRecords(ctx context.Context, ts []T) {
	ids := slice.Map(ts, func(idx int, src T) int64 { return src.Id() })
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var srcTs []T
	err := g.base.WithContext(dbCtx).Select("id").Where("id IN ?", ids).Find(&srcTs).Error
	switch {
	case len(srcTs) < 1, errors.Is(err, gorm.ErrRecordNotFound):
		// 说明 Id 全没有
		g.notifySrcMissing(ts)
	case err == nil:
		missing := slice.DiffSetFunc(ts, srcTs, func(src, dst T) bool { return src.Id() == dst.Id() })
		g.notifySrcMissing(missing)
	default:
		g.l.Error("dst to src 查询源表失败", logger.Error(err))
	}
}

func (g *GORMValidator[T]) notify(id int64, typ string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	evt := events.InconsistentEvent{Direction: g.direction, ID: id, Type: typ}
	if err := g.producer.ProduceInconsistentEvent(ctx, evt); err != nil {
		g.l.Error("上报失败", logger.Error(err), logger.Any("event", evt))
	}
}

func (g *GORMValidator[T]) notifySrcMissing(ts []T) {
	for _, t := range ts {
		g.notify(t.Id(), events.InconsistentEventTypeBaseMissing)
	}
}
