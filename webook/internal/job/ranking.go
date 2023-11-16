package job

import (
	"context"
	"time"

	rlock "github.com/gotomicro/redis-lock"

	"geektime-basic-go/webook/internal/service"
	"geektime-basic-go/webook/pkg/logger"
)

var _ Job = (*RankingJob)(nil)

type RankingJob struct {
	svc        service.RankingService
	timeout    time.Duration
	lockClient *rlock.Client
	l          logger.Logger
	key        string
}

func NewRankingJob(svc service.RankingService, lockClient *rlock.Client, l logger.Logger, timeout time.Duration) *RankingJob {
	return &RankingJob{svc: svc, lockClient: lockClient, timeout: timeout, l: l, key: "job:ranking"}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

func (r *RankingJob) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 本身我们这里设计的就是要在 r.timeout 内计算完成
	// 刚好也做成分布式锁的超时时间
	lock, err := r.lockClient.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{Interval: 100 * time.Millisecond, Max: 3}, time.Second)
	// 我们这里不需要处理 error，因为大部分情况下，可以相信别的节点会继续拿锁
	if err != nil {
		return err
	}

	defer func() {
		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err = lock.Unlock(ctx); err != nil {
			r.l.Error("释放分布式锁失败", logger.String("name", r.Name()), logger.Error(err))
		}
	}()
	return r.run()
}

func (r *RankingJob) run() error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.RankTopN(ctx)
}
