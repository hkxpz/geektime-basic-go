package ioc

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"geektime-basic-go/webook/internal/web"
	myjwt "geektime-basic-go/webook/internal/web/jwt"
	"geektime-basic-go/webook/internal/web/middleware/login"
	"geektime-basic-go/webook/pkg/ginx/middleware/accesslog"
	ginRatelimit "geektime-basic-go/webook/pkg/ginx/middleware/ratelimit"
	"geektime-basic-go/webook/pkg/logger"
	"geektime-basic-go/webook/pkg/ratelimit"
)

func InitWebServer(uh *web.UserHandler, ah *web.ArticleHandler, oh *web.OAuth2WechatHandler, fn []gin.HandlerFunc) *gin.Engine {
	server := gin.Default()
	server.Use(fn...)
	uh.RegisterRoutes(server)
	ah.RegisterRoutes(server)
	oh.RegisterRoutes(server)
	return server
}

func Middlewares(cmd redis.Cmdable, jwtHandler myjwt.Handler, l logger.Logger) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		ginRatelimit.NewBuilder(ratelimit.NewRedisSlideWindowLimiter(cmd, time.Minute, 100)).Build(),
		corsHandler(),
		login.NewJwtLoginMiddlewareBuilder(jwtHandler).
			SetIgnorePath("/users/signup", "/users/login", "/users/refresh_token").
			SetIgnorePath("/users/login_sms/code/send", "/users/login_sms").
			SetIgnorePath("/oauth2/wechat/authurl", "/oauth2/wechat/callback").
			Build(),
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
