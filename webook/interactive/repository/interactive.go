package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ecodeclub/ekit/slice"

	"geektime-basic-go/webook/interactive/domain"
	"geektime-basic-go/webook/interactive/repository/cache"
	"geektime-basic-go/webook/interactive/repository/dao"
	"geektime-basic-go/webook/pkg/logger"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizID int64) error
	IncrLike(ctx context.Context, biz string, bizID int64, uid int64) error
	DecrLike(ctx context.Context, biz string, bizID int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, bizID int64, cid int64, uid int64) error
	Get(ctx context.Context, biz string, bizID int64) (domain.Interactive, error)
	Liked(ctx context.Context, biz string, bizID int64, uid int64) (bool, error)
	Collected(ctx context.Context, biz string, bizID int64, uid int64) (bool, error)
	BatchIncrReadCnt(ctx context.Context, bizs []string, bizIDs []int64) error
	BatchIncrLike(ctx context.Context, biz string, bizIDs []int64, uids []int64) error
	BatchDecrLike(ctx context.Context, biz string, bizIDs []int64, uids []int64) error
	GetByIDs(ctx context.Context, biz string, bizIDs []int64) ([]domain.Interactive, error)
}

type cacheInteractiveRepository struct {
	cache cache.InteractiveCache
	dao   dao.InteractiveDAO
	l     logger.Logger
	m     sync.Map
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

func (repo *cacheInteractiveRepository) Get(ctx context.Context, biz string, bizID int64) (domain.Interactive, error) {
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
			repo.l.Error("回写缓存失败", logger.String("biz", biz), logger.Int("bizID", bizID), logger.Error(err))
		}
	}()
	return res, nil
}

func (repo *cacheInteractiveRepository) Liked(ctx context.Context, biz string, bizID int64, uid int64) (bool, error) {
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

func (repo *cacheInteractiveRepository) Collected(ctx context.Context, biz string, bizID int64, uid int64) (bool, error) {
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

func (repo *cacheInteractiveRepository) DecrLike(ctx context.Context, biz string, bizID int64, uid int64) error {
	if err := repo.dao.DeleteLikeInfo(ctx, biz, bizID, uid); err != nil {
		return err
	}
	return repo.cache.DecrLikeCntIfPresent(ctx, biz, bizID)
}

func (repo *cacheInteractiveRepository) IncrLike(ctx context.Context, biz string, bizID int64, uid int64) error {
	if err := repo.dao.InsertLikeInfo(ctx, biz, bizID, uid); err != nil {
		return err
	}
	return repo.cache.IncrLikeCntIfPresent(ctx, biz, bizID)
}

func (repo *cacheInteractiveRepository) AddCollectionItem(ctx context.Context, biz string, bizID int64, cid int64, uid int64) error {
	if err := repo.dao.InsertCollectionBiz(ctx, dao.UserCollectionBiz{CID: cid, BizID: bizID, Biz: biz, UID: uid}); err != nil {
		return err
	}
	return repo.cache.IncrCollectCntIfPresent(ctx, biz, bizID)
}

func (repo *cacheInteractiveRepository) toDomain(intr dao.Interactive) domain.Interactive {
	return domain.Interactive{
		BizID:      intr.BizID,
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
		ReadCnt:    intr.ReadCnt,
	}
}

func (repo *cacheInteractiveRepository) BatchIncrReadCnt(ctx context.Context, bizs []string, bizIDs []int64) error {
	return repo.dao.BatchIncrReadCnt(ctx, bizs, bizIDs)
}

func (repo *cacheInteractiveRepository) BatchIncrLike(ctx context.Context, biz string, bizIDs []int64, uids []int64) error {
	if err := repo.dao.BatchInsertLikeInfo(ctx, biz, bizIDs, uids); err != nil {
		return err
	}

	existBizIDs, doesNotExistBizIDs := repo.checkExists(biz, bizIDs)
	if err := repo.cache.BatchIncrLikeCntIfPresent(ctx, biz, existBizIDs); err != nil {
		return err
	}
	return repo.setLikeCntIfDoesNotExist(ctx, biz, doesNotExistBizIDs)
}

func (repo *cacheInteractiveRepository) BatchDecrLike(ctx context.Context, biz string, bizIDs []int64, uids []int64) error {
	if err := repo.dao.BatchDeleteLikeInfo(ctx, biz, bizIDs, uids); err != nil {
		return err
	}

	existBizIDs, doesNotExistBizIDs := repo.checkExists(biz, bizIDs)
	if err := repo.cache.BatchDecrLikeCntIfPresent(ctx, biz, existBizIDs); err != nil {
		return err
	}
	return repo.setLikeCntIfDoesNotExist(ctx, biz, doesNotExistBizIDs)
}

func (repo *cacheInteractiveRepository) setLikeCntIfDoesNotExist(ctx context.Context, biz string, bizIDs []int64) error {
	res, err := repo.dao.GetMultipleLikeCnt(ctx, biz, bizIDs)
	if err != nil {
		return err
	}

	cnts := make([]int64, len(res))
	setBizIDs := make([]int64, len(res))
	for idx := range res {
		cnts[idx] = res[idx].LikeCnt
		setBizIDs[idx] = res[idx].BizID
	}

	popKeys, err := repo.cache.BatchSetLikeCnt(ctx, biz, setBizIDs, cnts)
	for _, popKey := range popKeys {
		repo.m.Delete(popKey)
	}
	return err
}

func (repo *cacheInteractiveRepository) checkExists(biz string, bizIDs []int64) ([]int64, []int64) {
	existBizIDs := make([]int64, 0, len(bizIDs))
	doesNotExistBizIDs := make([]int64, 0, len(bizIDs))
	for _, bizID := range bizIDs {
		key := repo.key(biz, bizID)
		if _, ok := repo.m.Load(key); ok {
			existBizIDs = append(existBizIDs, bizID)
			continue
		}
		doesNotExistBizIDs = append(doesNotExistBizIDs, bizID)
		repo.m.Store(key, struct{}{})
	}
	return existBizIDs, doesNotExistBizIDs
}

func (repo *cacheInteractiveRepository) key(biz string, bizID int64) string {
	return fmt.Sprintf("interactive:%s:%d", biz, bizID)
}

func (repo *cacheInteractiveRepository) GetByIDs(ctx context.Context, biz string, bizIDs []int64) ([]domain.Interactive, error) {
	vals, err := repo.dao.GetByIDs(ctx, biz, bizIDs)
	if err != nil {
		return nil, err
	}
	return slice.Map(vals, func(idx int, src dao.Interactive) domain.Interactive {
		return repo.toDomain(src)
	}), nil
}
