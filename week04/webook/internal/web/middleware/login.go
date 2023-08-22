package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"

	"geektime-basic-go/week04/webook/internal/web"
	"github.com/gin-gonic/gin"
)

type JWTLoginMiddlewareBuilder struct{}

func (j *JWTLoginMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.URL.Path == "/users/signup" || ctx.Request.URL.Path == "/users/login" {
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
