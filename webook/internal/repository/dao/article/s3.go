package article

import (
	"bytes"
	"context"
	"geektime-basic-go/webook/internal/domain"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ecodeclub/ekit"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var statusPrivate = domain.ArticleStatusPrivate.ToUint8()

type S3DAO struct {
	oss *s3.S3
	gormDAO
	bucket *string
}

func NewS3DAO(oss *s3.S3, db *gorm.DB) DAO {
	return &S3DAO{
		oss:     oss,
		gormDAO: gormDAO{db: db},
		bucket:  ekit.ToPtr[string]("webook-123"),
	}
}

func (dao *S3DAO) Sync(ctx context.Context, art Article) (int64, error) {
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
	if err != nil {
		return 0, err
	}

	_, err = dao.oss.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      dao.bucket,
		Key:         ekit.ToPtr[string](strconv.FormatInt(art.ID, 10)),
		Body:        bytes.NewReader([]byte(art.Content)),
		ContentType: ekit.ToPtr[string]("text/plain;charset=utf-8"),
	})

	return art.ID, err
}

func (dao *S3DAO) SyncStatus(ctx context.Context, author, id int64, status uint8) error {
	err := dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&Article{}).Where("id = ? AND author_id = ?", id, author).Update("status", status)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}

		res = tx.Model(&PublishedArticle{}).Where("id = ? AND author_id = ?", id, author).Update("status", status)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrPossibleIncorrectAuthor
		}

		return nil
	})
	if err != nil {
		return err
	}

	if status == statusPrivate {
		_, err = dao.oss.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: dao.bucket,
			Key:    ekit.ToPtr[string](strconv.FormatInt(id, 10)),
		})
	}

	return err
}
