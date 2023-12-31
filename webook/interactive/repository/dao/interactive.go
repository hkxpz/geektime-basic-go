package dao

import (
	"context"
	"time"

	"geektime-basic-go/webook/pkg/migrator"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type InteractiveDAO interface {
	Get(ctx context.Context, biz string, bizID int64) (Interactive, error)
	GetLikeInfo(ctx context.Context, biz string, bizID int64, uid int64) (UserLikeBiz, error)
	GetCollectionInfo(ctx context.Context, biz string, bizID int64, uid int64) (UserCollectionBiz, error)
	InsertCollectionBiz(ctx context.Context, biz UserCollectionBiz) error
	IncrReadCnt(ctx context.Context, biz string, bizID int64) error
	BatchIncrReadCnt(ctx context.Context, bizs []string, bizIDs []int64) error
	InsertLikeInfo(ctx context.Context, biz string, bizID int64, uid int64) error
	BatchInsertLikeInfo(ctx context.Context, biz string, bizIDs []int64, uids []int64) error
	DeleteLikeInfo(ctx context.Context, biz string, bizID int64, uid int64) error
	BatchDeleteLikeInfo(ctx context.Context, biz string, bizIDs []int64, uids []int64) error
	GetMultipleLikeCnt(ctx context.Context, biz string, bizIDs []int64) ([]Interactive, error)
	GetByIDs(ctx context.Context, biz string, bizIDs []int64) ([]Interactive, error)
}

type gormDAO struct {
	db *gorm.DB
}

func NewInteractiveDAO(db *gorm.DB) InteractiveDAO {
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

func (i Interactive) Id() int64 {
	return i.ID
}

func (i Interactive) CompareTo(dst migrator.Entity) bool {
	if di, ok := dst.(Interactive); ok {
		return di == i
	}
	return false
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
	ID       int64  `gorm:"primaryKey,autoIncrement"`
	CID      int64  `gorm:"index"`
	BizID    int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	Biz      string `gorm:"type:varchar(128);uniqueIndex:biz_type_id_uid"`
	UID      int64  `gorm:"uniqueIndex:biz_type_id_uid"`
	CreateAt int64
	UpdateAt int64
}

// Collection 收藏夹
type Collection struct {
	Id       int64  `gorm:"primaryKey,autoIncrement"`
	Name     string `gorm:"type=varchar(1024)"`
	Uid      int64  `gorm:""`
	CreateAt int64
	UpdateAt int64
}

func (dao *gormDAO) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return dao.incrReadCnt(dao.db.WithContext(ctx), biz, bizId)
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
			"read_cnt":  gorm.Expr("`read_cnt`+1"),
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

func (dao *gormDAO) Get(ctx context.Context, biz string, bizID int64) (Interactive, error) {
	var res Interactive
	err := dao.db.WithContext(ctx).Where("biz = ? AND biz_id = ?", biz, bizID).Find(&res).Error
	return res, err
}

func (dao *gormDAO) GetLikeInfo(ctx context.Context, biz string, bizID int64, uid int64) (UserLikeBiz, error) {
	var res UserLikeBiz
	err := dao.db.WithContext(ctx).Where("biz=? AND biz_id = ? AND uid = ? AND status = ?", biz, bizID, uid, 1).First(&res).Error
	return res, err
}

func (dao *gormDAO) GetCollectionInfo(ctx context.Context, biz string, bizID int64, uid int64) (UserCollectionBiz, error) {
	var res UserCollectionBiz
	err := dao.db.WithContext(ctx).Where("biz = ? AND biz_id = ? AND uid = ?", biz, bizID, uid).First(&res).Error
	return res, err
}

func (dao *gormDAO) InsertLikeInfo(ctx context.Context, biz string, bizID int64, uid int64) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return dao.insertLikeInfo(tx, biz, bizID, uid)
	})
}

func (dao *gormDAO) BatchInsertLikeInfo(ctx context.Context, biz string, bizIDs []int64, uids []int64) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := 0; i < len(bizIDs); i++ {
			if err := dao.insertLikeInfo(tx, biz, bizIDs[i], uids[i]); err != nil {
				return err
			}
		}
		return nil
	})
}

func (dao *gormDAO) insertLikeInfo(db *gorm.DB, biz string, bizID int64, uid int64) error {
	now := time.Now().UnixMilli()
	err := db.Clauses(clause.OnConflict{
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

	return db.Clauses(clause.OnConflict{
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
}

func (dao *gormDAO) DeleteLikeInfo(ctx context.Context, biz string, bizID int64, uid int64) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return dao.deleteLikeInfo(tx, biz, bizID, uid)
	})
}

func (dao *gormDAO) BatchDeleteLikeInfo(ctx context.Context, biz string, bizIDs []int64, uids []int64) error {
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := 0; i < len(bizIDs); i++ {
			if err := dao.deleteLikeInfo(tx, biz, bizIDs[i], uids[i]); err != nil {
				return err
			}
		}
		return nil
	})
}

func (dao *gormDAO) deleteLikeInfo(db *gorm.DB, biz string, bizID int64, uid int64) error {
	now := time.Now().UnixMilli()
	err := db.Model(&UserLikeBiz{}).
		Where("biz = ? AND biz_id = ? AND uid = ?", biz, bizID, uid).
		Updates(map[string]any{
			"status":    0,
			"update_at": now,
		}).Error
	if err != nil {
		return err
	}
	return db.Clauses(clause.OnConflict{
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
}

func (dao *gormDAO) InsertCollectionBiz(ctx context.Context, cb UserCollectionBiz) error {
	now := time.Now().UnixMilli()
	cb.CreateAt, cb.UpdateAt = now, now
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := dao.db.WithContext(ctx).Create(&cb).Error; err != nil {
			return err
		}
		return tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]any{
				"collect_cnt": gorm.Expr("`collect_cnt`+1"),
				"update_at":   now,
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

func (dao *gormDAO) GetMultipleLikeCnt(ctx context.Context, biz string, bizIDs []int64) ([]Interactive, error) {
	res := make([]Interactive, 0, len(bizIDs))
	err := dao.db.WithContext(ctx).Model(UserLikeBiz{}).Select("biz_id,count(*) as like_cnt").
		Where("biz = ? AND biz_id IN ? AND status = 1", biz, bizIDs).
		Group("biz_id").
		Find(&res).Error
	return res, err
}

func (dao *gormDAO) GetByIDs(ctx context.Context, biz string, bizIDs []int64) ([]Interactive, error) {
	var res []Interactive
	err := dao.db.WithContext(ctx).Where("biz = ? AND id IN ?", biz, bizIDs).Find(&res).Error
	return res, err
}
