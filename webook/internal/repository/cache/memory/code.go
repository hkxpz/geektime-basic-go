package memory

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/allegro/bigcache/v3"

	"geektime-basic-go/webook/internal/repository/cache"
)

var (
	ErrCodeSendTooMany        = errors.New("发送验证码太频繁")
	ErrUnknownForCode         = errors.New("发送验证码遇到未知错误")
	ErrCodeVerifyTooManyTimes = errors.New("验证次数太多")
)

type codeCache struct {
	cmd *bigcache.BigCache
	l   sync.Mutex
}

func NewMemoryCodeCache(cmd *bigcache.BigCache) cache.CodeCache {
	return &codeCache{cmd: cmd}
}

// Set 如果该手机在该业务场景下，验证码不存在（都已经过期），那么发送
// 如果已经有一个验证码，但是发出去已经一分钟了，允许重发
// 如果已经有一个验证码，但是没有过期时间，说明有不知名错误
// 如果已经有一个验证码，但是发出去不到一分钟，不允许重发
// 验证码有效期 10 分钟
func (cc *codeCache) Set(ctx context.Context, biz, phone, code string) error {
	cc.l.Lock()
	defer cc.l.Unlock()

	key := cc.key(biz, phone)
	val, err := cc.cmd.Get(key)
	if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
		log.Println("查询验证码失败:", err)
		return ErrUnknownForCode
	}
	if val != nil {
		return ErrCodeSendTooMany
	}

	if err = cc.cmd.Set(key, []byte(code)); err != nil {
		log.Println("设置验证码失败:", err)
		return ErrUnknownForCode
	}
	if err = cc.cmd.Set(cc.keyCnt(biz, phone), cc.intToByteSlice(3)); err != nil {
		log.Println("设置验证码校验次数失败:", err)
		return ErrUnknownForCode
	}
	return nil
}

// Verify 验证验证码
// 如果验证码是一致的，那么删除
// 如果验证码不一致，那么保留的
func (cc *codeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	cc.l.Lock()
	defer cc.l.Unlock()

	key, keyCnt := cc.key(biz, phone), cc.keyCnt(biz, phone)
	val, resp, err := cc.cmd.GetWithInfo(keyCnt)
	cnt := cc.byteSliceToInt(val)
	switch {
	case err != nil && !errors.Is(err, bigcache.ErrEntryNotFound):
		log.Println("查询验证码失败:", err)
		return false, err
	case val == nil, resp.EntryStatus == bigcache.Expired:
		return false, nil
	case cnt < 1:
		_ = cc.cmd.Delete(key)
		_ = cc.cmd.Delete(keyCnt)
		return false, ErrCodeVerifyTooManyTimes
	}

	val, resp, err = cc.cmd.GetWithInfo(key)
	switch {
	case err != nil && !errors.Is(err, bigcache.ErrEntryNotFound):
		log.Println("查询验证码失败:", err)
		return false, err
	case resp.EntryStatus == bigcache.Expired, string(val) != code:
		if err = cc.cmd.Set(keyCnt, cc.intToByteSlice(cnt-1)); err != nil {
			log.Println("设置验证码校验次数失败:", err)
		}
		return false, nil
	}

	_ = cc.cmd.Delete(key)
	_ = cc.cmd.Delete(keyCnt)

	return true, nil
}

func (cc *codeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

func (cc *codeCache) keyCnt(biz, phone string) string {
	return cc.key(biz, phone) + ":cut"
}

func (cc *codeCache) byteSliceToInt(src []byte) int64 {
	if src == nil {
		return 0
	}
	bytesBuffer := bytes.NewBuffer(src)
	var num int64
	if err := binary.Read(bytesBuffer, binary.BigEndian, &num); err != nil {
		log.Println("解析缓存失败:", err)
	}
	return num
}

func (cc *codeCache) intToByteSlice(src int64) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	if err := binary.Write(bytesBuffer, binary.BigEndian, &src); err != nil {
		log.Println(err)
	}
	return bytesBuffer.Bytes()
}
