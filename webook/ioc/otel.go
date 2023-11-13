package ioc

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func InitOTEL() func(ctx context.Context) {
	res, err := newResource("demo", "v0.0.1")
	if err != nil {
		panic(fmt.Sprintf("opentelemetry 初始化 resource 失败: %s", err))
	}

	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	tp, err := newTraceProvider(res)
	if err != nil {
		panic(fmt.Sprintf("opentelemetry 初始化 TraceProvider 失败: %s", err))
	}
	otel.SetTracerProvider(tp)
	return func(ctx context.Context) { _ = tp.Shutdown(ctx) }
}

func newResource(serviceName string, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(), resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion(serviceVersion),
	))
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
}

func newTraceProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	var config struct {
		Addr string `yaml:"addrs"`
	}
	if err := viper.UnmarshalKey("zipkin", &config); err != nil {
		panic(fmt.Sprintf("opentelemetry 初始化获取 zipkin 地址失败: %s", err))
	}

	exporter, err := zipkin.New("http://" + config.Addr + "/api/v2/spans")
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(exporter, trace.WithBatchTimeout(time.Second)),
		trace.WithResource(res),
	)
	return traceProvider, nil
}
