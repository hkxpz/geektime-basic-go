package web

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/service"
	"geektime-basic-go/webook/internal/service/mocks"
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

		mock func(ctl *gomock.Controller) (service.UserService, service.CodeService)
		body io.Reader

		wantCode int
		wantBody string
	}{
		{
			name: "注册成功",
			mock: func(ctl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := mocks.NewMockUserService(ctl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{Email: "123@qq.com", Password: "hello@world123"}).Return(nil)
				return userSvc, nil
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123","confirmPassword":"hello@world123"}`)),
			wantCode: http.StatusOK,
			wantBody: `{"code":0,"msg":"你好，注册成功","data":null}`,
		},
		{
			name: "解析输入失败",
			mock: func(ctl *gomock.Controller) (service.UserService, service.CodeService) {
				return nil, nil
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123","confirmPassword":"hello@world123",}`)),
			wantCode: http.StatusBadRequest,
			wantBody: `{"code":5,"msg":"系统错误","data":null}`,
		},
		{
			name: "邮箱格式不正确",
			mock: func(ctl *gomock.Controller) (service.UserService, service.CodeService) {
				return nil, nil
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@","password":"hello@world123","confirmPassword":"hello@world123"}`)),
			wantCode: http.StatusOK,
			wantBody: `{"code":4,"msg":"邮箱不正确","data":null}`,
		},
		{
			name: "两次密码输入不同",
			mock: func(ctl *gomock.Controller) (service.UserService, service.CodeService) {
				return nil, nil
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123.","confirmPassword":"hello@world123"}`)),
			wantCode: http.StatusOK,
			wantBody: `{"code":4,"msg":"两次输入的密码不相同","data":null}`,
		},
		{
			name: "密码格式不对",
			mock: func(ctl *gomock.Controller) (service.UserService, service.CodeService) {
				return nil, nil
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello","confirmPassword":"hello"}`)),
			wantCode: http.StatusOK,
			wantBody: `{"code":4,"msg":"密码必须包括数字、字母两种字符，长度在8-15位之间","data":null}`,
		},
		{
			name: "邮箱冲突",
			mock: func(ctl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := mocks.NewMockUserService(ctl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{Email: "123@qq.com", Password: "hello@world123"}).Return(service.ErrUserDuplicate)
				return userSvc, nil
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123","confirmPassword":"hello@world123"}`)),
			wantCode: http.StatusOK,
			wantBody: `{"code":4,"msg":"重复邮箱，请换一个邮箱","data":null}`,
		},
		{
			name: "系统异常",
			mock: func(ctl *gomock.Controller) (service.UserService, service.CodeService) {
				userSvc := mocks.NewMockUserService(ctl)
				userSvc.EXPECT().Signup(gomock.Any(), domain.User{Email: "123@qq.com", Password: "hello@world123"}).Return(errors.New("模拟系统异常"))
				return userSvc, nil
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123","confirmPassword":"hello@world123"}`)),
			wantCode: http.StatusOK,
			wantBody: `{"code":5,"msg":"服务器异常，注册失败","data":null}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userSvc, codeSvc := tc.mock(ctrl)
			uh := NewUserHandler(userSvc, codeSvc)

			server := gin.New()
			uh.RegisterRoutes(server)
			req := reqBuilder(t, http.MethodPost, "/users/signup", tc.body)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.String())
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	now := time.Now()
	userDomain := domain.User{
		Id:       1,
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

		mock     func(ctrl *gomock.Controller) service.UserService
		body     io.Reader
		Id       int64
		useToken bool

		wantCode int
		wantBody string
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				us := mocks.NewMockUserService(ctrl)
				us.EXPECT().Login(gomock.Any(), "123@qq.com", "hello@world123").Return(userDomain, nil)
				return us
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123"}`)),
			Id:       1,
			useToken: true,
			wantCode: http.StatusOK,
			wantBody: `{"code":0,"msg":"登录成功","data":null}`,
		},
		{
			name: "解析输入失败",
			mock: func(ctl *gomock.Controller) service.UserService {
				return nil
			},
			body:     bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123",}`)),
			wantCode: http.StatusBadRequest,
			wantBody: `{"code":5,"msg":"系统错误","data":null}`,
		},
		{
			name: "用户名或密码不正确",
			mock: func(ctrl *gomock.Controller) service.UserService {
				us := mocks.NewMockUserService(ctrl)
				us.EXPECT().Login(gomock.Any(), "123@qq.com", "hello@world123").Return(domain.User{}, service.ErrInvalidUserOrPassword)
				return us
			},
			body: bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123"}`)),

			wantCode: http.StatusOK,
			wantBody: `{"code":4,"msg":"用户名或密码不正确，请重试","data":null}`,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				us := mocks.NewMockUserService(ctrl)
				us.EXPECT().Login(gomock.Any(), "123@qq.com", "hello@world123").Return(domain.User{}, errors.New("模拟系统错误"))
				return us
			},
			body: bytes.NewBuffer([]byte(`{"email":"123@qq.com","password":"hello@world123"}`)),

			wantCode: http.StatusOK,
			wantBody: `{"code":5,"msg":"系统错误","data":null}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			us := tc.mock(ctrl)
			uh := NewUserHandler(us, nil)
			req := reqBuilder(t, http.MethodPost, "/users/login", tc.body)
			recorder := httptest.NewRecorder()

			server := gin.New()
			uh.RegisterRoutes(server)
			server.ServeHTTP(recorder, req)

			require.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.String())

			if !tc.useToken {
				assert.Empty(t, recorder.Header().Get("x-jwt-token"))
				return
			}

			ctx := gin.CreateTestContextOnly(recorder, server)
			ctx.Request = req
			token, err := newJWTToken(ctx, tc.Id)
			require.NoError(t, err)
			assert.Equal(t, token, recorder.Header().Get("x-jwt-token"))
		})
	}
}

func TestUserHandler_Edit(t *testing.T) {
	birthday, err := time.Parse(time.DateOnly, "2000-01-01")
	require.NoError(t, err)
	userDomain := domain.User{
		Id:       1,
		Nickname: "泰裤辣",
		Birthday: birthday,
		AboutMe:  "泰裤辣",
	}
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) service.UserService
		body io.Reader
		Id   int64

		wantCode int
		wantBody string
	}{
		{
			name: "修改成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				us := mocks.NewMockUserService(ctrl)
				us.EXPECT().Edit(gomock.Any(), userDomain).Return(nil)
				return us
			},
			Id:       1,
			body:     bytes.NewBuffer([]byte(`{"nickname":"泰裤辣","birthday":"2000-01-01","aboutMe":"泰裤辣"}`)),
			wantCode: http.StatusOK,
			wantBody: `{"code":0,"msg":"OK","data":null}`,
		},
		{
			name: "解析输入失败",
			mock: func(ctrl *gomock.Controller) service.UserService {
				return nil
			},
			Id:       1,
			body:     bytes.NewBuffer([]byte(`{,"nickname":"泰裤辣","birthday":"2000-01-01","aboutMe":"泰裤辣"}`)),
			wantCode: http.StatusBadRequest,
			wantBody: `{"code":5,"msg":"系统错误","data":null}`,
		},
		{
			name: "空昵称",
			mock: func(ctrl *gomock.Controller) service.UserService {
				return nil
			},
			Id:       1,
			body:     bytes.NewBuffer([]byte(`{"birthday":"2000-01-01","aboutMe":"泰裤辣"}`)),
			wantCode: http.StatusOK,
			wantBody: `{"code":4,"msg":"昵称不能为空","data":null}`,
		},
		{
			name: "昵称过长",
			mock: func(ctrl *gomock.Controller) service.UserService {
				return nil
			},
			Id:       1,
			body:     bytes.NewBuffer([]byte(`{"nickname":"泰裤辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣","birthday":"2000-01-01","aboutMe":"泰裤辣"}`)),
			wantCode: http.StatusOK,
			wantBody: `{"code":4,"msg":"昵称过长","data":null}`,
		},
		{
			name: "关于我过长",
			mock: func(ctrl *gomock.Controller) service.UserService {
				return nil
			},
			Id:       1,
			body:     bytes.NewBuffer([]byte(`{"nickname":"泰裤辣","birthday":"2000-01-01","aboutMe":"泰裤辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣辣"}`)),
			wantCode: http.StatusOK,
			wantBody: `{"code":4,"msg":"关于我过长","data":null}`,
		},
		{
			name: "日期格式不对",
			mock: func(ctrl *gomock.Controller) service.UserService {
				return nil
			},
			Id:       1,
			body:     bytes.NewBuffer([]byte(`{"nickname":"泰裤辣","birthday":"2000-001-01","aboutMe":"泰裤辣"}`)),
			wantCode: http.StatusOK,
			wantBody: `{"code":4,"msg":"日期格式不对","data":null}`,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) service.UserService {
				us := mocks.NewMockUserService(ctrl)
				us.EXPECT().Edit(gomock.Any(), userDomain).Return(errors.New("模拟系统错误"))
				return us
			},
			Id:       1,
			body:     bytes.NewBuffer([]byte(`{"nickname":"泰裤辣","birthday":"2000-01-01","aboutMe":"泰裤辣"}`)),
			wantCode: http.StatusOK,
			wantBody: `{"code":5,"msg":"系统错误","data":null}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			us := tc.mock(ctrl)
			uh := NewUserHandler(us, nil)
			req := reqBuilder(t, http.MethodPost, "/users/edit", tc.body)
			recorder := httptest.NewRecorder()

			server := gin.New()
			server.Use(func(ctx *gin.Context) {
				ctx.Set("user", UserClaims{Id: tc.Id})
			})
			uh.RegisterRoutes(server)
			server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.String())
		})
	}
}

func TestUserHandler_Profile(t *testing.T) {
	now := time.Now()
	userDomain := domain.User{
		Id:       1,
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
		Id   int64

		wantCode int
		wantBody string
	}{
		{
			name: "成功",
			mock: func(ctrl *gomock.Controller) service.UserService {
				us := mocks.NewMockUserService(ctrl)
				us.EXPECT().Profile(gomock.Any(), int64(1)).Return(userDomain, nil)
				return us
			},
			Id:       1,
			wantCode: http.StatusOK,
			wantBody: fmt.Sprintf(`{"code":0,"msg":"OK","data":{"Id":1,"Email":"123@qq.com","Nickname":"泰裤辣","Password":"$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi","Phone":"13888888888","AboutMe":"泰裤辣","Birthday":"%[1]s","CreateAt":"%[1]s"}}`, now.Format(time.RFC3339Nano)),
		},
		{
			name: "失败",
			mock: func(ctrl *gomock.Controller) service.UserService {
				us := mocks.NewMockUserService(ctrl)
				us.EXPECT().Profile(gomock.Any(), int64(1)).Return(domain.User{}, errors.New("模拟系统错误"))
				return us
			},
			Id:       1,
			wantCode: http.StatusOK,
			wantBody: `{"code":5,"msg":"系统错误","data":null}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			us := tc.mock(ctrl)
			uh := NewUserHandler(us, nil)
			req := reqBuilder(t, http.MethodGet, "/users/profile", tc.body)
			recorder := httptest.NewRecorder()

			server := gin.New()
			server.Use(func(ctx *gin.Context) {
				ctx.Set("user", UserClaims{Id: tc.Id})
			})
			uh.RegisterRoutes(server)
			server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.String())
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
		wantBody string
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
			wantBody: `{"code":0,"msg":"发送成功","data":null}`,
		},
		{
			name: "解析输入失败",
			mock: func(ctrl *gomock.Controller) service.CodeService {
				return nil
			},
			body:     bytes.NewBuffer([]byte(`{"phone""13888888888"}`)),
			wantCode: http.StatusBadRequest,
			wantBody: `{"code":5,"msg":"系统错误","data":null}`,
		},
		{
			name: "手机号码错误",
			mock: func(ctrl *gomock.Controller) service.CodeService {
				return nil
			},
			body:     bytes.NewBuffer([]byte(`{"phone":"1388888888"}`)),
			wantCode: http.StatusOK,
			wantBody: `{"code":4,"msg":"手机号码错误","data":null}`,
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
			wantBody: `{"code":4,"msg":"短信发送太频繁，请稍后再试","data":null}`,
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
			wantBody: `{"code":5,"msg":"系统错误","data":null}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cs := tc.mock(ctrl)
			uh := NewUserHandler(nil, cs)
			req := reqBuilder(t, http.MethodPost, "/users/login_sms/code/send", tc.body)
			recorder := httptest.NewRecorder()

			server := gin.New()
			uh.RegisterRoutes(server)
			server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.String())
		})
	}
}

func TestUserHandler_LoginSMS(t *testing.T) {
	const biz = "login"
	now := time.Now()
	userDomain := domain.User{
		Id:       1,
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

		mock func(ctrl *gomock.Controller) (service.UserService, service.CodeService)
		body io.Reader

		wantCode int
		wantBody string
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				us := mocks.NewMockUserService(ctrl)
				cs := mocks.NewMockCodeService(ctrl)
				cs.EXPECT().Verify(gomock.Any(), biz, "13888888888", "123456").Return(true, nil)
				us.EXPECT().FindOrCreate(gomock.Any(), "13888888888").Return(userDomain, nil)
				return us, cs
			},
			body:     bytes.NewBuffer([]byte(`{"phone":"13888888888","code":"123456"}`)),
			wantCode: http.StatusOK,
			wantBody: `{"code":0,"msg":"登录成功","data":null}`,
		},
		{
			name: "解析输入失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				return nil, nil
			},
			body:     bytes.NewBuffer([]byte(`{"phone":"13888888888""code":"123456"}`)),
			wantCode: http.StatusBadRequest,
			wantBody: `{"code":5,"msg":"系统错误","data":null}`,
		},
		{
			name: "验证码校验失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				cs := mocks.NewMockCodeService(ctrl)
				cs.EXPECT().Verify(gomock.Any(), biz, "13888888888", "123456").Return(false, errors.New("模拟校验失败"))
				return nil, cs
			},
			body:     bytes.NewBuffer([]byte(`{"phone":"13888888888","code":"123456"}`)),
			wantCode: http.StatusOK,
			wantBody: `{"code":5,"msg":"系统错误","data":null}`,
		},
		{
			name: "验证码错误",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				cs := mocks.NewMockCodeService(ctrl)
				cs.EXPECT().Verify(gomock.Any(), biz, "13888888888", "123456").Return(false, nil)
				return nil, cs
			},
			body:     bytes.NewBuffer([]byte(`{"phone":"13888888888","code":"123456"}`)),
			wantCode: http.StatusOK,
			wantBody: `{"code":4,"msg":"验证码错误","data":null}`,
		},
		{
			name: "查找用户失败",
			mock: func(ctrl *gomock.Controller) (service.UserService, service.CodeService) {
				us := mocks.NewMockUserService(ctrl)
				cs := mocks.NewMockCodeService(ctrl)
				cs.EXPECT().Verify(gomock.Any(), biz, "13888888888", "123456").Return(true, nil)
				us.EXPECT().FindOrCreate(gomock.Any(), "13888888888").Return(domain.User{}, errors.New("模拟查找用户失败"))
				return us, cs
			},
			body:     bytes.NewBuffer([]byte(`{"phone":"13888888888","code":"123456"}`)),
			wantCode: http.StatusOK,
			wantBody: `{"code":5,"msg":"系统错误","data":null}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			us, cs := tc.mock(ctrl)
			uh := NewUserHandler(us, cs)
			req := reqBuilder(t, http.MethodPost, "/users/login_sms", tc.body)
			recorder := httptest.NewRecorder()

			server := gin.New()
			uh.RegisterRoutes(server)
			server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantBody, recorder.Body.String())
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

	uh := NewUserHandler(nil, nil)
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

	uh := NewUserHandler(nil, nil)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			match, err := uh.passwordRegexExp.MatchString(tc.password)
			require.NoError(t, err)
			assert.Equal(t, tc.match, match)
		})
	}
}
