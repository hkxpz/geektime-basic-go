package web

import (
	"github.com/gin-gonic/gin"

	"geektime-basic-go/webook/internal/web/middleware/handlefunc"
)

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type handler interface {
	RegisterRoutes(s *gin.Engine)
}

func InternalServerError() handlefunc.Response {
	return handlefunc.Response{Code: 5, Msg: "系统错误"}
}
