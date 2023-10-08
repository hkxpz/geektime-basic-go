package repository

import (
	"context"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository/dao/article"
	"geektime-basic-go/webook/pkg/logger"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
}

type cacheArticleRepository struct {
	dao    article.DAO
	logger logger.Logger
}

func NewCacheArticleRepository(dao article.DAO, logger logger.Logger) ArticleRepository {
	return &cacheArticleRepository{dao: dao, logger: logger}
}

func (repo *cacheArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return repo.dao.Insert(ctx, repo.toEntity(art))
}

func (repo *cacheArticleRepository) Update(ctx context.Context, art domain.Article) error {
	return repo.dao.UpdateById(ctx, repo.toEntity(art))
}

func (repo *cacheArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	return repo.dao.Sync(ctx, repo.toEntity(art))
}

func (repo *cacheArticleRepository) toEntity(art domain.Article) article.Article {
	return article.Article{
		ID:       art.ID,
		Title:    art.Title,
		Content:  art.Content,
		AuthorID: art.Author.ID,
		Status:   art.Status.ToUint8(),
	}
}

func (repo *cacheArticleRepository) toDomain(art article.Article) domain.Article {
	return domain.Article{
		ID:      art.ID,
		Title:   art.Title,
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
		Author:  domain.Author{ID: art.AuthorID},
	}
}
