package redis

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/redis/go-redis/v9"

	"geektime-basic-go/webook/internal/repository/cache"
)

var (
	//go:embed lua/set_code.lua
	luaSetCode string
	//go:embed lua/verify_code.lua
	luaVerifyCode string
)

type codeCache struct {
	cmd redis.Cmdable
}

func NewCodeCache(cmd redis.Cmdable) cache.CodeCache {
	return &codeCache{cmd: cmd}
}

// Set 如果该手机在该业务场景下，验证码不存在（都已经过期），那么发送
// 如果已经有一个验证码，但是发出去已经一分钟了，允许重发
// 如果已经有一个验证码，但是没有过期时间，说明有不知名错误
// 如果已经有一个验证码，但是发出去不到一分钟，不允许重发
// 验证码有效期 10 分钟
func (cc *codeCache) Set(ctx context.Context, biz, phone, code string) error {
	res, err := cc.cmd.Eval(ctx, luaSetCode, []string{cc.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch res {
	default:
		return cache.ErrUnknownForCode
	case 0:
		return nil
	case -1:
		return cache.ErrCodeSendTooMany
	}
}

// Verify 验证验证码
// 如果验证码是一致的，那么删除
// 如果验证码不一致，那么保留的
func (cc *codeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	res, err := cc.cmd.Eval(ctx, luaVerifyCode, []string{cc.key(biz, phone)}, code).Int()
	if err != nil {
		return false, err
	}
	switch res {
	default:
		return false, err
	case 0:
		return true, nil
	case -1:
		return false, cache.ErrCodeVerifyTooManyTimes
	}
}

func (cc *codeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
