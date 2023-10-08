package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/repository"
	"geektime-basic-go/webook/internal/repository/mocks"
)

func TestArticleService_Save(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) repository.ArticleRepository
		art     domain.Article
		wantErr error
		wantID  int64
	}{
		{
			name: "新建文章成功",
			mock: func(ctrl *gomock.Controller) repository.ArticleRepository {
				repo := mocks.NewMockArticleRepository(ctrl)
				repo.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "我的文章",
					Content: "我的内容",
					Status:  domain.ArticleStatusUnpublished,
					Author:  domain.Author{ID: 123},
				}).Return(int64(1), nil)
				return repo
			},
			art: domain.Article{
				Title:   "我的文章",
				Content: "我的内容",
				Author:  domain.Author{ID: 123},
			},
			wantID: 1,
		},
		{
			name: "新建文章失败",
			mock: func(ctrl *gomock.Controller) repository.ArticleRepository {
				repo := mocks.NewMockArticleRepository(ctrl)
				repo.EXPECT().Create(gomock.Any(), domain.Article{
					Title:   "我的文章",
					Content: "我的内容",
					Status:  domain.ArticleStatusUnpublished,
					Author:  domain.Author{ID: 123},
				}).Return(int64(0), errors.New("模拟新建失败"))
				return repo
			},
			art: domain.Article{
				Title:   "我的文章",
				Content: "我的内容",
				Author:  domain.Author{ID: 123},
			},
			wantErr: errors.New("模拟新建失败"),
		},
		{
			name: "更新文章成功",
			mock: func(ctrl *gomock.Controller) repository.ArticleRepository {
				repo := mocks.NewMockArticleRepository(ctrl)
				repo.EXPECT().Update(gomock.Any(), domain.Article{
					ID:      1,
					Title:   "我的文章",
					Content: "我的内容",
					Status:  domain.ArticleStatusUnpublished,
					Author:  domain.Author{ID: 123},
				}).Return(nil)
				return repo
			},
			art: domain.Article{
				ID:      1,
				Title:   "我的文章",
				Content: "我的内容",
				Author:  domain.Author{ID: 123},
			},
			wantID: 1,
		},
		{
			name: "更新文章失败",
			mock: func(ctrl *gomock.Controller) repository.ArticleRepository {
				repo := mocks.NewMockArticleRepository(ctrl)
				repo.EXPECT().Update(gomock.Any(), domain.Article{
					ID:      1,
					Title:   "我的文章",
					Content: "我的内容",
					Status:  domain.ArticleStatusUnpublished,
					Author:  domain.Author{ID: 123},
				}).Return(errors.New("模拟失败"))
				return repo
			},
			art: domain.Article{
				ID:      1,
				Title:   "我的文章",
				Content: "我的内容",
				Author:  domain.Author{ID: 123},
			},
			wantID:  1,
			wantErr: errors.New("模拟失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewArticleService(tc.mock(ctrl), nil)
			id, err := svc.Save(context.Background(), tc.art)
			require.Equal(t, tc.wantErr, err)
			require.Equal(t, tc.wantID, id)
		})
	}
}

func TestArticleService_Publish(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) repository.ArticleRepository
		art     domain.Article
		wantErr error
		wantID  int64
	}{
		{
			name: "成功",
			mock: func(ctrl *gomock.Controller) repository.ArticleRepository {
				repo := mocks.NewMockArticleRepository(ctrl)
				repo.EXPECT().Sync(gomock.Any(), domain.Article{
					Title:   "我的文章",
					Content: "我的内容",
					Status:  domain.ArticleStatusPublished,
					Author:  domain.Author{ID: 123},
				}).Return(int64(1), nil)
				return repo
			},
			art: domain.Article{
				Title:   "我的文章",
				Content: "我的内容",
				Author:  domain.Author{ID: 123},
			},
			wantID: 1,
		},
		{
			name: "失败",
			mock: func(ctrl *gomock.Controller) repository.ArticleRepository {
				repo := mocks.NewMockArticleRepository(ctrl)
				repo.EXPECT().Sync(gomock.Any(), domain.Article{
					ID:      1,
					Title:   "我的文章",
					Content: "我的内容",
					Status:  domain.ArticleStatusPublished,
					Author:  domain.Author{ID: 123},
				}).Return(int64(0), errors.New("模拟失败"))
				return repo
			},
			art: domain.Article{
				ID:      1,
				Title:   "我的文章",
				Content: "我的内容",
				Author:  domain.Author{ID: 123},
			},
			wantID:  0,
			wantErr: errors.New("模拟失败"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewArticleService(tc.mock(ctrl), nil)
			id, err := svc.Publish(context.Background(), tc.art)
			require.Equal(t, tc.wantErr, err)
			require.Equal(t, tc.wantID, id)
		})
	}
}
