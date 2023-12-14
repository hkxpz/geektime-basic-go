package redis

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository/cache"
	svcmocks "geektime-basic-go/webook/internal/repository/cache/redis/mocks"
)

func TestUserCache_Delete(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) redis.Cmdable
		id      int64
		wantErr error
	}{
		{
			name: "删除成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := svcmocks.NewMockCmdable(ctrl)
				res := redis.NewIntResult(1, nil)
				cmd.EXPECT().Del(gomock.Any(), "user:info:1").Return(res)
				return cmd
			},
			id: 1,
		},
		{
			name: "删除失败",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := svcmocks.NewMockCmdable(ctrl)
				res := redis.NewIntResult(0, redis.Nil)
				cmd.EXPECT().Del(gomock.Any(), "user:info:1").Return(res)
				return cmd
			},
			id:      1,
			wantErr: redis.Nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			uc := NewUserCache(tc.mock(ctrl))
			err := uc.Delete(context.Background(), tc.id)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestUserCache_Get(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) redis.Cmdable
		id       int64
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "查到用户",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := svcmocks.NewMockCmdable(ctrl)
				res := redis.NewStringResult(`{"id":1,"email":"123@qq.com","nickname":"泰裤辣","password":"$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi","phone":"13888888888","aboutMe":"泰裤辣"}`, nil)
				cmd.EXPECT().Get(gomock.Any(), "user:info:1").Return(res)
				return cmd
			},
			id: 1,
			wantUser: domain.User{
				ID:       1,
				Email:    "123@qq.com",
				Nickname: "泰裤辣",
				Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
				Phone:    "13888888888",
				AboutMe:  "泰裤辣",
			},
		},
		{
			name: "用户不存在",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := svcmocks.NewMockCmdable(ctrl)
				res := redis.NewStringResult("", redis.Nil)
				cmd.EXPECT().Get(gomock.Any(), "user:info:1").Return(res)
				return cmd
			},
			id:       1,
			wantUser: domain.User{},
			wantErr:  cache.ErrKeyNotExist,
		},
		{
			name: "查找用户失败",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := svcmocks.NewMockCmdable(ctrl)
				res := redis.NewStringResult("", errors.New("模拟查找用户失败"))
				cmd.EXPECT().Get(gomock.Any(), "user:info:1").Return(res)
				return cmd
			},
			id:       1,
			wantUser: domain.User{},
			wantErr:  errors.New("模拟查找用户失败"),
		},
		{
			name: "反序列化失败",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := svcmocks.NewMockCmdable(ctrl)
				res := redis.NewStringResult(`{"id":1,}`, nil)
				cmd.EXPECT().Get(gomock.Any(), "user:info:1").Return(res)
				return cmd
			},
			id:       1,
			wantUser: domain.User{},
			wantErr:  json.Unmarshal([]byte(`{"id":1,}`), &domain.User{}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			uc := NewUserCache(tc.mock(ctrl))
			user, err := uc.Get(context.Background(), tc.id)
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

func TestUserCache_Set(t *testing.T) {
	userDomain := domain.User{
		ID:       1,
		Email:    "123@qq.com",
		Nickname: "泰裤辣",
		Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
		Phone:    "13888888888",
		AboutMe:  "泰裤辣",
	}
	bs, err := json.Marshal(userDomain)
	require.NoError(t, err)
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) redis.Cmdable
		id      int64
		user    domain.User
		wantErr error
	}{
		{
			name: "设置成功",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := svcmocks.NewMockCmdable(ctrl)
				res := redis.NewStatusResult("Ok", nil)
				cmd.EXPECT().Set(gomock.Any(), "user:info:1", bs, 15*time.Minute).Return(res)
				return cmd
			},
			id:   1,
			user: userDomain,
		},
		{
			name: "设置失败",
			mock: func(ctrl *gomock.Controller) redis.Cmdable {
				cmd := svcmocks.NewMockCmdable(ctrl)
				res := redis.NewStatusResult("", errors.New("模拟设置失败"))
				cmd.EXPECT().Set(gomock.Any(), "user:info:1", bs, 15*time.Minute).Return(res)
				return cmd
			},
			id:      1,
			user:    userDomain,
			wantErr: errors.New("模拟设置失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			uc := NewUserCache(tc.mock(ctrl))
			err := uc.Set(context.Background(), tc.user)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func Test_userCache_key(t *testing.T) {
	testCases := []struct {
		name    string
		id      int64
		wantKey string
	}{
		{
			name:    "1",
			id:      1,
			wantKey: "user:info:1",
		},
	}

	uc := &userCache{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := uc.key(tc.id)
			assert.Equal(t, tc.wantKey, key)
		})
	}
}
