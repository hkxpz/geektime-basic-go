package article

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"geektime-basic-go/webook/internal/domain"
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
	res := dao.db.WithContext(ctx).Model(&Article{}).
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
		return ErrPossibleIncorrectAuthor
	}
	return nil
}

func (dao *gormDAO) Sync(ctx context.Context, art Article) (int64, error) {
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func (dao *gormDAO) SyncStatus(ctx context.Context, uid, id int64, status uint8) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id = ? AND author_id = ?", id, uid).Update("status", status)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}

		res = tx.Model(&PublishedArticle{}).Where("id = ? AND author_id = ?", id, uid).Update("status", status)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}

		return nil
	})
}

func (dao *gormDAO) GetPubByID(ctx context.Context, id int64) (PublishedArticle, error) {
	var pub PublishedArticle
	err := dao.db.WithContext(ctx).Where("id = ?", id).Find(&pub).Error
	return pub, err
}

func (dao *gormDAO) GetByAuthor(ctx context.Context, author int64, offset int, limit int) ([]Article, error) {
	var arts []Article
	err := dao.db.WithContext(ctx).Model(&Article{}).
		Where("author_id = ?", author).
		Offset(offset).
		Limit(limit).
		Order("update_at DESC").
		Find(&arts).Error
	return arts, err
}

func (dao *gormDAO) GetByID(ctx context.Context, id int64) (Article, error) {
	var art Article
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&art).Error
	return art, err
}

func (dao *gormDAO) List(ctx context.Context, author int64, offset int, limit int) ([]domain.Article, error) {
	//TODO implement me
	panic("implement me")
}
