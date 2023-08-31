package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/repository/cache"
	"geektime-basic-go/webook/internal/repository/cache/mocks"
)

func Test_cachedCodeRepository_Store(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) cache.CodeCache
		ctx     context.Context
		biz     string
		phone   string
		code    string
		wantErr error
	}{
		{
			name: "储存成功",
			mock: func(ctrl *gomock.Controller) cache.CodeCache {
				cc := mocks.NewMockCodeCache(ctrl)
				cc.EXPECT().Set(context.Background(), "login", "13888888888", "123456").Return(nil)
				return cc
			},
			ctx:   context.Background(),
			biz:   "login",
			phone: "13888888888",
			code:  "123456",
		},
		{
			name: "储存失败",
			mock: func(ctrl *gomock.Controller) cache.CodeCache {
				cc := mocks.NewMockCodeCache(ctrl)
				cc.EXPECT().Set(context.Background(), "login", "13888888888", "123456").Return(errors.New("模拟储存失败"))
				return cc
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "13888888888",
			code:    "123456",
			wantErr: errors.New("模拟储存失败"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cc := tc.mock(ctrl)
			repo := NewCodeRepository(cc)
			err := repo.Store(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func Test_cachedCodeRepository_Verify(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) cache.CodeCache
		ctx     context.Context
		biz     string
		phone   string
		code    string
		ok      bool
		wantErr error
	}{
		{
			name: "验证成功",
			mock: func(ctrl *gomock.Controller) cache.CodeCache {
				cc := mocks.NewMockCodeCache(ctrl)
				cc.EXPECT().Verify(context.Background(), "login", "13888888888", "123456").Return(true, nil)
				return cc
			},
			ctx:   context.Background(),
			biz:   "login",
			phone: "13888888888",
			code:  "123456",
			ok:    true,
		},
		{
			name: "验证失败",
			mock: func(ctrl *gomock.Controller) cache.CodeCache {
				cc := mocks.NewMockCodeCache(ctrl)
				cc.EXPECT().Verify(context.Background(), "login", "13888888888", "123456").Return(false, errors.New("模拟验证失败"))
				return cc
			},
			ctx:     context.Background(),
			biz:     "login",
			phone:   "13888888888",
			code:    "123456",
			wantErr: errors.New("模拟验证失败"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			cc := tc.mock(ctrl)
			repo := NewCodeRepository(cc)
			ok, err := repo.Verify(tc.ctx, tc.biz, tc.phone, tc.code)
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.ok, ok)
		})
	}
}
