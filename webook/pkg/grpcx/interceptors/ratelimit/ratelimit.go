// Package ratelimit 限流器
package ratelimit

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"geektime-basic-go/webook/pkg/logger"
	"geektime-basic-go/webook/pkg/ratelimit"
)

type InterceptorBuilder struct {
	limiter ratelimit.Limiter
	key     string
	l       logger.Logger
}

func NewInterceptorBuilder(limiter ratelimit.Limiter, key string, l logger.Logger) *InterceptorBuilder {
	return &InterceptorBuilder{limiter: limiter, key: key, l: l}
}

// Build "limiter:service" 整个应用、集群限流
func (b *InterceptorBuilder) Build() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		limited, err := b.limiter.Limit(ctx, b.key)
		if err != nil {
			b.l.Error("限流器故障", logger.Error(err))
			// 保守的限流策略
			return nil, status.Error(codes.ResourceExhausted, "触发限流")
			// 激进的限流策略
			//return handler(ctx,req)
		}
		if limited {
			return nil, status.Error(codes.ResourceExhausted, "触发限流")
		}
		return handler(ctx, req)
	}
}

// BuildV1 "limiter:service:UserService" UserService 限流
func (b *InterceptorBuilder) BuildV1() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if !strings.HasPrefix(info.FullMethod, "/UserService") {
			return handler(ctx, req)
		}

		limited, err := b.limiter.Limit(ctx, b.key+":UserService")
		if err != nil {
			b.l.Error("限流器故障", logger.Error(err))
			// 保守的限流策略
			return nil, status.Error(codes.ResourceExhausted, "触发限流")
			// 激进的限流策略
			//return handler(ctx,req)
		}
		if limited {
			return nil, status.Error(codes.ResourceExhausted, "触发限流")
		}
		return handler(ctx, req)
	}
}
