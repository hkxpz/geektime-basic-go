package repository

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository/cache"
	"geektime-basic-go/webook/internal/repository/dao"
	"geektime-basic-go/webook/pkg/logger"
)

//go:generate mockgen -source=interactive.go -package=mocks -destination=mocks/interactive_mock_gen.go InteractiveRepository
type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizID int64) error
	IncrLike(ctx *gin.Context, biz string, bizID int64, uid int64) error
	DecrLike(ctx *gin.Context, biz string, bizID int64, uid int64) error
	AddCollectionItem(ctx *gin.Context, biz string, bizID int64, cid int64, uid int64) error
	Get(ctx *gin.Context, biz string, bizID int64) (domain.Interactive, error)
	Liked(ctx *gin.Context, biz string, bizID int64, uid int64) (bool, error)
	Collected(ctx *gin.Context, biz string, bizID int64, uid int64) (bool, error)
	BatchIncrReadCnt(ctx context.Context, bizs []string, bizIds []int64) error
}

type cacheInteractiveRepository struct {
	cache cache.InteractiveCache
	dao   dao.InteractiveDAO
	l     logger.Logger
}

func NewInteractiveRepository(cache cache.InteractiveCache, dao dao.InteractiveDAO, l logger.Logger) InteractiveRepository {
	return &cacheInteractiveRepository{cache: cache, dao: dao, l: l}
}

func (repo *cacheInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizID int64) error {
	err := repo.dao.IncrReadCnt(ctx, biz, bizID)
	if err != nil {
		return err
	}
	return repo.cache.IncrReadCntIfPresent(ctx, biz, bizID)
}

func (repo *cacheInteractiveRepository) Get(ctx *gin.Context, biz string, bizID int64) (domain.Interactive, error) {
	intr, err := repo.cache.Get(ctx, biz, bizID)
	if err == nil {
		return intr, err
	}

	var ie dao.Interactive
	if ie, err = repo.dao.Get(ctx, biz, bizID); err != nil {
		return domain.Interactive{}, nil
	}
	res := repo.toDomain(ie)
	go func() {
		if err = repo.cache.Set(ctx, biz, bizID, res); err != nil {
			repo.l.Error("回写缓存失败", logger.String("biz", biz), logger.Int64("bizID", bizID), logger.Error(err))
		}
	}()
	return res, nil
}

func (repo *cacheInteractiveRepository) Liked(ctx *gin.Context, biz string, bizID int64, uid int64) (bool, error) {
	_, err := repo.dao.GetLikeInfo(ctx, biz, bizID, uid)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, dao.ErrDataNotFound):
		return false, nil
	default:
		return false, err
	}
}

func (repo *cacheInteractiveRepository) Collected(ctx *gin.Context, biz string, bizID int64, uid int64) (bool, error) {
	_, err := repo.dao.GetCollectionInfo(ctx, biz, bizID, uid)
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, dao.ErrDataNotFound):
		return false, nil
	default:
		return false, err
	}
}

func (repo *cacheInteractiveRepository) DecrLike(ctx *gin.Context, biz string, bizID int64, uid int64) error {
	if err := repo.dao.DeleteLikeInfo(ctx, biz, bizID, uid); err != nil {
		return err
	}
	return repo.cache.DecrLikeCntIfPresent(ctx, biz, bizID)
}

func (repo *cacheInteractiveRepository) IncrLike(ctx *gin.Context, biz string, bizID int64, uid int64) error {
	if err := repo.dao.InsertLikeInfo(ctx, biz, bizID, uid); err != nil {
		return err
	}
	return repo.cache.IncrLikeCntIfPresent(ctx, biz, bizID)
}

func (repo *cacheInteractiveRepository) AddCollectionItem(ctx *gin.Context, biz string, bizID int64, cid int64, uid int64) error {
	if err := repo.dao.InsertCollectionBiz(ctx, dao.UserCollectionBiz{CID: cid, BizID: bizID, Biz: biz, UID: uid}); err != nil {
		return err
	}
	return repo.cache.IncrCollectCntIfPresent(ctx, biz, bizID)
}

func (repo *cacheInteractiveRepository) toDomain(intr dao.Interactive) domain.Interactive {
	return domain.Interactive{
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
		ReadCnt:    intr.ReadCnt,
	}
}

func (repo *cacheInteractiveRepository) BatchIncrReadCnt(ctx context.Context, bizs []string, bizIDs []int64) error {
	return repo.dao.BatchIncrReadCnt(ctx, bizs, bizIDs)
}
