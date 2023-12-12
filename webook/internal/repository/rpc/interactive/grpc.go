package interactive

import (
	"context"

	"google.golang.org/grpc"

	intr "geektime-basic-go/webook/api/proto/gen/interactive"
)

type grpcInteractiveRPC struct {
	rpc intr.InteractiveServiceClient
}

func NewGRPCInteractiveRPC(rpc intr.InteractiveServiceClient) intr.InteractiveServiceClient {
	return &grpcInteractiveRPC{rpc: rpc}
}

func (g *grpcInteractiveRPC) IncrReadCnt(ctx context.Context, in *intr.IncrReadCntRequest, opts ...grpc.CallOption) (*intr.IncrReadCntResponse, error) {
	return g.rpc.IncrReadCnt(ctx, in)
}

func (g *grpcInteractiveRPC) Like(ctx context.Context, in *intr.LikeRequest, opts ...grpc.CallOption) (*intr.LikeResponse, error) {
	return g.rpc.Like(ctx, in)
}

func (g *grpcInteractiveRPC) Collect(ctx context.Context, in *intr.CollectRequest, opts ...grpc.CallOption) (*intr.CollectResponse, error) {
	return g.rpc.Collect(ctx, in)
}

func (g *grpcInteractiveRPC) Get(ctx context.Context, in *intr.GetRequest, opts ...grpc.CallOption) (*intr.GetResponse, error) {
	return g.rpc.Get(ctx, in)
}

func (g *grpcInteractiveRPC) GetByIDs(ctx context.Context, in *intr.GetByIDsRequest, opts ...grpc.CallOption) (*intr.GetByIDsResponse, error) {
	return g.rpc.GetByIDs(ctx, in)
}
