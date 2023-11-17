package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/service/mocks"
)

func TestBatchRankingService_rankTopN(t *testing.T) {
	const batchSize = 2
	now := time.Now()
	mockErr := errors.New("模拟失败")
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (ArticleService, InteractiveService)
		wantErr error
		wantRes []domain.Article
	}{
		{
			name: "计算成功-单批",
			mock: func(ctrl *gomock.Controller) (ArticleService, InteractiveService) {
				artSvc := mocks.NewMockArticleService(ctrl)
				intrSvc := mocks.NewMockInteractiveService(ctrl)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, batchSize).Return([]domain.Article{
					{ID: 1, UpdateAt: now},
					{ID: 2, UpdateAt: now},
				}, nil)
				intrSvc.EXPECT().GetByIDs(gomock.Any(), "article", []int64{1, 2}).Return(map[int64]domain.Interactive{
					1: {LikeCnt: 1},
					2: {LikeCnt: 2},
				}, nil)

				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 2, batchSize).Return([]domain.Article{}, nil)
				return artSvc, intrSvc
			},
			wantRes: []domain.Article{
				{ID: 2, UpdateAt: now},
				{ID: 1, UpdateAt: now},
			},
		},
		{
			name: "计算成功-两批次",
			mock: func(ctrl *gomock.Controller) (ArticleService, InteractiveService) {
				artSvc := mocks.NewMockArticleService(ctrl)
				intrSvc := mocks.NewMockInteractiveService(ctrl)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, batchSize).Return([]domain.Article{
					{ID: 1, UpdateAt: now},
					{ID: 2, UpdateAt: now},
				}, nil)
				intrSvc.EXPECT().GetByIDs(gomock.Any(), "article", []int64{1, 2}).Return(map[int64]domain.Interactive{
					1: {LikeCnt: 1},
					2: {LikeCnt: 2},
				}, nil)

				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 2, batchSize).Return([]domain.Article{
					{ID: 3, UpdateAt: now},
					{ID: 4, UpdateAt: now},
				}, nil)
				intrSvc.EXPECT().GetByIDs(gomock.Any(), "article", []int64{3, 4}).Return(map[int64]domain.Interactive{
					3: {LikeCnt: 3},
					4: {LikeCnt: 4},
				}, nil)

				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 4, batchSize).Return([]domain.Article{}, nil)
				return artSvc, intrSvc
			},
			wantRes: []domain.Article{
				{ID: 4, UpdateAt: now},
				{ID: 3, UpdateAt: now},
				{ID: 2, UpdateAt: now},
			},
		},
		{
			name: "计算成功-队列最小值大于分数",
			mock: func(ctrl *gomock.Controller) (ArticleService, InteractiveService) {
				artSvc := mocks.NewMockArticleService(ctrl)
				intrSvc := mocks.NewMockInteractiveService(ctrl)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, batchSize).Return([]domain.Article{
					{ID: 1, UpdateAt: now},
					{ID: 2, UpdateAt: now},
				}, nil)
				intrSvc.EXPECT().GetByIDs(gomock.Any(), "article", []int64{1, 2}).Return(map[int64]domain.Interactive{
					1: {LikeCnt: 5},
					2: {LikeCnt: 2},
				}, nil)

				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 2, batchSize).Return([]domain.Article{
					{ID: 3, UpdateAt: now},
					{ID: 4, UpdateAt: now},
				}, nil)
				intrSvc.EXPECT().GetByIDs(gomock.Any(), "article", []int64{3, 4}).Return(map[int64]domain.Interactive{
					3: {LikeCnt: 5},
					4: {LikeCnt: 1},
				}, nil)

				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 4, batchSize).Return([]domain.Article{
					{ID: 5, UpdateAt: now},
				}, nil)
				intrSvc.EXPECT().GetByIDs(gomock.Any(), "article", []int64{5}).Return(map[int64]domain.Interactive{
					5: {LikeCnt: 3},
				}, nil)

				return artSvc, intrSvc
			},
			wantRes: []domain.Article{
				{ID: 1, UpdateAt: now},
				{ID: 3, UpdateAt: now},
				{ID: 5, UpdateAt: now},
			},
		},
		{
			name: "计算成功-最小值大于分数",
			mock: func(ctrl *gomock.Controller) (ArticleService, InteractiveService) {
				artSvc := mocks.NewMockArticleService(ctrl)
				intrSvc := mocks.NewMockInteractiveService(ctrl)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, batchSize).Return([]domain.Article{
					{ID: 1, UpdateAt: now},
					{ID: 2, UpdateAt: now},
				}, nil)
				intrSvc.EXPECT().GetByIDs(gomock.Any(), "article", []int64{1, 2}).Return(map[int64]domain.Interactive{
					1: {LikeCnt: 5},
					2: {LikeCnt: 2},
				}, nil)

				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 2, batchSize).Return([]domain.Article{
					{ID: 3, UpdateAt: now},
					{ID: 4, UpdateAt: now},
				}, nil)
				intrSvc.EXPECT().GetByIDs(gomock.Any(), "article", []int64{3, 4}).Return(map[int64]domain.Interactive{
					3: {LikeCnt: 5},
					4: {LikeCnt: 1},
				}, nil)

				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 4, batchSize).Return([]domain.Article{}, nil)
				return artSvc, intrSvc
			},
			wantRes: []domain.Article{
				{ID: 1, UpdateAt: now},
				{ID: 3, UpdateAt: now},
				{ID: 2, UpdateAt: now},
			},
		},
		{
			name: "art失败",
			mock: func(ctrl *gomock.Controller) (ArticleService, InteractiveService) {
				artSvc := mocks.NewMockArticleService(ctrl)
				intrSvc := mocks.NewMockInteractiveService(ctrl)

				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, batchSize).Return([]domain.Article{
					{ID: 1, UpdateAt: now},
					{ID: 2, UpdateAt: now},
				}, nil)
				intrSvc.EXPECT().GetByIDs(gomock.Any(), "article", []int64{1, 2}).Return(map[int64]domain.Interactive{
					1: {LikeCnt: 1},
					2: {LikeCnt: 2},
				}, nil)

				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 2, batchSize).Return([]domain.Article{}, mockErr)
				return artSvc, intrSvc
			},
			wantErr: mockErr,
		},
		{
			name: "intr失败",
			mock: func(ctrl *gomock.Controller) (ArticleService, InteractiveService) {
				artSvc := mocks.NewMockArticleService(ctrl)
				intrSvc := mocks.NewMockInteractiveService(ctrl)

				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, batchSize).Return([]domain.Article{
					{ID: 1, UpdateAt: now},
					{ID: 2, UpdateAt: now},
				}, nil)
				intrSvc.EXPECT().GetByIDs(gomock.Any(), "article", []int64{1, 2}).Return(map[int64]domain.Interactive{
					1: {LikeCnt: 1},
					2: {LikeCnt: 2},
				}, nil)

				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 2, batchSize).Return([]domain.Article{
					{ID: 3, UpdateAt: now},
					{ID: 4, UpdateAt: now},
				}, nil)
				intrSvc.EXPECT().GetByIDs(gomock.Any(), "article", []int64{3, 4}).Return(map[int64]domain.Interactive{}, mockErr)
				return artSvc, intrSvc
			},
			wantErr: mockErr,
		},
		{
			name: "intr不存在",
			mock: func(ctrl *gomock.Controller) (ArticleService, InteractiveService) {
				artSvc := mocks.NewMockArticleService(ctrl)
				intrSvc := mocks.NewMockInteractiveService(ctrl)
				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 0, batchSize).Return([]domain.Article{
					{ID: 1, UpdateAt: now},
					{ID: 2, UpdateAt: now},
				}, nil)
				intrSvc.EXPECT().GetByIDs(gomock.Any(), "article", []int64{1, 2}).Return(map[int64]domain.Interactive{
					1: {LikeCnt: 1},
					2: {LikeCnt: 2},
				}, nil)

				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 2, batchSize).Return([]domain.Article{
					{ID: 3, UpdateAt: now},
					{ID: 4, UpdateAt: now},
				}, nil)
				intrSvc.EXPECT().GetByIDs(gomock.Any(), "article", []int64{3, 4}).Return(map[int64]domain.Interactive{
					3: {LikeCnt: 3},
				}, nil)

				artSvc.EXPECT().ListPub(gomock.Any(), gomock.Any(), 4, batchSize).Return([]domain.Article{}, nil)
				return artSvc, intrSvc
			},
			wantRes: []domain.Article{
				{ID: 3, UpdateAt: now},
				{ID: 2, UpdateAt: now},
				{ID: 1, UpdateAt: now},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			artSvc, intrSvc := tc.mock(ctrl)
			svc := &batchRankingService{
				artSvc:    artSvc,
				intrSvc:   intrSvc,
				BatchSize: batchSize,
				N:         3,
			}
			svc.soreFunc = svc.score
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			res, err := svc.rankTopN(ctx)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
