package ratelimit

import (
	_ "embed"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"geektime-basic-go/webook/pkg/ratelimit"
)

type Builder struct {
	prefix  string
	limiter ratelimit.Limiter
}

func NewBuilder(limiter ratelimit.Limiter) *Builder {
	return &Builder{
		prefix:  "ip-limiter",
		limiter: limiter,
	}
}

func (b *Builder) SetPrefix(prefix string) *Builder {
	b.prefix = prefix
	return b
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limited, err := b.limit(ctx)
		if err != nil {
			log.Println(err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if limited {
			ctx.AbortWithStatus(http.StatusMisdirectedRequest)
			return
		}
		ctx.Next()
	}
}

func (b *Builder) limit(ctx *gin.Context) (bool, error) {
	key := b.prefix + ":" + ctx.ClientIP()
	return b.limiter.Limit(ctx, key)
}
