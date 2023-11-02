package repository

import (
	"context"

	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository/cache"
	"geektime-basic-go/webook/internal/repository/dao/article"
	"geektime-basic-go/webook/pkg/logger"
)

//go:generate mockgen -source=article.go -package=mocks -destination=mocks/acticle_mock_gen.go ArticleRepository
type ArticleRepository interface {
	Create(ctx context.Context, art domain.Article) (int64, error)
	Update(ctx context.Context, art domain.Article) error
	Sync(ctx context.Context, art domain.Article) (int64, error)
	SyncStatus(ctx context.Context, uid, id int64, status domain.ArticleStatus) error
	GetPublishedById(ctx *gin.Context, id int64) (domain.Article, error)
	GetById(ctx *gin.Context, id int64) (domain.Article, error)
	List(ctx *gin.Context, author int64, offset int, limit int) ([]domain.Article, error)
}

type cacheArticleRepository struct {
	dao      article.DAO
	userRepo UserRepository
	cache    cache.ArticleCache
	l        logger.Logger
}

func NewCacheArticleRepository(dao article.DAO,
	userRepo UserRepository,
	cache cache.ArticleCache,
	l logger.Logger,
) ArticleRepository {
	return &cacheArticleRepository{
		dao:      dao,
		userRepo: userRepo,
		cache:    cache,
		l:        l,
	}
}

func (repo *cacheArticleRepository) Create(ctx context.Context, art domain.Article) (int64, error) {
	id, err := repo.dao.Insert(ctx, repo.toEntity(art))
	if err != nil {
		return 0, err
	}

	if err = repo.cache.DelFirstPage(ctx, art.Author.ID); err != nil {
		repo.l.Error(
			"删除缓存失败",
			logger.Int("author", art.Author.ID),
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
		repo.l.Error(
			"删除缓存失败",
			logger.Int("author", art.Author.ID),
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
			repo.l.Error(
				"删除缓存失败",
				logger.Int("author", art.Author.ID),
				logger.Error(err),
			)
		}

		user, err := repo.userRepo.FindByID(ctx, authorID)
		if err != nil {
			repo.l.Error(
				"提前设置缓存准备用户信息失败",
				logger.Int("uid", authorID),
				logger.Error(err),
			)
		}
		art.ID = id
		art.Author = domain.Author{ID: user.ID, Name: user.Nickname}
		if err = repo.cache.SetPub(ctx, art); err != nil {
			repo.l.Error(
				"提前设置缓存失败",
				logger.Int("author", authorID),
				logger.Error(err),
			)
		}
	}()

	return id, nil
}

func (repo *cacheArticleRepository) SyncStatus(ctx context.Context, uid, id int64, status domain.ArticleStatus) error {
	return repo.dao.SyncStatus(ctx, uid, id, status.ToUint8())
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

func (repo *cacheArticleRepository) GetPublishedById(ctx *gin.Context, id int64) (domain.Article, error) {
	res, err := repo.cache.GetPub(ctx, id)
	if err == nil {
		return res, err
	}
	art, err := repo.dao.GetPubByID(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	user, err := repo.userRepo.FindByID(ctx, art.AuthorID)
	if err != nil {
		return domain.Article{}, err
	}
	res = domain.Article{
		ID:      art.ID,
		Title:   art.Title,
		Status:  domain.ArticleStatus(art.Status),
		Content: art.Content,
		Author:  domain.Author{ID: user.ID, Name: user.Nickname},
	}

	go func() {
		if err = repo.cache.SetPub(ctx, res); err != nil {
			repo.l.Error("缓存已发表文章失败", logger.Error(err), logger.Int("aid", res.ID))
		}
	}()
	return res, nil
}

func (repo *cacheArticleRepository) GetById(ctx *gin.Context, id int64) (domain.Article, error) {
	cachedArt, err := repo.cache.Get(ctx, id)
	if err == nil {
		return cachedArt, nil
	}
	art, err := repo.dao.GetByID(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	return repo.toDomain(art), nil
}

func (repo *cacheArticleRepository) List(ctx *gin.Context, author int64, offset int, limit int) ([]domain.Article, error) {
	if offset == 0 && limit == 100 {
		data, err := repo.cache.GetFirstPage(ctx, author)
		if err != nil {
			repo.l.Error("查询缓存文章失败", logger.Int64("author", author), logger.Error(err))
		}

		go func() { repo.preCache(ctx, data) }()
		return data, nil
	}

	arts, err := repo.dao.GetByAuthor(ctx, author, offset, limit)
	if err != nil {
		return nil, err
	}
	res := slice.Map[article.Article, domain.Article](arts, func(idx int, src article.Article) domain.Article {
		return repo.toDomain(arts[idx])
	})
	go func() { repo.preCache(ctx, res) }()
	if err = repo.cache.SetFirstPage(ctx, author, res); err != nil {
		repo.l.Error("刷新第一页文章的缓存失败", logger.Int64("author", author), logger.Error(err))
	}
	return res, nil
}

func (repo *cacheArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	const contentSizeThreshold = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) <= contentSizeThreshold {
		if err := repo.cache.Set(ctx, arts[0]); err != nil {
			repo.l.Error("提前准备缓存失败", logger.Error(err))
		}
	}
}
