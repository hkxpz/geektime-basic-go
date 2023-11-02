//go:build wireinject

package main

import (
	"github.com/google/wire"

	events "geektime-basic-go/webook/internal/events/article"
	"geektime-basic-go/webook/internal/repository"
	cache "geektime-basic-go/webook/internal/repository/cache/redis"
	"geektime-basic-go/webook/internal/repository/dao"
	"geektime-basic-go/webook/internal/repository/dao/article"
	"geektime-basic-go/webook/internal/service"
	"geektime-basic-go/webook/internal/web"
	webarticle "geektime-basic-go/webook/internal/web/article"
	"geektime-basic-go/webook/internal/web/jwt"
	"geektime-basic-go/webook/ioc"
	"geektime-basic-go/webook/ioc/sms"
)

var thirdProvider = wire.NewSet(
	ioc.InitDB,
	ioc.InitRedis,
	ioc.InitZapLogger,
	ioc.InitKafka,
	ioc.NewSyncProducer,
)

var userSvcProvider = wire.NewSet(
	service.NewUserService,
	repository.NewUserRepository,
	dao.NewUserDAO,
	cache.NewUserCache,
)

var codeSvcProvider = wire.NewSet(
	sms.InitSmsSvc,
	service.NewSMSCodeService,
	cache.NewCodeCache,
	repository.NewCodeRepository,
)

var articleSvcProvider = wire.NewSet(
	service.NewArticleService,
	repository.NewCacheArticleRepository,
	article.NewGormArticleDAO,
	cache.NewArticleCache,
)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewInteractiveRepository,
	dao.NewInteractiveDAO,
	cache.NewInteractiveCache,
)

var HandlerProvider = wire.NewSet(
	jwt.NewJWTHandler,
	web.NewUserHandler,
	web.NewOAuth2WechatHandler,
	webarticle.NewArticleHandler,
)

var eventsProvider = wire.NewSet(
	events.NewSaramaSyncProducer,
	events.NewInteractiveReadEventConsumer,
	ioc.NewConsumers,
)

func InitApp() *App {
	wire.Build(
		thirdProvider,

		// events 部分
		eventsProvider,

		// service
		userSvcProvider,
		codeSvcProvider,
		articleSvcProvider,
		interactiveSvcProvider,
		ioc.InitLocalWechatService,

		// handler 部分
		HandlerProvider,

		// gin 的中间件
		ioc.Middlewares,

		// Web 服务器
		ioc.InitWebServer,

		wire.Struct(new(App), "*"),
	)

	return new(App)
}
