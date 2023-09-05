package async

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"time"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/internal/service/sms"
)

type service struct {
	svc  sms.Service
	repo repository.Repository

	// 当前重试间隔
	curRetryInterval int64
	// 重试间隔
	retryInterval int64
	// 最大重试间隔
	maxRetryInterval int64
	// 重试次数
	maxReTryCnt uint32
}

func NewService(svc sms.Service, repo repository.Repository,
	retryInterval, maxRetryInterval time.Duration, maxReTryCnt uint32) sms.Service {
	return &service{
		svc:              svc,
		repo:             repo,
		curRetryInterval: int64(retryInterval),
		retryInterval:    int64(retryInterval),
		maxRetryInterval: int64(maxRetryInterval),
		maxReTryCnt:      maxReTryCnt,
	}
}

func (s *service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	go func() {
		for i := s.maxReTryCnt; i > 0; i++ {
			err := s.svc.Send(ctx, tplId, args, numbers...)
			switch {
			default:
				log.Println(err)
			case err == nil:
				return
			case errors.Is(err, sms.ErrLimited), errors.Is(err, sms.ErrServiceProviderException):
				break
			}

			time.Sleep(time.Duration(atomic.LoadInt64(&s.curRetryInterval)))
			curRetryInterval := atomic.LoadInt64(&s.curRetryInterval)
			newRetryInterval := curRetryInterval + s.retryInterval
			if newRetryInterval > s.maxRetryInterval {
				atomic.CompareAndSwapInt64(&s.curRetryInterval, curRetryInterval, s.maxRetryInterval)
				continue
			}
			atomic.CompareAndSwapInt64(&s.curRetryInterval, curRetryInterval, newRetryInterval)
		}

		if err := s.repo.Store(ctx, domain.SMS{}); err != nil {
			log.Println(err)
		}
	}()

	return nil
}
