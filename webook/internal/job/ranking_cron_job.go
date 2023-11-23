package job

import (
	"context"
	"errors"
	"time"

	"golang.org/x/sync/semaphore"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/service"
	"geektime-basic-go/webook/pkg/logger"
)

type ExecFunc func(ctx context.Context, j domain.CronJob) error

type Executor interface {
	Name() string
	// Exec ctx 是整个任务调度的上下文
	// 当从 ctx.Done 有信号的时候，就需要考虑结束执行
	// 具体实现来控制
	Exec(ctx context.Context, j domain.CronJob) error
}

type LocalFuncExecutor struct {
	funcs map[string]ExecFunc
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{funcs: make(map[string]ExecFunc)}
}

func (l *LocalFuncExecutor) AddLocalFunc(name string, fn ExecFunc) {
	l.funcs[name] = fn
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.CronJob) error {
	fn, ok := l.funcs[j.Name]
	if !ok {
		return errors.New("是不是忘记注册本地方法了？")
	}
	return fn(ctx, j)
}

// Scheduler 调度器
type Scheduler struct {
	svc       service.CronJobService
	execs     map[string]Executor
	l         logger.Logger
	dbTimeout time.Duration
	interval  time.Duration
	limiter   *semaphore.Weighted
}

func NewScheduler(svc service.CronJobService, l logger.Logger) *Scheduler {
	return &Scheduler{
		svc:       svc,
		execs:     make(map[string]Executor, 8),
		l:         l,
		dbTimeout: 3 * time.Second,
		interval:  time.Second,
		// 假如说最多只有 100 个在运行
		limiter: semaphore.NewWeighted(100),
	}
}

func (s *Scheduler) RegisterExecutor(exec Executor) {
	s.execs[exec.Name()] = exec
}

func (s *Scheduler) Start(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			// 已经超时了，或者被取消运行，大多数时候，都是被取消了，或者说关闭了
			return ctx.Err()
		}
		if err := s.limiter.Acquire(ctx, 1); err != nil {
			// 正常来说，只有 ctx 超时或者取消才会进来这里
			return err
		}
		dbCtx, cancel := context.WithTimeout(ctx, s.dbTimeout)
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			// 你也可以进一步细分不同的错误，如果是可以容忍的错误，
			// 没有抢占到，进入下一个循环
			// 这里可以考虑睡眠一段时间
			// 不然就直接 return
			time.Sleep(s.interval)
			continue
		}

		// 执行job
		exec, ok := s.execs[j.Executor]
		if !ok {
			s.l.Error("支持的 Executor 方式")
			j.CancelFunc()
			continue
		}
		// 要单独开一个 goroutine 来执行，这样我们就可以进入下一个循环了
		go func() {
			defer func() {
				s.limiter.Release(1)
				j.CancelFunc()
			}()

			if e := exec.Exec(ctx, j); e != nil {
				s.l.Error("调度任务失败", logger.Int("id", j.ID), logger.Error(e))
				return
			}
			if e := s.svc.ResetNextTime(ctx, j); e != nil {
				s.l.Error("更新下次执行失败", logger.Int("id", j.ID), logger.Error(e))
			}
		}()
	}
}
