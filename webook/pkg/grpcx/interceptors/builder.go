package interceptors

import (
	"context"
	"net"
	"strings"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type Builder struct{}

func NewBuilder() Builder { return Builder{} }

// PeerName 获取对端应用名称
func (b *Builder) PeerName(ctx context.Context) string {
	return b.grpcHeaderValue(ctx, "app")
}

// PeerIP 获取对端ip
func (b *Builder) PeerIP(ctx context.Context) string {
	// 如果 ctx 里传入, 或者说客户端里面设置了, 直接使用设置的
	// 有的时候你经过网关子类的东西, 就需要在客户端主动设置, 防止拿到网关的 IP
	clientIP := b.grpcHeaderValue(ctx, "client-ip")
	if clientIP != "" {
		return clientIP
	}

	// 从 grpc 里取对端 IP
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}
	if pr.Addr == net.Addr(nil) {
		return ""
	}
	addrSlice := strings.SplitN(pr.Addr.String(), ":", 2)
	if len(addrSlice) == 2 {
		return addrSlice[0]
	}
	return ""
}

func (b *Builder) grpcHeaderValue(ctx context.Context, key string) string {
	if key == "" {
		return ""
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	return strings.Join(md.Get(key), ";")
}

func (b *Builder) SplitMethodName(fullMethodName string) (string, string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/")
	if idx := strings.Index(fullMethodName, "/"); idx > 0 {
		return fullMethodName[:idx], fullMethodName[idx+1:]
	}
	return "unknown", "unknown"
}
