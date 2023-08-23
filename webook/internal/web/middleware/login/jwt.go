package login

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ecodeclub/ekit/set"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"geektime-basic-go/webook/internal/web"
)

type jwtLoginMiddlewareBuilder struct {
	publicPaths set.Set[string]
}

func NewJwtLoginMiddlewareBuilder() MiddlewareBuilder {
	return &jwtLoginMiddlewareBuilder{publicPaths: set.NewMapSet[string](3)}
}

func (j *jwtLoginMiddlewareBuilder) SetIgnorePath(paths ...string) MiddlewareBuilder {
	for _, path := range paths {
		j.publicPaths.Add(path)
	}
	return j
}

func (j *jwtLoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if j.publicPaths.Exist(ctx.Request.URL.Path) {
			return
		}

		authCode := ctx.GetHeader("Authorization")
		if authCode == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		authSegments := strings.SplitN(authCode, " ", 2)
		if len(authSegments) != 2 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenStr := authSegments[1]
		uc := web.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return web.JWTKey, nil
		})
		if err != nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		expireTime, err := uc.GetExpirationTime()
		if err != nil || expireTime.Before(time.Now()) {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}

		if ctx.Request.UserAgent() != uc.UserAgent {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if expireTime.Sub(time.Now()) < 50*time.Second {
			uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))
			newToken, err := token.SignedString(web.JWTKey)
			if err != nil {
				log.Println(err)
			} else {
				ctx.Header("x-jwt-token", newToken)
			}
		}

		ctx.Set("user", uc)
	}
}
