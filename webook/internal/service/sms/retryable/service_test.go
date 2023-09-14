package retryable

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/service/sms"
	"geektime-basic-go/webook/internal/service/sms/mocks"
)

func TestService_Send(t *testing.T) {
	const (
		maxReTry         = int64(5)
		retryInterval    = 20 * time.Millisecond
		maxRetryInterval = 60 * time.Millisecond
	)
	var failed = errors.New("发送失败")

	testCases := []struct {
		name             string
		mock             func(ctrl *gomock.Controller) sms.Service
		retryInterval    time.Duration
		maxRetryInterval time.Duration
		maxReTry         int64
		wantSvc          *service
		wantErr          error
	}{
		{
			name: "发送成功",
			mock: func(ctrl *gomock.Controller) sms.Service {
				svc := mocks.NewMockService(ctrl)
				svc.EXPECT().Send(gomock.Any(), "", []string{}, "13888888888")
				return svc
			},
			maxReTry: maxReTry,
			wantSvc:  &service{maxReTry: maxReTry},
		},
		{
			name: "全都失败",
			mock: func(ctrl *gomock.Controller) sms.Service {
				svc := mocks.NewMockService(ctrl)
				svc.EXPECT().Send(gomock.Any(), "", []string{}, "13888888888").Return(failed)
				svc.EXPECT().Send(gomock.Any(), "", []string{}, "13888888888").Return(failed)
				svc.EXPECT().Send(gomock.Any(), "", []string{}, "13888888888").Return(failed)
				svc.EXPECT().Send(gomock.Any(), "", []string{}, "13888888888").Return(failed)
				svc.EXPECT().Send(gomock.Any(), "", []string{}, "13888888888").Return(failed)
				return svc
			},
			maxReTry:         maxReTry,
			retryInterval:    retryInterval,
			maxRetryInterval: maxRetryInterval,
			wantSvc: &service{
				maxReTry:         5,
				curRetryInterval: int64(maxRetryInterval),
				retryInterval:    int64(retryInterval),
				maxRetryInterval: int64(maxRetryInterval),
			},
			wantErr: errors.New("重试都失败了"),
		},
		{
			name: "失败一次",
			mock: func(ctrl *gomock.Controller) sms.Service {
				svc := mocks.NewMockService(ctrl)
				svc.EXPECT().Send(gomock.Any(), "", []string{}, "13888888888").Return(failed)
				svc.EXPECT().Send(gomock.Any(), "", []string{}, "13888888888")
				return svc
			},
			maxReTry:         maxReTry,
			retryInterval:    retryInterval,
			maxRetryInterval: maxRetryInterval,
			wantSvc: &service{
				maxReTry:         5,
				curRetryInterval: int64(retryInterval),
				retryInterval:    int64(retryInterval),
				maxRetryInterval: int64(maxRetryInterval),
			},
		},
		{
			name: "限流",
			mock: func(ctrl *gomock.Controller) sms.Service {
				svc := mocks.NewMockService(ctrl)
				svc.EXPECT().Send(gomock.Any(), "", []string{}, "13888888888").Return(sms.ErrLimited)
				return svc
			},
			maxReTry:         maxReTry,
			retryInterval:    retryInterval,
			maxRetryInterval: maxRetryInterval,
			wantSvc: &service{
				maxReTry:         5,
				curRetryInterval: int64(maxRetryInterval),
				retryInterval:    int64(retryInterval),
				maxRetryInterval: int64(maxRetryInterval),
			},
			wantErr: sms.ErrLimited,
		},
		{
			name: "服务商异常",
			mock: func(ctrl *gomock.Controller) sms.Service {
				svc := mocks.NewMockService(ctrl)
				svc.EXPECT().Send(gomock.Any(), "", []string{}, "13888888888").Return(sms.ErrServiceProviderException)
				return svc
			},
			maxReTry:         maxReTry,
			retryInterval:    retryInterval,
			maxRetryInterval: maxRetryInterval,
			wantSvc: &service{
				maxReTry:         5,
				curRetryInterval: int64(maxRetryInterval),
				retryInterval:    int64(retryInterval),
				maxRetryInterval: int64(maxRetryInterval),
			},
			wantErr: sms.ErrServiceProviderException,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s := tc.mock(ctrl)
			svc := NewService(s, tc.retryInterval, tc.maxRetryInterval, tc.maxReTry)
			err := svc.Send(context.Background(), "", []string{}, "13888888888")
			assert.Equal(t, tc.wantErr, err)
			tc.wantSvc.svc = s
			assert.Equal(t, tc.wantSvc, svc)
		})
	}
}
