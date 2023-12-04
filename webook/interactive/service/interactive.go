package service

import (
	"context"

	"golang.org/x/sync/errgroup"

	"geektime-basic-go/webook/interactive/domain"
	events "geektime-basic-go/webook/interactive/events/article"
	"geektime-basic-go/webook/interactive/repository"
	"geektime-basic-go/webook/pkg/logger"
)

//go:generate mockgen -source=interactive.go -package=mocks -destination=mocks/interactive_mock_gen.go InteractiveService
type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizID int64) error
	Get(ctx context.Context, biz string, bizID int64, uid int64) (domain.Interactive, error)
	Like(ctx context.Context, biz string, bizID int64, uid int64, like bool) error
	Collect(ctx context.Context, biz string, bizID int64, cid int64, uid int64) error
	GetByIDs(ctx context.Context, biz string, bizIDs []int64) (map[int64]domain.Interactive, error)
}

type interactiveService struct {
	repo     repository.InteractiveRepository
	producer events.ChangeLikeProducer
	l        logger.Logger
}

func NewInteractiveService(repo repository.InteractiveRepository, producer events.ChangeLikeProducer, l logger.Logger) InteractiveService {
	return &interactiveService{repo: repo, producer: producer, l: l}
}

func (svc *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizID int64) error {
	return svc.repo.IncrReadCnt(ctx, biz, bizID)
}

func (svc *interactiveService) Get(ctx context.Context, biz string, bizID int64, uid int64) (domain.Interactive, error) {
	intr, err := svc.repo.Get(ctx, biz, bizID)
	if err != nil {
		return domain.Interactive{}, err
	}

	var eg errgroup.Group
	eg.Go(func() error {
		intr.Liked, err = svc.repo.Liked(ctx, biz, bizID, uid)
		return err
	})
	eg.Go(func() error {
		intr.Collected, err = svc.repo.Collected(ctx, biz, bizID, uid)
		return err
	})
	if err = eg.Wait(); err != nil {
		svc.l.Error("查询用户是否点赞信息失败",
			logger.String("biz", biz),
			logger.Int("bizID", bizID),
			logger.Int("uid", uid),
			logger.Error(err),
		)
	}
	return intr, err
}

func (svc *interactiveService) Like(ctx context.Context, biz string, bizID int64, uid int64, like bool) error {
	return svc.producer.ProduceChangeLikeEvent(ctx, events.ChangeLikeEvent{BizID: bizID, Uid: uid, Liked: like})
}

func (svc *interactiveService) Collect(ctx context.Context, biz string, bizID int64, cid int64, uid int64) error {
	return svc.repo.AddCollectionItem(ctx, biz, bizID, cid, uid)
}

func (svc *interactiveService) GetByIDs(ctx context.Context, biz string, bizIDs []int64) (map[int64]domain.Interactive, error) {
	intrs, err := svc.repo.GetByIDs(ctx, biz, bizIDs)
	if err != nil {
		return nil, err
	}
	res := make(map[int64]domain.Interactive, len(intrs))
	for _, intr := range intrs {
		res[intr.BizID] = intr
	}
	return res, nil
}
