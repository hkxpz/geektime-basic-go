//go:build prd

package ioc

import (
	"os"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	txsms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"

	"geektime-basic-go/webook/internal/service/sms"
	"geektime-basic-go/webook/internal/service/sms/tencent"
)

func InitSmsSvc() sms.Service {
	return initSmsTencentService()
}

func initSmsTencentService() sms.Service {
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		panic("没有找到环境变量 SMS_SECRET_ID")
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
	if !ok {
		panic("没有找到环境变量 SMS_SECRET_KEY")
	}
	region, ok := os.LookupEnv("REGION")
	if !ok {
		panic("没有找到环境变量 REGION")
	}
	appId, ok := os.LookupEnv("APP_ID")
	if !ok {
		panic("没有找到环境变量 APP_ID")
	}
	signName, ok := os.LookupEnv("SIGN_NAME")
	if !ok {
		panic("没有找到环境变量 SIGN_NAME")
	}

	c, err := txsms.NewClient(common.NewCredential(secretId, secretKey), region, profile.NewClientProfile())
	if err != nil {
		panic(err)
	}
	return tencent.NewService(c, appId, signName)
}
