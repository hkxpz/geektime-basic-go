package web

import (
	"github.com/gin-gonic/gin"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/service"
	myjwt "geektime-basic-go/webook/internal/web/jwt"
	"geektime-basic-go/webook/internal/web/middleware/handlefunc"
	"geektime-basic-go/webook/pkg/logger"
)

type ArticleHandler struct {
	svc     service.ArticleService
	logFunc handlefunc.LogFunc
}

func NewArticleHandler(svc service.ArticleService, l logger.Logger) *ArticleHandler {
	return &ArticleHandler{svc: svc, logFunc: handlefunc.DefaultLogFunc(l)}
}

func (ah *ArticleHandler) RegisterRoutes(s *gin.Engine) {
	g := s.Group("/articles")
	g.POST("/edit", handlefunc.WrapReqWithLog[ArticleReq](ah.Edit, ah.logFunc))
	g.POST("/publish", handlefunc.WrapReqWithLog[ArticleReq](ah.Publish, ah.logFunc))
}

func (ah *ArticleHandler) Edit(ctx *gin.Context, req ArticleReq, uc myjwt.UserClaims) (Response, error) {
	id, err := ah.svc.Save(ctx, req.toDomain(uc.ID))
	if err != nil {
		return InternalServerError(), err
	}

	return Response{Data: id}, nil
}

func (ah *ArticleHandler) Publish(ctx *gin.Context, req ArticleReq, uc myjwt.UserClaims) (Response, error) {
	id, err := ah.svc.Publish(ctx, req.toDomain(uc.ID))
	if err != nil {
		return InternalServerError(), err
	}

	return Response{Data: id}, nil
}

type ArticleReq struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (req *ArticleReq) toDomain(uid int64) domain.Article {
	return domain.Article{
		ID:      req.ID,
		Title:   req.Title,
		Content: req.Content,
		Author:  domain.Author{ID: uid},
	}
}
