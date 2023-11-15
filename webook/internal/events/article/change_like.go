package article

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"

	"geektime-basic-go/webook/internal/events"
	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/pkg/logger"
	"geektime-basic-go/webook/pkg/saramax"
)

const topicChangeLike = "article_change_like_event"

type ChangeLikeEvent struct {
	BizID int64
	Uid   int64
	Liked bool
}

type ChangeLikeProducer interface {
	ProduceChangeLikeEvent(ctx context.Context, evt ChangeLikeEvent) error
}

type changeLikeSaramaSyncProducer struct {
	producer sarama.SyncProducer
}

func NewChangeLikeSaramaSyncProducer(producer sarama.SyncProducer) ChangeLikeProducer {
	return &changeLikeSaramaSyncProducer{producer: producer}
}

func (p *changeLikeSaramaSyncProducer) ProduceChangeLikeEvent(ctx context.Context, evt ChangeLikeEvent) error {
	tracer := otel.GetTracerProvider().Tracer("webook/internal/events/article/change_like")
	_, span := tracer.Start(ctx, "produceChangeLikeEvent", trace.WithSpanKind(trace.SpanKindProducer))
	defer span.End()
	val, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topicChangeLike,
		Value: sarama.ByteEncoder(val),
	})
	return err
}

var _ events.Consumer = (*ChangeLikeEventConsumer)(nil)

type ChangeLikeEventConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.Logger
}

func NewInteractiveLikeEventConsumer(client sarama.Client, repo repository.InteractiveRepository, l logger.Logger) *ChangeLikeEventConsumer {
	return &ChangeLikeEventConsumer{client: client, repo: repo, l: l}
}

func (c *ChangeLikeEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("change_like", c.client)
	if err != nil {
		c.l.Error("获取消费者组失败", logger.Error(err))
	}
	go func() {
		err = cg.Consume(context.Background(), []string{topicChangeLike}, saramax.NewBatchHandler[ChangeLikeEvent](c.l, c.BatchConsume))
		if err != nil {
			c.l.Error("退出消费者循环异常", logger.Error(err))
		}
	}()
	return err
}

func (c *ChangeLikeEventConsumer) Consume(msg *sarama.ConsumerMessage, evt ChangeLikeEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if evt.Liked {
		return c.repo.IncrLike(ctx, "article", evt.BizID, evt.Uid)
	}
	return c.repo.DecrLike(ctx, "article", evt.BizID, evt.Uid)
}

func (c *ChangeLikeEventConsumer) BatchConsume(msgs []*sarama.ConsumerMessage, evts []ChangeLikeEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	likeAids := make([]int64, 0, len(msgs))
	likeUids := make([]int64, 0, len(msgs))
	unlikeAids := make([]int64, 0, len(msgs))
	unlikeUids := make([]int64, 0, len(msgs))

	for _, evt := range evts {
		if evt.Liked {
			likeAids = append(likeAids, evt.BizID)
			likeUids = append(likeUids, evt.Uid)
			continue
		}

		unlikeAids = append(unlikeAids, evt.BizID)
		unlikeUids = append(unlikeUids, evt.Uid)
	}

	var eg errgroup.Group
	eg.Go(func() error {
		return c.repo.BatchIncrLike(ctx, "article", likeAids, likeUids)
	})

	eg.Go(func() error {
		return c.repo.BatchDecrLike(ctx, "article", unlikeAids, unlikeUids)
	})

	return eg.Wait()
}
