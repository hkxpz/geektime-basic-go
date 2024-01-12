package log

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"geektime-basic-go/webook/pkg/grpcx/interceptors"
	"geektime-basic-go/webook/pkg/logger"
)

type InterceptorBuilder struct {
	l logger.Logger
	interceptors.Builder
}

func NewInterceptorBuilder(l logger.Logger) *InterceptorBuilder {
	return &InterceptorBuilder{l: l, Builder: interceptors.NewBuilder()}
}

func (b *InterceptorBuilder) Build() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if info.FullMethod == "grpc.health.v1.Health/Check" {
			return handler(ctx, req)
		}
		var start = time.Now()
		var event = "normal"
		defer func() {
			cost := time.Since(start)
			if rec := recover(); rec != nil {
				switch recType := rec.(type) {
				case error:
					err = recType
				default:
					err = fmt.Errorf("%v", err)
				}
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				event = "recover"
				err = status.New(codes.Internal, "panic, err"+err.Error()).Err()
			}
			var code = "OK"
			st, _ := status.FromError(err)
			if st != nil {
				code = st.Code().String()
			}
			b.l.Info(
				"RPC调用",
				logger.String("type", "unary"),
				logger.String("code", code),
				logger.String("code_msg", st.Message()),
				logger.String("event", event),
				logger.String("method", info.FullMethod),
				logger.Int("cost", cost.Milliseconds()),
				logger.String("peer", b.PeerName(ctx)),
				logger.String("peer_ip", b.PeerIP(ctx)),
			)
		}()
		return handler(ctx, req)
	}
}
