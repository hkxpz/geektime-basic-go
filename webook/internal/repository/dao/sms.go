package dao

import (
	"context"
	"time"

	"gorm.io/gorm"
)

//go:generate mockgen -source=sms.go -package=svcmocks -destination=mocks/sms_mock_gen.go SMSDao
type SMSDao interface {
	Insert(ctx context.Context, sms SMS) error
	FindByMaxRetryAndStatus(ctx context.Context, maxRetry int64, status int64) ([]SMS, error)
	UpdateStatus(ctx context.Context, IDs []int64) error
	UpdateRetryCnt(ctx context.Context, ds []int64) error
}

type gormSMSDAO struct {
	db *gorm.DB
}

func NewGormSMSDAO(db *gorm.DB) SMSDao {
	return &gormSMSDAO{db: db.Model(&SMS{})}
}

type SMS struct {
	ID       int64  `gorm:"primaryKey,autoIncrement"`
	Biz      string `gorm:"comment:业务签名"`
	Args     string `gorm:"comment:参数"`
	Numbers  string `gorm:"comment:发送号码"`
	Status   int64  `gorm:"comment:状态 1 发送成功 2 发送失败"`
	RetryCnt int64  `gorm:"comment:发送次数"`
	CreateAt int64  `gorm:"comment:创建时间"`
	UpdateAt int64  `gorm:"comment:更新时间"`
}

func (sd *gormSMSDAO) Insert(ctx context.Context, sms SMS) error {
	now := time.Now().UnixMilli()
	sms.UpdateAt, sms.CreateAt = now, now
	return sd.db.WithContext(ctx).Create(&sms).Error
}

func (sd *gormSMSDAO) FindByMaxRetryAndStatus(ctx context.Context, maxRetry int64, status int64) ([]SMS, error) {
	res := make([]SMS, 0, 100)
	err := sd.db.WithContext(ctx).Find(&res, "status = ? and retry_cnt <= ?", status, maxRetry).Error
	return res, err
}

func (sd *gormSMSDAO) UpdateStatus(ctx context.Context, IDs []int64) error {
	return sd.db.WithContext(ctx).Where("id in ?", IDs).Updates(map[string]interface{}{
		"status":    1,
		"update_at": time.Now().UnixMilli(),
	}).Error
}

func (sd *gormSMSDAO) UpdateRetryCnt(ctx context.Context, IDs []int64) error {
	return sd.db.WithContext(ctx).Where("id in ?", IDs).Updates(map[string]interface{}{
		"retry_cnt": gorm.Expr("retry_cnt + ?", 1),
		"update_at": time.Now().UnixMilli(),
	}).Error
}
