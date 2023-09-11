package jwt

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var AccessTokenKey = []byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixm")
var RefreshTokenKey = []byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixA")

var _ Handler = (*RedisHandler)(nil)

type RedisHandler struct {
	cmd          redis.Cmdable
	rtExpiration time.Duration
}

func NewRedisHandler(cmd redis.Cmdable) Handler {
	return &RedisHandler{
		cmd:          cmd,
		rtExpiration: 7 * 24 * time.Hour,
	}
}

func (rh *RedisHandler) ClearToken(ctx *gin.Context) error {
	//TODO implement me
	panic("implement me")
}

func (rh *RedisHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	err := rh.SetJWTToken(ctx, ssid, uid)
	if err != nil {
		return err
	}
	return rh.setRefreshToken(ctx, ssid, uid)
}

func (rh *RedisHandler) SetJWTToken(ctx *gin.Context, ssid string, uid int64) error {
	token, err := newJWTToken(ctx, uid, ssid)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", token)
	return nil
}

func (rh *RedisHandler) CheckSession(ctx *gin.Context, ssid string) error {
	//TODO implement me
	panic("implement me")
}

func (rh *RedisHandler) ExtractTokenString(ctx *gin.Context) string {
	//TODO implement me
	panic("implement me")
}

func (rh *RedisHandler) setRefreshToken(ctx *gin.Context, ssid string, uid int64) error {
	return nil
}

func newJWTToken(ctx *gin.Context, uid int64, ssid string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, UserClaims{
		ID:        uid,
		SSID:      ssid,
		UserAgent: ctx.Request.UserAgent(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
		},
	})
	return token.SignedString(AccessTokenKey)
}
