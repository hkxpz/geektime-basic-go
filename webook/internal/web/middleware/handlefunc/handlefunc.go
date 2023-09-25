package handlefunc

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	myjwt "geektime-basic-go/webook/internal/web/jwt"
	"geektime-basic-go/webook/pkg/logger"
)

func WrapReqWithLog[T any](fn func(ctx *gin.Context, req T, uc myjwt.UserClaims) (Response, error), logfn LogFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp, err := wrapReq[T](ctx, fn)
		if err != nil && logfn != nil {
			logfn(ctx.Request.Context(), err)
		}
		ctx.JSON(http.StatusOK, resp)
	}
}

func wrapReq[T any](ctx *gin.Context, fn func(ctx *gin.Context, req T, uc myjwt.UserClaims) (Response, error)) (Response, error) {
	var req T
	if err := ctx.ShouldBind(&req); err != nil {
		return InternalServerError(), err
	}

	uc := ctx.MustGet("user").(myjwt.UserClaims)
	return fn(ctx, req, uc)
}

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func InternalServerError() Response {
	return Response{Code: 5, Msg: "系统错误"}
}

type ReqLog struct {
}

type LogFunc func(ctx context.Context, args any)

func DefaultLogFunc(l logger.Logger) LogFunc {
	return func(ctx context.Context, args any) {
		// 设置为 DEBUG 级别
		l.Debug("error", logger.Field{Key: "error", Value: args})
	}
}
