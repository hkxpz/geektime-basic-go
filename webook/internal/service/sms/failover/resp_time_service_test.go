package failover

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/service/sms"
	"geektime-basic-go/webook/internal/service/sms/mocks"
)

func TestRespTimeService_Send(t *testing.T) {
	testCases := []struct {
		name                   string
		mock                   func(ctrl *gomock.Controller, curRespTime time.Duration) sms.Service
		before                 func(service2 sms.Service)
		curcnt                 int64
		curWindowsRespTime     int64
		curWindowsAvgRespTime  int64
		lastWindowsAvgRespTime int64
		curRespTime            time.Duration

		wantCurCnt                 int64
		wantCurWindowsRespTime     int64
		wantCurWindowsAvgRespTime  int64
		wantLastWindowsAvgRespTime int64
		wantErr                    error
	}{
		{
			name: "首次发送",
			mock: func(ctrl *gomock.Controller, curRespTime time.Duration) sms.Service {
				svc := mocks.NewMockService(ctrl)
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(any, any, any, any) {
					time.Sleep(curRespTime * time.Millisecond)
				})
				return svc
			},
			before:                    func(svc sms.Service) {},
			curRespTime:               20,
			wantCurCnt:                1,
			wantCurWindowsRespTime:    20,
			wantCurWindowsAvgRespTime: 20,
		},
		{
			name: "平均响应涨幅超过预期",
			mock: func(ctrl *gomock.Controller, curRespTime time.Duration) sms.Service {
				svc := mocks.NewMockService(ctrl)
				return svc
			},
			before: func(svc sms.Service) {
				svc.(*respTimeService).curcnt = 1
				svc.(*respTimeService).curWindowsRespTime = 40
				svc.(*respTimeService).lastWindowsAvgRespTime = 30
				svc.(*respTimeService).curWindowsAvgRespTime = 40
			},
			curRespTime:                20,
			wantCurCnt:                 1,
			wantCurWindowsRespTime:     40,
			wantCurWindowsAvgRespTime:  40,
			wantLastWindowsAvgRespTime: 40,
			wantErr:                    sms.ErrServiceProviderException,
		},
		{
			name: "平均响应涨幅没过预期, 且当前窗口小于最大窗口",
			mock: func(ctrl *gomock.Controller, curRespTime time.Duration) sms.Service {
				svc := mocks.NewMockService(ctrl)
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(any, any, any, any) {
					time.Sleep(curRespTime * time.Millisecond)
				})
				return svc
			},
			before: func(svc sms.Service) {
				svc.(*respTimeService).curcnt = 1
				svc.(*respTimeService).curWindowsRespTime = 20
				svc.(*respTimeService).lastWindowsAvgRespTime = 30
				svc.(*respTimeService).curWindowsAvgRespTime = 20
			},
			curRespTime:                20,
			wantCurCnt:                 2,
			wantCurWindowsRespTime:     40,
			wantCurWindowsAvgRespTime:  30,
			wantLastWindowsAvgRespTime: 30,
		},
		{
			name: "平均响应涨幅没过预期, 且当前窗口达到最大",
			mock: func(ctrl *gomock.Controller, curRespTime time.Duration) sms.Service {
				svc := mocks.NewMockService(ctrl)
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(any, any, any, any) {
					time.Sleep(curRespTime * time.Millisecond)
				})
				return svc
			},
			before: func(svc sms.Service) {
				svc.(*respTimeService).curcnt = 9
				svc.(*respTimeService).curWindowsRespTime = 180
				svc.(*respTimeService).lastWindowsAvgRespTime = 30
				svc.(*respTimeService).curWindowsAvgRespTime = 20
			},
			curRespTime:                20,
			wantCurWindowsAvgRespTime:  20,
			wantLastWindowsAvgRespTime: 20,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mockSvc := tc.mock(ctrl, tc.curRespTime)
			svc := NewRespTimeService(mockSvc, 10, 20)
			tc.before(svc)
			err := svc.Send(context.Background(), "", nil, "")
			require.Equal(t, tc.wantErr, err)
			require.Equal(t, tc.wantCurCnt, svc.(*respTimeService).curcnt)
			require.Equal(t, tc.wantCurWindowsRespTime, svc.(*respTimeService).curWindowsRespTime)
			require.Equal(t, tc.wantLastWindowsAvgRespTime, svc.(*respTimeService).lastWindowsAvgRespTime)
		})
	}
}
