package auth

import (
	"context"

	"github.com/golang-jwt/jwt/v5"

	"geektime-basic-go/webook/internal/service/sms"
)

type service struct {
	svc sms.Service
	key []byte
}

func newService(svc sms.Service, key []byte) sms.Service {
	return &service{svc: svc, key: key}
}

func (s *service) Send(ctx context.Context, token string, args []string, numbers ...string) error {
	var c claims
	_, err := jwt.ParseWithClaims(token, &c, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	return s.svc.Send(ctx, c.tpl, args, numbers...)
}

type claims struct {
	jwt.RegisteredClaims
	tpl string
}
