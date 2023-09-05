package sms

import (
	"errors"
)

var (
	ErrLimited                  = errors.New("短信服务触发限流")
	ErrServiceProviderException = errors.New("短信服务提供商异常")
)
