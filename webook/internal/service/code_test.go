package service

import (
	"context"
	"errors"
	smsMocks "geektime-basic-go/webook/internal/service/sms/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/internal/repository/mocks"
	"geektime-basic-go/webook/internal/service/sms"
)

func TestSmsCodeService_Send(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (sms.Service, repository.CodeRepository)
		wantErr error
	}{
		{
			name: "发送成功",
			mock: func(ctrl *gomock.Controller) (sms.Service, repository.CodeRepository) {
				ss := smsMocks.NewMockService(ctrl)
				repo := mocks.NewMockCodeRepository(ctrl)
				repo.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				ss.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return ss, repo
			},
		},
		{
			name: "验证码存储失败",
			mock: func(ctrl *gomock.Controller) (sms.Service, repository.CodeRepository) {
				repo := mocks.NewMockCodeRepository(ctrl)
				repo.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("模拟验证码储存失败"))
				return nil, repo
			},
			wantErr: errors.New("模拟验证码储存失败"),
		},
		{
			name: "发送失败",
			mock: func(ctrl *gomock.Controller) (sms.Service, repository.CodeRepository) {
				ss := smsMocks.NewMockService(ctrl)
				repo := mocks.NewMockCodeRepository(ctrl)
				repo.EXPECT().Store(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				ss.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("模拟发送失败"))
				return ss, repo
			},
			wantErr: errors.New("模拟发送失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ss, repo := tc.mock(ctrl)
			svc := NewSMSCodeService(ss, repo)
			err := svc.Send(context.Background(), "login", "13888888888")
			require.Equal(t, tc.wantErr, err)
		})
	}
}

func TestSmsCodeService_Verify(t *testing.T) {
	testCases := []struct {
		name       string
		mock       func(ctrl *gomock.Controller) repository.CodeRepository
		wantVerify bool
		wantErr    error
	}{
		{
			name: "验证成功",
			mock: func(ctrl *gomock.Controller) repository.CodeRepository {
				repo := mocks.NewMockCodeRepository(ctrl)
				repo.EXPECT().Verify(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
				return repo
			},
			wantVerify: true,
		},
		{
			name: "验证失败",
			mock: func(ctrl *gomock.Controller) repository.CodeRepository {
				repo := mocks.NewMockCodeRepository(ctrl)
				repo.EXPECT().Verify(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, errors.New("模拟验证失败"))
				return repo
			},
			wantErr: errors.New("模拟验证失败"),
		},
		{
			name: "验证次数过多",
			mock: func(ctrl *gomock.Controller) repository.CodeRepository {
				repo := mocks.NewMockCodeRepository(ctrl)
				repo.EXPECT().Verify(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false, repository.ErrCodeVerifyTooManyTimes)
				return repo
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := tc.mock(ctrl)
			svc := NewSMSCodeService(nil, repo)
			ok, err := svc.Verify(context.Background(), "login", "13888888888", "123456")
			require.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantVerify, ok)
		})
	}
}
