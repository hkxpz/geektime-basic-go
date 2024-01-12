// Package circuitbreaker 熔断器
package circuitbreaker

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/go-kratos/aegis/circuitbreaker"
	"google.golang.org/grpc"
)

type InterceptorBuilder struct {
	breaker circuitbreaker.CircuitBreaker
}

func (b *InterceptorBuilder) Build() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if err = b.breaker.Allow(); err != nil {
			// 触发了熔断器
			b.breaker.MarkFailed()
			return nil, err
		}

		resp, err = handler(ctx, req)
		if st, ok := status.FromError(err); ok && st.Code() == codes.Unavailable {
			// 是否是业务错误
			b.breaker.MarkFailed()
			return
		}
		b.breaker.MarkSuccess()
		return
	}
}
