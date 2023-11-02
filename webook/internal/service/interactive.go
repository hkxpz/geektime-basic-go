package service

import (
	"context"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/pkg/logger"
)

//go:generate mockgen -source=interactive.go -package=mocks -destination=mocks/interactive_mock_gen.go InteractiveService
type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizID int64) error
	Get(ctx *gin.Context, biz string, bizID int64, uid int64) (domain.Interactive, error)
	CancelLike(ctx *gin.Context, biz string, bizID int64, uid int64) error
	Like(ctx *gin.Context, biz string, bizID int64, uid int64) error
	Collect(ctx *gin.Context, biz string, bizID int64, cid int64, uid int64) error
}

type interactiveService struct {
	repo repository.InteractiveRepository
	l    logger.Logger
}

func NewInteractiveService(repo repository.InteractiveRepository, l logger.Logger) InteractiveService {
	return &interactiveService{repo: repo, l: l}
}

func (svc *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizID int64) error {
	return svc.repo.IncrReadCnt(ctx, biz, bizID)
}

func (svc *interactiveService) Get(ctx *gin.Context, biz string, bizID int64, uid int64) (domain.Interactive, error) {
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

func (svc *interactiveService) CancelLike(ctx *gin.Context, biz string, bizID int64, uid int64) error {
	return svc.repo.DecrLike(ctx, biz, bizID, uid)
}

func (svc *interactiveService) Like(ctx *gin.Context, biz string, bizID int64, uid int64) error {
	return svc.repo.IncrLike(ctx, biz, bizID, uid)
}

func (svc *interactiveService) Collect(ctx *gin.Context, biz string, bizID int64, cid int64, uid int64) error {
	return svc.repo.AddCollectionItem(ctx, biz, bizID, cid, uid)
}
