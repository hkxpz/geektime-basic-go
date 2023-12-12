package service

import (
	"context"
	"time"

	"geektime-basic-go/webook/internal/domain"
	events "geektime-basic-go/webook/internal/events/article"
	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/pkg/logger"
)

//go:generate mockgen -source=article.go -package=svcmocks -destination=mocks/article_mock_gen.go ArticleService
type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid, id int64) error
	GetPublishedByID(ctx context.Context, id, uid int64) (domain.Article, error)
	GetByID(ctx context.Context, id int64) (domain.Article, error)
	List(ctx context.Context, id int64, offset int, limit int) ([]domain.Article, error)
	ListPub(ctx context.Context, updateAt time.Time, offset, limit int) ([]domain.Article, error)
	PubDetail(ctx context.Context, bizID int64, uid int64) (domain.Vo, error)
	Collect(ctx context.Context, biz string, bizID int64, cid int64, uid int64) error
	Like(ctx context.Context, biz string, bizID int64, uid int64, like bool) error
}

type articleService struct {
	repo     repository.ArticleRepository
	logger   logger.Logger
	producer events.Producer
}

func NewArticleService(repo repository.ArticleRepository, logger logger.Logger, producer events.Producer) ArticleService {
	return &articleService{repo: repo, logger: logger, producer: producer}
}

func (svc *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.ID > 0 {
		return art.ID, svc.repo.Update(ctx, art)
	}
	return svc.repo.Create(ctx, art)
}

func (svc *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return svc.repo.Sync(ctx, art)
}

func (svc *articleService) Withdraw(ctx context.Context, uid, id int64) error {
	return svc.repo.SyncStatus(ctx, uid, id, domain.ArticleStatusPrivate)
}

func (svc *articleService) GetPublishedByID(ctx context.Context, id, uid int64) (domain.Article, error) {
	res, err := svc.repo.GetPublishedByID(ctx, id)
	defer func() {
		go func() {
			if err != nil {
				return
			}

			if e := svc.producer.ProduceReadEvent(events.ReadEvent{Aid: id, Uid: uid}); err != nil {
				svc.logger.Error("发送消息失败", logger.Int("uid", uid), logger.Int("aid", id), logger.Error(e))
			}
		}()
	}()
	return res, err
}

func (svc *articleService) GetByID(ctx context.Context, id int64) (domain.Article, error) {
	return svc.repo.GetById(ctx, id)
}

func (svc *articleService) List(ctx context.Context, author int64, offset int, limit int) ([]domain.Article, error) {
	return svc.repo.List(ctx, author, offset, limit)
}

func (svc *articleService) ListPub(ctx context.Context, updateAt time.Time, offset, limit int) ([]domain.Article, error) {
	return svc.repo.ListPub(ctx, updateAt, offset, limit)
}

func (svc *articleService) PubDetail(ctx context.Context, bizID int64, uid int64) (domain.Vo, error) {
	res, err := svc.repo.PubDetail(ctx, bizID, uid)
	defer func() {
		go func() {
			if err != nil {
				return
			}

			if e := svc.producer.ProduceReadEvent(events.ReadEvent{Aid: bizID, Uid: uid}); err != nil {
				svc.logger.Error("发送消息失败", logger.Int("uid", uid), logger.Int("aid", bizID), logger.Error(e))
			}
		}()
	}()

	return res, err
}

func (svc *articleService) Collect(ctx context.Context, biz string, bizID int64, cid int64, uid int64) error {
	return svc.repo.Collect(ctx, biz, bizID, cid, uid)
}

func (svc *articleService) Like(ctx context.Context, biz string, bizID int64, uid int64, like bool) error {
	return svc.repo.Like(ctx, biz, bizID, uid, like)
}
