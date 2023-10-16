package service

import (
	"context"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid, id int64) error
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
