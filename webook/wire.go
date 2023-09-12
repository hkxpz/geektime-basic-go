//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"geektime-basic-go/webook/internal/repository"
	cache "geektime-basic-go/webook/internal/repository/cache/redis"
	"geektime-basic-go/webook/internal/repository/dao"
	"geektime-basic-go/webook/internal/service"
	"geektime-basic-go/webook/internal/web"
	"geektime-basic-go/webook/internal/web/jwt"
	"geektime-basic-go/webook/ioc"
	"geektime-basic-go/webook/ioc/sms"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,

		dao.NewUserDAO,

		cache.NewUserCache,
		cache.NewCodeCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,

		sms.InitSmsSvc,
		service.NewUserService,
		service.NewSMSCodeService,
		ioc.InitWechatService,

		jwt.NewRedisHandler,
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,

		ioc.Middlewares,

		ioc.InitWebServer,
	)

	return new(gin.Engine)
}
