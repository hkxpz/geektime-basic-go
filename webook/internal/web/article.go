package web

import (
	"net/http"

	"github.com/gin-gonic/gin"

	myjwt "geektime-basic-go/webook/internal/web/jwt"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/service"
	"geektime-basic-go/webook/pkg/logger"
)

type ArticleHandler struct {
	svc service.ArticleService
	l   logger.Logger
}

func NewArticleHandler(svc service.ArticleService, l logger.Logger) *ArticleHandler {
	return &ArticleHandler{svc: svc, l: l}
}

func (ah *ArticleHandler) RegisterRoutes(s *gin.Engine) {
	g := s.Group("/articles")
	g.POST("/edit", ah.Edit)
	g.POST("/publish", ah.Publish)
}

func (ah *ArticleHandler) Edit(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		ah.l.Error("反序列化请求失败", logger.Error(err))
		return
	}

	uc, ok := ctx.MustGet("user").(myjwt.UserClaims)
	if !ok {
		ah.l.Error("获得用户会话信息失败")
		ctx.JSON(http.StatusOK, InternalServerError())
		return
	}

	id, err := ah.svc.Save(ctx, req.toDomain(uc.ID))
	if err != nil {
		ah.l.Error("保存数据失败", logger.Error(err))
		ctx.JSON(http.StatusOK, InternalServerError())
	}

	ctx.JSON(http.StatusOK, Response{Data: id})
}

func (ah *ArticleHandler) Publish(ctx *gin.Context) {
	var req ArticleReq
	if err := ctx.Bind(&req); err != nil {
		ah.l.Error("反序列化请求失败", logger.Error(err))
		return
	}

	uc, ok := ctx.MustGet("user").(myjwt.UserClaims)
	if !ok {
		ah.l.Error("获得用户会话信息失败")
		ctx.JSON(http.StatusOK, InternalServerError())
		return
	}

	id, err := ah.svc.Publish(ctx, req.toDomain(uc.ID))
	if err != nil {
		ah.l.Error("发表失败", logger.Error(err))
		ctx.JSON(http.StatusOK, InternalServerError())
	}

	ctx.JSON(http.StatusOK, Response{Data: id})
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
