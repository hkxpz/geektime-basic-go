package article

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/integration"
	"geektime-basic-go/webook/internal/integration/startup"
	"geektime-basic-go/webook/internal/repository/dao/article"
	myjwt "geektime-basic-go/webook/internal/web/jwt"
	"geektime-basic-go/webook/ioc"
)

type ArticleMongoHandlerTestSuite struct {
	suite.Suite
	server  *gin.Engine
	mdb     *mongo.Database
	col     *mongo.Collection
	liveCol *mongo.Collection
}

func TestMongoArticle(t *testing.T) {
	suite.Run(t, new(ArticleMongoHandlerTestSuite))
}

func (s *ArticleMongoHandlerTestSuite) SetupSuite() {
	gin.SetMode(gin.ReleaseMode)
	s.server = gin.New()
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("user", myjwt.UserClaims{ID: 123})
	})
	startup.InitViper()
	node, err := snowflake.NewNode(1)
	require.NoError(s.T(), err)
	s.mdb = ioc.InitMongoDB()
	err = article.InitCollections(s.mdb)
	require.NoError(s.T(), err)
	s.col = s.mdb.Collection("articles")
	s.liveCol = s.mdb.Collection("published_articles")
	ah := startup.InitArticleHandler(article.NewMongoDBDAO(s.mdb, node))
	ah.RegisterRoutes(s.server)
}

func (s *ArticleMongoHandlerTestSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err := s.col.DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
	_, err = s.liveCol.DeleteMany(ctx, bson.D{})
	assert.NoError(s.T(), err)
}

func (s *ArticleMongoHandlerTestSuite) TestArticleMongoHandler_Edit() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)
		req    Article

		wantCode int
		wantRes  integration.Response
	}{
		{
			name:   "新建贴子",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				var art article.Article
				err := s.col.FindOne(ctx, bson.M{"author_id": 123}).Decode(&art)
				require.NoError(t, err)
				assert.True(t, art.CreateAt > 0)
				assert.True(t, art.UpdateAt > 0)
				assert.True(t, art.ID > 0)
				assert.Equal(t, article.Article{
					ID:       art.ID,
					Title:    "hello，你好",
					Content:  "随便试试",
					AuthorID: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
					CreateAt: art.CreateAt,
					UpdateAt: art.UpdateAt,
				}, art)
			},
			req: Article{
				Title:   "hello，你好",
				Content: "随便试试",
			},
			wantCode: http.StatusOK,
			wantRes:  integration.Response{Data: 1},
		},
		{
			name: "更新帖子",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				_, err := s.col.InsertOne(ctx, &article.Article{
					ID:       1,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorID: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					CreateAt: 111,
					UpdateAt: 111,
				})
				require.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				var art article.Article
				err := s.col.FindOne(ctx, bson.M{"id": 1}).Decode(&art)
				require.NoError(t, err)
				require.True(t, art.UpdateAt > 111)
				assert.Equal(t, article.Article{
					ID:       1,
					Title:    "hello，你好",
					Content:  "随便试试",
					AuthorID: 123,
					Status:   domain.ArticleStatusUnpublished.ToUint8(),
					CreateAt: 111,
					UpdateAt: art.UpdateAt,
				}, art)
			},
			req: Article{
				ID:      1,
				Title:   "hello，你好",
				Content: "随便试试",
			},
			wantCode: http.StatusOK,
			wantRes:  integration.Response{Data: 1},
		},
		{
			name: "更新别人帖子",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				_, err := s.col.InsertOne(ctx, &article.Article{
					ID:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorID: 456,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					CreateAt: 111,
					UpdateAt: 111,
				})
				require.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				var art article.Article
				err := s.col.FindOne(ctx, bson.M{"id": 2}).Decode(&art)
				require.NoError(t, err)
				assert.Equal(t, article.Article{
					ID:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorID: 456,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					CreateAt: 111,
					UpdateAt: 111,
				}, art)
			},
			req: Article{
				ID:      2,
				Title:   "hello，你好",
				Content: "随便试试",
			},
			wantCode: http.StatusOK,
			wantRes:  integration.Response{Code: 5, Msg: "系统错误"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			data, err := json.Marshal(tc.req)
			require.NoError(t, err)

			req := integration.ReqBuilder(t, http.MethodPost, "/articles/edit", data)
			recorder := httptest.NewRecorder()

			s.server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			require.True(t, recorder.Code == http.StatusOK)

			var res integration.Response
			err = json.NewDecoder(recorder.Body).Decode(&res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes.Code, res.Code)
			if res.Code == 0 && tc.wantRes.Data.(int) > 0 {
				assert.True(t, res.Data.(float64) > 0)
			}

		})
	}
}

func (s *ArticleMongoHandlerTestSuite) TestArticle_Publish() {
	t := s.T()
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)
		req    Article

		wantCode int
		wantRes  integration.Response
	}{
		{
			name:   "新建帖子并发表",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				wantArt := article.Article{
					Title:    "hello，你好",
					Content:  "随便试试",
					AuthorID: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
				}
				var art article.Article
				err := s.col.FindOne(ctx, bson.M{"author_id": 123}).Decode(&art)
				require.NoError(t, err)
				assert.True(t, art.CreateAt > 0)
				assert.True(t, art.UpdateAt > 0)
				assert.True(t, art.ID > 0)
				wantArt.ID = art.ID
				wantArt.CreateAt = art.CreateAt
				wantArt.UpdateAt = art.UpdateAt
				assert.Equal(t, wantArt, art)

				var publishedArt article.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.M{"author_id": 123}).Decode(&publishedArt)
				assert.True(t, art.CreateAt > 0)
				assert.True(t, art.UpdateAt > 0)
				assert.True(t, art.ID > 0)
				wantArt.ID = publishedArt.ID
				wantArt.CreateAt = publishedArt.CreateAt
				wantArt.UpdateAt = publishedArt.UpdateAt
				assert.Equal(t, article.PublishedArticle(wantArt), publishedArt)
			},
			req: Article{
				Title:   "hello，你好",
				Content: "随便试试",
			},
			wantCode: http.StatusOK,
			wantRes:  integration.Response{Data: 1},
		},
		{
			name: "更新帖子并新发表",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				art := article.Article{
					ID:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorID: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					CreateAt: 111,
					UpdateAt: 111,
				}
				_, err := s.col.InsertOne(ctx, &art)
				require.NoError(t, err)
				part := article.PublishedArticle(art)
				_, err = s.liveCol.InsertOne(ctx, &part)
				require.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				wantArt := article.Article{
					ID:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorID: 123,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					CreateAt: 111,
				}
				var art article.Article
				err := s.col.FindOne(ctx, bson.M{"id": 2}).Decode(&art)
				require.NoError(t, err)
				assert.True(t, art.UpdateAt > 0)
				wantArt.UpdateAt = art.UpdateAt
				assert.Equal(t, wantArt, art)

				var publishedArt article.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.M{"id": 2}).Decode(&publishedArt)
				assert.True(t, art.UpdateAt > 0)
				wantArt.UpdateAt = publishedArt.UpdateAt
				assert.Equal(t, article.PublishedArticle(wantArt), publishedArt)
			},
			req: Article{
				ID:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes:  integration.Response{Data: 1},
		},
		{
			name: "更新别人帖子",
			before: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				art := article.Article{
					ID:       3,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorID: 456,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					CreateAt: 111,
					UpdateAt: 111,
				}
				_, err := s.col.InsertOne(ctx, &art)
				require.NoError(t, err)
				part := article.PublishedArticle(art)
				_, err = s.liveCol.InsertOne(ctx, &part)
				require.NoError(t, err)
			},
			after: func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()

				wantArt := article.Article{
					ID:       3,
					Title:    "新的标题",
					Content:  "新的内容",
					AuthorID: 456,
					Status:   domain.ArticleStatusPublished.ToUint8(),
					CreateAt: 111,
					UpdateAt: 111,
				}
				var art article.Article
				err := s.col.FindOne(ctx, bson.M{"id": 3}).Decode(&art)
				require.NoError(t, err)
				assert.Equal(t, wantArt, art)

				var publishedArt article.PublishedArticle
				err = s.liveCol.FindOne(ctx, bson.M{"id": 3}).Decode(&publishedArt)
				assert.Equal(t, article.PublishedArticle(wantArt), publishedArt)
			},
			req: Article{
				ID:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantRes:  integration.Response{Data: 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			data, err := json.Marshal(tc.req)
			require.NoError(t, err)

			req := integration.ReqBuilder(t, http.MethodPost, "/articles/publish", data)
			recorder := httptest.NewRecorder()

			s.server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			require.True(t, recorder.Code == http.StatusOK)

			var res integration.Response
			err = json.NewDecoder(recorder.Body).Decode(&res)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes.Code, res.Code)
			if res.Code == 0 && tc.wantRes.Data.(int) > 0 {
				assert.True(t, res.Data.(float64) > 0)
			}

		})
	}
}
