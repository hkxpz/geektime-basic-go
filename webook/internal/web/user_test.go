package web

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/service"
	"geektime-basic-go/webook/internal/service/mocks"
)

func TestUserHandler_SignUp(t *testing.T) {
	const signupUrl = "/users/signup"
	testCases := []struct {
		name string

		mock       func(ctl *gomock.Controller) (service.UserService, service.CodeService)
		reqBuilder func(t *testing.T) *http.Request

		wantCode int
		wantBody []byte
	}{
		{
			name: "注册成功",
			mock: func(ctl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := mocks.NewMockUserService(ctl)
				userSvc.EXPECT().Signup(gomock.Any(), gomock.Any()).Return(nil)

				return userSvc, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				body := bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123","confirmPassword":"hello@world123"}`))
				req, err := http.NewRequest(http.MethodPost, signupUrl, body)
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			wantCode: http.StatusOK,
			wantBody: []byte(`{"code":0,"msg":"你好，注册成功","data":null}`),
		},
		{
			name: "非 JSON 输入",
			mock: func(ctl *gomock.Controller) (service.UserService, service.CodeService) {
				return nil, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				body := bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123","confirmPassword":"hello@world123",}`))
				req, err := http.NewRequest(http.MethodPost, signupUrl, body)
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			wantCode: http.StatusBadRequest,
			wantBody: []byte(`{"code":5,"msg":"系统错误","data":null}`),
		},
		{
			name: "邮箱格式不正确",
			mock: func(ctl *gomock.Controller) (service.UserService, service.CodeService) {
				return nil, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				body := bytes.NewBuffer([]byte(`{"email":"123@","password":"hello@world123","confirmPassword":"hello@world123"}`))
				req, err := http.NewRequest(http.MethodPost, signupUrl, body)
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			wantCode: http.StatusOK,
			wantBody: []byte(`{"code":4,"msg":"邮箱不正确","data":null}`),
		},
		{
			name: "两次密码输入不同",
			mock: func(ctl *gomock.Controller) (service.UserService, service.CodeService) {
				return nil, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				body := bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123.","confirmPassword":"hello@world123"}`))
				req, err := http.NewRequest(http.MethodPost, signupUrl, body)
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			wantCode: http.StatusOK,
			wantBody: []byte(`{"code":4,"msg":"两次输入的密码不相同","data":null}`),
		},
		{
			name: "密码格式不对",
			mock: func(ctl *gomock.Controller) (service.UserService, service.CodeService) {
				return nil, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				body := bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello","confirmPassword":"hello"}`))
				req, err := http.NewRequest(http.MethodPost, signupUrl, body)
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			wantCode: http.StatusOK,
			wantBody: []byte(`{"code":4,"msg":"密码必须包括数字、字母两种字符，长度在8-15位之间","data":null}`),
		},
		{
			name: "邮箱冲突",
			mock: func(ctl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := mocks.NewMockUserService(ctl)
				userSvc.EXPECT().Signup(gomock.Any(), gomock.Any()).Return(service.ErrUserDuplicate)
				return userSvc, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				body := bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123","confirmPassword":"hello@world123"}`))
				req, err := http.NewRequest(http.MethodPost, signupUrl, body)
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			wantCode: http.StatusOK,
			wantBody: []byte(`{"code":4,"msg":"重复邮箱，请换一个邮箱","data":null}`),
		},
		{
			name: "系统异常",
			mock: func(ctl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := mocks.NewMockUserService(ctl)
				userSvc.EXPECT().Signup(gomock.Any(), gomock.Any()).Return(errors.New("模拟系统异常"))
				return userSvc, nil
			},
			reqBuilder: func(t *testing.T) *http.Request {
				body := bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123","confirmPassword":"hello@world123"}`))
				req, err := http.NewRequest(http.MethodPost, signupUrl, body)
				require.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			wantCode: http.StatusOK,
			wantBody: []byte(`{"code":5,"msg":"服务器异常，注册失败","data":null}`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userSvc, codeSvc := tc.mock(ctrl)
			uh := NewUserHandler(userSvc, codeSvc)

			gin.SetMode(gin.ReleaseMode)
			server := gin.New()
			uh.RegisterRoutes(server)
			req := tc.reqBuilder(t)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.Bytes())
		})
	}
}

func TestEmailPattern(t *testing.T) {
	testCases := []struct {
		name  string
		email string
		match bool
	}{
		{
			name:  "不带@",
			email: "123456",
			match: false,
		},
		{
			name:  "带@ 但是没后缀",
			email: "123456@",
			match: false,
		},
		{
			name:  "合法邮箱",
			email: "123456@qq.com",
			match: true,
		},
	}

	uh := NewUserHandler(nil, nil)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, err := uh.emailRegexExp.MatchString(tc.email)
			assert.NoError(t, err)
			assert.Equal(t, tc.match, match)
		})
	}
}

func TestPasswordPattern(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		match    bool
	}{
		{
			name:     "合法密码",
			password: "Hello#world123",
			match:    true,
		},
		{
			name:     "没有数字",
			password: "Hello#world",
			match:    false,
		},
		{
			name:     "没有字母",
			password: "123123123",
			match:    false,
		},
		{
			name:     "长度不足",
			password: "he!123",
			match:    false,
		},
	}

	uh := NewUserHandler(nil, nil)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, err := uh.passwordRegexExp.MatchString(tc.password)
			require.NoError(t, err)
			assert.Equal(t, tc.match, match)
		})
	}
}
