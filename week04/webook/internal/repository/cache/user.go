package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"geektime-basic-go/week04/webook/internal/domain"
)

var ErrKeyNotExists = redis.Nil

type UserCache interface {
	Delete(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
}

type RedisUserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewRedisUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{cmd: cmd, expiration: time.Minute * 15}
}

func (rc *RedisUserCache) Delete(ctx context.Context, id int64) error {
	return rc.cmd.Del(ctx, rc.key(id)).Err()
}

func (rc *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	data, err := rc.cmd.Get(ctx, rc.key(id)).Result()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal([]byte(data), &u)
	return u, err
}

func (rc *RedisUserCache) Set(ctx context.Context, u domain.User) error {
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return rc.cmd.Set(ctx, rc.key(u.Id), data, rc.expiration).Err()
}

func (rc *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
