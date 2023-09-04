package ratelimit

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed lua/slide_window.lua
var luaScript string

type RedisSlideWindowLimiter struct {
	cmd      redis.Cmdable
	interval time.Duration
	rate     int
}

func NewRedisSlideWindowLimiter(cmd redis.Cmdable, interval time.Duration, rate int) Limiter {
	return &RedisSlideWindowLimiter{cmd: cmd, interval: interval, rate: rate}
}

func (r *RedisSlideWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return r.cmd.Eval(ctx, luaScript, []string{key}, r.interval.Microseconds(), r.rate, time.Now().UnixMilli()).Bool()
}
