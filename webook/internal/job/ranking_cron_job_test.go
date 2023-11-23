package job

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/integration/startup"
	"geektime-basic-go/webook/internal/service"
	svcmocks "geektime-basic-go/webook/internal/service/mocks"
)

func TestScheduler_Start(t *testing.T) {
	startup.InitViper()
	l := startup.InitZapLogger()
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) service.CronJobService

		wantErr error
		wantJob *testJob
	}{
		{
			name: "调度了一个任务",
			mock: func(ctrl *gomock.Controller) service.CronJobService {
				svc := svcmocks.NewMockCronJobService(ctrl)
				svc.EXPECT().Preempt(gomock.Any()).Return(domain.CronJob{
					ID:         1,
					Name:       "test_job",
					Executor:   "local",
					Cfg:        "hello, world",
					Expression: "my cron expression",
					CancelFunc: func() {
						t.Log("取消了")
					},
				}, nil)
				svc.EXPECT().Preempt(gomock.Any()).AnyTimes().Return(domain.CronJob{}, errors.New("db 错误"))
				svc.EXPECT().ResetNextTime(gomock.Any(), gomock.Any()).Return(nil)
				return svc
			},
			wantErr: context.DeadlineExceeded,
			wantJob: &testJob{cnt: 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := tc.mock(ctrl)
			scheduler := NewScheduler(svc, l)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			executor := NewLocalFuncExecutor()
			tj := &testJob{}
			executor.AddLocalFunc("test_job", tj.Do)
			scheduler.RegisterExecutor(executor)
			err := scheduler.Start(ctx)
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantJob, tj)
		})
	}
}

type testJob struct {
	cnt int
}

func (tj *testJob) Do(ctx context.Context, j domain.CronJob) error {
	tj.cnt++
	return nil
}
