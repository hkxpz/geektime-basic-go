package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"

	"geektime-basic-go/webook/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/lithammer/shortuuid/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/service"
	"geektime-basic-go/webook/internal/service/mocks"
	myjwt "geektime-basic-go/webook/internal/web/jwt"
	jwtmocks "geektime-basic-go/webook/internal/web/jwt/mocks"
	"geektime-basic-go/webook/internal/web/middleware/handlefunc"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func reqBuilder(t *testing.T, method, url string, body io.Reader, headers ...[]string) *http.Request {
	req, err := http.NewRequest(method, url, body)
	require.NoError(t, err)

	for _, header := range headers {
		req.Header.Set(header[0], header[1])
	}

	if len(headers) < 1 {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func TestUserHandler_SignUp(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctl *gomock.Controller) service.UserService
		body io.Reader

		wantCode int
		wantRes  handlefunc.Response
	}{
		{
			name: "注册成功",
			mock: func(ctl *gomock.Controller) service.UserService {
				userSvc := mocks.NewMockUserService(ctl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{Email: "123@qq.com", Password: "hello@world123"}).Return(nil)
				return userSvc
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123","confirmPassword":"hello@world123"}`)),
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Code: 0, Msg: "你好，注册成功"},
		},
		{
			name: "解析输入失败",
			mock: func(ctl *gomock.Controller) service.UserService {
				return nil
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123","confirmPassword":"hello@world123",}`)),
			wantCode: http.StatusBadRequest,
			wantRes:  InternalServerError(),
		},
		{
			name: "邮箱格式不正确",
			mock: func(ctl *gomock.Controller) service.UserService {
				return nil
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@","password":"hello@world123","confirmPassword":"hello@world123"}`)),
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Code: 4, Msg: "邮箱不正确"},
		},
		{
			name: "两次密码输入不同",
			mock: func(ctl *gomock.Controller) service.UserService {
				return nil
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123.","confirmPassword":"hello@world123"}`)),
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Code: 4, Msg: "两次输入的密码不相同"},
		},
		{
			name: "密码格式不对",
			mock: func(ctl *gomock.Controller) service.UserService {
				return nil
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello","confirmPassword":"hello"}`)),
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Code: 4, Msg: "密码必须包括数字、字母两种字符，长度在8-15位之间"},
		},
		{
			name: "邮箱冲突",
			mock: func(ctl *gomock.Controller) service.UserService {
				userSvc := mocks.NewMockUserService(ctl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{Email: "123@qq.com", Password: "hello@world123"}).Return(service.ErrUserDuplicate)
				return userSvc
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123","confirmPassword":"hello@world123"}`)),
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Code: 4, Msg: "重复邮箱，请换一个邮箱"},
		},
		{
			name: "系统异常",
			mock: func(ctl *gomock.Controller) service.UserService {
				userSvc := mocks.NewMockUserService(ctl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{Email: "123@qq.com", Password: "hello@world123"}).Return(errors.New("模拟系统异常"))
				return userSvc
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123","confirmPassword":"hello@world123"}`)),
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Code: 5, Msg: "服务器异常，注册失败"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uh := NewUserHandler(tc.mock(ctrl), nil, nil, nil)
			server := gin.New()
			uh.RegisterRoutes(server)
			req := reqBuilder(t, http.MethodPost, "/users/signup", tc.body)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)

			var webRes handlefunc.Response
			err := json.NewDecoder(recorder.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantRes, webRes)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	now := time.Now()
	userDomain := domain.User{
		ID:       1,
		Email:    "123@qq.com",
		Nickname: "泰裤辣",
		Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
		Phone:    "13888888888",
		AboutMe:  "泰裤辣",
		Birthday: now,
		CreateAt: now,
	}
	testCases := []struct {
		name string

		mock     func(ctrl *gomock.Controller) (service.UserService, myjwt.Handler)
		body     io.Reader
		ID       int64
		useToken bool

		wantCode int
		wantRes  handlefunc.Response
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, myjwt.Handler) {
				us := mocks.NewMockUserService(ctrl)
				jh := jwtmocks.NewMockHandler(ctrl)
				us.EXPECT().Login(gomock.Any(), "123@qq.com", "hello@world123").Return(userDomain, nil)
				jh.EXPECT().SetLoginToken(gomock.Any(), int64(1))
				return us, jh
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123"}`)),
			ID:       1,
			useToken: true,
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Code: 0, Msg: "登录成功"},
		},
		{
			name: "登录成功,设置token失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, myjwt.Handler) {
				us := mocks.NewMockUserService(ctrl)
				jh := jwtmocks.NewMockHandler(ctrl)
				us.EXPECT().Login(gomock.Any(), "123@qq.com", "hello@world123").Return(userDomain, nil)
				jh.EXPECT().SetLoginToken(gomock.Any(), int64(1)).Return(errors.New("设置token失败"))
				return us, jh
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123"}`)),
			ID:       1,
			useToken: true,
			wantCode: http.StatusOK,
			wantRes:  InternalServerError(),
		},
		{
			name: "解析输入失败",
			mock: func(ctl *gomock.Controller) (service.UserService, myjwt.Handler) {
				return nil, nil
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123",}`)),
			wantCode: http.StatusBadRequest,
			wantRes:  InternalServerError(),
		},
		{
			name: "用户名或密码不正确",
			mock: func(ctrl *gomock.Controller) (service.UserService, myjwt.Handler) {
				us := mocks.NewMockUserService(ctrl)
				us.EXPECT().Login(gomock.Any(), "123@qq.com", "hello@world123").Return(domain.User{}, service.ErrInvalidUserOrPassword)
				return us, nil
			},
			body: bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123"}`)),

			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Code: 4, Msg: "用户名或密码不正确，请重试"},
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, myjwt.Handler) {
				us := mocks.NewMockUserService(ctrl)
				us.EXPECT().Login(gomock.Any(), "123@qq.com", "hello@world123").Return(domain.User{}, errors.New("模拟系统错误"))
				return us, nil
			},
			body: bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123"}`)),

			wantCode: http.StatusOK,
			wantRes:  InternalServerError(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			us, jh := tc.mock(ctrl)
			uh := NewUserHandler(us, nil, jh, nil)
			req := reqBuilder(t, http.MethodPost, "/users/login", tc.body)
			recorder := httptest.NewRecorder()

			server := gin.New()
			uh.RegisterRoutes(server)
			server.ServeHTTP(recorder, req)

			require.Equal(t, tc.wantCode, recorder.Code)
			var webRes handlefunc.Response
			err := json.NewDecoder(recorder.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, webRes)

			if !tc.useToken {
				assert.Empty(t, recorder.Header().Get("x-jwt-token"))
				return
			}

			ctx := gin.CreateTestContextOnly(httptest.NewRecorder(), server)
			ctx.Request = req
			err = myjwt.NewRedisHandler(nil).SetLoginToken(ctx, 1)
			require.NoError(t, err)
			assert.Equal(t, ctx.GetHeader("x-jwt-token"), recorder.Header().Get("x-jwt-token"))
		})
	}
}

func TestUserHandler_Edit(t *testing.T) {
	birthday, err := time.Parse(time.DateOnly, "2000-01-01")
	require.NoError(t, err)
	userDomain := domain.User{
		ID:       1,
		Nickname: "泰裤辣",
		Birthday: birthday,
		AboutMe:  "泰裤辣",
	}
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) service.UserService
		body io.Reader
		ID   int64

		wantCode int
		wantRes  handlefunc.Response
	}{
		{
			name: "修改成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				us := mocks.NewMockUserService(ctrl)
				us.EXPECT().Edit(gomock.Any(), userDomain).Return(nil)
				return us
			},
			ID:       1,
			body:     bytes.NewBuffer([]byte(`{"nickname":"泰裤辣","birthday":"2000-01-01","aboutMe":"泰裤辣"}`)),
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Code: 0, Msg: "OK"},
		},
		{
			name: "空昵称",
			mock: func(ctrl *gomock.Controller) service.UserService {
				return nil
			},
			ID:       1,
			body:     bytes.NewBuffer([]byte(`{"birthday":"2000-01-01","aboutMe":"泰裤辣"}`)),
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Code: 4, Msg: "昵称不能为空"},
		},
		{
			name: "昵称过长",
			mock: func(ctrl *gomock.Controller) service.UserService {
				return nil
			},
			ID:       1,
			body:     bytes.NewBuffer([]byte(`{"nickname":"泰裤辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣","birthday":"2000-01-01","aboutMe":"泰裤辣"}`)),
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Code: 4, Msg: "昵称过长"},
		},
		{
			name: "关于我过长",
			mock: func(ctrl *gomock.Controller) service.UserService {
				return nil
			},
			ID:       1,
			body:     bytes.NewBuffer([]byte(`{"nickname":"泰裤辣","birthday":"2000-01-01","aboutMe":"泰裤辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣"}`)),
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Code: 4, Msg: "关于我过长"},
		},
		{
			name: "日期格式不对",
			mock: func(ctrl *gomock.Controller) service.UserService {
				return nil
			},
			ID:       1,
			body:     bytes.NewBuffer([]byte(`{"nickname":"泰裤辣","birthday":"2000-001-01","aboutMe":"泰裤辣"}`)),
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Code: 4, Msg: "日期格式不对"},
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				us := mocks.NewMockUserService(ctrl)
				us.EXPECT().Edit(gomock.Any(), userDomain).Return(errors.New("模拟系统错误"))
				return us
			},
			ID:       1,
			body:     bytes.NewBuffer([]byte(`{"nickname":"泰裤辣","birthday":"2000-01-01","aboutMe":"泰裤辣"}`)),
			wantCode: http.StatusOK,
			wantRes:  InternalServerError(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			us := tc.mock(ctrl)
			uh := NewUserHandler(us, nil, nil, logger.NewZapLogger(zap.NewExample()))
			req := reqBuilder(t, http.MethodPost, "/users/edit", tc.body)
			recorder := httptest.NewRecorder()

			server := gin.New()
			server.Use(func(ctx *gin.Context) {
				ctx.Set("user", myjwt.UserClaims{ID: tc.ID})
			})
			uh.RegisterRoutes(server)
			server.ServeHTTP(recorder, req)

			var webRes handlefunc.Response
			err = json.NewDecoder(recorder.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantRes, webRes)
		})
	}
}

func TestUserHandler_Profile(t *testing.T) {
	now, err := time.Parse(time.DateOnly, "2023-09-11")
	require.NoError(t, err)
	userDomain := domain.User{
		ID:       1,
		Email:    "123@qq.com",
		Nickname: "泰裤辣",
		Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
		Phone:    "13888888888",
		AboutMe:  "泰裤辣",
		Birthday: now,
		CreateAt: now,
	}
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) service.UserService
		body io.Reader
		ID   int64

		wantCode int
		wantRes  handlefunc.Response
	}{
		{
			name: "成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				us := mocks.NewMockUserService(ctrl)
				us.EXPECT().Profile(gomock.Any(), int64(1)).Return(userDomain, nil)
				return us
			},
			ID:       1,
			wantCode: http.StatusOK,
			wantRes: handlefunc.Response{Code: 0, Msg: "OK", Data: map[string]interface{}{
				"aboutMe":  "泰裤辣",
				"birthday": "2023-09-11",
				"email":    "123@qq.com",
				"nickname": "泰裤辣",
				"phone":    "13888888888",
			}},
		},
		{
			name: "失败",
			mock: func(ctrl *gomock.Controller) service.UserService {
				us := mocks.NewMockUserService(ctrl)
				us.EXPECT().Profile(gomock.Any(), int64(1)).Return(domain.User{}, errors.New("模拟系统错误"))
				return us
			},
			ID:       1,
			wantCode: http.StatusOK,
			wantRes:  InternalServerError(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			us := tc.mock(ctrl)
			uh := NewUserHandler(us, nil, nil, nil)
			req := reqBuilder(t, http.MethodGet, "/users/profile", tc.body)
			recorder := httptest.NewRecorder()

			server := gin.New()
			server.Use(func(ctx *gin.Context) {
				ctx.Set("user", myjwt.UserClaims{ID: tc.ID})
			})
			uh.RegisterRoutes(server)
			server.ServeHTTP(recorder, req)

			var webRes handlefunc.Response
			err := json.NewDecoder(recorder.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantRes, webRes)
		})
	}
}

func TestUserHandler_SendSMSLoginCode(t *testing.T) {
	const biz = "login"
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) service.CodeService
		body io.Reader

		wantCode int
		wantRes  handlefunc.Response
	}{
		{
			name: "发送成功",
			mock: func(ctrl *gomock.Controller) service.CodeService {
				cs := mocks.NewMockCodeService(ctrl)
				cs.EXPECT().Send(gomock.Any(), biz, "13888888888").Return(nil)
				return cs
			},
			body:     bytes.NewBuffer([]byte(`{"phone":"13888888888"}`)),
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Code: 0, Msg: "发送成功"},
		},
		{
			name: "解析输入失败",
			mock: func(ctrl *gomock.Controller) service.CodeService {
				return nil
			},
			body:     bytes.NewBuffer([]byte(`{"phone""13888888888"}`)),
			wantCode: http.StatusBadRequest,
			wantRes:  InternalServerError(),
		},
		{
			name: "手机号码错误",
			mock: func(ctrl *gomock.Controller) service.CodeService {
				return nil
			},
			body:     bytes.NewBuffer([]byte(`{"phone":"1388888888"}`)),
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Code: 4, Msg: "手机号码错误"},
		},
		{
			name: "短信发送太频繁",
			mock: func(ctrl *gomock.Controller) service.CodeService {
				cs := mocks.NewMockCodeService(ctrl)
				cs.EXPECT().Send(gomock.Any(), biz, "13888888888").Return(service.ErrCodeSendTooMany)
				return cs
			},
			body:     bytes.NewBuffer([]byte(`{"phone":"13888888888"}`)),
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Code: 4, Msg: "短信发送太频繁，请稍后再试"},
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) service.CodeService {
				cs := mocks.NewMockCodeService(ctrl)
				cs.EXPECT().Send(gomock.Any(), biz, "13888888888").Return(errors.New("模拟系统错误"))
				return cs
			},
			body:     bytes.NewBuffer([]byte(`{"phone":"13888888888"}`)),
			wantCode: http.StatusOK,
			wantRes:  InternalServerError(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cs := tc.mock(ctrl)
			uh := NewUserHandler(nil, cs, nil, nil)
			req := reqBuilder(t, http.MethodPost, "/users/login_sms/code/send", tc.body)
			recorder := httptest.NewRecorder()

			server := gin.New()
			uh.RegisterRoutes(server)
			server.ServeHTTP(recorder, req)

			var webRes handlefunc.Response
			err := json.NewDecoder(recorder.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantRes, webRes)
		})
	}
}

func TestUserHandler_LoginSMS(t *testing.T) {
	const biz = "login"
	now := time.Now()
	userDomain := domain.User{
		ID:       1,
		Email:    "123@qq.com",
		Nickname: "泰裤辣",
		Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
		Phone:    "13888888888",
		AboutMe:  "泰裤辣",
		Birthday: now,
		CreateAt: now,
	}
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (service.UserService, service.CodeService, myjwt.Handler)
		body io.Reader

		wantCode int
		wantRes  handlefunc.Response
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, myjwt.Handler) {
				us := mocks.NewMockUserService(ctrl)
				cs := mocks.NewMockCodeService(ctrl)
				jh := jwtmocks.NewMockHandler(ctrl)
				cs.EXPECT().Verify(gomock.Any(), biz, "13888888888", "123456").Return(true, nil)
				us.EXPECT().FindOrCreate(gomock.Any(), "13888888888").Return(userDomain, nil)
				jh.EXPECT().SetLoginToken(gomock.Any(), int64(1))
				return us, cs, jh
			},
			body:     bytes.NewBuffer([]byte(`{"phone":"13888888888","code":"123456"}`)),
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Code: 0, Msg: "登录成功"},
		},
		{
			name: "登录成功,设置token失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, myjwt.Handler) {
				us := mocks.NewMockUserService(ctrl)
				cs := mocks.NewMockCodeService(ctrl)
				jh := jwtmocks.NewMockHandler(ctrl)
				cs.EXPECT().Verify(gomock.Any(), biz, "13888888888", "123456").Return(true, nil)
				us.EXPECT().FindOrCreate(gomock.Any(), "13888888888").Return(userDomain, nil)
				jh.EXPECT().SetLoginToken(gomock.Any(), int64(1)).Return(errors.New("模拟设置token失败"))
				return us, cs, jh
			},
			body:     bytes.NewBuffer([]byte(`{"phone":"13888888888","code":"123456"}`)),
			wantCode: http.StatusOK,
			wantRes:  InternalServerError(),
		},
		{
			name: "验证码校验失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, myjwt.Handler) {
				cs := mocks.NewMockCodeService(ctrl)
				cs.EXPECT().Verify(gomock.Any(), biz, "13888888888", "123456").Return(false, errors.New("模拟校验失败"))
				return nil, cs, nil
			},
			body:     bytes.NewBuffer([]byte(`{"phone":"13888888888","code":"123456"}`)),
			wantCode: http.StatusOK,
			wantRes:  InternalServerError(),
		},
		{
			name: "验证码错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, myjwt.Handler) {
				cs := mocks.NewMockCodeService(ctrl)
				cs.EXPECT().Verify(gomock.Any(), biz, "13888888888", "123456").Return(false, nil, nil)
				return nil, cs, nil
			},
			body:     bytes.NewBuffer([]byte(`{"phone":"13888888888","code":"123456"}`)),
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Code: 4, Msg: "验证码错误"},
		},
		{
			name: "查找用户失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService, myjwt.Handler) {
				us := mocks.NewMockUserService(ctrl)
				cs := mocks.NewMockCodeService(ctrl)
				cs.EXPECT().Verify(gomock.Any(), biz, "13888888888", "123456").Return(true, nil, nil)
				us.EXPECT().FindOrCreate(gomock.Any(), "13888888888").Return(domain.User{}, errors.New("模拟查找用户失败"))
				return us, cs, nil
			},
			body:     bytes.NewBuffer([]byte(`{"phone":"13888888888","code":"123456"}`)),
			wantCode: http.StatusOK,
			wantRes:  InternalServerError(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			us, cs, jh := tc.mock(ctrl)
			uh := NewUserHandler(us, cs, jh, logger.NewZapLogger(zap.NewExample()))
			req := reqBuilder(t, http.MethodPost, "/users/login_sms", tc.body)
			recorder := httptest.NewRecorder()

			server := gin.New()
			uh.RegisterRoutes(server)
			server.ServeHTTP(recorder, req)

			var webRes handlefunc.Response
			err := json.NewDecoder(recorder.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantRes, webRes)
		})
	}
}

func TestUserHandler_RefreshToken(t *testing.T) {
	newRefreshToken := func(t *testing.T, expiration time.Duration, tokenKey []byte) (ssid, token string) {
		ssid = uuid.New()
		uc := myjwt.UserClaims{
			ID:   1,
			SSID: ssid,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			},
		}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, uc).SignedString(tokenKey)
		require.NoError(t, err)
		return ssid, token
	}

	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) myjwt.Handler
		wantCode int
		wantRes  handlefunc.Response
	}{
		{
			name: "刷新成功",
			mock: func(ctrl *gomock.Controller) myjwt.Handler {
				ssid, token := newRefreshToken(t, 30*time.Minute, myjwt.RefreshTokenKey)
				hdl := jwtmocks.NewMockHandler(ctrl)
				hdl.EXPECT().ExtractTokenString(gomock.Any()).Return(token)
				hdl.EXPECT().CheckSession(gomock.Any(), ssid).Return(nil, nil)
				hdl.EXPECT().SetJWTToken(gomock.Any(), ssid, int64(1))
				return hdl
			},
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Msg: "刷新成功"},
		},
		{
			name: "解析token失败",
			mock: func(ctrl *gomock.Controller) myjwt.Handler {
				_, token := newRefreshToken(t, 30*time.Minute, nil)
				hdl := jwtmocks.NewMockHandler(ctrl)
				hdl.EXPECT().ExtractTokenString(gomock.Any()).Return(token)
				return hdl
			},
			wantCode: http.StatusUnauthorized,
			wantRes:  handlefunc.Response{Code: 4, Msg: "请登录"},
		},
		{
			name: "非法token",
			mock: func(ctrl *gomock.Controller) myjwt.Handler {
				_, token := newRefreshToken(t, 30*time.Minute, myjwt.AccessTokenKey)
				hdl := jwtmocks.NewMockHandler(ctrl)
				hdl.EXPECT().ExtractTokenString(gomock.Any()).Return(token)
				return hdl
			},
			wantCode: http.StatusUnauthorized,
			wantRes:  handlefunc.Response{Code: 4, Msg: "请登录"},
		},
		{
			name: "token过期",
			mock: func(ctrl *gomock.Controller) myjwt.Handler {
				_, token := newRefreshToken(t, 1, myjwt.RefreshTokenKey)
				hdl := jwtmocks.NewMockHandler(ctrl)
				hdl.EXPECT().ExtractTokenString(gomock.Any()).Return(token)
				return hdl
			},
			wantCode: http.StatusUnauthorized,
			wantRes:  handlefunc.Response{Code: 4, Msg: "请登录"},
		},
		{
			name: "token不存在过期时间",
			mock: func(ctrl *gomock.Controller) myjwt.Handler {
				ssid := uuid.New()
				uc := myjwt.UserClaims{
					ID:   1,
					SSID: ssid,
				}
				token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, uc).SignedString(myjwt.RefreshTokenKey)
				require.NoError(t, err)
				hdl := jwtmocks.NewMockHandler(ctrl)
				hdl.EXPECT().ExtractTokenString(gomock.Any()).Return(token)
				return hdl
			},
			wantCode: http.StatusUnauthorized,
			wantRes:  handlefunc.Response{Code: 4, Msg: "请登录"},
		},
		{
			name: "用户主动退出",
			mock: func(ctrl *gomock.Controller) myjwt.Handler {
				ssid, token := newRefreshToken(t, 30*time.Minute, myjwt.RefreshTokenKey)
				hdl := jwtmocks.NewMockHandler(ctrl)
				hdl.EXPECT().ExtractTokenString(gomock.Any()).Return(token)
				hdl.EXPECT().CheckSession(gomock.Any(), ssid).Return(errors.New("用户已经退出登录"))
				return hdl
			},
			wantCode: http.StatusUnauthorized,
			wantRes:  handlefunc.Response{Code: 4, Msg: "请登录"},
		},
		{
			name: "设置token失败",
			mock: func(ctrl *gomock.Controller) myjwt.Handler {
				ssid, token := newRefreshToken(t, 30*time.Minute, myjwt.RefreshTokenKey)
				hdl := jwtmocks.NewMockHandler(ctrl)
				hdl.EXPECT().ExtractTokenString(gomock.Any()).Return(token)
				hdl.EXPECT().CheckSession(gomock.Any(), ssid).Return(nil, nil)
				hdl.EXPECT().SetJWTToken(gomock.Any(), ssid, int64(1)).Return(errors.New("模拟设置token失败"))
				return hdl
			},
			wantCode: http.StatusOK,
			wantRes:  InternalServerError(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uh := NewUserHandler(nil, nil, tc.mock(ctrl), nil)

			req := reqBuilder(t, http.MethodPost, "/users/refresh_token", nil, nil)
			recorder := httptest.NewRecorder()

			server := gin.New()
			uh.RegisterRoutes(server)
			server.ServeHTTP(recorder, req)

			var webRes handlefunc.Response
			err := json.NewDecoder(recorder.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantRes, webRes)
		})
	}
}

func TestUserHandler_Logout(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) myjwt.Handler
		wantCode int
		wantRes  handlefunc.Response
	}{
		{
			name: "登出成功",
			mock: func(ctrl *gomock.Controller) myjwt.Handler {
				hdl := jwtmocks.NewMockHandler(ctrl)
				hdl.EXPECT().ClearToken(gomock.Any()).Return(nil, nil)
				return hdl
			},
			wantCode: http.StatusOK,
			wantRes:  handlefunc.Response{Msg: "OK"},
		},
		{
			name: "登出失败",
			mock: func(ctrl *gomock.Controller) myjwt.Handler {
				hdl := jwtmocks.NewMockHandler(ctrl)
				hdl.EXPECT().ClearToken(gomock.Any()).Return(errors.New("模拟失败"))
				return hdl
			},
			wantCode: http.StatusOK,
			wantRes:  InternalServerError(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uh := NewUserHandler(nil, nil, tc.mock(ctrl), nil)

			req := reqBuilder(t, http.MethodPost, "/users/logout", nil, nil)
			recorder := httptest.NewRecorder()

			server := gin.New()
			uh.RegisterRoutes(server)
			server.ServeHTTP(recorder, req)

			var webRes handlefunc.Response
			err := json.NewDecoder(recorder.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantRes, webRes)
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

	uh := NewUserHandler(nil, nil, nil, nil)
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
		name  string
		phone string
		match bool
	}{
		{
			name:  "合法手机号",
			phone: "13888888888",
			match: true,
		},
		{
			name:  "手机号过短",
			phone: "1238",
			match: false,
		},
		{
			name:  "手机号过长",
			phone: "13888888888888888888888",
			match: false,
		},
		{
			name:  "非法手机号",
			phone: "12388888888",
			match: false,
		},
	}

	uh := NewUserHandler(nil, nil, nil, nil)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, err := uh.phoneRegexExp.MatchString(tc.phone)
			require.NoError(t, err)
			assert.Equal(t, tc.match, match)
		})
	}
}

func TestPhonePattern(t *testing.T) {
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

	uh := NewUserHandler(nil, nil, nil, nil)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, err := uh.passwordRegexExp.MatchString(tc.password)
			require.NoError(t, err)
			assert.Equal(t, tc.match, match)
		})
	}
}
