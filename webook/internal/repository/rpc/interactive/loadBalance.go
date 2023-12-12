package interactive

import (
	"context"
	"math/rand"

	"github.com/ecodeclub/ekit/syncx/atomicx"
	"google.golang.org/grpc"

	intr "geektime-basic-go/webook/api/proto/gen/interactive"
)

type LoadBalanceRPC struct {
	remote intr.InteractiveServiceClient
	local  *LocalRPCAdapter

	threshold *atomicx.Value[int32]
}

func NewInteractiveClient(remote intr.InteractiveServiceClient, local *LocalRPCAdapter, threshold int32) *LoadBalanceRPC {
	return &LoadBalanceRPC{remote: remote, local: local, threshold: atomicx.NewValueOf(threshold)}
}

func (rpc *LoadBalanceRPC) IncrReadCnt(ctx context.Context, in *intr.IncrReadCntRequest, opts ...grpc.CallOption) (*intr.IncrReadCntResponse, error) {
	return rpc.selectClient().IncrReadCnt(ctx, in)
}

func (rpc *LoadBalanceRPC) Like(ctx context.Context, in *intr.LikeRequest, opts ...grpc.CallOption) (*intr.LikeResponse, error) {
	return rpc.selectClient().Like(ctx, in)
}

func (rpc *LoadBalanceRPC) Collect(ctx context.Context, in *intr.CollectRequest, opts ...grpc.CallOption) (*intr.CollectResponse, error) {
	return rpc.selectClient().Collect(ctx, in)
}

func (rpc *LoadBalanceRPC) Get(ctx context.Context, in *intr.GetRequest, opts ...grpc.CallOption) (*intr.GetResponse, error) {
	return rpc.selectClient().Get(ctx, in)
}

func (rpc *LoadBalanceRPC) GetByIDs(ctx context.Context, in *intr.GetByIDsRequest, opts ...grpc.CallOption) (*intr.GetByIDsResponse, error) {
	return rpc.selectClient().GetByIDs(ctx, in)
}

func (rpc *LoadBalanceRPC) selectClient() intr.InteractiveServiceClient {
	num := rand.Int31n(100)
	if num < rpc.threshold.Load() {
		return rpc.remote
	}
	return rpc.local
}

func (rpc *LoadBalanceRPC) UpdateThreshold(threshold int32) {
	rpc.threshold.Store(threshold)
}
