package job

import (
	"fmt"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/integration/startup"
	svcmocks "geektime-basic-go/webook/internal/service/mocks"
)

func TestRun(t *testing.T) {
	startup.InitViper()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for i := 0; i < 3; i++ {
		go func(i int) {
			job := initRankingJob(ctrl, fmt.Sprintf("node-%d", i))
			for {
				time.Sleep(3 * time.Second)
				_ = job.Run()
			}
		}(i)
	}

	time.Sleep(30 * time.Second)
}

func initRankingJob(ctrl *gomock.Controller, nodeName string) *RankingJob {
	l := startup.InitZapLogger()
	cmd := startup.InitRedis()
	rc := startup.InitRLockClient(cmd)
	svc := svcmocks.NewMockRankingService(ctrl)
	svc.EXPECT().RankTopN(gomock.Any()).AnyTimes()
	return NewRankingJob(svc, rc, cmd, l, 3*time.Second, nodeName)
}
