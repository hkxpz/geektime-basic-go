package article

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/integration"
	"geektime-basic-go/webook/internal/integration/startup"
	"geektime-basic-go/webook/internal/repository/dao/article"
	myjwt "geektime-basic-go/webook/internal/web/jwt"
)

type ArticleGORMHandlerTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func TestGORMArticle(t *testing.T) {
	suite.Run(t, new(ArticleGORMHandlerTestSuite))
}

func (s *ArticleGORMHandlerTestSuite) SetupSuite() {
	startup.InitViper()
	gin.SetMode(gin.ReleaseMode)
	s.server = gin.New()
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("user", myjwt.UserClaims{ID: 123})
	})
	s.db = startup.InitDB()
	ah := startup.InitArticleHandler(article.NewGormArticleDAO(s.db))
	ah.RegisterRoutes(s.server)
}

func (s *ArticleGORMHandlerTestSuite) TearDownTest() {
	err := s.db.Exec("TRUNCATE TABLE `articles`").Error
	assert.NoError(s.T(), err)
	err = s.db.Exec("TRUNCATE TABLE `published_articles`").Error
	assert.NoError(s.T(), err)
}

func (s *ArticleGORMHandlerTestSuite) TestArticleHandler_Edit() {
	t := s.T()
	testCases := []struct {
		name   string
		before func()
		after  func()

		req Article

		wantCode int
		wantRes  integration.Response
	}{
		{
			name:   "新建帖子",
			before: func() {},
			after: func() {
				var art article.Article
				s.db.Where("author_id = ?", 123).First(&art)
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
			req: Article{
				Title:   "hello, 你好",
				Content: "随便试试",
			},
			wantCode: http.StatusOK,
			wantRes:  integration.Response{Data: float64(1)},
		},
		{
			name: "更新文章",
			before: func() {
				err := s.db.Exec("TRUNCATE TABLE `articles`").Error
				require.NoError(t, err)

				s.db.Create(article.Article{
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
				s.db.Where("id = ?", 1).First(&art)
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
			req: Article{
				ID:      1,
				Title:   "更新标题",
				Content: "更新内容",
			},
			wantCode: http.StatusOK,
			wantRes:  integration.Response{Data: float64(1)},
		},
		{
			name: "更新别人文章",
			before: func() {
				err := s.db.Exec("TRUNCATE TABLE `articles`").Error
				require.NoError(t, err)

				s.db.Create(article.Article{
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
				s.db.Where("id = ?", 1).First(&art)
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
			req: Article{
				ID:      1,
				Title:   "更新标题",
				Content: "更新内容",
			},
			wantCode: http.StatusOK,
			wantRes:  integration.Response{Code: 5, Msg: "系统错误"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before()
			defer tc.after()

			data, err := json.Marshal(tc.req)
			require.NoError(t, err)
			req := integration.ReqBuilder(t, http.MethodPost, "/articles/edit", data)
			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantCode, recorder.Code)
			var res integration.Response
			err = json.NewDecoder(recorder.Body).Decode(&res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func (s *ArticleGORMHandlerTestSuite) TestArticleHandler_Publish() {
	t := s.T()
	startup.InitViper()
	gin.SetMode(gin.ReleaseMode)
	s.server = gin.New()
	db := startup.InitDB()
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("user", myjwt.UserClaims{ID: 123})
	})
	startup.InitArticleHandler(article.NewGormArticleDAO(db)).RegisterRoutes(s.server)

	testCases := []struct {
		name   string
		before func()
		after  func()

		req Article

		wantCode int
		wantRes  integration.Response
	}{
		{
			name: "新建帖子并发表成功",
			before: func() {
				err := s.db.Exec("TRUNCATE TABLE `articles`").Error
				require.NoError(t, err)
				err = s.db.Exec("TRUNCATE TABLE `published_articles`").Error
				require.NoError(t, err)
			},
			after: func() {
				var art article.Article
				s.db.Where("author_id = ?", 123).First(&art)
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
				s.db.Where("author_id = ?", 123).First(&publishedArt)
				assert.True(t, art.CreateAt > 0)
				assert.True(t, art.UpdateAt > 0)
				resPublishedArt := article.PublishedArticle(resArt)
				resPublishedArt.CreateAt = publishedArt.CreateAt
				resPublishedArt.UpdateAt = publishedArt.UpdateAt
				assert.Equal(t, resPublishedArt, publishedArt)
			},
			req: Article{
				Title:   "hello, 你好",
				Content: "随便试试",
			},
			wantCode: http.StatusOK,
			wantRes:  integration.Response{Data: float64(1)},
		},
		{
			name: "更新帖子并发表成功",
			before: func() {
				err := s.db.Exec("TRUNCATE TABLE `articles`").Error
				require.NoError(t, err)
				err = s.db.Exec("TRUNCATE TABLE `published_articles`").Error
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
				require.NoError(t, s.db.Create(art).Error)
				require.NoError(t, s.db.Create(article.PublishedArticle(art)).Error)
			},
			after: func() {
				var art article.Article
				s.db.Model(art).Where("author_id = ?", 123).First(&art)
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
				s.db.Model(publishedArt).Where("author_id = ?", 123).First(&publishedArt)
				assert.True(t, art.CreateAt > 0)
				assert.True(t, art.UpdateAt > 0)
				resPublishedArt := article.PublishedArticle(resArt)
				resPublishedArt.CreateAt = publishedArt.CreateAt
				resPublishedArt.UpdateAt = publishedArt.UpdateAt
				assert.Equal(t, resPublishedArt, publishedArt)
			},
			req: Article{
				ID:      1,
				Title:   "更新啦",
				Content: "更新啦",
			},
			wantCode: http.StatusOK,
			wantRes:  integration.Response{Data: float64(1)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before()
			defer tc.after()

			data, err := json.Marshal(tc.req)
			require.NoError(t, err)
			req := integration.ReqBuilder(t, http.MethodPost, "/articles/publish", data)
			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantCode, recorder.Code)
			var res integration.Response
			err = json.NewDecoder(recorder.Body).Decode(&res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
