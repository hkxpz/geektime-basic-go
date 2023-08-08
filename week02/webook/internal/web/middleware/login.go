package middleware

import (
	"encoding/gob"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddlewareBuilder struct {
	ignorePath []string
}

func (l *LoginMiddlewareBuilder) IgnorePath(path ...string) *LoginMiddlewareBuilder {
	l.ignorePath = append(l.ignorePath, path...)
	return l
}

func (l *LoginMiddlewareBuilder) Build() gin.HandlerFunc {
	gob.Register(time.Time{})
	return func(ctx *gin.Context) {
		for _, path := range l.ignorePath {
			if ctx.Request.URL.Path == path {
				return
			}
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
