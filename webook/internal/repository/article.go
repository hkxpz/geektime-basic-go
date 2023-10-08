package repository

import (
	"context"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository/cache"
	"geektime-basic-go/webook/internal/repository/dao/article"
	"geektime-basic-go/webook/pkg/logger"
)

type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
}

type cacheArticleRepository struct {
	dao      article.DAO
	userRepo UserRepository
	cache    cache.ArticleCache
	logger   logger.Logger
}

func NewCacheArticleRepository(dao article.DAO,
	userRepo UserRepository,
	cache cache.ArticleCache,
	logger logger.Logger,
) ArticleRepository {
	return &cacheArticleRepository{
		dao:      dao,
		userRepo: userRepo,
		cache:    cache,
		logger:   logger,
	}
}

func (repo *cacheArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	id, err := repo.dao.Insert(ctx, repo.toEntity(art))
	if err != nil {
		return 0, err
	}

	if err = repo.cache.DelFirstPage(ctx, art.Author.ID); err != nil {
		repo.logger.Error(
			"删除缓存失败",
			logger.Int64("author", art.Author.ID),
			logger.Error(err),
		)
	}

	return id, nil
}

func (repo *cacheArticleRepository) Update(ctx context.Context, art domain.Article) error {
	if err := repo.dao.UpdateById(ctx, repo.toEntity(art)); err != nil {
		return err
	}

	if err := repo.cache.DelFirstPage(ctx, art.Author.ID); err != nil {
		repo.logger.Error(
			"删除缓存失败",
			logger.Int64("author", art.Author.ID),
			logger.Error(err),
		)
	}

	return nil
}

func (repo *cacheArticleRepository) Sync(ctx context.Context, art domain.Article) (int64, error) {
	id, err := repo.dao.Sync(ctx, repo.toEntity(art))
	if err != nil {
		return 0, err
	}
	art.ID = id
	go func() {
		authorID := art.Author.ID
		if err := repo.cache.DelFirstPage(ctx, authorID); err != nil {
			repo.logger.Error(
				"删除缓存失败",
				logger.Int64("author", art.Author.ID),
				logger.Error(err),
			)
		}

		user, err := repo.userRepo.FindByID(ctx, authorID)
		if err != nil {
			repo.logger.Error(
				"提前设置缓存准备用户信息失败",
				logger.Int64("uid", authorID),
				logger.Error(err),
			)
		}
		art.ID = id
		art.Author = domain.Author{ID: user.ID, Name: user.Nickname}
		if err = repo.cache.SetPub(ctx, art); err != nil {
			repo.logger.Error(
				"提前设置缓存失败",
				logger.Int64("author", authorID),
				logger.Error(err),
			)
		}
	}()

	return id, nil
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
