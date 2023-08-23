//go:build local

package ioc

import (
	"geektime-basic-go/webook/internal/service/sms"
	"geektime-basic-go/webook/internal/service/sms/local"
)

func InitSmsSvc() sms.Service {
	return local.NewService()
}
