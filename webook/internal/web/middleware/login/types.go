package login

import "github.com/gin-gonic/gin"

type MiddlewareBuilder interface {
	SetIgnorePath(paths ...string) MiddlewareBuilder
	Build() gin.HandlerFunc
}
