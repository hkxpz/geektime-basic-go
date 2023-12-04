package grpc

import (
	"context"

	"google.golang.org/grpc"

	intr "geektime-basic-go/webook/api/proto/gen/interactive"
	"geektime-basic-go/webook/interactive/domain"
	"geektime-basic-go/webook/interactive/service"
)

type InteractiveServiceServer struct {
	intr.UnimplementedInteractiveServiceServer
	svc service.InteractiveService
}

func NewInteractiveServiceServer(svc service.InteractiveService) *InteractiveServiceServer {
	return &InteractiveServiceServer{svc: svc}
}

func (i *InteractiveServiceServer) Register(server grpc.ServiceRegistrar) {
	intr.RegisterInteractiveServiceServer(server, i)
}

func (i *InteractiveServiceServer) IncrReadCnt(ctx context.Context, request *intr.IncrReadCntRequest) (*intr.IncrReadCntResponse, error) {
	err := i.svc.IncrReadCnt(ctx, request.Biz, request.BizId)
	return &intr.IncrReadCntResponse{}, err
}

func (i *InteractiveServiceServer) Get(ctx context.Context, request *intr.GetRequest) (*intr.GetResponse, error) {
	res, err := i.svc.Get(ctx, request.GetBiz(), request.BizId, request.Uid)
	if err != nil {
		return &intr.GetResponse{}, err
	}
	return &intr.GetResponse{Intr: i.toDTO(res)}, err
}

func (i *InteractiveServiceServer) Like(ctx context.Context, request *intr.LikeRequest) (*intr.LikeResponse, error) {
	err := i.svc.Like(ctx, request.Biz, request.BizId, request.Uid, request.Liked)
	return &intr.LikeResponse{}, err
}

func (i *InteractiveServiceServer) Collect(ctx context.Context, request *intr.CollectRequest) (*intr.CollectResponse, error) {
	err := i.svc.Collect(ctx, request.Biz, request.BizId, request.Cid, request.Uid)
	return &intr.CollectResponse{}, err
}

func (i *InteractiveServiceServer) GetByIDs(ctx context.Context, request *intr.GetByIDsRequest) (*intr.GetByIDsResponse, error) {
	if len(request.Ids) == 0 {
		return &intr.GetByIDsResponse{}, nil
	}
	data, err := i.svc.GetByIDs(ctx, request.Biz, request.GetIds())
	if err != nil {
		return &intr.GetByIDsResponse{}, err
	}
	res := make(map[int64]*intr.Interactive, len(data))
	for k, v := range data {
		res[k] = i.toDTO(v)
	}
	return &intr.GetByIDsResponse{Intrs: res}, err
}

func (i *InteractiveServiceServer) toDTO(res domain.Interactive) *intr.Interactive {
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
