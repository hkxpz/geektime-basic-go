package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/internal/service/sms"
)

var ErrCodeSendTooMany = repository.ErrCodeSendTooMany

const codeTplId = ""

//go:generate mockgen -source=code.go -package=mocks -destination=mocks/code_mock_gen.go CodeService
type CodeService interface {
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type smsCodeService struct {
	sms  sms.Service
	repo repository.CodeRepository
}

func NewSMSCodeService(svc sms.Service, repo repository.CodeRepository) CodeService {
	return &smsCodeService{sms: svc, repo: repo}
}

func (s *smsCodeService) Send(ctx context.Context, biz, phone string) error {
	code := s.generate()
	err := s.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	return s.sms.Send(ctx, codeTplId, []string{code}, phone)
}

func (s *smsCodeService) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	ok, err := s.repo.Verify(ctx, biz, phone, code)
	if errors.Is(err, repository.ErrCodeVerifyTooManyTimes) {
		return false, nil
	}
	return ok, err
}

func (s *smsCodeService) generate() string {
	return fmt.Sprintf("%06d", rand.Intn(999999))
}
