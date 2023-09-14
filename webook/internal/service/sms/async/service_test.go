package async

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository"
	repomocks "geektime-basic-go/webook/internal/repository/mocks"
	"geektime-basic-go/webook/internal/service/sms"
	smsmocks "geektime-basic-go/webook/internal/service/sms/mocks"
)

var failed = errors.New("模拟失败")

func TestService_Send(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (sms.Service, repository.SMSRepository)
		wantErr error
	}{
		{
			name: "发送成功",
			mock: func(ctrl *gomock.Controller) (sms.Service, repository.SMSRepository) {
				svc := smsmocks.NewMockService(ctrl)
				svc.EXPECT().Send(gomock.Any(), "", []string{}, "13888888888")
				return svc, nil
			},
		},
		{
			name: "发送失败",
			mock: func(ctrl *gomock.Controller) (sms.Service, repository.SMSRepository) {
				svc := smsmocks.NewMockService(ctrl)
				svc.EXPECT().Send(gomock.Any(), "", []string{}, "13888888888").Return(failed)
				return svc, nil
			},
			wantErr: failed,
		},
		{
			name: "限流",
			mock: func(ctrl *gomock.Controller) (sms.Service, repository.SMSRepository) {
				svc := smsmocks.NewMockService(ctrl)
				repo := repomocks.NewMockSMSRepository(ctrl)
				svc.EXPECT().Send(gomock.Any(), "", []string{}, "13888888888").Return(sms.ErrLimited)
				repo.EXPECT().Store(gomock.Any(), domain.SMS{
					Biz:     "",
					Args:    `[]`,
					Numbers: "13888888888",
					Status:  2,
				})
				return svc, repo
			},
		},
		{
			name: "服务商异常",
			mock: func(ctrl *gomock.Controller) (sms.Service, repository.SMSRepository) {
				svc := smsmocks.NewMockService(ctrl)
				repo := repomocks.NewMockSMSRepository(ctrl)
				svc.EXPECT().Send(gomock.Any(), "", []string{}, "13888888888").Return(sms.ErrServiceProviderException)
				repo.EXPECT().Store(gomock.Any(), domain.SMS{
					Biz:     "",
					Args:    `[]`,
					Numbers: "13888888888",
					Status:  2,
				})
				return svc, repo
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			s, repo := tc.mock(ctrl)
			svc := NewService(s, repo, time.Minute, 3)
			err := svc.Send(context.Background(), "", []string{}, "13888888888")
			assert.Equal(t, tc.wantErr, err)
		})
	}
}
