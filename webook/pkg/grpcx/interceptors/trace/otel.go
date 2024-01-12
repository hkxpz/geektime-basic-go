package trace

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	"geektime-basic-go/webook/pkg/grpcx/interceptors"
)

type InterceptorBuilder struct {
	tracer      trace.Tracer
	propagator  propagation.TextMapPropagator
	serviceName string
	interceptors.Builder
}

func NewInterceptorBuilder(tracer trace.Tracer, propagator propagation.TextMapPropagator, serviceName string) *InterceptorBuilder {
	return &InterceptorBuilder{tracer: tracer, propagator: propagator, serviceName: serviceName, Builder: interceptors.NewBuilder()}
}

func (b *InterceptorBuilder) Build() grpc.UnaryServerInterceptor {
	tracer := b.tracer
	if tracer == nil {
		tracer = otel.GetTracerProvider().Tracer("geektime-basic-go/webook/pkg/grpcx")
	}
	propagator := b.propagator
	if propagator == nil {
		propagator = propagation.NewCompositeTextMapPropagator(propagation.TraceContext{})
	}
	attrs := []attribute.KeyValue{
		semconv.RPCServiceKey.String("grpc"),
		attribute.Key("rpc.grpc.kind").String("unary"),
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		spanCtx, span := tracer.Start(ctx, info.FullMethod, trace.WithAttributes(attrs...))
		span.SetAttributes(
			semconv.RPCMethodKey.String(info.FullMethod),
			semconv.NetPeerNameKey.String(b.PeerName(spanCtx)),
			attribute.Key("net.per.ip").String(b.PeerIP(spanCtx)),
		)
		defer func() {
			if err != nil {
				span.RecordError(err)
				if e := errors.FromError(err); e != nil {
					span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(e.Code)))
				}
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "OK")
			}

			span.End()
		}()

		return handler(ctx, req)
	}
}
