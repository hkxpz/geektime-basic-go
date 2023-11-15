package saramax

import (
	"context"
	"encoding/json"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/IBM/sarama"

	"geektime-basic-go/webook/pkg/logger"
)

type BatchHandler[T any] struct {
	l                   logger.Logger
	fn                  func(msg []*sarama.ConsumerMessage, t []T) error
	consumerOffsetGauge prometheus.Gauge
	errorGauge          prometheus.Gauge
}

type options[T any] func(hdl *BatchHandler[T])

func (b *BatchHandler[T]) SetConsumerOffsetGauge() *BatchHandler[T] {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   "hkxpz",
		Subsystem:   "webook",
		Name:        "kafka_consumer_offset",
		ConstLabels: map[string]string{"instance_id": "instance-1"},
	})
	prometheus.MustRegister(gauge)
	b.consumerOffsetGauge = gauge
	return b
}

func (b *BatchHandler[T]) SetErrorGauge() *BatchHandler[T] {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   "hkxpz",
		Subsystem:   "webook",
		Name:        "kafka_consumer_error",
		ConstLabels: map[string]string{"instance_id": "instance-1"},
	})
	prometheus.MustRegister(gauge)
	b.errorGauge = gauge
	return b
}

func NewBatchHandler[T any](l logger.Logger, fn func(msg []*sarama.ConsumerMessage, t []T) error, opts ...options[T]) *BatchHandler[T] {
	hdl := &BatchHandler[T]{l: l, fn: fn}
	for _, opt := range opts {
		opt(hdl)
	}
	return hdl
}

func (b *BatchHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	latestConsumerOffset := claim.InitialOffset()
	msgCh := claim.Messages()
	var lastMsg *sarama.ConsumerMessage
	const batchSize = 20
	for {
		b.consumerOffsetGauge.Set(float64(claim.HighWaterMarkOffset() - latestConsumerOffset))
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
				lastMsg = msg
			}
		}

		cancel()
		if len(msgs) < 1 {
			continue
		}
		if err := b.fn(msgs, ts); err != nil {
			b.errorGauge.Set(-1)
			// 这里可以考虑重试，也可以在具体的业务逻辑里面重试
			// 也就是 eg.Go 里面重试
			continue
		}
		b.errorGauge.Set(1)
		latestConsumerOffset = lastMsg.Offset
		session.MarkMessage(lastMsg, "")
	}
}
