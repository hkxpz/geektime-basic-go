package article

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/errs"
	"geektime-basic-go/webook/internal/service"
	hf "geektime-basic-go/webook/pkg/ginx/handlefunc"
	"geektime-basic-go/webook/pkg/logger"
)

type Handler struct {
	svc service.ArticleService
	l   logger.Logger
	biz string
}

func NewArticleHandler(svc service.ArticleService, l logger.Logger) *Handler {
	return &Handler{svc: svc, l: l, biz: "article"}
}

func (ah *Handler) RegisterRoutes(s *gin.Engine) {
	g := s.Group("/articles")
	g.GET("/detail/:id", hf.WrapClaims(ah.Detail))
	g.POST("/list", hf.WrapClaimsAndReq[LimitReq](ah.List))

	g.POST("/edit", hf.WrapClaimsAndReq[Req](ah.Edit))
	g.POST("/publish", hf.WrapClaimsAndReq[Req](ah.Publish))
	g.POST("/withdraw", hf.WrapClaimsAndReq[Req](ah.Withdraw))

	pub := g.Group("/pub")
	pub.GET("/:id", hf.WrapClaims(ah.PubDetail))
	pub.POST("/like", hf.WrapClaimsAndReq[LikeReq](ah.Like))
	pub.POST("/collect", hf.WrapClaimsAndReq[CollectReq](ah.Collect))
}

func (ah *Handler) Edit(ctx *gin.Context, req Req, uc hf.UserClaims) (hf.Response, error) {
	id, err := ah.svc.Save(ctx, req.toDomain(uc.ID))
	if err != nil {
		return hf.InternalServerErrorWith(errs.ArticleInternalServerError), fmt.Errorf("保存数据失败: %w", err)
	}
	return hf.Response{Data: id}, nil
}

func (ah *Handler) Publish(ctx *gin.Context, req Req, uc hf.UserClaims) (hf.Response, error) {
	id, err := ah.svc.Publish(ctx, req.toDomain(uc.ID))
	if err != nil {
		return hf.InternalServerErrorWith(errs.ArticleInternalServerError), fmt.Errorf("发表失败: %w", err)
	}
	return hf.Response{Data: id}, nil
}

func (ah *Handler) Withdraw(ctx *gin.Context, req Req, uc hf.UserClaims) (hf.Response, error) {
	if err := ah.svc.Withdraw(ctx, uc.ID, req.ID); err != nil {
		ah.l.Error("设置为尽自己可见失败", logger.Error(err), logger.Field{Key: "id", Value: req.ID})
		return hf.InternalServerErrorWith(errs.ArticleInternalServerError), errors.New("设置为尽自己可见失败")
	}
	return hf.Response{Msg: "OK"}, nil
}

func (ah *Handler) PubDetail(ctx *gin.Context, uc hf.UserClaims) (hf.Response, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return hf.Response{Code: errs.ArticleInvalidInput, Msg: "参数错误"}, fmt.Errorf("查询文章详情的 ID %s 不正确, %w", idStr, err)
	}

	art, err := ah.svc.PubDetail(ctx, id, uc.ID)
	if err != nil {
		return hf.InternalServerErrorWith(errs.ArticleInternalServerError), fmt.Errorf("获取文章信息失败: %w", err)
	}
	return hf.Response{Data: art}, nil

}

func (ah *Handler) Detail(ctx *gin.Context, uc hf.UserClaims) (hf.Response, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return hf.Response{Code: errs.ArticleInvalidInput, Msg: "参数错误"}, fmt.Errorf("查询文章详情的 ID %s 不正确, %w", idStr, err)
	}

	art, err := ah.svc.GetByID(ctx, id)
	if err != nil {
		return hf.InternalServerErrorWith(errs.ArticleInternalServerError), fmt.Errorf("获得文章信息失败: %w", err)
	}

	if art.Author.ID != uc.ID {
		return hf.Response{Code: errs.ArticleInvalidInput, Msg: "输入错误"}, fmt.Errorf("非法访问文章，创作者 ID 不匹配, uid %d", uc.ID)
	}
	return hf.Response{Data: Vo{
		ID:       art.ID,
		Title:    art.Title,
		Status:   art.Status.ToUint8(),
		Content:  art.Content,
		CreateAt: art.CreateAt.Format(time.DateTime),
		UpdateAt: art.UpdateAt.Format(time.DateTime),
	}}, nil
}

func (ah *Handler) List(ctx *gin.Context, req LimitReq, uc hf.UserClaims) (hf.Response, error) {
	if req.Limit > 100 {
		return hf.BadRequestError("请求错误"), fmt.Errorf("分页过大 %d", req.Limit)
	}

	arts, err := ah.svc.List(ctx, uc.ID, req.Offset, req.Limit)
	if err != nil {
		return hf.InternalServerErrorWith(errs.ArticleInternalServerError), fmt.Errorf("获得用户会话信息失败: %w", err)
	}

	fn := func(idx int, src domain.Article) Vo {
		return Vo{
			ID:       src.ID,
			Title:    src.Title,
			Abstract: src.Abstract(),
			Status:   src.Status.ToUint8(),
			CreateAt: src.CreateAt.Format(time.DateTime),
			UpdateAt: src.UpdateAt.Format(time.DateTime),
		}
	}
	return hf.Response{Data: slice.Map[domain.Article, Vo](arts, fn)}, nil
}

func (ah *Handler) Like(ctx *gin.Context, req LikeReq, uc hf.UserClaims) (hf.Response, error) {
	if err := ah.svc.Like(ctx.Request.Context(), ah.biz, req.ID, uc.ID, req.Like); err != nil {
		return hf.InternalServerErrorWith(errs.ArticleInternalServerError), err
	}
	return hf.RespSuccess("OK"), nil
}

func (ah *Handler) Collect(ctx *gin.Context, req CollectReq, uc hf.UserClaims) (hf.Response, error) {
	if err := ah.svc.Collect(ctx.Request.Context(), ah.biz, req.ID, req.Cid, uc.ID); err != nil {
		return hf.InternalServerErrorWith(errs.ArticleInternalServerError), err
	}
	return hf.RespSuccess("OK"), nil
}
