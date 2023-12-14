package fixer

import (
	"context"
	"errors"

	"github.com/IBM/sarama"
	"gorm.io/gorm"

	"geektime-basic-go/webook/pkg/logger"
	"geektime-basic-go/webook/pkg/migrator"
	"geektime-basic-go/webook/pkg/migrator/events"
	"geektime-basic-go/webook/pkg/migrator/fixer"
	"geektime-basic-go/webook/pkg/saramax"
)

type Consumer[T migrator.Entity] struct {
	client   sarama.Client
	l        logger.Logger
	srcFirst *fixer.OverrideFixer[T]
	dstFirst *fixer.OverrideFixer[T]
	topic    string
}

func NewConsumer[T migrator.Entity](client sarama.Client, l logger.Logger, src *gorm.DB, dst *gorm.DB, topic string) (*Consumer[T], error) {
	srcFirst, err := fixer.NewOverrideFixer[T](src, dst)
	if err != nil {
		return nil, err
	}
	dstFirst, err := fixer.NewOverrideFixer[T](dst, src)
	if err != nil {
		return nil, err
	}
	return &Consumer[T]{client: client, l: l, srcFirst: srcFirst, dstFirst: dstFirst, topic: topic}, nil
}

func (c *Consumer[T]) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("migrator-fix", c.client)
	if err != nil {
		return err
	}
	go func() {
		err = cg.Consume(context.Background(), []string{c.topic}, saramax.NewHandler[events.InconsistentEvent](c.l, c.Consume))
		if err != nil {
			c.l.Error("退出消费循环异常", logger.Error(err))
		}
	}()
	return nil
}

func (c *Consumer[T]) Consume(msg *sarama.ConsumerMessage, evt events.InconsistentEvent) error {
	switch evt.Direction {
	default:
		return errors.New("未知检验方向")
	case "src":
		return c.srcFirst.Fix(evt)
	case "dst":
		return c.dstFirst.Fix(evt)
	}
}

func (c *Consumer[T]) StartBatch() error {
	//TODO implement me
	panic("implement me")
}
