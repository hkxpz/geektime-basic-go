package article

import (
	"context"
	"time"

	"github.com/IBM/sarama"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/pkg/logger"
	"geektime-basic-go/webook/pkg/saramax"
)

type HistoryConsumer struct {
	client sarama.Client
	repo   repository.HistoryRecordRepository
	l      logger.Logger
}

func (hc *HistoryConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("history", hc.client)
	if err != nil {
		return err
	}

	go func() {
		if err = cg.Consume(context.Background(), []string{topicReadEvent}, saramax.NewHandler[ReadEvent](hc.l, hc.Consume)); err != nil {
			hc.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (hc *HistoryConsumer) Consume(msg *sarama.ConsumerMessage, evt ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return hc.repo.AddRecord(ctx, domain.HistoryRecord{
		Uid:   evt.Uid,
		Biz:   "article",
		BizID: evt.Aid,
	})
}
