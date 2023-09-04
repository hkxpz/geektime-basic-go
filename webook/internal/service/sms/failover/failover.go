package failover

import (
	"context"
	"errors"
	"log"
	"sync/atomic"

	"geektime-basic-go/webook/internal/service/sms"
)

type service struct {
	svcs []sms.Service
	idx  uint64
}

func NewService(svcs []sms.Service) sms.Service {
	return &service{svcs: svcs}
}

func (s *service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	idx := atomic.AddUint64(&s.idx, 1)
	length := uint64(len(s.svcs))
	for i := idx; i < idx+length; i++ {
		svc := s.svcs[i%length]
		err := svc.Send(ctx, tplId, args, numbers...)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled):
			return err
		}
		log.Printf("发送失败: %v", svc)
	}
	return errors.New("发送失败，所有服务商都尝试过了")
}

func (s *service) SendV1(ctx context.Context, tplId string, args []string, numbers ...string) error {
	for _, svc := range s.svcs {
		err := svc.Send(ctx, tplId, args, numbers...)
		if err == nil {
			return nil
		}
		log.Printf("发送失败: %v", svc)
	}
	return errors.New("发送失败，所有服务商都尝试过了")
}
