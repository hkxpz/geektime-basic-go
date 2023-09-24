package repository

import (
	"context"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository/dao/article"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
}

type articleRepository struct {
	dao article.DAO
}

func NewArticleRepository(dao article.DAO) ArticleRepository {
	return &articleRepository{dao: dao}
}

func (repo *articleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	return repo.dao.Create(ctx, repo.toEntity(art))
}

func (repo *articleRepository) Update(ctx context.Context, art domain.Article) error {
	return repo.dao.UpdateByID(ctx, repo.toEntity(art))
}

func (repo *articleRepository) toEntity(art domain.Article) article.Article {
	return article.Article{
		ID:       art.ID,
		Title:    art.Title,
		Content:  art.Content,
		AuthorID: art.Author.ID,
		Status:   art.Status.ToUint8(),
	}
}
