package service

import (
	"context"
	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
}

type articleService struct {
	repo repository.ArticleRepository
}

func NewArticleService(repo repository.ArticleRepository) ArticleService {
	return &articleService{repo: repo}
}

func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.ID > 0 {
		err := a.repo.Update(ctx, art)
		return art.ID, err
	}

	return a.repo.Create(ctx, art)
}
