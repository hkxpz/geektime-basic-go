package alibaba

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	"github.com/ecodeclub/ekit"

	"geektime-basic-go/homework/week04/service/sms"
)

type codeService struct {
	client   *client.Client
	signName *string
}

func NewCodeService(client *client.Client, signName string) sms.Service {
	return &codeService{
		client:   client,
		signName: ekit.ToPtr[string](signName),
	}
}

func (s *codeService) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	req := &client.SendSmsRequest{
		SignName:     s.signName,
		TemplateCode: ekit.ToPtr[string]("SMS_154950909"),
		PhoneNumbers: ekit.ToPtr[string](strings.Join(numbers, ",")),
	}

	req.TemplateParam = ekit.ToPtr[string](fmt.Sprintf(`{"code":"%s"}`, args[0]))
	resp, err := s.client.SendSms(req)
	if err != nil {
		log.Println("发送短信失败:", err)
		return err
	}
	if *(resp.Body.Code) != "OK" {
		log.Println(resp.Body.String())
	}
	return nil
}
