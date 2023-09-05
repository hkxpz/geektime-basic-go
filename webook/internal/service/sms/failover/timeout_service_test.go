package failover

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/service/sms"
	"geektime-basic-go/webook/internal/service/sms/mocks"
)

func TestTimeoutService_Send(t *testing.T) {
	testCases := []struct {
		name      string
		mock      func(ctrl *gomock.Controller) []sms.Service
		threshold uint32
		idx       uint64
		cnt       uint32

		wantErr error
		wantIdx uint64
		wanCnt  uint32
	}{
		{
			name: "超时,但没有连续超时",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := mocks.NewMockService(ctrl)
				svc0.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)
				return []sms.Service{svc0}
			},
			threshold: 3,
			cnt:       0,
			idx:       0,
			wanCnt:    1,
			wantIdx:   0,
			wantErr:   context.DeadlineExceeded,
		},
		{
			name: "触发了切换,切换之后成功了",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := mocks.NewMockService(ctrl)
				svc1 := mocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return []sms.Service{svc0, svc1}
			},
			threshold: 3,
			cnt:       3,
			wanCnt:    0,
			wantIdx:   1,
			wantErr:   nil,
		},
		{
			name: "触发了切换,切换之后失败",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := mocks.NewMockService(ctrl)
				svc1 := mocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("发送失败"))
				return []sms.Service{svc0, svc1}
			},
			threshold: 3,
			cnt:       3,
			wanCnt:    0,
			wantIdx:   1,
			wantErr:   errors.New("发送失败"),
		},
		{
			name: "触发了切换,切换之后依旧超时",
			mock: func(ctrl *gomock.Controller) []sms.Service {
				svc0 := mocks.NewMockService(ctrl)
				svc1 := mocks.NewMockService(ctrl)
				svc1.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded)
				return []sms.Service{svc0, svc1}
			},
			threshold: 3,
			cnt:       3,
			wanCnt:    1,
			wantIdx:   1,
			wantErr:   context.DeadlineExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := NewTimeoutService(tc.mock(ctrl), tc.threshold).(*timeoutService)
			svc.idx = tc.idx
			svc.cnt = tc.cnt

			err := svc.Send(context.Background(), "", []string{"123456"}, "13888888888")
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantIdx, svc.idx)
			assert.Equal(t, tc.wanCnt, svc.cnt)
		})
	}
}
