package article

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
)

var ErrPossibleIncorrectAuthor = errors.New("用户在尝试操作非本人数据")

//go:generate mockgen -source=types.go -package=mocks -destination=mocks/types_mock_gen.go DAO
type DAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, uid, id int64, status uint8) error
	GetPubByID(ctx *gin.Context, id int64) (PublishedArticle, error)
	GetByAuthor(ctx *gin.Context, author int64, offset int, limit int) ([]Article, error)
	GetByID(ctx *gin.Context, id int64) (Article, error)
}
