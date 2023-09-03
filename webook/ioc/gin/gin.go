//go:build !local

package gin

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"geektime-basic-go/webook/internal/web"
	"geektime-basic-go/webook/internal/web/middleware/login"
	"geektime-basic-go/webook/pkg/ginx/middleware/ratelimit"
)

func InitWebServer(uh *web.UserHandler, fn []gin.HandlerFunc) *gin.Engine {
	server := gin.Default()
	server.Use(fn...)
	uh.RegisterRoutes(server)
	return server
}

func Middlewares(cmd redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		corsHandler(),
		login.NewJwtLoginMiddlewareBuilder().SetIgnorePath("/users/signup", "/users/login_sms/code/send",
			"/users/login_sms", "/users/login").Build(),
		ratelimit.NewBuilder(cmd, time.Minute, 100).Build(),
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
