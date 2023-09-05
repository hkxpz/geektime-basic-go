//go:build wireinject

package integration

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"geektime-basic-go/webook/internal/repository"
	redisCache "geektime-basic-go/webook/internal/repository/cache/redis"
	"geektime-basic-go/webook/internal/repository/dao"
	"geektime-basic-go/webook/internal/service"
	"geektime-basic-go/webook/internal/web"
	"geektime-basic-go/webook/ioc"
	ginServer "geektime-basic-go/webook/ioc/gin"
	"geektime-basic-go/webook/ioc/sms"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitRedis, ioc.InitDB,

		dao.NewUserDAO,

		redisCache.NewUserCache, redisCache.NewCodeCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,

		sms.InitSmsSvc,
		service.NewSMSCodeService,
		service.NewUserService,

		web.NewUserHandler,

		ginServer.Middlewares,

		ginServer.InitWebServer,
	)

	return gin.Default()
}