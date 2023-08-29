package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository/cache"
)

type userCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewUserCache(cmd redis.Cmdable) cache.UserCache {
	return &userCache{cmd: cmd, expiration: time.Minute * 15}
}

func (uc *userCache) Delete(ctx context.Context, id int64) error {
	return uc.cmd.Del(ctx, uc.key(id)).Err()
}

func (uc *userCache) Get(ctx context.Context, id int64) (domain.User, error) {
	data, err := uc.cmd.Get(ctx, uc.key(id)).Result()
	if errors.Is(err, redis.Nil) {
		return domain.User{}, cache.ErrKeyNotExist
	}
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal([]byte(data), &u)
	return u, err
}

func (uc *userCache) Set(ctx context.Context, u domain.User) error {
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return uc.cmd.Set(ctx, uc.key(u.Id), data, uc.expiration).Err()
}

func (uc *userCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
