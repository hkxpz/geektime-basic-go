package integration

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

	"geektime-basic-go/webook/internal/integration/startup"
	"geektime-basic-go/webook/internal/repository/dao/article"
	myjwt "geektime-basic-go/webook/internal/web/jwt"
	"geektime-basic-go/webook/ioc"
)

type ArticleHandlerTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func (s *ArticleHandlerTestSuite) SetupSuite() {
	startup.InitViper()
	gin.SetMode(gin.ReleaseMode)
	s.server = gin.New()
	s.db = ioc.InitDB()
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("user", myjwt.UserClaims{ID: 123})
	})
	startup.InitArticleHandler().RegisterRoutes(s.server)
}

func (s *ArticleHandlerTestSuite) TearDownTest() {
	err := s.db.Exec("TRUNCATE TABLE `articles`").Error
	require.NoError(s.T(), err)
	s.db.Exec("TRUNCATE TABLE `published_articles`")
}

func (s *ArticleHandlerTestSuite) TestArticleHandler_Edit() {
	t := s.T()
	testCases := []struct {
		name   string
		before func()
		after  func()

		req Article

		wantCode int
		wantRes  Result[int64]
	}{
		{
			name:   "新建帖子并发表",
			before: func() {},
			after: func() {
				var art article.Article
				s.db.Where("author_id = ?", 123).First(&art)
				assert.Equal(t, "hello, 你好", art.Title)
				assert.Equal(t, "随便试试", art.Content)
				assert.Equal(t, int64(123), art.AuthorID)
				assert.True(t, art.CreateAt > 0)
				assert.True(t, art.UpdateAt > 0)

				//var publishedArt article.PublishedArticle
				//s.db.Where("author_id = ?", 123).First(&publishedArt)
				//assert.Equal(t, "hello, 你好", publishedArt.Title)
				//assert.Equal(t, "随便试试", publishedArt.Content)
				//assert.Equal(t, int64(123), publishedArt.AuthorID)
				//assert.True(t, publishedArt.CreateAt > 0)
				//assert.True(t, publishedArt.UpdateAt > 0)
			},
			req: Article{
				Title:   "hello, 你好",
				Content: "随便试试",
			},
			wantCode: http.StatusOK,
			wantRes:  Result[int64]{Data: 1},
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
			s.server.ServeHTTP(recorder, req)

			assert.Equal(t, tc.wantCode, recorder.Code)
			var res Result[int64]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func TestArticle(t *testing.T) {
	suite.Run(t, new(ArticleHandlerTestSuite))
}

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
