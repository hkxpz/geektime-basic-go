package ioc

import (
	"github.com/redis/go-redis/v9"

	"geektime-basic-go/webook/config"
)

func InitRedis() redis.Cmdable {
	cfg := config.Config.Redis
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}
