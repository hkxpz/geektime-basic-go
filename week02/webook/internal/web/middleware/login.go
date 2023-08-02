package middleware

import (
	"encoding/gob"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddlewareBuilder struct{}

func (*LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	gob.Register(time.Time{})
	return func(ctx *gin.Context) {
		if ctx.Request.URL.Path == "/users/signup" ||
			ctx.Request.URL.Path == "/users/login" {
			return
		}
		sess := sessions.Default(ctx)
		if sess.Get("userId") == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		const timeKey = "update_time"
		if val := sess.Get(timeKey); val != nil {
			if updateTime, ok := val.(time.Time); ok && time.Now().Sub(updateTime) > 10*time.Second {
				sess.Options(sessions.Options{MaxAge: 60})
				sess.Set(timeKey, time.Now())
				if err := sess.Save(); err != nil {
					panic(err)
				}
			}
		}
	}
}
