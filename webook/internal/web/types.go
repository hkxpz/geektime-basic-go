package web

import (
	"github.com/gin-gonic/gin"

	"geektime-basic-go/webook/internal/errs"
	"geektime-basic-go/webook/pkg/ginx/handlefunc"
)

type handler interface {
	RegisterRoutes(s *gin.Engine)
}

type Response = handlefunc.Response

var InternalServerError = handlefunc.InternalServerErrorWith(errs.UserInternalServerError)
