package tencent

import (
	"context"
	"fmt"

	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	txsms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"

	"geektime-basic-go/webook/internal/service/sms"
)

type smsService struct {
	client   *txsms.Client
	appId    *string
	signName *string
}

func NewService(c *txsms.Client, appId string, signName string) sms.Service {
	return &smsService{
		client:   c,
		appId:    ekit.ToPtr[string](appId),
		signName: ekit.ToPtr[string](signName),
	}
}

func (s *smsService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	req := txsms.NewSendSmsRequest()
	req.PhoneNumberSet = toStringPtrSlice(numbers)
	req.SmsSdkAppId = s.appId
	req.SetContext(ctx)
	req.TemplateParamSet = toStringPtrSlice(args)
	req.TemplateId = ekit.ToPtr[string](tplId)
	req.SignName = s.signName
	resp, err := s.client.SendSms(req)
	if err != nil {
		return err
	}
	for _, status := range resp.Response.SendStatusSet {
		if status.Code == nil || *(status.Code) != "Ok" {
			return fmt.Errorf("发送失败，code: %s, 原因：%s", *status.Code, *status.Message)
		}
	}
	return nil
}

func toStringPtrSlice(src []string) []*string {
	return slice.Map[string, *string](src, func(idx int, src string) *string {
		return &src
	})
}
