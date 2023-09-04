//go:build alibaba

package sms

import (
	"os"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	alibabasms "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	"github.com/ecodeclub/ekit"

	"geektime-basic-go/webook/internal/service/sms"
	"geektime-basic-go/webook/internal/service/sms/alibaba"
)

func InitSmsSvc() sms.Service {
	return initSmsTencentService()
}

func initSmsTencentService() sms.Service {
	accessKeyId, ok := os.LookupEnv("ALIBABA_CLOUD_ACCESS_KEY_ID")
	if !ok {
		panic("没有找到环境变量 ALIBABA_CLOUD_ACCESS_KEY_ID")
	}
	accessKeySecret, ok := os.LookupEnv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")
	if !ok {
		panic("没有找到环境变量 ALIBABA_CLOUD_ACCESS_KEY_SECRET")
	}
	endpoint, ok := os.LookupEnv("ENDPOINT")
	if !ok {
		panic("没有找到环境变量 ENDPOINT")
	}
	signName, ok := os.LookupEnv("SIGN_NAME")
	if !ok {
		panic("没有找到环境变量 SIGN_NAME")
	}

	config := &openapi.Config{
		AccessKeyId:     ekit.ToPtr[string](accessKeyId),
		AccessKeySecret: ekit.ToPtr[string](accessKeySecret),
		Endpoint:        ekit.ToPtr[string](endpoint),
	}
	c, err := alibabasms.NewClient(config)
	if err != nil {
		panic(err)
	}
	return alibaba.NewCodeService(c, signName)
}
