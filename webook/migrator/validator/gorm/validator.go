package gorm

import (
	"context"
	"errors"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"

	"geektime-basic-go/webook/migrator"
	"geektime-basic-go/webook/migrator/events"
	"geektime-basic-go/webook/pkg/logger"
)

type Validator[T migrator.Entity] struct {
	base   *gorm.DB
	target *gorm.DB

	// 这边需要告知是以 src 为准 还是以 dst 为准
	direction string

	batchSize int

	l        logger.Logger
	producer events.Producer
}

func NewValidator[T migrator.Entity](base *gorm.DB, target *gorm.DB, direction string, l logger.Logger, producer events.Producer) *Validator[T] {
	return &Validator[T]{base: base, target: target, direction: direction, l: l, producer: producer, batchSize: 100}
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	if err := v.baseToTarget(ctx); err != nil {
		return err
	}
	return v.targetToBase(ctx)
}

// baseToTarget 从 base 到 target 的验证
// 找出 dst 中错误的数据
func (v *Validator[T]) baseToTarget(ctx context.Context) error {
	offset := -1
	base := v.base.WithContext(ctx)
	target := v.target.WithContext(ctx)
	for {
		offset++
		var src T
		err := base.Order("id").Offset(offset).First(&src).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			v.l.Error("src to dst 查询源表失败", logger.Error(err))
			continue
		}

		var dst T
		err = target.Where("id = ?", src.ID()).First(&dst).Error
		switch {
		default:
			v.l.Error("src to dst 查询目标表失败", logger.Error(err))
		case errors.Is(err, gorm.ErrRecordNotFound):
			v.notify(src.ID(), events.InconsistentEventTypeTargetMissing)
		case err == nil:
			if equal := src.CompareTo(dst); !equal {
				v.notify(src.ID(), events.InconsistentEventTypeNotEqual)
			}
		}
	}
}

// targetToBase 从 target 到 base 的验证
// 找出 dst 中多余的数据
func (v *Validator[T]) targetToBase(ctx context.Context) error {
	offset := -v.batchSize
	base := v.base.WithContext(ctx)
	target := v.target.WithContext(ctx)
	for {
		offset += v.batchSize
		var dstTs []T
		err := target.Model(new(T)).Select("id").Offset(offset).Limit(v.batchSize).Find(&dstTs).Error
		if errors.Is(err, gorm.ErrRecordNotFound) || len(dstTs) < 1 {
			return nil
		}
		if err != nil {
			v.l.Error("dst to src 查询目标表失败", logger.Error(err))
			continue
		}
		ids := slice.Map(dstTs, func(idx int, src T) int64 { return src.ID() })

		var srcTs []T
		err = base.Select("id").Where("id IN ?", ids).Find(&srcTs).Error
		switch {
		default:
			v.l.Error("dst to src 查询源表失败", logger.Error(err))
		case errors.Is(err, gorm.ErrRecordNotFound):
			// 说明 ID 全没有
			v.notifySrcMissing(dstTs)
		case err == nil:
			missing := slice.DiffSetFunc(dstTs, srcTs, func(src, dst T) bool { return src.ID() == dst.ID() })
			v.notifySrcMissing(missing)
		}

		if len(dstTs) < v.batchSize {
			return nil
		}
	}
}

func (v *Validator[T]) notify(id int64, typ string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	evt := events.InconsistentEvent{Direction: v.direction, ID: id, Type: typ}
	if err := v.producer.ProduceInconsistentEvent(ctx, evt); err != nil {
		v.l.Error("上报失败", logger.Error(err), logger.Any("event", evt))
	}
}

func (v *Validator[T]) notifySrcMissing(ts []T) {
	for _, t := range ts {
		v.notify(t.ID(), events.InconsistentEventTypeBaseMissing)
	}
}
