package async

import (
	"context"
	"errors"
	"log"

	"geektime-basic-go/homework/week04/domain"
	"geektime-basic-go/homework/week04/repository"
	"geektime-basic-go/homework/week04/service/sms"
)

type service struct {
	svc  sms.Service
	repo repository.Repository

	// 重试次数
	maxReTryCnt uint32
}

func NewService(svc sms.Service, repo repository.Repository, maxReTryCnt uint32) sms.Service {
	return &service{svc: svc, repo: repo, maxReTryCnt: maxReTryCnt}
}

func (s *service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	go func() {
		for i := s.maxReTryCnt; i > 0; i++ {
			err := s.svc.Send(ctx, tplId, args, numbers...)
			switch {
			default:
				log.Println(err)
			case err == nil:
				return
			case errors.Is(err, sms.ErrLimited), errors.Is(err, sms.ErrServiceProviderException):
				break
			}
		}

		if err := s.repo.Store(ctx, domain.SMS{}); err != nil {
			log.Println(err)
		}
	}()

	return nil
}
