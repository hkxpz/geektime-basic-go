package redis

import (
	"context"
	"errors"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/repository/cache"
	svcmocks "geektime-basic-go/webook/internal/repository/cache/redis/mocks"
)

func TestCodeCache_Set(t *testing.T) {
	const biz = "login"
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) redis.Cmdable

		ctx   context.Context
		phone string
		code  string

		wantErr error
	}{
		{
			name: "设置成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := svcmocks.NewMockCmdable(ctrl)
				mockRedis := redis.NewCmdResult(int64(0), nil)
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:13888888888"}, "123456").Return(mockRedis)
				return cmd
			},
			ctx:   context.Background(),
			phone: "13888888888",
			code:  "123456",
		},
		{
			name: "发送太频繁",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := svcmocks.NewMockCmdable(ctrl)
				mockRedis := redis.NewCmdResult(int64(-1), nil)
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:13888888888"}, "123456").Return(mockRedis)
				return cmd
			},
			ctx:     context.Background(),
			phone:   "13888888888",
			code:    "123456",
			wantErr: cache.ErrCodeSendTooMany,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := svcmocks.NewMockCmdable(ctrl)
				mockRedis := redis.NewCmdResult(int64(2), nil)
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:13888888888"}, "123456").Return(mockRedis)
				return cmd
			},
			ctx:     context.Background(),
			phone:   "13888888888",
			code:    "123456",
			wantErr: cache.ErrUnknownForCode,
		},
		{
			name: "设置失败",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := svcmocks.NewMockCmdable(ctrl)
				mockRedis := redis.NewCmdResult(int64(0), errors.New("模拟设置失败"))
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, []string{"phone_code:login:13888888888"}, "123456").Return(mockRedis)
				return cmd
			},
			ctx:     context.Background(),
			phone:   "13888888888",
			code:    "123456",
			wantErr: errors.New("模拟设置失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			c := NewCodeCache(tc.mock(ctrl))
			err := c.Set(tc.ctx, biz, tc.phone, tc.code)
			require.Equal(t, tc.wantErr, err)
		})
	}
}

func TestCodeCache_Verify(t *testing.T) {
	const biz = "login"
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) redis.Cmdable

		ctx   context.Context
		phone string
		code  string

		ok      bool
		wantErr error
	}{
		{
			name: "验证通过",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := svcmocks.NewMockCmdable(ctrl)
				mockRedis := redis.NewCmdResult(int64(0), nil)
				cmd.EXPECT().Eval(gomock.Any(), luaVerifyCode, []string{"phone_code:login:13888888888"}, "123456").Return(mockRedis)
				return cmd
			},
			ctx:   context.Background(),
			phone: "13888888888",
			code:  "123456",
			ok:    true,
		},
		{
			name: "验证次数太多",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := svcmocks.NewMockCmdable(ctrl)
				mockRedis := redis.NewCmdResult(int64(-1), nil)
				cmd.EXPECT().Eval(gomock.Any(), luaVerifyCode, []string{"phone_code:login:13888888888"}, "123456").Return(mockRedis)
				return cmd
			},
			ctx:     context.Background(),
			phone:   "13888888888",
			code:    "123456",
			wantErr: cache.ErrCodeVerifyTooManyTimes,
		},
		{
			name: "系统错误",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := svcmocks.NewMockCmdable(ctrl)
				mockRedis := redis.NewCmdResult(int64(2), nil)
				cmd.EXPECT().Eval(gomock.Any(), luaVerifyCode, []string{"phone_code:login:13888888888"}, "123456").Return(mockRedis)
				return cmd
			},
			ctx:   context.Background(),
			phone: "13888888888",
			code:  "123456",
		},
		{
			name: "校验失败",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := svcmocks.NewMockCmdable(ctrl)
				mockRedis := redis.NewCmdResult(int64(0), errors.New("模拟校验失败"))
				cmd.EXPECT().Eval(gomock.Any(), luaVerifyCode, []string{"phone_code:login:13888888888"}, "123456").Return(mockRedis)
				return cmd
			},
			ctx:     context.Background(),
			phone:   "13888888888",
			code:    "123456",
			wantErr: errors.New("模拟校验失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			c := NewCodeCache(tc.mock(ctrl))
			ok, err := c.Verify(tc.ctx, biz, tc.phone, tc.code)
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.ok, ok)
		})
	}
}

func TestCodeCache_key(t *testing.T) {
	testCases := []struct {
		name    string
		phone   string
		biz     string
		wantKey string
	}{
		{
			name:    "login_biz",
			biz:     "login",
			phone:   "13888888888",
			wantKey: "phone_code:login:13888888888",
		},
		{
			name:    "signup_biz",
			biz:     "signup",
			phone:   "13888888886",
			wantKey: "phone_code:signup:13888888886",
		},
	}

	cc := &codeCache{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := cc.key(tc.biz, tc.phone)
			assert.Equal(t, tc.wantKey, key)
		})
	}
}
