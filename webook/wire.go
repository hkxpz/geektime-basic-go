//go:build wireinject

package main

import (
	"github.com/google/wire"

	intrevents "geektime-basic-go/webook/interactive/events/article"
	intrioc "geektime-basic-go/webook/interactive/ioc"
	intrrepo "geektime-basic-go/webook/interactive/repository"
	intrcache "geektime-basic-go/webook/interactive/repository/cache"
	intrdao "geektime-basic-go/webook/interactive/repository/dao"
	intrscv "geektime-basic-go/webook/interactive/service"

	events "geektime-basic-go/webook/internal/events/article"
	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/internal/repository/cache/memory"
	cache "geektime-basic-go/webook/internal/repository/cache/redis"
	"geektime-basic-go/webook/internal/repository/dao"
	"geektime-basic-go/webook/internal/repository/dao/article"
	"geektime-basic-go/webook/internal/service"
	"geektime-basic-go/webook/internal/web"
	webarticle "geektime-basic-go/webook/internal/web/article"
	myjwt "geektime-basic-go/webook/internal/web/jwt"
	"geektime-basic-go/webook/ioc"
	"geektime-basic-go/webook/ioc/sms"
)

// todo 后续移除
var interactiveSvcProvider = wire.NewSet(
	intrscv.NewInteractiveService,
	intrrepo.NewInteractiveRepository,
	intrdao.NewInteractiveDAO,
	intrcache.NewInteractiveCache,
	intrevents.NewChangeLikeSaramaSyncProducer,
	intrevents.NewInteractiveLikeEventConsumer,
	intrevents.NewInteractiveReadEventConsumer,
	intrioc.NewConsumers,
)

var thirdProvider = wire.NewSet(
	ioc.InitDB,
	ioc.InitRedis,
	ioc.InitZapLogger,
	ioc.InitKafka,
	ioc.InitRLockClient,
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

var rankServiceProvider = wire.NewSet(
	service.NewBatchRankingService,
	repository.NewCacheRankingRepository,
	cache.NewRankingCache,
	memory.NewRankingCache,
)

var HandlerProvider = wire.NewSet(
	myjwt.NewJWTHandler,
	web.NewUserHandler,
	web.NewOAuth2WechatHandler,
	webarticle.NewArticleHandler,
)

var producerProvider = wire.NewSet(
	ioc.NewSyncProducer,
	events.NewSaramaSyncProducer,
)

var grpcClientProvider = wire.NewSet(
	ioc.InitInteractiveGRPCClient,
)

var jobProvider = wire.NewSet(
	ioc.InitJobs,
	ioc.InitRankingJob,
)

func InitApp() *App {
	wire.Build(
		thirdProvider,
		grpcClientProvider,

		// todo 后续移除
		interactiveSvcProvider,

		// events 部分
		producerProvider,

		// job 部分
		jobProvider,

		// service
		userSvcProvider,
		codeSvcProvider,
		articleSvcProvider,
		rankServiceProvider,
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
