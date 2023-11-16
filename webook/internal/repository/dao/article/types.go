package article

import (
	"context"
	"errors"
	"time"
)

var ErrPossibleIncorrectAuthor = errors.New("用户在尝试操作非本人数据")

type DAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, uid, id int64, status uint8) error
	GetPubByID(ctx context.Context, id int64) (PublishedArticle, error)
	GetByID(ctx context.Context, id int64) (Article, error)
	GetByAuthor(ctx context.Context, author int64, offset, limit int) ([]Article, error)
	ListPubByCreateAt(ctx context.Context, updateAt time.Time, offset int, limit int) ([]PublishedArticle, error)
}
