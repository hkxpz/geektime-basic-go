package ioc

import (
	"fmt"
	"time"

	rlock "github.com/gotomicro/redis-lock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"

	"geektime-basic-go/webook/internal/job"
	"geektime-basic-go/webook/internal/service"
	"geektime-basic-go/webook/pkg/logger"
)

func InitRankingJob(svc service.RankingService, client *rlock.Client, l logger.Logger) *job.RankingJob {
	return job.NewRankingJob(svc, client, l, 30*time.Second)
}

func InitJobs(l logger.Logger, rankingJob *job.RankingJob) *cron.Cron {
	bd := job.NewCronJobBuilder(l, prometheus.SummaryOpts{
		Namespace: "hkxpz",
		Subsystem: "webook",
		Name:      "cron_job",
		Help:      "定时任务",
	})
	expr := cron.New(cron.WithSeconds())
	_, err := expr.AddJob("@every 1m", bd.Build(rankingJob))
	if err != nil {
		panic(fmt.Sprintf("初始化定时任务失败: %s", err))
	}
	return expr
}
