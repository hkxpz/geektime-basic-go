package web

import (
	"github.com/gin-gonic/gin"

	"geektime-basic-go/webook/internal/web/middleware/handlefunc"
)

type handler interface {
	RegisterRoutes(s *gin.Engine)
}

type Response = handlefunc.Response

func InternalServerError() Response {
	return Response{Code: 5, Msg: "系统错误"}
}
