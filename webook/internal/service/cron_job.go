package service

import (
	"context"
	"time"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/pkg/logger"
)

//go:generate mockgen -source=cron_job.go -package=svcmocks -destination=mocks/cron_job_mock_gen.go CronJobService
type CronJobService interface {
	Preempt(ctx context.Context) (domain.CronJob, error)
	AddJob(ctx context.Context, j domain.CronJob) error
	ResetNextTime(ctx context.Context, j domain.CronJob) error
}

type cronJobService struct {
	repo            repository.CronJobRepository
	l               logger.Logger
	refreshInterval time.Duration
}

func (svc *cronJobService) AddJob(ctx context.Context, j domain.CronJob) error {
	j.NextTime = j.Next(time.Now())
	return svc.repo.AddJob(ctx, j)
}

func (svc *cronJobService) ResetNextTime(ctx context.Context, j domain.CronJob) error {
	if t := j.Next(time.Now()); !t.IsZero() {
		return svc.repo.UpdateNextTime(ctx, j.ID, t)
	}
	return nil
}

func NewCronJobService(repo repository.CronJobRepository, l logger.Logger) CronJobService {
	return &cronJobService{repo: repo, l: l, refreshInterval: 10 * time.Second}
}

func (svc *cronJobService) Preempt(ctx context.Context) (domain.CronJob, error) {
	j, err := svc.repo.Preempt(ctx)
	if err != nil {
		return domain.CronJob{}, err
	}

	ticker := time.NewTicker(svc.refreshInterval)
	go func() {
		// 这边要启动一个 goroutine 开始续约，也就是在持续占有期间
		// 假定说我们这里是十秒钟续约一次
		for range ticker.C {
			svc.refresh(j.ID)
		}
	}()

	// 只能调用一次，也就是放弃续约。这时候要把状态还原回去
	j.CancelFunc = func() {
		ticker.Stop()
		releaseCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if e := svc.repo.Release(releaseCtx, j.ID); e != nil {
			svc.l.Error("释放任务失败", logger.Int("id", j.ID), logger.Error(e))
		}
	}
	return j, nil
}

func (svc *cronJobService) refresh(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := svc.repo.UpdateUpdateTime(ctx, id); err != nil {
		svc.l.Error("续约失败", logger.Int("id", id), logger.Error(err))
	}
}
