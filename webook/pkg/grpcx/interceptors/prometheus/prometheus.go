package prometheus

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"geektime-basic-go/webook/pkg/grpcx/interceptors"
)

type InterceptorBuilder struct {
	NameSpace string
	Subsystem string
	interceptors.Builder
}

func NewInterceptorBuilder(NameSpace, Subsystem string) *InterceptorBuilder {
	return &InterceptorBuilder{NameSpace: NameSpace, Subsystem: Subsystem, Builder: interceptors.NewBuilder()}
}

func (i *InterceptorBuilder) defaultUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  i.NameSpace,
			Subsystem:  i.Subsystem,
			Name:       "server_handle_seconds",
			Objectives: map[float64]float64{0.5: 0.01, 0.9: 0.01, 0.95: 0.01, 0.99: 0.001, 0.999: 0.0001}},
		[]string{"type", "service", "method", "peer", "code"},
	)
	prometheus.MustRegister(summary)

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		start := time.Now()
		defer func() {
			end := time.Since(start)
			serviceName, method := i.SplitMethodName(info.FullMethod)
			st, _ := status.FromError(err)
			code := "OK"
			if st != nil {
				code = st.Code().String()
			}
			summary.WithLabelValues("unary", serviceName, method, i.PeerName(ctx), code).Observe(float64(end.Milliseconds()))
		}()
		return handler(ctx, req)
	}
}
