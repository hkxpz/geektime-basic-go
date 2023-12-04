package client

import (
	"context"
	"math/rand"

	"github.com/ecodeclub/ekit/syncx/atomicx"
	"google.golang.org/grpc"

	intr "geektime-basic-go/webook/api/proto/gen/interactive"
)

type InteractiveClient struct {
	remote intr.InteractiveServiceClient
	local  *InteractiveLocalAdapter

	threshold *atomicx.Value[int32]
}

func NewInteractiveClient(remote intr.InteractiveServiceClient, local *InteractiveLocalAdapter, threshold int32) *InteractiveClient {
	return &InteractiveClient{remote: remote, local: local, threshold: atomicx.NewValueOf(threshold)}
}

func (i *InteractiveClient) IncrReadCnt(ctx context.Context, in *intr.IncrReadCntRequest, opts ...grpc.CallOption) (*intr.IncrReadCntResponse, error) {
	return i.selectClient().IncrReadCnt(ctx, in)
}

func (i *InteractiveClient) Like(ctx context.Context, in *intr.LikeRequest, opts ...grpc.CallOption) (*intr.LikeResponse, error) {
	return i.selectClient().Like(ctx, in)
}

func (i *InteractiveClient) Collect(ctx context.Context, in *intr.CollectRequest, opts ...grpc.CallOption) (*intr.CollectResponse, error) {
	return i.selectClient().Collect(ctx, in)
}

func (i *InteractiveClient) Get(ctx context.Context, in *intr.GetRequest, opts ...grpc.CallOption) (*intr.GetResponse, error) {
	return i.selectClient().Get(ctx, in)
}

func (i *InteractiveClient) GetByIDs(ctx context.Context, in *intr.GetByIDsRequest, opts ...grpc.CallOption) (*intr.GetByIDsResponse, error) {
	return i.selectClient().GetByIDs(ctx, in)
}

func (i *InteractiveClient) selectClient() intr.InteractiveServiceClient {
	num := rand.Int31n(100)
	if num < i.threshold.Load() {
		return i.remote
	}
	return i.local
}

func (i *InteractiveClient) UpdateThreshold(threshold int32) {
	i.threshold.Store(threshold)
}
