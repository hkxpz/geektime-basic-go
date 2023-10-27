package service

import (
	"context"

	"github.com/gin-gonic/gin"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid, id int64) error
	GetPublishedByID(ctx *gin.Context, id int64) (domain.Article, error)
	GetByID(ctx *gin.Context, id int64) (domain.Article, error)
	List(ctx *gin.Context, id int64, offset int, limit int) ([]domain.Article, error)
}

type articleService struct {
	repo   repository.ArticleRepository
	logger logger.Logger
}

func NewArticleService(repo repository.ArticleRepository, logger logger.Logger) ArticleService {
	return &articleService{repo: repo, logger: logger}
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

func (svc *articleService) GetPublishedByID(ctx *gin.Context, id int64) (domain.Article, error) {
	return svc.repo.GetPublishedById(ctx, id)
}

func (svc *articleService) GetByID(ctx *gin.Context, id int64) (domain.Article, error) {
	//TODO implement me
	panic("implement me")
}

func (svc *articleService) List(ctx *gin.Context, id int64, offset int, limit int) ([]domain.Article, error) {
	//TODO implement me
	panic("implement me")
}
