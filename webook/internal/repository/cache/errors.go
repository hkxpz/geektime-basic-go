package cache

import "errors"

var (
	ErrCodeSendTooMany        = errors.New("发送验证码太频繁")
	ErrUnknownForCode         = errors.New("发送验证码遇到未知错误")
	ErrCodeVerifyTooManyTimes = errors.New("验证次数太多")
)
