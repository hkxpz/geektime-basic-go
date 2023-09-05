package ratelimit

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/service/sms"
	"geektime-basic-go/webook/internal/service/sms/mocks"
	"geektime-basic-go/webook/pkg/ratelimit"
	limitmocks "geektime-basic-go/webook/pkg/ratelimit/mocks"
)

func TestService_Send(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter)
		ctx     context.Context
		phone   string
		code    string
		wantErr error
	}{
		{
			name: "正常发送",
			mock: func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter) {
				svc := mocks.NewMockService(ctrl)
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(ctx, "sms_alibaba").Return(false, nil)
				svc.EXPECT().Send(ctx, gomock.Any(), []string{"123456"}, "13888888888").Return(nil)
				return svc, limiter
			},
			ctx:   ctx,
			phone: "13888888888",
			code:  "123456",
		},
		{
			name: "触发限流",
			mock: func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter) {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(ctx, "sms_alibaba").Return(true, nil)
				return nil, limiter
			},
			ctx:     ctx,
			phone:   "13888888888",
			code:    "123456",
			wantErr: sms.ErrLimited,
		},
		{
			name: "限流异常",
			mock: func(ctrl *gomock.Controller) (sms.Service, ratelimit.Limiter) {
				limiter := limitmocks.NewMockLimiter(ctrl)
				limiter.EXPECT().Limit(ctx, "sms_alibaba").Return(false, errors.New("限流器异常"))
				return nil, limiter
			},
			ctx:     ctx,
			phone:   "13888888888",
			code:    "123456",
			wantErr: fmt.Errorf("短信服务限流异常 %w", errors.New("限流器异常")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc, limiter := tc.mock(ctrl)
			limitSvc := NewService(svc, limiter, "sms_alibaba")
			err := limitSvc.Send(tc.ctx, "", []string{tc.code}, tc.phone)
			assert.Equal(t, tc.wantErr, err)
		})
	}

}
