package local

import (
	"context"
	"log"

	"geektime-basic-go/homework/week04/service/sms"
)

type service struct{}

func NewService() sms.Service {
	return &service{}
}

func (s service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	log.Println("验证码", args)
	return nil
}
