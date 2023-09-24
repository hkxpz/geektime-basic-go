package article

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type DAO interface {
	Create(ctx context.Context, art Article) (int64, error)
	UpdateByID(ctx context.Context, art Article) error
}

type gormArticleDAO struct {
	db *gorm.DB
}

func NewGormArticleDAO(db *gorm.DB) DAO {
	return &gormArticleDAO{db: db}
}

func (dao *gormArticleDAO) Create(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.CreateAt, art.UpdateAt = now, now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.ID, err
}

func (dao *gormArticleDAO) UpdateByID(ctx context.Context, art Article) error {
	res := dao.db.Model(&Article{}).WithContext(ctx).
		Where("id = ? AND author_id = ?", art.ID, art.AuthorID).
		Updates(map[string]any{
			"title":     art.Title,
			"content":   art.Content,
			"status":    art.Status,
			"update_at": time.Now().UnixMilli(),
		})

	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return errors.New("更新数据失败")
	}
	return nil
}

type Article struct {
	ID       int64  `gorm:"primaryKey,autoIncrement"`
	Title    string `gorm:"type=varchar(4096)"`
	Content  string `gorm:"type=BLOB"`
	AuthorID int64  `gorm:"index"`
	Status   uint8  `gorm:"default=1"`
	CreateAt int64
	UpdateAt int64
}
