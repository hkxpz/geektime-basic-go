package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"geektime-basic-go/webook/pkg/logger"

	"geektime-basic-go/webook/internal/domain"

	"geektime-basic-go/webook/internal/service/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"geektime-basic-go/webook/internal/service"
	myjwt "geektime-basic-go/webook/internal/web/jwt"
)

func TestArticleHandler_Edit(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) service.ArticleService
		reqBody []byte

		wantCode int
		wantRes  Response
	}{
		{
			name: "新建帖子",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := mocks.NewMockArticleService(ctrl)
				svc.EXPECT().Save(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author:  domain.Author{ID: 123},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody:  []byte(`{"title":"我的标题","content":"我的内容"}`),
			wantCode: http.StatusOK,
			wantRes:  Response{Data: float64(1)},
		},
		{
			name: "更新文章",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := mocks.NewMockArticleService(ctrl)
				svc.EXPECT().Save(gomock.Any(), domain.Article{
					ID:      1,
					Title:   "我的标题",
					Content: "我的内容",
					Author:  domain.Author{ID: 123},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody:  []byte(`{"id":1,"title":"我的标题","content":"我的内容"}`),
			wantCode: http.StatusOK,
			wantRes:  Response{Data: float64(1)},
		},
		{
			name: "更新别人文章",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := mocks.NewMockArticleService(ctrl)
				svc.EXPECT().Save(gomock.Any(), domain.Article{
					ID:      1,
					Title:   "我的标题",
					Content: "我的内容",
					Author:  domain.Author{ID: 123},
				}).Return(int64(0), errors.New("模拟失败"))
				return svc
			},
			reqBody:  []byte(`{"id":1,"title":"我的标题","content":"我的内容"}`),
			wantCode: http.StatusOK,
			wantRes:  Response{Code: 5, Msg: "系统错误"},
		},
		{
			name: "Bind错误",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				return nil
			},
			reqBody:  []byte(`{"id":1,"title""我的标题","content":"我的内容"}`),
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.New()
			gin.SetMode(gin.ReleaseMode)
			server.Use(func(ctx *gin.Context) {
				ctx.Set("user", myjwt.UserClaims{ID: 123})
			})
			uh := NewArticleHandler(tc.mock(ctrl), logger.NewZapLogger(zap.NewNop()))
			uh.RegisterRoutes(server)

			req := reqBuilder(t, http.MethodPost, "/articles/edit", bytes.NewBuffer(tc.reqBody))
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				return
			}
			var webRes Response
			err := json.NewDecoder(recorder.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantRes, webRes)
		})
	}

}

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) service.ArticleService
		reqBody []byte

		wantCode int
		wantRes  Response
	}{
		{
			name: "新建立刻发表成功",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := mocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author:  domain.Author{ID: 123},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody:  []byte(`{"title":"我的标题","content":"我的内容"}`),
			wantCode: http.StatusOK,
			wantRes:  Response{Data: float64(1)},
		},
		{
			name: "已有帖子发表成功",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := mocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					ID:      1,
					Title:   "我的标题",
					Content: "我的内容",
					Author:  domain.Author{ID: 123},
				}).Return(int64(1), nil)
				return svc
			},
			reqBody:  []byte(`{"ID":1,"title":"我的标题","content":"我的内容"}`),
			wantCode: http.StatusOK,
			wantRes:  Response{Data: float64(1)},
		},
		{
			name: "发表失败",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := mocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author:  domain.Author{ID: 123},
				}).Return(int64(0), errors.New("模拟错误"))
				return svc
			},
			reqBody:  []byte(`{"title":"我的标题","content":"我的内容"}`),
			wantCode: http.StatusOK,
			wantRes:  Response{Code: 5, Msg: "系统错误"},
		},
		{
			name: "Bind错误",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				return nil
			},
			reqBody:  []byte(`{"title:"我的标题","content":"我的内容"}`),
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := gin.New()
			gin.SetMode(gin.ReleaseMode)
			server.Use(func(ctx *gin.Context) {
				ctx.Set("user", myjwt.UserClaims{ID: 123})
			})
			uh := NewArticleHandler(tc.mock(ctrl), logger.NewZapLogger(zap.NewNop()))
			uh.RegisterRoutes(server)

			req := reqBuilder(t, http.MethodPost, "/articles/publish", bytes.NewBuffer(tc.reqBody))
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				return
			}
			var webRes Response
			err := json.NewDecoder(recorder.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantCode, recorder.Code)
			assert.Equal(t, tc.wantRes, webRes)
		})
	}
}
