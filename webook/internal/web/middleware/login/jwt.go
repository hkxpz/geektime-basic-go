package login

import (
	"net/http"

	"github.com/ecodeclub/ekit/set"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	myjwt "geektime-basic-go/webook/internal/web/jwt"
)

type JwtMiddlewareBuilder struct {
	publicPaths set.Set[string]
	myjwt.Handler
}

func NewJwtLoginMiddlewareBuilder(jwtHandler myjwt.Handler) *JwtMiddlewareBuilder {
	return &JwtMiddlewareBuilder{publicPaths: set.NewMapSet[string](3), Handler: jwtHandler}
}

func (j *JwtMiddlewareBuilder) SetIgnorePath(paths ...string) MiddlewareBuilder {
	for _, path := range paths {
		j.publicPaths.Add(path)
	}
	return j
}

func (j *JwtMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if j.publicPaths.Exist(ctx.Request.URL.Path) {
			return
		}

		tokenStr := j.ExtractTokenString(ctx)
		var uc myjwt.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return myjwt.AccessTokenKey, nil
		})
		if err != nil || token == nil || !token.Valid {
			// 不正确的 token
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		expireTime, err := uc.GetExpirationTime()
		if err != nil || expireTime == nil {
			// 拿不到过期时间或者token过期
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}
		if ctx.Request.UserAgent() != uc.UserAgent {
			// 换了一个 User-Agent，可能是攻击者
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if err = j.CheckSession(ctx, uc.SSID); err != nil {
			// 系统错误或者用户已经主动退出登录了
			// 也可以考虑说, 如果 redis 崩溃的时候, 就不去验证用户是不是主动退出
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set("user", uc)
	}
}
