package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/service"
	"geektime-basic-go/webook/internal/service/mocks"
	"geektime-basic-go/webook/internal/service/oauth2/wechat"
	wechatmocks "geektime-basic-go/webook/internal/service/oauth2/wechat/mocks"
	myjwt "geektime-basic-go/webook/internal/web/jwt"
	jwtmocks "geektime-basic-go/webook/internal/web/jwt/mocks"
)

var errFailed = errors.New("模拟失败")

func TestOAuth2WechatHandler_OAuth2URL(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) wechat.Service
		wantCode int
	}{
		{
			name: "获取成功",
			mock: func(ctrl *gomock.Controller) wechat.Service {
				svc := wechatmocks.NewMockService(ctrl)
				svc.EXPECT().AuthURL(gomock.Any(), gomock.Any())
				return svc
			},
			wantCode: http.StatusOK,
		},
		{
			name: "获取失败",
			mock: func(ctrl *gomock.Controller) wechat.Service {
				svc := wechatmocks.NewMockService(ctrl)
				svc.EXPECT().AuthURL(gomock.Any(), gomock.Any()).Return("", errFailed)
				return svc
			},
			wantCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			oh := NewOAuth2WechatHandler(tc.mock(ctrl), nil, nil)
			server := gin.New()
			oh.RegisterRoutes(server)
			req := reqBuilder(t, http.MethodGet, "/oauth2/wechat/authurl", nil)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
		})
	}

}

func TestOAuth2WechatHandler_Callback(t *testing.T) {
	ssid := uuid.New()
	stateTokenKey := []byte("moyn8y9abnd7q4zkq2m73yw8tu9j5ixB")
	domainWechatInfo := domain.WechatInfo{UnionID: ssid, OpenID: ssid}
	domainUser := domain.User{ID: 1, WechatInfo: domainWechatInfo}
	newToken := func(ssid string, tokenKey []byte) string {
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, StateClaims{State: ssid}).SignedString(tokenKey)
		require.NoError(t, err)
		return token
	}
	newCookie := func(token string) *http.Cookie {
		return &http.Cookie{
			Name:     "jwt-state",
			Value:    url.QueryEscape(token),
			Path:     "/outh2/wechat/callback",
			MaxAge:   600,
			Secure:   false,
			HttpOnly: true,
		}
	}
	testCases := []struct {
		name      string
		mock      func(ctrl *gomock.Controller) (wechat.Service, service.UserService, myjwt.Handler)
		addCookie func(req *http.Request)
		wantCode  int
		wantRse   Response
	}{
		{
			name: "登陆成功",
			mock: func(ctrl *gomock.Controller) (wechat.Service, service.UserService, myjwt.Handler) {
				svc := wechatmocks.NewMockService(ctrl)
				userSvc := mocks.NewMockUserService(ctrl)
				hdl := jwtmocks.NewMockHandler(ctrl)
				svc.EXPECT().VerifyCode(gomock.Any(), ssid).Return(domainWechatInfo, nil)
				userSvc.EXPECT().FindOrCreateByWechat(gomock.Any(), domainWechatInfo).Return(domainUser, nil)
				hdl.EXPECT().SetLoginToken(gomock.Any(), int64(1))
				return svc, userSvc, hdl
			},
			addCookie: func(req *http.Request) {
				cookie := newCookie(newToken(ssid, stateTokenKey))
				req.AddCookie(cookie)
			},
			wantCode: http.StatusOK,
			wantRse:  Response{Msg: "登陆成功"},
		},
		{
			name: "没有cookie",
			mock: func(ctrl *gomock.Controller) (wechat.Service, service.UserService, myjwt.Handler) {
				return nil, nil, nil
			},
			addCookie: func(req *http.Request) {},
			wantCode:  http.StatusOK,
			wantRse:   InternalServerError(),
		},
		{
			name: "非法cookie token",
			mock: func(ctrl *gomock.Controller) (wechat.Service, service.UserService, myjwt.Handler) {
				return nil, nil, nil
			},
			addCookie: func(req *http.Request) {
				req.AddCookie(newCookie(newToken(ssid, myjwt.AccessTokenKey)))
			},
			wantCode: http.StatusOK,
			wantRse:  InternalServerError(),
		},
		{
			name: "state被篡改",
			mock: func(ctrl *gomock.Controller) (wechat.Service, service.UserService, myjwt.Handler) {
				return nil, nil, nil
			},
			addCookie: func(req *http.Request) {
				req.AddCookie(newCookie(newToken(uuid.New(), stateTokenKey)))
			},
			wantCode: http.StatusOK,
			wantRse:  InternalServerError(),
		},
		{
			name: "授权码验证失败",
			mock: func(ctrl *gomock.Controller) (wechat.Service, service.UserService, myjwt.Handler) {
				svc := wechatmocks.NewMockService(ctrl)
				svc.EXPECT().VerifyCode(gomock.Any(), ssid).Return(domain.WechatInfo{}, errFailed)
				return svc, nil, nil
			},
			addCookie: func(req *http.Request) {
				req.AddCookie(newCookie(newToken(ssid, stateTokenKey)))
			},
			wantCode: http.StatusOK,
			wantRse:  InternalServerError(),
		},
		{
			name: "查找/创建用户失败",
			mock: func(ctrl *gomock.Controller) (wechat.Service, service.UserService, myjwt.Handler) {
				svc := wechatmocks.NewMockService(ctrl)
				userSvc := mocks.NewMockUserService(ctrl)
				svc.EXPECT().VerifyCode(gomock.Any(), ssid).Return(domainWechatInfo, nil)
				userSvc.EXPECT().FindOrCreateByWechat(gomock.Any(), domainWechatInfo).Return(domain.User{}, errFailed)
				return svc, userSvc, nil
			},
			addCookie: func(req *http.Request) {
				req.AddCookie(newCookie(newToken(ssid, stateTokenKey)))
			},
			wantCode: http.StatusOK,
			wantRse:  InternalServerError(),
		},
		{
			name: "设置token失败",
			mock: func(ctrl *gomock.Controller) (wechat.Service, service.UserService, myjwt.Handler) {
				svc := wechatmocks.NewMockService(ctrl)
				userSvc := mocks.NewMockUserService(ctrl)
				hdl := jwtmocks.NewMockHandler(ctrl)
				svc.EXPECT().VerifyCode(gomock.Any(), ssid).Return(domainWechatInfo, nil)
				userSvc.EXPECT().FindOrCreateByWechat(gomock.Any(), domainWechatInfo).Return(domainUser, nil)
				hdl.EXPECT().SetLoginToken(gomock.Any(), int64(1)).Return(errFailed)
				return svc, userSvc, hdl
			},
			addCookie: func(req *http.Request) {
				req.AddCookie(newCookie(newToken(ssid, stateTokenKey)))
			},
			wantCode: http.StatusOK,
			wantRse:  InternalServerError(),
		},
	}

	callbackURL := fmt.Sprintf("/oauth2/wechat/callback?code=%[1]s&state=%[1]s", ssid)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc, userSvc, hdl := tc.mock(ctrl)
			oh := NewOAuth2WechatHandler(svc, userSvc, hdl)
			server := gin.New()
			oh.RegisterRoutes(server)
			req := reqBuilder(t, http.MethodGet, callbackURL, nil)
			tc.addCookie(req)
			recorder := httptest.NewRecorder()

			server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			var res Response
			err := json.NewDecoder(recorder.Body).Decode(&res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRse, res)
		})
	}

}
