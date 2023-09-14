package async

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/internal/service/sms"
)

type service struct {
	svc           sms.Service
	repo          repository.SMSRepository
	retryInterval time.Duration
	maxRetry      int64
	ticker        *time.Ticker
}

func NewService(svc sms.Service, repo repository.SMSRepository, retryInterval time.Duration, maxRetry int64) sms.Service {
	s := &service{svc: svc, repo: repo, ticker: time.NewTicker(retryInterval), maxRetry: maxRetry}
	s.startAsync()
	return s
}

func (s *service) startAsync() {
	go func() {
		for range s.ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 4*s.retryInterval/5)
			reqs, err := s.repo.FindRetryWithMaxRetry(ctx, s.maxRetry, 2)
			if err != nil {
				//todo log
				cancel()
				continue
			}

			success := make([]int64, 0, len(reqs))
			failed := make([]int64, 0, len(reqs))
			for _, req := range reqs {
				args := strings.Split(req.Args, ";")
				numbers := strings.Split(req.Numbers, ",")
				if err = s.svc.Send(ctx, req.Biz, args, numbers...); err != nil {
					failed = append(failed, req.ID)
					continue
				}
				success = append(success, req.ID)
			}

			go func() {
				if err = s.repo.UpdateRetryCnt(ctx, failed); err != nil {
					//todo log
				}
			}()
			go func() {
				if err = s.repo.UpdateStatus(ctx, success, 1); err != nil {
					//todo log
				}
			}()
			cancel()
		}
	}()
}

func (s *service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	err := s.svc.Send(ctx, tplId, args, numbers...)
	if errors.Is(err, sms.ErrLimited) || errors.Is(err, sms.ErrServiceProviderException) {
		bs, er := json.Marshal(&args)
		if er != nil {
			return er
		}
		return s.repo.Store(ctx, domain.SMS{
			Biz:     tplId,
			Args:    string(bs),
			Numbers: strings.Join(numbers, ","),
			Status:  2,
		})
	}
	return err
}
