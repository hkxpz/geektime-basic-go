package job

import (
	"context"
	"sync"
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

	localLock sync.Mutex
	lock      *rlock.Lock
}

func NewRankingJob(svc service.RankingService, lockClient *rlock.Client, l logger.Logger, timeout time.Duration) *RankingJob {
	return &RankingJob{svc: svc, lockClient: lockClient, timeout: timeout, l: l, key: "job:ranking"}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

func (r *RankingJob) Run() error {
	r.localLock.Lock()
	defer r.localLock.Unlock()

	if r.lock != nil {
		return r.run()
	}

	var err error
	// 试着拿锁
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 本身我们这里设计的就是要在 r.timeout 内计算完成
	// 刚好也做成分布式锁的超时时间
	r.lock, err = r.lockClient.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{Interval: 100 * time.Millisecond, Max: 3}, time.Second)
	// 我们这里不需要处理 error，因为大部分情况下，可以相信别的节点会继续拿锁
	if err != nil {
		return nil
	}
	// 自动续约，也就是延长分布式锁的过期时间
	// r.timeout 的一半作为刷新间隔。你这边可以设置为几秒钟，因为访问 Redis 是很快的
	// 每次续约 r.timeout 的时间（也就是分布式锁的过期时间重置为 r.timeout
	go func() {
		if err = r.lock.AutoRefresh(r.timeout/2, r.timeout); err != nil {
			r.l.Error("续约失败", logger.Error(err))
			r.localLock.Lock()
			r.lock = nil
			r.localLock.Unlock()
		}
	}()
	return r.run()
}

func (r *RankingJob) run() error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.RankTopN(ctx)
}

func (r *RankingJob) Close() error {
	r.localLock.Lock()
	defer r.localLock.Unlock()
	// 释放锁
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	// 释放锁失败，但是也不需要作什么，因为这个分布式锁会在过期时间之后自动释放
	// unlock 的时候就会触发退出 AutoRefresh
	// 这个时候是否把 r.lock 置为 nil 都可以了
	return r.lock.Unlock(ctx)
}
