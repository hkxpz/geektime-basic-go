package saramax

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"

	"geektime-basic-go/webook/pkg/logger"
)

type BatchHandler[T any] struct {
	l  logger.Logger
	fn func(msg []*sarama.ConsumerMessage, t []T) error
}

func NewBatchHandler[T any](l logger.Logger, fn func(msg []*sarama.ConsumerMessage, t []T) error) *BatchHandler[T] {
	return &BatchHandler[T]{l: l, fn: fn}
}

func (b *BatchHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgCh := claim.Messages()
	const batchSize = 20
	for {
		msgs := make([]*sarama.ConsumerMessage, 0, batchSize)
		ts := make([]T, 0, batchSize)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	down:
		for i := 0; i < batchSize; i++ {
			select {
			case <-ctx.Done():
				break down
			case msg, ok := <-msgCh:
				if !ok {
					// channel 被关闭了
					cancel()
					return nil
				}

				var t T
				if err := json.Unmarshal(msg.Value, &t); err != nil {
					b.l.Error(
						"反序列化消息体失败",
						logger.String("topic", msg.Topic),
						logger.Int("partition", msg.Partition),
						logger.Int("offset", msg.Offset),
						logger.Error(err),
					)
					session.MarkMessage(msg, "")
					continue
				}
				msgs = append(msgs, msg)
				ts = append(ts, t)
			}
		}

		cancel()
		if len(msgs) < 1 {
			continue
		}
		if err := b.fn(msgs, ts); err != nil {
			// 这里可以考虑重试，也可以在具体的业务逻辑里面重试
			// 也就是 eg.Go 里面重试
			continue
		}

		// 这边就要都提交了
		for _, msg := range msgs {
			session.MarkMessage(msg, "")
		}
	}
}
