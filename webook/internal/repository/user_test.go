package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository/cache"
	cachemocks "geektime-basic-go/webook/internal/repository/cache/mocks"
	"geektime-basic-go/webook/internal/repository/dao"
	daomocks "geektime-basic-go/webook/internal/repository/dao/mocks"
)

func TestUserRepository_Create(t *testing.T) {
	userDao := dao.User{
		Email:    sql.NullString{String: "123@qq.com", Valid: true},
		Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
		Phone:    sql.NullString{String: "13888888888", Valid: true},
	}
	userDomain := domain.User{
		Email:    "123@qq.com",
		Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
		Phone:    "13888888888",
	}
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) dao.UserDAO
		user    domain.User
		wantErr error
	}{
		{
			name: "创建成功",
			mock: func(ctrl *gomock.Controller) dao.UserDAO {
				ud := daomocks.NewMockUserDAO(ctrl)
				ud.EXPECT().Insert(gomock.Any(), userDao).Return(nil)
				return ud
			},
			user: userDomain,
		},
		{
			name: "创建失败",
			mock: func(ctrl *gomock.Controller) dao.UserDAO {
				ud := daomocks.NewMockUserDAO(ctrl)
				ud.EXPECT().Insert(gomock.Any(), userDao).Return(errors.New("模拟创建失败"))
				return ud
			},
			user:    userDomain,
			wantErr: errors.New("模拟创建失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ud := tc.mock(ctrl)
			repo := NewUserRepository(ud, nil)
			err := repo.Create(context.Background(), tc.user)
			require.Equal(t, tc.wantErr, err)
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	now := time.UnixMilli(nowMs)
	userDomain := domain.User{
		ID:       1,
		Email:    "123@qq.com",
		Nickname: "泰裤辣",
		Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
		Phone:    "13888888888",
		AboutMe:  "泰裤辣",
		Birthday: now,
	}
	userDao := dao.User{
		Id:       1,
		Email:    sql.NullString{String: "123@qq.com", Valid: true},
		Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
		Phone:    sql.NullString{String: "13888888888", Valid: true},
		Nickname: sql.NullString{String: "泰裤辣", Valid: true},
		AboutMe:  sql.NullString{String: "泰裤辣", Valid: true},
		Birthday: sql.NullInt64{Int64: nowMs, Valid: true},
	}
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)
		user    domain.User
		wantErr error
	}{
		{
			name: "更新成功",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				ud := daomocks.NewMockUserDAO(ctrl)
				uc := cachemocks.NewMockUserCache(ctrl)
				ud.EXPECT().Update(gomock.Any(), userDao).Return(nil)
				uc.EXPECT().Delete(gomock.Any(), int64(1)).Return(nil)
				return ud, uc
			},
			user: userDomain,
		},
		{
			name: "更新失败",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				ud := daomocks.NewMockUserDAO(ctrl)
				uc := cachemocks.NewMockUserCache(ctrl)
				ud.EXPECT().Update(gomock.Any(), userDao).Return(errors.New("模拟更新失败"))
				return ud, uc
			},
			user:    userDomain,
			wantErr: errors.New("模拟更新失败"),
		},
		{
			name: "删除缓存失败",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				ud := daomocks.NewMockUserDAO(ctrl)
				uc := cachemocks.NewMockUserCache(ctrl)
				ud.EXPECT().Update(gomock.Any(), userDao).Return(nil)
				uc.EXPECT().Delete(gomock.Any(), int64(1)).Return(errors.New("模拟删除缓存失败"))
				return ud, uc
			},
			user:    userDomain,
			wantErr: errors.New("模拟删除缓存失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ud, uc := tc.mock(ctrl)
			repo := NewUserRepository(ud, uc)
			err := repo.Update(context.Background(), tc.user)
			require.Equal(t, tc.wantErr, err)
		})
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	now := time.UnixMilli(nowMs)
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
	userDao := dao.User{
		Id:       1,
		Email:    sql.NullString{String: "123@qq.com", Valid: true},
		Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
		Phone:    sql.NullString{String: "13888888888", Valid: true},
		Nickname: sql.NullString{String: "泰裤辣", Valid: true},
		AboutMe:  sql.NullString{String: "泰裤辣", Valid: true},
		Birthday: sql.NullInt64{Int64: nowMs, Valid: true},
		CreateAt: nowMs,
		UpdateAt: nowMs,
	}
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)
		id       int64
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "找到了用户,未命中缓存",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				ud := daomocks.NewMockUserDAO(ctrl)
				uc := cachemocks.NewMockUserCache(ctrl)
				uc.EXPECT().Get(gomock.Any(), int64(1)).Return(domain.User{}, cache.ErrKeyNotExist)
				ud.EXPECT().FindByID(gomock.Any(), int64(1)).Return(userDao, nil)
				uc.EXPECT().Set(gomock.Any(), userDomain).Return(nil)
				return ud, uc
			},
			id:       1,
			wantUser: userDomain,
		},
		{
			name: "找到用户,直接命中缓存",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uc := cachemocks.NewMockUserCache(ctrl)
				uc.EXPECT().Get(gomock.Any(), int64(1)).Return(userDomain, nil)
				return nil, uc
			},
			id:       1,
			wantUser: userDomain,
		},
		{
			name: "没找到用户",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				ud := daomocks.NewMockUserDAO(ctrl)
				uc := cachemocks.NewMockUserCache(ctrl)
				uc.EXPECT().Get(gomock.Any(), int64(1)).Return(domain.User{}, cache.ErrKeyNotExist)
				ud.EXPECT().FindByID(gomock.Any(), int64(1)).Return(dao.User{}, dao.ErrDataNotFound)
				return ud, uc
			},
			id:       1,
			wantUser: domain.User{},
			wantErr:  ErrUserNotFound,
		},
		{
			name: "缓存获取用户失败",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uc := cachemocks.NewMockUserCache(ctrl)
				uc.EXPECT().Get(gomock.Any(), int64(1)).Return(domain.User{}, errors.New("模拟缓存获取用户失败"))
				return nil, uc
			},
			id:       1,
			wantUser: domain.User{},
			wantErr:  errors.New("模拟缓存获取用户失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ud, uc := tc.mock(ctrl)
			repo := NewUserRepository(ud, uc)
			u, err := repo.FindByID(context.Background(), tc.id)
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, u)
		})
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	now := time.UnixMilli(nowMs)
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
	userDao := dao.User{
		Id:       1,
		Email:    sql.NullString{String: "123@qq.com", Valid: true},
		Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
		Phone:    sql.NullString{String: "13888888888", Valid: true},
		Nickname: sql.NullString{String: "泰裤辣", Valid: true},
		AboutMe:  sql.NullString{String: "泰裤辣", Valid: true},
		Birthday: sql.NullInt64{Int64: nowMs, Valid: true},
		CreateAt: nowMs,
		UpdateAt: nowMs,
	}
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)
		email    string
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "找到用户",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				ud := daomocks.NewMockUserDAO(ctrl)
				ud.EXPECT().FindByEmail(gomock.Any(), "123@qq.com").Return(userDao, nil)
				return ud, nil
			},
			email:    "123@qq.com",
			wantUser: userDomain,
		},
		{
			name: "没找到用户",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				ud := daomocks.NewMockUserDAO(ctrl)
				ud.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(dao.User{}, dao.ErrDataNotFound)
				return ud, nil
			},
			email:    "123@qq.com",
			wantUser: domain.User{CreateAt: time.UnixMilli(0)},
			wantErr:  dao.ErrDataNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ud, uc := tc.mock(ctrl)
			repo := NewUserRepository(ud, uc)
			u, err := repo.FindByEmail(context.Background(), tc.email)
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, u)
		})
	}
}

func TestUserRepository_FindByPhone(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	now := time.UnixMilli(nowMs)
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
	userDao := dao.User{
		Id:       1,
		Email:    sql.NullString{String: "123@qq.com", Valid: true},
		Password: "$2a$10$s51GBcU20dkNUVTpUAQqpe6febjXkRYvhEwa5OkN5rU6rw2KTbNUi",
		Phone:    sql.NullString{String: "13888888888", Valid: true},
		Nickname: sql.NullString{String: "泰裤辣", Valid: true},
		AboutMe:  sql.NullString{String: "泰裤辣", Valid: true},
		Birthday: sql.NullInt64{Int64: nowMs, Valid: true},
		CreateAt: nowMs,
		UpdateAt: nowMs,
	}
	testCases := []struct {
		name     string
		mock     func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)
		phone    string
		wantUser domain.User
		wantErr  error
	}{
		{
			name: "找到用户",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				ud := daomocks.NewMockUserDAO(ctrl)
				ud.EXPECT().FindByPhone(gomock.Any(), "13888888888").Return(userDao, nil)
				return ud, nil
			},
			phone:    "13888888888",
			wantUser: userDomain,
		},
		{
			name: "没找到用户",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				ud := daomocks.NewMockUserDAO(ctrl)
				ud.EXPECT().FindByPhone(gomock.Any(), "13888888888").Return(dao.User{}, dao.ErrDataNotFound)
				return ud, nil
			},
			phone:    "13888888888",
			wantUser: domain.User{CreateAt: time.UnixMilli(0)},
			wantErr:  dao.ErrDataNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ud, uc := tc.mock(ctrl)
			repo := NewUserRepository(ud, uc)
			u, err := repo.FindByPhone(context.Background(), tc.phone)
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, u)
		})
	}
}
