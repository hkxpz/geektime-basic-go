package sms

import "context"

//go:generate mockgen -source=types.go -package=mocks -destination=mocks/types_mock_gen.go Service
type Service interface {
	Send(ctx context.Context, tplId string, args []string, numbers ...string) error
}
