package handlefunc

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"

	"geektime-basic-go/webook/pkg/logger"
)

var log = logger.NewNoOpLogger()

var vector *prometheus.CounterVec

func InitCounter(opt prometheus.CounterOpts) {
	vector = prometheus.NewCounterVec(opt, []string{"code"})
	prometheus.MustRegister(vector)
}

func SetLogger(l logger.Logger) { log = l }

func WrapClaimsAndReq[Req any](fn func(*gin.Context, Req, UserClaims) (Response, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			log.Error("解析请求失败", logger.Error(err))
			return
		}

		rawVal, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			log.Error("无法获取 claims", logger.String("path", ctx.Request.URL.Path))
			return
		}
		claims, ok := rawVal.(UserClaims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			log.Error("无法获取 claims", logger.String("path", ctx.Request.URL.Path))
			return
		}

		res, err := fn(ctx, req, claims)
		vector.WithLabelValues(strconv.Itoa(res.Code)).Inc()
		if err != nil {
			log.Error("执行业务逻辑失败", logger.String("path", ctx.Request.URL.Path), logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapReq[Req any](fn func(*gin.Context, Req) (Response, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			log.Error("解析请求失败", logger.Error(err))
			return
		}

		res, err := fn(ctx, req)
		if err != nil {
			log.Error("执行业务逻辑失败", logger.String("path", ctx.Request.URL.Path), logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func Wrap(fn func(*gin.Context) (Response, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res, err := fn(ctx)
		if err != nil {
			log.Error("执行业务逻辑失败", logger.String("path", ctx.Request.URL.Path), logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}

func WrapClaims(fn func(*gin.Context, UserClaims) (Response, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		rawVal, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			log.Error("无法获取 claims", logger.String("path", ctx.Request.URL.Path))
			return
		}
		claims, ok := rawVal.(UserClaims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			log.Error("无法获取 claims", logger.String("path", ctx.Request.URL.Path))
			return
		}

		res, err := fn(ctx, claims)
		if err != nil {
			log.Error("执行业务逻辑失败", logger.String("path", ctx.Request.URL.Path), logger.Error(err))
		}
		ctx.JSON(http.StatusOK, res)
	}
}
