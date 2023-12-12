package interactive

import (
	"context"

	"google.golang.org/grpc"

	intr "geektime-basic-go/webook/api/proto/gen/interactive"
	"geektime-basic-go/webook/interactive/domain"
	"geektime-basic-go/webook/interactive/service"
)

type LocalRPCAdapter struct {
	svc service.InteractiveService
}

func NewInteractiveServiceAdapter(svc service.InteractiveService) *LocalRPCAdapter {
	return &LocalRPCAdapter{svc: svc}
}

func (local *LocalRPCAdapter) IncrReadCnt(ctx context.Context, in *intr.IncrReadCntRequest, opts ...grpc.CallOption) (*intr.IncrReadCntResponse, error) {
	err := local.svc.IncrReadCnt(ctx, in.GetBiz(), in.GetBizId())
	return &intr.IncrReadCntResponse{}, err
}

func (local *LocalRPCAdapter) Get(ctx context.Context, in *intr.GetRequest, opts ...grpc.CallOption) (*intr.GetResponse, error) {
	res, err := local.svc.Get(ctx, in.GetBiz(), in.GetBizId(), in.GetUid())
	if err != nil {
		return &intr.GetResponse{}, err
	}
	return &intr.GetResponse{Intr: local.toDTO(res)}, nil
}

func (local *LocalRPCAdapter) Like(ctx context.Context, in *intr.LikeRequest, opts ...grpc.CallOption) (*intr.LikeResponse, error) {
	err := local.svc.Like(ctx, in.Biz, in.BizId, in.Uid, in.Liked)
	return &intr.LikeResponse{}, err
}

func (local *LocalRPCAdapter) Collect(ctx context.Context, in *intr.CollectRequest, opts ...grpc.CallOption) (*intr.CollectResponse, error) {
	err := local.svc.Collect(ctx, in.Biz, in.BizId, in.Cid, in.Uid)
	return &intr.CollectResponse{}, err
}

func (local *LocalRPCAdapter) GetByIDs(ctx context.Context, in *intr.GetByIDsRequest, opts ...grpc.CallOption) (*intr.GetByIDsResponse, error) {
	if len(in.Ids) == 0 {
		return &intr.GetByIDsResponse{}, nil
	}
	data, err := local.svc.GetByIDs(ctx, in.Biz, in.GetIds())
	if err != nil {
		return &intr.GetByIDsResponse{}, nil
	}
	res := make(map[int64]*intr.Interactive, len(data))
	for k, v := range data {
		res[k] = local.toDTO(v)
	}
	return &intr.GetByIDsResponse{Intrs: res}, nil
}

func (local *LocalRPCAdapter) toDTO(res domain.Interactive) *intr.Interactive {
	return &intr.Interactive{
		Biz:        res.Biz,
		BizId:      res.BizID,
		ReadCnt:    res.ReadCnt,
		LikeCnt:    res.LikeCnt,
		CollectCnt: res.CollectCnt,
		Liked:      res.Liked,
		Collected:  res.Collected,
	}
}
