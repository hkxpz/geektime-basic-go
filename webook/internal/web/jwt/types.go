package jwt

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"geektime-basic-go/webook/pkg/ginx/handlefunc"
)

type Handler interface {
	ClearToken(ctx *gin.Context) error
	SetLoginToken(ctx *gin.Context, uid int64) error
	SetJWTToken(ctx *gin.Context, ssid string, uid int64) error
	CheckSession(ctx *gin.Context, ssid string) error
	ExtractTokenString(ctx *gin.Context) string
}

type UserClaims = handlefunc.UserClaims

type RefreshClaims struct {
	ID   int64
	SSID string
	jwt.RegisteredClaims
}
