package article

import (
	"context"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"

	"geektime-basic-go/webook/interactive/events"
	"geektime-basic-go/webook/interactive/repository"
	"geektime-basic-go/webook/pkg/logger"
	"geektime-basic-go/webook/pkg/saramax"
)

const topicReadEvent = "article_read_event"

type ReadEvent struct {
	Aid int64
	Uid int64
}

type Producer interface {
	ProduceReadEvent(evt ReadEvent) error
}

type saramaSyncProducer struct {
	producer sarama.SyncProducer
}

func NewSaramaSyncProducer(producer sarama.SyncProducer) Producer {
	return &saramaSyncProducer{producer: producer}
}

func (p *saramaSyncProducer) ProduceReadEvent(evt ReadEvent) error {
	val, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topicReadEvent,
		Value: sarama.ByteEncoder(val),
	})
	return err
}

var _ events.Consumer = (*InteractiveReadEventConsumer)(nil)

type InteractiveReadEventConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.Logger
}

func NewInteractiveReadEventConsumer(client sarama.Client, repo repository.InteractiveRepository, l logger.Logger) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{client: client, repo: repo, l: l}
}

func (c *InteractiveReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", c.client)
	if err != nil {
		c.l.Error("获取消费者组失败", logger.Error(err))
		return err
	}
	go func() {
		err = cg.Consume(context.Background(), []string{topicReadEvent}, saramax.NewHandler[ReadEvent](c.l, c.Consume))
		if err != nil {
			c.l.Error("退出消费者循环异常", logger.Error(err))
		}
	}()
	return err
}

func (c *InteractiveReadEventConsumer) StartBatch() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", c.client)
	if err != nil {
		c.l.Error("获取消费者组失败", logger.Error(err))
	}
	go func() {
		err = cg.Consume(context.Background(), []string{topicReadEvent}, saramax.NewBatchHandler[ReadEvent](c.l, c.BatchConsume))
		if err != nil {
			c.l.Error("退出消费者循环异常", logger.Error(err))
		}
	}()
	return err
}

func (c *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage, evt ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return c.repo.IncrReadCnt(ctx, "article", evt.Aid)
}

func (c *InteractiveReadEventConsumer) BatchConsume(msgs []*sarama.ConsumerMessage, evts []ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	bizs := make([]string, len(msgs))
	ids := make([]int64, len(msgs))
	for i := range evts {
		bizs[i] = "article"
		ids[i] = evts[i].Uid
	}
	return c.repo.BatchIncrReadCnt(ctx, bizs, ids)
}
