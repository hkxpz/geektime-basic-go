package ratelimit

import (
	"context"
	"errors"
	"fmt"

	"geektime-basic-go/webook/internal/service/sms"
	"geektime-basic-go/webook/pkg/ratelimit"
)

const key = "sms_alibaba"

var errLimited = errors.New("短信服务触发限流")

type service struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewService(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &service{svc: svc, limiter: limiter}
}

func (s *service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	limited, err := s.limiter.Limit(ctx, key)
	if err != nil {
		return fmt.Errorf("短信服务判断是否限流异常 %w", err)
	}
	if limited {
		return errLimited
	}
	return s.svc.Send(ctx, tplId, args, numbers...)
}
