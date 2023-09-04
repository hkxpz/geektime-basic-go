package redis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"geektime-basic-go/webook/internal/repository/cache"
	"geektime-basic-go/webook/ioc"
)

func TestCodeCache_Set_e2e(t *testing.T) {
	rdb := ioc.InitRedis()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, rdb.Ping(ctx).Err())
	c := NewCodeCache(rdb).(*codeCache)

	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		ctx   context.Context
		biz   string
		phone string
		code  string

		wantErr error
	}{
		{
			name:   "验证码储存成功",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				key := c.key("login", "13888888888")
				val, err := rdb.Get(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", val)
				ttl, err := rdb.TTL(ctx, key).Result()
				assert.NoError(t, err)
				assert.True(t, ttl > 9*time.Minute)
				err = rdb.Del(ctx, key).Err()
				assert.NoError(t, err)
			},
			ctx:   ctx,
			biz:   "login",
			phone: "13888888888",
			code:  "123456",
		},
		{
			name: "发送太频繁",
			before: func(t *testing.T) {
				key := c.key("login", "13888888888")
				err := rdb.Set(ctx, key, "123456", 10*time.Minute).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				key := c.key("login", "13888888888")
				val, err := rdb.Get(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", val)
				err = rdb.Del(ctx, key).Err()
				assert.NoError(t, err)
			},
			ctx:     ctx,
			biz:     "login",
			phone:   "13888888888",
			code:    "123456",
			wantErr: cache.ErrCodeSendTooMany,
		},
		{
			name: "未知错误",
			before: func(t *testing.T) {
				key := c.key("login", "13888888888")
				err := rdb.Set(ctx, key, "123456", 0).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				key := c.key("login", "13888888888")
				val, err := rdb.Get(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", val)
				err = rdb.Del(ctx, key).Err()
				assert.NoError(t, err)
			},
			ctx:     ctx,
			biz:     "login",
			phone:   "13888888888",
			code:    "123456",
			wantErr: cache.ErrUnknownForCode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			err := c.Set(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
			tc.after(t)
		})
	}
}

func TestCodeCache_Verify_e2e(t *testing.T) {
	rdb := ioc.InitRedis()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, rdb.Ping(ctx).Err())
	c := NewCodeCache(rdb).(*codeCache)

	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		ctx   context.Context
		biz   string
		phone string
		code  string

		wantRes bool
		wantErr error
	}{
		{
			name: "验证成功",
			before: func(t *testing.T) {
				key := c.key("login", "13888888888")
				keyCnt := key + ":cnt"
				err := rdb.Set(ctx, key, "123456", 10*time.Minute).Err()
				assert.NoError(t, err)
				err = rdb.Set(ctx, keyCnt, 3, 10*time.Minute).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				key := c.key("login", "13888888888")
				keyCnt := key + ":cnt"
				ttl, err := rdb.TTL(ctx, keyCnt).Result()
				assert.NoError(t, err)
				assert.Equal(t, time.Duration(-1), ttl)
				err = rdb.Del(ctx, key).Err()
				assert.NoError(t, err)
				err = rdb.Del(ctx, keyCnt).Err()
				assert.NoError(t, err)
			},
			ctx:     ctx,
			biz:     "login",
			phone:   "13888888888",
			code:    "123456",
			wantRes: true,
		},
		{
			name: "验证失败",
			before: func(t *testing.T) {
				key := c.key("login", "13888888888")
				keyCnt := key + ":cnt"
				err := rdb.Set(ctx, key, "123456", 10*time.Minute).Err()
				assert.NoError(t, err)
				err = rdb.Set(ctx, keyCnt, 3, 10*time.Minute).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				key := c.key("login", "13888888888")
				keyCnt := key + ":cnt"
				val, err := rdb.Get(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", val)
				val, err = rdb.Get(ctx, keyCnt).Result()
				assert.NoError(t, err)
				assert.Equal(t, "2", val)
				err = rdb.Del(ctx, key).Err()
				assert.NoError(t, err)
				err = rdb.Del(ctx, keyCnt).Err()
				assert.NoError(t, err)
			},
			ctx:     ctx,
			biz:     "login",
			phone:   "13888888888",
			code:    "1234561",
			wantErr: nil,
		},
		{
			name: "验证次数耗尽",
			before: func(t *testing.T) {
				key := c.key("login", "13888888888")
				keyCnt := key + ":cnt"
				err := rdb.Set(ctx, key, "123456", 10*time.Minute).Err()
				assert.NoError(t, err)
				err = rdb.Set(ctx, keyCnt, 0, 10*time.Minute).Err()
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				key := c.key("login", "13888888888")
				keyCnt := key + ":cnt"
				val, err := rdb.Get(ctx, key).Result()
				assert.NoError(t, err)
				assert.Equal(t, "123456", val)
				val, err = rdb.Get(ctx, keyCnt).Result()
				assert.NoError(t, err)
				assert.Equal(t, "0", val)
				err = rdb.Del(ctx, key).Err()
				assert.NoError(t, err)
				err = rdb.Del(ctx, keyCnt).Err()
				assert.NoError(t, err)
			},
			ctx:     ctx,
			biz:     "login",
			phone:   "13888888888",
			code:    "123456",
			wantErr: cache.ErrCodeVerifyTooManyTimes,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			ok, err := c.Verify(tc.ctx, tc.biz, tc.phone, tc.code)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantRes, ok)
			tc.after(t)
		})
	}
}
