//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/internal/repository/cache/memory"
	"geektime-basic-go/webook/internal/repository/cache/redis"
	"geektime-basic-go/webook/internal/repository/dao"
	"geektime-basic-go/webook/internal/service"
	"geektime-basic-go/webook/internal/web"
	"geektime-basic-go/webook/ioc"
	ginServer "geektime-basic-go/webook/ioc/gin"
	"geektime-basic-go/webook/ioc/sms"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,

		// dao
		dao.NewUserDAO,

		// cache
		redis.NewUserCache,
		//redis.NewCodeCache,
		ioc.InitMemory,
		memory.NewCodeCache,

		repository.NewUserRepository,
		repository.NewCodeRepository,

		// svc
		sms.InitSmsSvc,
		service.NewUserService,
		service.NewSMSCodeService,

		// handler
		web.NewUserHandler,

		// middleware
		ginServer.Middlewares,

		// web
		ginServer.InitWebServer,
	)

	return new(gin.Engine)
}
