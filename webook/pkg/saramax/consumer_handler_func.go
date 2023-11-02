package saramax

import (
	"encoding/json"

	"github.com/IBM/sarama"

	"geektime-basic-go/webook/pkg/logger"
)

type Handler[T any] struct {
	l  logger.Logger
	fn func(msg *sarama.ConsumerMessage, t T) error
}

func NewHandler[T any](l logger.Logger, fn func(msg *sarama.ConsumerMessage, t T) error) *Handler[T] {
	return &Handler[T]{l: l, fn: fn}
}

func (h *Handler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var t T
		if err := json.Unmarshal(msg.Value, &t); err != nil {
			h.l.Error(
				"反序列化消息体失败",
				logger.String("topic", msg.Topic),
				logger.Int("partition", msg.Partition),
				logger.Int("offset", msg.Offset),
				logger.Error(err),
			)
			session.MarkMessage(msg, "")
			continue
		}

		if err := h.fn(msg, t); err != nil {
			h.l.Error(
				"处理消息失败",
				logger.String("topic", msg.Topic),
				logger.Int("partition", msg.Partition),
				logger.Int("offset", msg.Offset),
				logger.Error(err),
			)
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
