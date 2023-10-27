package handlefunc

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"geektime-basic-go/webook/pkg/logger"
)

var log = logger.NewNoOpLogger()

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
