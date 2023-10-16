package article

import (
	"context"
	"errors"
)

var ErrPossibleIncorrectAuthor = errors.New("用户在尝试操作非本人数据")

type DAO interface {
	Insert(ctx context.Context, art Article) (int64, error)
	UpdateById(ctx context.Context, art Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, uid, id int64, status uint8) error
}
