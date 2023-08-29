package redis

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/repository/cache"
	"geektime-basic-go/webook/internal/repository/cache/redis/mocks"
)

func TestRedisCodeCache_Set(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) redis.Cmdable

		ctx   context.Context
		biz   string
		phone string
		code  string

		wantErr error
	}{
		{
			name: "设置成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				mockRedis := redis.NewCmdResult(int64(0), nil)

				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, gomock.Any(), gomock.Any()).Return(mockRedis)
				return cmd
			},
			ctx:   context.Background(),
			phone: "13888888888",
			code:  "123456",
		},
		{
			name: "发送太频繁",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := mocks.NewMockCmdable(ctrl)
				mockRedis := redis.NewCmdResult(int64(-1), nil)
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, gomock.Any(), gomock.Any()).Return(mockRedis)
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
				cmd := mocks.NewMockCmdable(ctrl)
				mockRedis := redis.NewCmdResult(int64(2), nil)
				cmd.EXPECT().Eval(gomock.Any(), luaSetCode, gomock.Any(), gomock.Any()).Return(mockRedis)
				return cmd
			},
			ctx:     context.Background(),
			phone:   "13888888888",
			code:    "123456",
			wantErr: cache.ErrUnknownForCode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			c := NewCodeCache(tc.mock(ctrl))
			err := c.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
