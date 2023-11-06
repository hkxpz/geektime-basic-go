package ioc

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"geektime-basic-go/webook/internal/web"
	"geektime-basic-go/webook/internal/web/article"
	myjwt "geektime-basic-go/webook/internal/web/jwt"
	"geektime-basic-go/webook/internal/web/middleware/login"
	"geektime-basic-go/webook/pkg/ginx/handlefunc"
	"geektime-basic-go/webook/pkg/ginx/metrics"
	"geektime-basic-go/webook/pkg/ginx/middleware/accesslog"
	"geektime-basic-go/webook/pkg/logger"
)

func InitWebServer(fn []gin.HandlerFunc,
	uh *web.UserHandler,
	ah *article.Handler,
	oh *web.OAuth2WechatHandler,
	l logger.Logger,
) *gin.Engine {
	handlefunc.SetLogger(l)
	server := gin.Default()
	server.Use(fn...)
	uh.RegisterRoutes(server)
	ah.RegisterRoutes(server)
	oh.RegisterRoutes(server)
	return server
}

func Middlewares(cmd redis.Cmdable, jwtHandler myjwt.Handler, l logger.Logger) []gin.HandlerFunc {
	pb := &metrics.PrometheusBuilder{
		NameSpace:  "hkxpz",
		Subsystem:  "webook",
		Name:       "gin_http",
		InstanceID: "instance-1",
		Help:       "GIN HTTP 请求",
	}
	return []gin.HandlerFunc{
		//ginRatelimit.NewBuilder(ratelimit.NewRedisSlideWindowLimiter(cmd, time.Minute, 100)).Build(),
		corsHandler(),
		pb.BuildResponseTime(),
		pb.BuildActiveRequest(),
		login.NewJwtLoginMiddlewareBuilder(jwtHandler).Build(),
		accesslog.NewBuilder(accesslog.DefaultLogFunc(l)).AllowReqBody().AllowRespBody().Build(),
	}
}

func corsHandler() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"X-Jwt-Token"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}

			return strings.Contains(origin, "your_company.com")
		},
		MaxAge: 12 * time.Hour,
	})
}
