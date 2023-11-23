package repository

import (
	"context"
	"time"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository/dao"
)

type CronJobRepository interface {
	Preempt(ctx context.Context) (domain.CronJob, error)
	Release(ctx context.Context, id int64) error
	UpdateUpdateTime(ctx context.Context, id int64) error
	AddJob(ctx context.Context, j domain.CronJob) error
	UpdateNextTime(ctx context.Context, id int64, nextTime time.Time) error
}

type preemptCronJobRepository struct {
	dao dao.CronJobDAO
}

func NewPreemptCronJobRepository(dao dao.CronJobDAO) CronJobRepository {
	return &preemptCronJobRepository{dao: dao}
}

func (repo *preemptCronJobRepository) AddJob(ctx context.Context, j domain.CronJob) error {
	return repo.dao.Insert(ctx, repo.toEntity(j))
}

func (repo *preemptCronJobRepository) UpdateNextTime(ctx context.Context, id int64, nextTime time.Time) error {
	return repo.dao.UpdateNextTime(ctx, id, nextTime)
}

func (repo *preemptCronJobRepository) Release(ctx context.Context, id int64) error {
	return repo.dao.Release(ctx, id)
}

func (repo *preemptCronJobRepository) UpdateUpdateTime(ctx context.Context, id int64) error {
	return repo.dao.UpdateUpdateTime(ctx, id)
}

func (repo *preemptCronJobRepository) Preempt(ctx context.Context) (domain.CronJob, error) {
	j, err := repo.dao.Preempt(ctx)
	if err != nil {
		return domain.CronJob{}, nil
	}
	return repo.toDomain(j), nil
}

func (repo *preemptCronJobRepository) toEntity(j domain.CronJob) dao.Job {
	return dao.Job{
		ID:         j.ID,
		Name:       j.Name,
		Expression: j.Expression,
		Cfg:        j.Cfg,
		Executor:   j.Executor,
		NextTime:   j.NextTime.UnixMilli(),
	}
}

func (repo *preemptCronJobRepository) toDomain(j dao.Job) domain.CronJob {
	return domain.CronJob{
		ID:         j.ID,
		Name:       j.Name,
		Expression: j.Expression,
		Cfg:        j.Cfg,
		Executor:   j.Executor,
		NextTime:   time.UnixMilli(j.NextTime),
	}
}
