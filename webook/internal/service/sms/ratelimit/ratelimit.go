package ratelimit

import (
	"context"
	"fmt"

	"geektime-basic-go/webook/internal/service/sms"
	"geektime-basic-go/webook/pkg/ratelimit"
)

type service struct {
	svc     sms.Service
	limiter ratelimit.Limiter
	key     string
}

func NewService(svc sms.Service, limiter ratelimit.Limiter, key string) sms.Service {
	return &service{svc: svc, limiter: limiter, key: key}
}

func (s *service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	limited, err := s.limiter.Limit(ctx, s.key)
	if err != nil {
		return fmt.Errorf("短信服务限流异常 %w", err)
	}
	if limited {
		return sms.ErrLimited
	}
	return s.svc.Send(ctx, tplId, args, numbers...)
}
