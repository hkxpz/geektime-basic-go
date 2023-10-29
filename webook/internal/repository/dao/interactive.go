package dao

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:generate mockgen -source=interactive.go -package=mocks -destination=mocks/interactive_mock_gen.go InteractiveDAO
type InteractiveDAO interface {
	IncrReadCnt(ctx context.Context, biz string, bizID int64) error
	Get(ctx *gin.Context, biz string, bizID int64) (Interactive, error)
	GetLikeInfo(ctx *gin.Context, biz string, bizID int64, uid int64) (UserLikeBiz, error)
	GetCollectionInfo(ctx *gin.Context, biz string, bizID int64, uid int64) (UserCollectionBiz, error)
	InsertLikeInfo(ctx *gin.Context, biz string, bizID int64, uid int64) error
	DeleteLikeInfo(ctx *gin.Context, biz string, bizID int64, uid int64) error
	InsertCollectionBiz(ctx *gin.Context, biz UserCollectionBiz) error
	BatchIncrReadCnt(ctx context.Context, bizs []string, ds []int64) error
}

type gormDAO struct {
	db *gorm.DB
}

func NewGormDAO(db *gorm.DB) InteractiveDAO {
	return &gormDAO{db: db}
}

type Interactive struct {
	ID         int64  `gorm:"primaryKey,autoIncrement"`
	BizID      int64  `gorm:"uniqueIndex:biz_type_id"`
	Biz        string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`
	ReadCnt    int64
	CollectCnt int64
	LikeCnt    int64
	CreateAt   int64
	UpdateAt   int64
}

// UserLikeBiz 命名无能，用户点赞的某个东西
type UserLikeBiz struct {
	ID       int64  `gorm:"primaryKey,autoIncrement"`
	BizID    int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Biz      string `gorm:"type:varchar(128);uniqueIndex:biz_type_id_uid"`
	UID      int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Status   uint8  // 1- 有效，0-无效。软删除的用法
	CreateAt int64
	UpdateAt int64
}

// UserCollectionBiz 收藏的东西
type UserCollectionBiz struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	CID      int64  `gorm:"index"`
	BizID    int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Biz      string `gorm:"type:varchar(128);uniqueIndex:biz_type_id_uid"`
	UID      int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	CreateAt int64
	UpdateAt int64
}

func (dao *gormDAO) IncrReadCnt(ctx context.Context, biz string, bizID int64) error {
	return dao.incrReadCnt(dao.db.WithContext(ctx), biz, bizID)
}

func (dao *gormDAO) Get(ctx *gin.Context, biz string, bizID int64) (Interactive, error) {
	var res Interactive
	err := dao.db.WithContext(ctx).Where("biz = ? AND biz_id = ?", biz, bizID).Find(&res).Error
	return res, err
}

func (dao *gormDAO) GetLikeInfo(ctx *gin.Context, biz string, bizID int64, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := dao.db.WithContext(ctx).Where("biz = ? AND biz_id = ? AND status = ?", biz, bizID, 1).Error
	return res, err
}

func (dao *gormDAO) GetCollectionInfo(ctx *gin.Context, biz string, bizID int64, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := dao.db.WithContext(ctx).Where("biz = ? AND biz_id = ? AND uid = ?", biz, bizID, uid).Error
	return res, err
}

func (dao *gormDAO) InsertLikeInfo(ctx *gin.Context, biz string, bizID int64, uid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"status":    1,
				"update_at": now,
			}),
		}).Create(&UserLikeBiz{
			BizID:    bizID,
			Biz:      biz,
			UID:      uid,
			Status:   1,
			CreateAt: now,
			UpdateAt: now,
		}).Error
		if err != nil {
			return err
		}

		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"like_cnt":  gorm.Expr("`like_cnt`+1"),
				"update_at": now,
			}),
		}).Create(&Interactive{
			BizID:    bizID,
			Biz:      biz,
			LikeCnt:  1,
			CreateAt: now,
			UpdateAt: now,
		}).Error
	})
}

func (dao *gormDAO) DeleteLikeInfo(ctx *gin.Context, biz string, bizID int64, uid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&UserLikeBiz{}).
			Where("biz = ? AND biz_id = ? AND uid = ?", biz, bizID, uid).
			Updates(map[string]any{
				"status":    0,
				"update_at": now,
			}).Error
		if err != nil {
			return err
		}
		return dao.db.WithContext(ctx).Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"like_cnt":  gorm.Expr("`like_cnt`-1"),
				"update_at": now,
			}),
		}).Create(&Interactive{
			LikeCnt:  1,
			CreateAt: now,
			UpdateAt: now,
			Biz:      biz,
			BizID:    bizID,
		}).Error
	})
}

func (dao *gormDAO) InsertCollectionBiz(ctx *gin.Context, cb UserCollectionBiz) error {
	now := time.Now().UnixMilli()
	cb.CreateAt, cb.UpdateAt = now, now
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := dao.db.WithContext(ctx).Create(&cb).Error; err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"like_cnt":  gorm.Expr("`like_cnt`+1"),
				"update_at": now,
			}),
		}).Create(&Interactive{
			CollectCnt: 1,
			CreateAt:   now,
			UpdateAt:   now,
			Biz:        cb.Biz,
			BizID:      cb.BizID,
		}).Error
	})
}

func (dao *gormDAO) BatchIncrReadCnt(ctx context.Context, bizs []string, bizIDs []int64) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := 0; i < len(bizs); i++ {
			if err := dao.incrReadCnt(tx, bizs[i], bizIDs[i]); err != nil {
				return err
			}
		}
		return nil
	})
}

func (dao *gormDAO) incrReadCnt(db *gorm.DB, biz string, bizID int64) error {
	now := time.Now().UnixMilli()
	return db.Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			"read_cut":  gorm.Expr("`read_cnt`+1"),
			"update_at": now,
		}),
	}).Create(&Interactive{
		ReadCnt:  1,
		CreateAt: now,
		UpdateAt: now,
		Biz:      biz,
		BizID:    bizID,
	}).Error
}
