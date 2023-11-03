package service

import (
	"context"

	"github.com/gin-gonic/gin"

	"geektime-basic-go/webook/internal/domain"
	events "geektime-basic-go/webook/internal/events/article"
	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/pkg/logger"
)

//go:generate mockgen -source=article.go -package=mocks -destination=mocks/article_mock_gen.go ArticleService
type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid, id int64) error
	GetPublishedByID(ctx *gin.Context, id, uid int64) (domain.Article, error)
	GetByID(ctx *gin.Context, id int64) (domain.Article, error)
	List(ctx *gin.Context, id int64, offset int, limit int) ([]domain.Article, error)
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

func (svc *articleService) GetPublishedByID(ctx *gin.Context, id, uid int64) (domain.Article, error) {
	res, err := svc.repo.GetPublishedById(ctx, id)
	go func() {
		if err != nil {
			return
		}

		if e := svc.producer.ProduceReadEvent(events.ReadEvent{Aid: id, Uid: uid}); err != nil {
			svc.logger.Error("发送消息失败", logger.Int("uid", uid), logger.Int("aid", id), logger.Error(e))
		}
	}()
	return res, err
}

func (svc *articleService) GetByID(ctx *gin.Context, id int64) (domain.Article, error) {
	return svc.repo.GetById(ctx, id)
}

func (svc *articleService) List(ctx *gin.Context, author int64, offset int, limit int) ([]domain.Article, error) {
	return svc.repo.List(ctx, author, offset, limit)
}
