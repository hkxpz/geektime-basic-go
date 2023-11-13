package otel

import (
	"context"

	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"geektime-basic-go/webook/internal/service/sms"
)

type service struct {
	svc    sms.Service
	tracer trace.Tracer
}

func (s *service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	ctx, span := s.tracer.Start(ctx, "sms_send")
	defer span.End()
	span.SetAttributes(attribute.String("tplID", tplId))
	err := s.svc.Send(ctx, tplId, args, numbers...)
	if err != nil {
		span.RecordError(err)
	}
	return err
}

func NewService(svc sms.Service) sms.Service {
	return &service{svc: svc, tracer: otel.GetTracerProvider().Tracer("sms_service")}
}
