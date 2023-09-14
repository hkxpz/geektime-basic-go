package retryable

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"geektime-basic-go/webook/internal/service/sms"
)

type service struct {
	svc sms.Service
	// 当前重试间隔
	curRetryInterval int64
	// 重试间隔
	retryInterval int64
	// 最大重试间隔
	maxRetryInterval int64
	// 重试次数
	maxReTry int64
}

func NewService(svc sms.Service, retryInterval, maxRetryInterval time.Duration, maxReTry int64) sms.Service {
	return &service{
		svc:              svc,
		curRetryInterval: int64(retryInterval),
		retryInterval:    int64(retryInterval),
		maxRetryInterval: int64(maxRetryInterval),
		maxReTry:         maxReTry,
	}
}

func (s *service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	for i := s.maxReTry; i > 0; i-- {
		err := s.svc.Send(ctx, tplId, args, numbers...)
		switch {
		case err == nil:
			atomic.SwapInt64(&s.curRetryInterval, s.retryInterval)
			return nil
		case errors.Is(err, sms.ErrLimited), errors.Is(err, sms.ErrServiceProviderException):
			atomic.SwapInt64(&s.curRetryInterval, s.maxRetryInterval)
			return err
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

	return errors.New("重试都失败了")
}
