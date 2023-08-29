package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository/cache"
	cachemocks "geektime-basic-go/webook/internal/repository/cache/mocks"
	"geektime-basic-go/webook/internal/repository/dao"
	daomocks "geektime-basic-go/webook/internal/repository/dao/mocks"
)

func TestCachedUserRepository_FindById(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	now := time.UnixMilli(nowMs)

	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache)
		ctx  context.Context
		Id   int64

		wantUser domain.User
		wantErr  error
	}{
		{
			name: "找到了用户, 未命中缓存",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				ud := daomocks.NewMockUserDAO(ctrl)
				uc := cachemocks.NewMockUserCache(ctrl)

				uc.EXPECT().Get(gomock.Any(), int64(12)).Return(domain.User{}, cache.ErrKeyNotExist)

				ud.EXPECT().FindByID(gomock.Any(), int64(12)).Return(dao.User{
					Id:       12,
					Email:    sql.NullString{String: "123@qq.com", Valid: true},
					Password: "123456",
					Phone:    sql.NullString{String: "13888888888", Valid: true},
					Nickname: sql.NullString{String: "泰裤辣", Valid: true},
					AboutMe:  sql.NullString{String: "泰裤辣", Valid: true},
					Birthday: sql.NullInt64{Int64: nowMs, Valid: true},
					CreateAt: nowMs,
					UpdateAt: nowMs,
				}, nil)
				uc.EXPECT().Set(gomock.Any(), domain.User{
					Id:       12,
					Email:    "123@qq.com",
					Password: "123456",
					Phone:    "13888888888",
					Nickname: "泰裤辣",
					AboutMe:  "泰裤辣",
					Birthday: now,
					CreateAt: now,
				}).Return(nil)

				return ud, uc
			},
			ctx: context.Background(),
			Id:  12,
			wantUser: domain.User{
				Id:       12,
				Email:    "123@qq.com",
				Password: "123456",
				Phone:    "13888888888",
				Nickname: "泰裤辣",
				AboutMe:  "泰裤辣",
				Birthday: now,
				CreateAt: now,
			},
		},
		{
			name: "找到用户, 直接命中缓存",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				uc := cachemocks.NewMockUserCache(ctrl)
				uc.EXPECT().Get(gomock.Any(), int64(12)).Return(domain.User{
					Id:       12,
					Email:    "123@qq.com",
					Password: "123456",
					Phone:    "13888888888",
					Nickname: "泰裤辣",
					AboutMe:  "泰裤辣",
					Birthday: now,
					CreateAt: now,
				}, nil)

				return nil, uc
			},
			ctx: context.Background(),
			Id:  12,
			wantUser: domain.User{
				Id:       12,
				Email:    "123@qq.com",
				Password: "123456",
				Phone:    "13888888888",
				Nickname: "泰裤辣",
				AboutMe:  "泰裤辣",
				Birthday: now,
				CreateAt: now,
			},
			wantErr: nil,
		},
		{
			name: "没找到用户",
			mock: func(ctrl *gomock.Controller) (dao.UserDAO, cache.UserCache) {
				ud := daomocks.NewMockUserDAO(ctrl)
				uc := cachemocks.NewMockUserCache(ctrl)

				uc.EXPECT().Get(gomock.Any(), int64(12)).Return(domain.User{}, cache.ErrKeyNotExist)
				ud.EXPECT().FindByID(gomock.Any(), int64(12)).Return(dao.User{}, dao.ErrDataNotFound)

				return ud, uc
			},
			ctx:      context.Background(),
			Id:       12,
			wantUser: domain.User{},
			wantErr:  ErrUserNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ud, uc := tc.mock(ctrl)
			repo := NewUserRepository(ud, uc)
			u, err := repo.FindByID(tc.ctx, tc.Id)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantUser, u)
		})
	}
}
