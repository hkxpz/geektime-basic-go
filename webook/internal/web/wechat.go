package web

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"

	"geektime-basic-go/webook/internal/service"
	"geektime-basic-go/webook/internal/service/oauth2/wechat"
	myjwt "geektime-basic-go/webook/internal/web/jwt"
)

var _ handler = (*OAuth2WechatHandler)(nil)

type OAuth2WechatHandler struct {
	svc             wechat.Service
	userSvc         service.UserService
	stateTokenKey   []byte
	stateCookieName string
	myjwt.Handler
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserService, handler myjwt.Handler) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:             svc,
		userSvc:         userSvc,
		Handler:         handler,
		stateCookieName: "jwt-state",
		stateTokenKey:   []byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixB"),
	}
}

func (oh *OAuth2WechatHandler) RegisterRoutes(s *gin.Engine) {
	g := s.Group("/oauth2/wechat")
	g.GET("/authurl", oh.OAuth2URL)
	g.Any("/callback", oh.Callback)
}

func (oh *OAuth2WechatHandler) OAuth2URL(ctx *gin.Context) {
	state := uuid.New()
	url, err := oh.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, InternalServerError)
		return
	}
	if err = oh.setStateCookie(ctx, state); err != nil {
		ctx.JSON(http.StatusOK, InternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, Response{Data: url})
}

func (oh *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	if err := oh.verifyState(ctx); err != nil {
		ctx.JSON(http.StatusOK, InternalServerError)
		return
	}

	code := ctx.Query("code")
	info, err := oh.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, InternalServerError)
		return
	}
	u, err := oh.userSvc.FindOrCreateByWechat(ctx, info)
	if err != nil {
		ctx.JSON(http.StatusOK, InternalServerError)
		return
	}
	if err = oh.SetLoginToken(ctx, u.ID); err != nil {
		ctx.JSON(http.StatusOK, InternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, Response{Msg: "登陆成功"})
}

func (oh *OAuth2WechatHandler) setStateCookie(ctx *gin.Context, state string) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, StateClaims{State: state})
	tokenStr, err := token.SignedString(oh.stateTokenKey)
	if err != nil {
		return err
	}
	ctx.SetCookie("jwt-state", tokenStr, 600, "/outh2/wechat/callback", "", false, true)
	return nil
}

func (oh *OAuth2WechatHandler) verifyState(ctx *gin.Context) interface{} {
	stata := ctx.Query("state")
	ck, err := ctx.Cookie(oh.stateCookieName)
	if err != nil {
		return fmt.Errorf("%w, 无法获得 cookie", err)
	}

	var sc StateClaims
	token, err := jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return oh.stateTokenKey, nil
	})
	if err != nil || token == nil || !token.Valid {
		return fmt.Errorf("%w, cookie 不是合法 JWT token", err)
	}
	if sc.State != stata {
		return errors.New("state 被篡改了")
	}
	return nil
}

type StateClaims struct {
	State string
	jwt.RegisteredClaims
}
