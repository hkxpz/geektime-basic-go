package ioc

import (
	"os"

	"geektime-basic-go/webook/internal/service/oauth2/wechat"
	"geektime-basic-go/webook/pkg/logger"
)

func InitWechatService(logger logger.Logger) wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("没有找到环境变量 WECHAT_APP_ID ")
	}
	appKey, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("没有找到环境变量 WECHAT_APP_SECRET")
	}
	return wechat.NewService(appId, appKey, logger)
}

func InitLocalWechatService(logger logger.Logger) wechat.Service {
	return wechat.NewService("", "", logger)
}
