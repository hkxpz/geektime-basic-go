package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/internal/repository/mocks"
)

func TestUserService_Signup(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) repository.UserRepository
		ctx     context.Context
		user    domain.User
		wantErr error
	}{
		{
			name: "注册成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := mocks.NewMockUserRepository(ctrl)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
				return repo
			},
			ctx:     context.Background(),
			user:    domain.User{Password: "123123"},
			wantErr: nil,
		},
		{
			name: "加密失败",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				return nil
			},
			ctx:     context.Background(),
			user:    domain.User{Password: "你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好你好"},
			wantErr: bcrypt.ErrPasswordTooLong,
		},
		{
			name: "注册失败",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := mocks.NewMockUserRepository(ctrl)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("模拟注册失败"))
				return repo
			},
			ctx:     context.Background(),
			user:    domain.User{Password: "123123"},
			wantErr: errors.New("模拟注册失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tc.mock(ctrl)
			svc := NewUserService(repo)
			err := svc.Signup(tc.ctx, tc.user)
			require.Equal(t, tc.wantErr, err)
		})
	}
}

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
				repo.EXPECT().FindByEmail(context.Background(), "123@qq.com").Return(domain.User{}, repository.ErrUserNotFound)
				return repo
			},
			ctx:     context.Background(),
			email:   "123@qq.com",
			wantErr: ErrInvalidUserOrPassword,
		},
		{
			name: "查找用户失败",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := mocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByEmail(context.Background(), "123@qq.com").Return(domain.User{}, errors.New("模拟查找用户失败"))
				return repo
			},
			ctx:     context.Background(),
			email:   "123@qq.com",
			wantErr: errors.New("模拟查找用户失败"),
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
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

func TestUserService_Profile(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) repository.UserRepository
		ctx      context.Context
		id       int64
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "获取成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := mocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(domain.User{Id: 1}, nil)
				return repo
			},
			ctx:      context.Background(),
			id:       1,
			wantUser: domain.User{Id: 1},
			wantErr:  nil,
		},
		{
			name: "获取失败",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := mocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByID(gomock.Any(), gomock.Any()).Return(domain.User{}, errors.New("模拟创建失败"))
				return repo
			},
			ctx:      context.Background(),
			id:       1,
			wantUser: domain.User{},
			wantErr:  errors.New("模拟创建失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tc.mock(ctrl)
			svc := NewUserService(repo)
			user, err := svc.Profile(tc.ctx, tc.id)
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

func TestUserService_Edit(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) repository.UserRepository
		ctx     context.Context
		user    domain.User
		wantErr error
	}{
		{
			name: "更新成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := mocks.NewMockUserRepository(ctrl)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
				return repo
			},
			ctx:     context.Background(),
			user:    domain.User{},
			wantErr: nil,
		},
		{
			name: "获取失败",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := mocks.NewMockUserRepository(ctrl)
				repo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("模拟更新失败"))
				return repo
			},
			ctx:     context.Background(),
			user:    domain.User{},
			wantErr: errors.New("模拟更新失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tc.mock(ctrl)
			svc := NewUserService(repo)
			err := svc.Edit(tc.ctx, tc.user)
			require.Equal(t, tc.wantErr, err)
		})
	}
}

func TestUserService_FindOrCreate(t *testing.T) {
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) repository.UserRepository
		ctx      context.Context
		phone    string
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "找到用户",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := mocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByPhone(gomock.Any(), gomock.Any()).Return(domain.User{Phone: "13888888888"}, nil)
				return repo
			},
			ctx:      context.Background(),
			phone:    "13888888888",
			wantUser: domain.User{Phone: "13888888888"},
		},
		{
			name: "查找用户失败",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := mocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByPhone(gomock.Any(), gomock.Any()).Return(domain.User{}, errors.New("模拟查找用户失败"))
				return repo
			},
			ctx:     context.Background(),
			phone:   "13888888888",
			wantErr: errors.New("模拟查找用户失败"),
		},
		{
			name: "没找到用户,创建成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := mocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByPhone(gomock.Any(), gomock.Any()).Return(domain.User{}, repository.ErrUserNotFound)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
				repo.EXPECT().FindByPhone(gomock.Any(), gomock.Any()).Return(domain.User{Phone: "13888888888"}, nil)
				return repo
			},
			ctx:      context.Background(),
			phone:    "13888888888",
			wantUser: domain.User{Phone: "13888888888"},
		},
		{
			name: "没找到用户,没创建成功",
			mock: func(ctrl *gomock.Controller) repository.UserRepository {
				repo := mocks.NewMockUserRepository(ctrl)
				repo.EXPECT().FindByPhone(gomock.Any(), gomock.Any()).Return(domain.User{}, repository.ErrUserNotFound)
				repo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errors.New("模拟创建失败"))
				return repo
			},
			ctx:     context.Background(),
			phone:   "13888888888",
			wantErr: errors.New("模拟创建失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tc.mock(ctrl)
			svc := NewUserService(repo)
			user, err := svc.FindOrCreate(tc.ctx, tc.phone)
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}
