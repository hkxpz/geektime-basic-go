//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	events "geektime-basic-go/webook/internal/events/article"
	"geektime-basic-go/webook/internal/repository"
	redisCache "geektime-basic-go/webook/internal/repository/cache/redis"
	"geektime-basic-go/webook/internal/repository/dao"
	"geektime-basic-go/webook/internal/repository/dao/article"
	"geektime-basic-go/webook/internal/service"
	"geektime-basic-go/webook/internal/web"
	webarticle "geektime-basic-go/webook/internal/web/article"
	myjwt "geektime-basic-go/webook/internal/web/jwt"
)

var thirdProvider = wire.NewSet(
	InitDB,
	InitZapLogger,
	InitRedis,
	InitKafka,
	NewSyncProducer,
)

var userSvcProvider = wire.NewSet(
	service.NewUserService,
	repository.NewUserRepository,
	dao.NewUserDAO,
	redisCache.NewUserCache,
)

var articleSvcProvider = wire.NewSet(
	service.NewArticleService,
	repository.NewCacheArticleRepository,
	article.NewGormArticleDAO,
	redisCache.NewArticleCache,
)

var codeSvcProvider = wire.NewSet(
	InitSmsSvc,
	service.NewSMSCodeService,
	repository.NewCodeRepository,
	redisCache.NewCodeCache,
)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewInteractiveRepository,
	dao.NewInteractiveDAO,
	redisCache.NewInteractiveCache,
)

var eventsProvider = wire.NewSet(
	events.NewSaramaSyncProducer,
	events.NewInteractiveReadEventConsumer,
	events.NewInteractiveLikeEventConsumer,
	events.NewChangeLikeSaramaSyncProducer,
	NewConsumers,
)

var HandlerProvider = wire.NewSet(
	myjwt.NewJWTHandler,
	web.NewUserHandler,
	web.NewOAuth2WechatHandler,
	webarticle.NewArticleHandler,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdProvider,

		// events 部分
		eventsProvider,

		// service
		userSvcProvider,
		codeSvcProvider,
		articleSvcProvider,
		interactiveSvcProvider,
		InitLocalWechatService,

		// handler 部分
		HandlerProvider,

		// gin 的中间件
		Middlewares,

		// Web 服务器
		InitWeb,
	)

	return gin.Default()
}

func InitArticleHandlerWithDAO(dao article.DAO) *webarticle.Handler {
	wire.Build(
		thirdProvider,
		userSvcProvider,
		interactiveSvcProvider,
		events.NewSaramaSyncProducer,
		events.NewChangeLikeSaramaSyncProducer,
		service.NewArticleService,
		repository.NewCacheArticleRepository,
		redisCache.NewArticleCache,
		webarticle.NewArticleHandler,
	)
	return new(webarticle.Handler)
}

func InitArticleHandlerWithKafka() *webarticle.Handler {
	wire.Build(
		thirdProvider,
		userSvcProvider,
		interactiveSvcProvider,
		articleSvcProvider,
		events.NewSaramaSyncProducer,
		events.NewChangeLikeSaramaSyncProducer,
		webarticle.NewArticleHandler,
	)
	return new(webarticle.Handler)
}

func InitInteractiveLikeEventConsumer() *events.ChangeLikeEventConsumer {
	wire.Build(
		thirdProvider,
		interactiveSvcProvider,
		events.NewInteractiveLikeEventConsumer,
	)

	return new(events.ChangeLikeEventConsumer)
}

func InitInteractiveService() service.InteractiveService {
	wire.Build(
		thirdProvider,
		interactiveSvcProvider,
		events.NewChangeLikeSaramaSyncProducer,
	)
	return service.NewInteractiveService(nil, nil, nil)
}
