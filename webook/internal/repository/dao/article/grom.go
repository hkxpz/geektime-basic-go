package article

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormDAO struct {
	db *gorm.DB
}

func NewGormArticleDAO(db *gorm.DB) DAO {
	return &gormDAO{db: db}
}

func (dao *gormDAO) Insert(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.CreateAt, art.UpdateAt = now, now
	err := dao.db.WithContext(ctx).Create(&art).Error
	return art.ID, err
}

func (dao *gormDAO) UpdateById(ctx context.Context, art Article) error {
	now := time.Now().UnixMilli()
	res := dao.db.Model(&Article{}).WithContext(ctx).
		Where("id= ? AND author_id = ? ", art.ID, art.AuthorID).
		Updates(map[string]any{
			"title":     art.Title,
			"content":   art.Content,
			"status":    art.Status,
			"update_at": now,
		})

	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return errors.New("更新数据失败")
	}
	return nil
}

func (dao *gormDAO) Sync(ctx context.Context, art Article) (int64, error) {
	err := dao.db.Transaction(func(tx *gorm.DB) error {
		var err error
		if art.ID == 0 {
			art.ID, err = NewGormArticleDAO(tx).Insert(ctx, art)
		} else {
			err = NewGormArticleDAO(tx).UpdateById(ctx, art)
		}
		if err != nil {
			return err
		}

		now := time.Now().UnixMilli()
		publishArt := PublishedArticle(art)
		publishArt.CreateAt, publishArt.UpdateAt = now, now
		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":     art.Title,
				"content":   art.Content,
				"status":    art.Status,
				"update_at": now,
			}),
		}).Create(&publishArt).Error
	})

	return art.ID, err
}
