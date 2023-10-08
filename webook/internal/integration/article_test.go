package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/integration/startup"
	"geektime-basic-go/webook/internal/repository/dao/article"
	myjwt "geektime-basic-go/webook/internal/web/jwt"
)

type ArticleReq struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func TestArticleHandler_Edit(t *testing.T) {
	startup.InitViper()
	gin.SetMode(gin.ReleaseMode)
	server := gin.New()
	db := startup.InitDB()
	server.Use(func(ctx *gin.Context) {
		ctx.Set("user", myjwt.UserClaims{ID: 123})
	})
	startup.InitArticleHandler().RegisterRoutes(server)

	testCases := []struct {
		name   string
		before func()
		after  func()

		req ArticleReq

		wantCode int
		wantRes  Response
	}{
		{
			name: "新建帖子",
			before: func() {
				err := db.Exec("TRUNCATE TABLE `articles`").Error
				require.NoError(t, err)
			},
			after: func() {
				var art article.Article
				db.Where("author_id = ?", 123).First(&art)
				assert.True(t, art.CreateAt > 0)
				assert.True(t, art.UpdateAt > 0)
				assert.Equal(t, article.Article{
					ID:       1,
					Title:    "hello, 你好",
					Content:  "随便试试",
					AuthorID: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
					CreateAt: art.CreateAt,
					UpdateAt: art.UpdateAt,
				}, art)
			},
			req: ArticleReq{
				Title:   "hello, 你好",
				Content: "随便试试",
			},
			wantCode: http.StatusOK,
			wantRes:  Response{Data: float64(1)},
		},
		{
			name: "更新文章",
			before: func() {
				err := db.Exec("TRUNCATE TABLE `articles`").Error
				require.NoError(t, err)

				db.Create(article.Article{
					ID:       1,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorID: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					CreateAt: 123,
					UpdateAt: 123,
				})
			},
			after: func() {
				var art article.Article
				db.Where("id = ?", 1).First(&art)
				assert.True(t, art.UpdateAt > 123)
				assert.Equal(t, article.Article{
					ID:       1,
					Title:    "更新标题",
					Content:  "更新内容",
					AuthorID: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
					CreateAt: 123,
					UpdateAt: art.UpdateAt,
				}, art)
			},
			req: ArticleReq{
				ID:      1,
				Title:   "更新标题",
				Content: "更新内容",
			},
			wantCode: http.StatusOK,
			wantRes:  Response{Data: float64(1)},
		},
		{
			name: "更新别人文章",
			before: func() {
				err := db.Exec("TRUNCATE TABLE `articles`").Error
				require.NoError(t, err)

				db.Create(article.Article{
					ID:       1,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorID: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					CreateAt: 123,
					UpdateAt: 123,
				})
			},
			after: func() {
				var art article.Article
				db.Where("id = ?", 1).First(&art)
				assert.Equal(t, article.Article{
					ID:       1,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorID: 789,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					CreateAt: 123,
					UpdateAt: 123,
				}, art)
			},
			req: ArticleReq{
				ID:      1,
				Title:   "更新标题",
				Content: "更新内容",
			},
			wantCode: http.StatusOK,
			wantRes:  Response{Code: 5, Msg: "系统错误"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before()
			defer tc.after()

			data, err := json.Marshal(tc.req)
			require.NoError(t, err)
			req := reqBuilder(t, http.MethodPost, "/articles/edit", data)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantCode, recorder.Code)
			var res Response
			err = json.NewDecoder(recorder.Body).Decode(&res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func TestArticleHandler_Publish(t *testing.T) {
	startup.InitViper()
	gin.SetMode(gin.ReleaseMode)
	server := gin.New()
	db := startup.InitDB()
	server.Use(func(ctx *gin.Context) {
		ctx.Set("user", myjwt.UserClaims{ID: 123})
	})
	startup.InitArticleHandler().RegisterRoutes(server)

	testCases := []struct {
		name   string
		before func()
		after  func()

		req ArticleReq

		wantCode int
		wantRes  Response
	}{
		{
			name: "新建帖子并发表成功",
			before: func() {
				err := db.Exec("TRUNCATE TABLE `articles`").Error
				require.NoError(t, err)
				err = db.Exec("TRUNCATE TABLE `published_articles`").Error
				require.NoError(t, err)
			},
			after: func() {
				var art article.Article
				db.Where("author_id = ?", 123).First(&art)
				assert.True(t, art.CreateAt > 0)
				assert.True(t, art.UpdateAt > 0)
				resArt := article.Article{
					ID:       1,
					Title:    "hello, 你好",
					Content:  "随便试试",
					AuthorID: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					CreateAt: art.CreateAt,
					UpdateAt: art.UpdateAt,
				}
				assert.Equal(t, resArt, art)

				var publishedArt article.PublishedArticle
				db.Where("author_id = ?", 123).First(&publishedArt)
				assert.True(t, art.CreateAt > 0)
				assert.True(t, art.UpdateAt > 0)
				resPublishedArt := article.PublishedArticle(resArt)
				resPublishedArt.CreateAt = publishedArt.CreateAt
				resPublishedArt.UpdateAt = publishedArt.UpdateAt
				assert.Equal(t, resPublishedArt, publishedArt)
			},
			req: ArticleReq{
				Title:   "hello, 你好",
				Content: "随便试试",
			},
			wantCode: http.StatusOK,
			wantRes:  Response{Data: float64(1)},
		},
		{
			name: "更新帖子并发表成功",
			before: func() {
				err := db.Exec("TRUNCATE TABLE `articles`").Error
				require.NoError(t, err)
				err = db.Exec("TRUNCATE TABLE `published_articles`").Error
				require.NoError(t, err)

				art := article.Article{
					ID:       1,
					Title:    "hello, 你好",
					Content:  "随便试试",
					AuthorID: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					CreateAt: 123,
					UpdateAt: 123,
				}
				require.NoError(t, db.Create(art).Error)
				require.NoError(t, db.Create(article.PublishedArticle(art)).Error)
			},
			after: func() {
				var art article.Article
				db.Model(art).Where("author_id = ?", 123).First(&art)
				assert.True(t, art.CreateAt > 0)
				assert.True(t, art.UpdateAt > 0)
				resArt := article.Article{
					ID:       1,
					Title:    "更新啦",
					Content:  "更新啦",
					AuthorID: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					CreateAt: art.CreateAt,
					UpdateAt: art.UpdateAt,
				}
				assert.Equal(t, resArt, art)

				var publishedArt article.PublishedArticle
				db.Model(publishedArt).Where("author_id = ?", 123).First(&publishedArt)
				assert.True(t, art.CreateAt > 0)
				assert.True(t, art.UpdateAt > 0)
				resPublishedArt := article.PublishedArticle(resArt)
				resPublishedArt.CreateAt = publishedArt.CreateAt
				resPublishedArt.UpdateAt = publishedArt.UpdateAt
				assert.Equal(t, resPublishedArt, publishedArt)
			},
			req: ArticleReq{
				ID:      1,
				Title:   "更新啦",
				Content: "更新啦",
			},
			wantCode: http.StatusOK,
			wantRes:  Response{Data: float64(1)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before()
			defer tc.after()

			data, err := json.Marshal(tc.req)
			require.NoError(t, err)
			req := reqBuilder(t, http.MethodPost, "/articles/publish", data)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantCode, recorder.Code)
			var res Response
			err = json.NewDecoder(recorder.Body).Decode(&res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
