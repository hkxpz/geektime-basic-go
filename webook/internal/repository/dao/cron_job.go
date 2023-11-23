package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type CronJobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	Insert(ctx context.Context, j Job) error
	UpdateNextTime(ctx context.Context, id int64, nextTime time.Time) error
	Release(ctx context.Context, id int64) error
	UpdateUpdateTime(ctx context.Context, id int64) error
}

type gormCronJobDAO struct {
	db *gorm.DB
}

func NewGormCronJobDAO(db *gorm.DB) CronJobDAO {
	return &gormCronJobDAO{db: db}
}

func (dao *gormCronJobDAO) Insert(ctx context.Context, j Job) error {
	now := time.Now().UnixMilli()
	j.CreateAt, j.UpdateAt = now, now
	return dao.db.WithContext(ctx).Create(&j).Error
}

func (dao *gormCronJobDAO) UpdateNextTime(ctx context.Context, id int64, nextTime time.Time) error {
	return dao.db.WithContext(ctx).Model(&Job{}).Where("id = ?", id).Updates(map[string]any{
		"next_time":   nextTime.UnixMilli(),
		"update_time": time.Now().UnixMilli(),
	}).Error
}

func (dao *gormCronJobDAO) Release(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Model(&Job{}).Where("id = ?", id).Updates(map[string]any{
		"status":    jobStatusWaiting,
		"update_at": time.Now().UnixMilli(),
	}).Error
}

func (dao *gormCronJobDAO) UpdateUpdateTime(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Model(&Job{}).Where("id = ?", id).Updates(map[string]any{
		"update_at": time.Now().UnixMilli(),
	}).Error
}

func (dao *gormCronJobDAO) Preempt(ctx context.Context) (Job, error) {
	db := dao.db.WithContext(ctx)
	for {
		// 每一个循环都重新计算 time.Now，因为之前可能已经花了一些时间了
		now := time.Now().UnixMilli()
		var j Job
		// 到了调度的时间
		if err := db.Where("next_time <= ? AND status = ?", now, jobStatusWaiting).First(&j).Error; err != nil {
			return Job{}, err
		}

		// 然后要开始抢占
		// 这里利用 update_at 来执行 CAS 操作
		// 其它一些公司可能会有一些 version 之类的字段
		res := db.Model(&Job{}).Where("id = ? AND version = ?", j.ID, j.Version).Updates(map[string]any{
			"update_at": now,
			"version":   j.Version + 1,
			"status":    jobStatusRunning,
		})
		if res.Error != nil {
			return Job{}, res.Error
		}
		// 抢占成功
		if res.RowsAffected == 1 {
			return j, nil
		}
		// 没有抢占到，也就是同一时刻被人抢走了，那么就下一个循环
	}
}

type Job struct {
	ID         int64 `gorm:"primaryKey,autoIncrement"`
	Name       string
	Executor   string
	Cfg        string
	Expression string
	Version    int64
	NextTime   int64 `gorm:"index"`
	Status     int
	CreateAt   int64
	UpdateAt   int64
}

const (
	// 等待被调度，意思就是没有人正在调度
	jobStatusWaiting = iota + 1
	// 已经被 goroutine 抢占了
	jobStatusRunning
	// 不再需要调度了，比如说被终止了，或者被删除了。
	jobStatusEnd
)
