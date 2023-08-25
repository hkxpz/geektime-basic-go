package ratelimit

import (
	_ "embed"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Builder struct {
	prefix   string
	cmd      redis.Cmdable
	interval time.Duration
	// 阈值
	rate int
}

//go:embed slide_window.lua
var luaScript string

func NewBuilder(cmd redis.Cmdable, interval time.Duration, rate int) *Builder {
	return &Builder{
		cmd:      cmd,
		prefix:   "ip-limiter",
		interval: interval,
		rate:     rate,
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
	return b.cmd.Eval(ctx, luaScript, []string{key}, b.interval.Microseconds(), b.rate, time.Now().UnixMilli()).Bool()
}
