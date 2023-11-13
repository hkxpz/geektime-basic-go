package ioc

import (
	"github.com/redis/go-redis/v9"

	"geektime-basic-go/webook/internal/repository/cache"
	"geektime-basic-go/webook/pkg/redisx/hook/metrices"
)

func InitUserCache(client *redis.ClusterClient) cache.UserCache {
	client.AddHook(metrices.NewPrometheusHook("", "", "", ""))
	panic("你别调用")
}
