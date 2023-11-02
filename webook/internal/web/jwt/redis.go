package jwt

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var AccessTokenKey = []byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixm")
var RefreshTokenKey = []byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixA")

var _ Handler = (*redisHandler)(nil)

type redisHandler struct {
	cmd          redis.Cmdable
	rtExpiration time.Duration
}

func NewJWTHandler(cmd redis.Cmdable) Handler {
	return &redisHandler{cmd: cmd, rtExpiration: 7 * 24 * time.Hour}
}

func (rh *redisHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	uc := ctx.MustGet("user").(UserClaims)
	return rh.cmd.SetNX(ctx, rh.key(uc.SSID), "", rh.rtExpiration).Err()
}

func (rh *redisHandler) SetLoginToken(ctx *gin.Context, uid int64) error {
	ssid := uuid.New().String()
	if err := rh.SetJWTToken(ctx, ssid, uid); err != nil {
		return err
	}
	return rh.setRefreshToken(ctx, ssid, uid)
}

func (rh *redisHandler) SetJWTToken(ctx *gin.Context, ssid string, uid int64) error {
	uc := UserClaims{
		ID:        uid,
		SSID:      ssid,
		UserAgent: ctx.Request.UserAgent(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * time.Minute)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, uc).SignedString(AccessTokenKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", token)
	return nil
}

func (rh *redisHandler) CheckSession(ctx *gin.Context, ssid string) error {
	logout, err := rh.cmd.Exists(ctx, rh.key(ssid)).Result()
	if err != nil {
		return err
	}
	if logout > 0 {
		return errors.New("用户已经退出登录")
	}
	return nil
}

func (rh *redisHandler) ExtractTokenString(ctx *gin.Context) string {
	authCode := ctx.GetHeader("Authorization")
	if authCode == "" {
		return ""
	}
	authSegs := strings.SplitN(authCode, " ", 2)
	if len(authSegs) != 2 {
		// 格式不对
		return ""
	}
	return authSegs[1]
}

func (rh *redisHandler) setRefreshToken(ctx *gin.Context, ssid string, uid int64) error {
	rc := RefreshClaims{
		ID:   uid,
		SSID: ssid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, rc).SignedString(RefreshTokenKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", token)
	return nil
}

func (rh *redisHandler) key(ssid string) string {
	return "users:ssid:" + ssid
}
