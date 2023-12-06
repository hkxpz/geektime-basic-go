package wechat

import (
	"context"

	"geektime-basic-go/webook/internal/domain"
)

//go:generate mockgen -source=types.go -package=svcmocks -destination=mocks/types_mock_gen.go Service
type Service interface {
	AuthURL(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error)
}
