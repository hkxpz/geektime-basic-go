package memory

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

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

type value struct {
	Val        any
	ExpireTime time.Time
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
	bs, err := cc.cmd.Get(key)
	now := time.Now()
	if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
		log.Println("memory: 查询验证码失败 ", err)
		return ErrUnknownForCode
	}
	if err == nil {
		var val value
		err = json.Unmarshal(bs, &val)
		if err != nil {
			log.Println("memory: 反序列化失败 ", err)
			return ErrUnknownForCode
		}
		if now.Sub(val.ExpireTime) < 60*time.Second {
			return ErrCodeSendTooMany
		}
	}

	cb, err := json.Marshal(value{Val: code, ExpireTime: now})
	if err != nil {
		log.Println("memory: 序列化失败 ", err)
		return ErrUnknownForCode
	}
	if err = cc.cmd.Set(key, cb); err != nil {
		log.Println("memory: 设置验证码失败 ", err)
		return ErrUnknownForCode
	}
	nb, err := json.Marshal(value{Val: 3, ExpireTime: now})
	if err != nil {
		log.Println("memory: 序列化失败 ", err)
		return ErrUnknownForCode
	}
	if err = cc.cmd.Set(cc.keyCnt(biz, phone), nb); err != nil {
		log.Println("memory: 设置验证码校验次数失败 ", err)
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
	if err != nil && !errors.Is(err, bigcache.ErrEntryNotFound) {
		log.Println("memory: 查询验证码失败 ", err)
		return false, err
	}
	if errors.Is(err, bigcache.ErrEntryNotFound) || resp.EntryStatus == bigcache.Expired {
		return false, nil
	}

	var cnt value
	err = json.Unmarshal(val, &cnt)
	if err != nil {
		log.Println("memory: 反序列化失败 ", err)
		return false, ErrUnknownForCode
	}
	if cnt.Val.(float64) < 1 {
		return false, ErrCodeVerifyTooManyTimes
	}

	val, resp, err = cc.cmd.GetWithInfo(key)
	if resp.EntryStatus == bigcache.Expired || errors.Is(err, bigcache.ErrEntryNotFound) {
		return false, nil
	}
	if err != nil {
		log.Println("memory: 查询验证码失败 ", err)
		return false, err
	}

	var expectedCode value
	if err = json.Unmarshal(val, &expectedCode); err != nil {
		log.Println("memory: 序列化失败 ", err)
		return false, ErrUnknownForCode
	}
	if expectedCode.Val != code {
		cnt.Val = cnt.Val.(float64) - 1
		nb, cntErr := json.Marshal(cnt)
		if cntErr != nil {
			log.Println("memory: 序列化失败 ", cntErr)
			return false, ErrUnknownForCode
		}
		if err = cc.cmd.Set(keyCnt, nb); err != nil {
			log.Println("memory: 设置验证码校验次数失败 ", err)
		}
		return false, nil
	}

	cnt.Val = -1
	nb, cntErr := json.Marshal(cnt)
	if cntErr != nil {
		log.Println("memory: 序列化失败 ", cntErr)
		return false, ErrUnknownForCode
	}
	if err = cc.cmd.Set(keyCnt, nb); err != nil {
		log.Println("memory: 设置验证码校验次数失败 ", err)
	}
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
		log.Println("memory: 解析缓存失败 ", err)
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
