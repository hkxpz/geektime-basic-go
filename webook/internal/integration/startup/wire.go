//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"

	"geektime-basic-go/webook/internal/repository"
	redisCache "geektime-basic-go/webook/internal/repository/cache/redis"
	"geektime-basic-go/webook/internal/repository/dao"
	"geektime-basic-go/webook/internal/repository/dao/article"
	"geektime-basic-go/webook/internal/service"
	"geektime-basic-go/webook/internal/web"
	"geektime-basic-go/webook/internal/web/jwt"
	"geektime-basic-go/webook/ioc"
	"geektime-basic-go/webook/ioc/sms"
)

var thirdProvider = wire.NewSet(ioc.InitRedis, ioc.InitDB, ioc.InitZapLogger)

var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	redisCache.NewUserCache,
	repository.NewUserRepository,
	service.NewUserService,
)

var articleSvcProvider = wire.NewSet(
	article.NewGormArticleDAO,
	repository.NewArticleRepository,
	service.NewArticleService,
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdProvider,
		userSvcProvider,
		articleSvcProvider,

		repository.NewCodeRepository,
		redisCache.NewCodeCache,

		// svc
		sms.InitSmsSvc,
		service.NewSMSCodeService,
		ioc.InitLocalWechatService,

		// handler
		jwt.NewRedisHandler,
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,

		ioc.Middlewares,
		ioc.InitWebServer,
	)

	return gin.Default()
}

func InitArticleHandler() *web.ArticleHandler {
	wire.Build(thirdProvider, articleSvcProvider, web.NewArticleHandler)
	return new(web.ArticleHandler)
}
