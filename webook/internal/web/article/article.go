package article

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"

	"geektime-basic-go/webook/internal/domain"
	"geektime-basic-go/webook/internal/service"
	hf "geektime-basic-go/webook/pkg/ginx/handlefunc"
	"geektime-basic-go/webook/pkg/logger"
)

type Handler struct {
	svc     service.ArticleService
	intrSvc service.InteractiveService
	l       logger.Logger
	biz     string
}

func NewArticleHandler(svc service.ArticleService, l logger.Logger) *Handler {
	return &Handler{svc: svc, l: l}
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
		return hf.InternalServerError(), fmt.Errorf("保存数据失败: %w", err)
	}
	return hf.Response{Data: id}, nil
}

func (ah *Handler) Publish(ctx *gin.Context, req Req, uc hf.UserClaims) (hf.Response, error) {
	id, err := ah.svc.Publish(ctx, req.toDomain(uc.ID))
	if err != nil {
		return hf.InternalServerError(), fmt.Errorf("发表失败: %w", err)
	}
	return hf.Response{Data: id}, nil
}

func (ah *Handler) Withdraw(ctx *gin.Context, req Req, uc hf.UserClaims) (hf.Response, error) {
	if err := ah.svc.Withdraw(ctx, uc.ID, req.ID); err != nil {
		ah.l.Error("设置为尽自己可见失败", logger.Error(err), logger.Field{Key: "id", Value: req.ID})
		return hf.InternalServerError(), errors.New("设置为尽自己可见失败")
	}
	return hf.Response{Msg: "OK"}, nil
}

func (ah *Handler) PubDetail(ctx *gin.Context, uc hf.UserClaims) (hf.Response, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return hf.Response{Code: 4, Msg: "参数错误"}, fmt.Errorf("查询文章详情的 ID %s 不正确, %w", idStr, err)
	}

	var (
		eg   errgroup.Group
		art  domain.Article
		intr domain.Interactive
	)

	eg.Go(func() (err error) {
		art, err = ah.svc.GetPublishedByID(ctx, id)
		return
	})
	eg.Go(func() (err error) {
		intr, err = ah.intrSvc.Get(ctx, ah.biz, id, uc.ID)
		return
	})
	if err = eg.Wait(); err != nil {
		return hf.InternalServerError(), fmt.Errorf("获取文章信息失败: %w", err)
	}

	go func() {
		if err = ah.intrSvc.IncrReadCnt(ctx, ah.biz, art.ID); err != nil {
			ah.l.Error("增加文章阅读数失败", logger.Error(err))
		}
	}()
	return hf.Response{Data: Vo{
		ID:      art.ID,
		Title:   art.Title,
		Status:  art.Status.ToUint8(),
		Content: art.Content,
		// 要把作者信息带出去
		Author:     art.Author.Name,
		CreateAt:   art.CreateAt.Format(time.DateTime),
		UpdateAt:   art.UpdateAt.Format(time.DateTime),
		ReadCnt:    intr.ReadCnt,
		CollectCnt: intr.CollectCnt,
		LikeCnt:    intr.LikeCnt,
		Liked:      intr.Liked,
		Collected:  intr.Collected,
	}}, nil
}

func (ah *Handler) Detail(ctx *gin.Context, uc hf.UserClaims) (hf.Response, error) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return hf.Response{Code: 4, Msg: "参数错误"}, fmt.Errorf("查询文章详情的 ID %s 不正确, %w", idStr, err)
	}

	art, err := ah.svc.GetByID(ctx, id)
	if err != nil {
		return hf.InternalServerError(), fmt.Errorf("获得文章信息失败: %w", err)
	}

	if art.Author.ID != uc.ID {
		return hf.Response{Code: 4, Msg: "输入错误"}, fmt.Errorf("非法访问文章，创作者 ID 不匹配, uid %d", uc.ID)
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
		return hf.InternalServerError(), fmt.Errorf("获得用户会话信息失败: %w", err)
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
	var err error
	if req.Like {
		err = ah.intrSvc.Like(ctx, ah.biz, req.ID, uc.ID)
	} else {
		err = ah.intrSvc.CancelLike(ctx, ah.biz, req.ID, uc.ID)
	}

	if err != nil {
		return hf.InternalServerError(), err
	}
	return hf.RespSuccess("OK"), nil
}

func (ah *Handler) Collect(ctx *gin.Context, req CollectReq, uc hf.UserClaims) (hf.Response, error) {
	if err := ah.intrSvc.Collect(ctx, ah.biz, req.ID, req.CID, uc.ID); err != nil {
		return hf.InternalServerError(), err
	}
	return hf.RespSuccess("OK"), nil
}
