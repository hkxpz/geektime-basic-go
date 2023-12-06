package job

import (
	"context"
	_ "embed"
	"time"

	"golang.org/x/exp/rand"

	"github.com/redis/go-redis/v9"

	rlock "github.com/gotomicro/redis-lock"

	"geektime-basic-go/webook/internal/service"
	"geektime-basic-go/webook/pkg/logger"
)

var _ Job = (*RankingJob)(nil)

type RankingJob struct {
	svc     service.RankingService
	timeout time.Duration
	l       logger.Logger
	key     string

	lockClient *rlock.Client
	cmd        redis.Cmdable

	lock       *rlock.Lock
	changeNode chan struct{}
	nodeName   string
}

func NewRankingJob(svc service.RankingService, lockClient *rlock.Client, cmd redis.Cmdable, l logger.Logger, timeout time.Duration, nodeName string) *RankingJob {
	return &RankingJob{
		svc:        svc,
		lockClient: lockClient,
		cmd:        cmd,
		timeout:    timeout,
		l:          l,
		key:        "job:ranking",
		changeNode: make(chan struct{}),
		nodeName:   nodeName,
	}
}

func (r *RankingJob) Name() string {
	return "ranking"
}

func (r *RankingJob) Run() error {
	go r.uploadStatus()
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
		timeout := r.timeout / 2
		ticker := time.NewTicker(timeout)
		for {
			select {
			case <-ticker.C:
				rfCtx, rfCancel := context.WithTimeout(context.Background(), timeout)
				if err = r.lock.Refresh(rfCtx); err != nil {
					r.l.Error("续约失败", logger.Error(err))
					r.lock = nil
					rfCancel()
					return
				}
				rfCancel()
			case <-r.changeNode:
				if r.lock == nil {
					return
				}
				rfCtx, rfCancel := context.WithTimeout(context.Background(), timeout)
				if err = r.lock.Unlock(rfCtx); err != nil {
					r.l.Error("解锁失败", logger.Error(err))
				}
				rfCancel()
				r.lock = nil
				return
			}
		}
	}()
	return r.run()
}

//func (r *RankingJob) RunV1() error {
//	if r.lock != nil {
//		return r.run()
//	}
//
//	var err error
//	// 试着拿锁
//	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
//	defer cancel()
//
//	// 本身我们这里设计的就是要在 r.timeout 内计算完成
//	// 刚好也做成分布式锁的超时时间
//	r.lock, err = r.lockClient.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{Interval: 100 * time.Millisecond, Max: 3}, time.Second)
//	// 我们这里不需要处理 error，因为大部分情况下，可以相信别的节点会继续拿锁
//	if err != nil {
//		return nil
//	}
//	// 自动续约，也就是延长分布式锁的过期时间
//	// r.timeout 的一半作为刷新间隔。你这边可以设置为几秒钟，因为访问 Redis 是很快的
//	// 每次续约 r.timeout 的时间（也就是分布式锁的过期时间重置为 r.timeout
//	go func() {
//		if err = r.lock.AutoRefresh(r.timeout/2, r.timeout); err != nil {
//			r.l.Error("续约失败", logger.Error(err))
//			r.lock = nil
//		}
//	}()
//	return r.run()
//}

func (r *RankingJob) run() error {
	r.l.Info("当前工作节点", logger.String("current_node", r.nodeName))
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.RankTopN(ctx)
}

var (
	//go:embed lua/upload_status.lua
	luaUploadStatus string
)

func (r *RankingJob) uploadStatus() {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout/2)
	defer cancel()
	status := r.currentStatus()
	r.l.Info("上报负载", logger.String("current_node", r.nodeName), logger.Int("负载", status))
	res, err := r.cmd.Eval(ctx, luaUploadStatus, []string{"ranking:node"}, status, 5*time.Second.Seconds()).Int()
	if r.lock != nil && res == 0 {
		select {
		default:
		case r.changeNode <- struct{}{}:
			r.l.Info("下一次调度更换节点", logger.String("current_node", r.nodeName), logger.Int("负载", status))
		}
	}
	if err != nil {
		r.l.Error("上报状态失败", logger.Int("负载", status), logger.Error(err))
	}
}

func (r *RankingJob) currentStatus() int {
	rand.Seed(uint64(time.Now().UnixNano()))
	return rand.Intn(100)
}
