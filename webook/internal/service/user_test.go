package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/internal/repository/mocks"
)

func TestUserService_Login(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) repository.UserRepository

		ctx      context.Context
		email    string
		password string

		wantErr  error
		wantUser domain.User
	}{
		{
			name: "登录成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := mocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(context.Background(), "123@qq.com").Return(domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Nickname: "泰裤辣",
					Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
					Phone:    "13888888888",
					AboutMe:  "泰裤辣",
					Birthday: now,
					CreateAt: now,
				}, nil)
				return repo
			},
			ctx:      context.Background(),
			email:    "123@qq.com",
			password: "hello#world123",
			wantUser: domain.User{
				Id:       123,
				Email:    "123@qq.com",
				Nickname: "泰裤辣",
				Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
				Phone:    "13888888888",
				AboutMe:  "泰裤辣",
				Birthday: now,
				CreateAt: now,
			},
		},
		{
			name: "未找到用户",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := mocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(context.Background(), "123@qq.com").Return(domain.User{}, ErrInvalidUserOrPassword)
				return repo
			},
			ctx:     context.Background(),
			email:   "123@qq.com",
			wantErr: ErrInvalidUserOrPassword,
		},
		{
			name: "密码错误",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := mocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(context.Background(), "123@qq.com").Return(domain.User{
					Id:       123,
					Email:    "123@qq.com",
					Nickname: "泰裤辣",
					Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
					Phone:    "13888888888",
					AboutMe:  "泰裤辣",
					Birthday: now,
					CreateAt: now,
				}, nil)
				return repo
			},
			ctx:      context.Background(),
			email:    "123@qq.com",
			password: "123@qq.com",
			wantErr:  ErrInvalidUserOrPassword,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tc.mock(ctrl)
			svc := NewUserService(repo)
			user, err := svc.Login(tc.ctx, tc.email, tc.password)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}
