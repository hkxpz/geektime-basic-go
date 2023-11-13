package wechat

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"geektime-basic-go/webook/internal/domain"
)

type PrometheusDecorator struct {
	service
	sum prometheus.Summary
}

func NewPrometheusDecorator(svc service, namespace, subsystem, instanceID, name string) *PrometheusDecorator {
	sum := prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        name,
		ConstLabels: map[string]string{"instance_id": instanceID},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.95:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})
	prometheus.MustRegister(sum)
	return &PrometheusDecorator{service: svc, sum: sum}
}

func (p *PrometheusDecorator) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	start := time.Now()
	defer func() {
		p.sum.Observe(float64(time.Since(start).Milliseconds()))
	}()
	return p.service.VerifyCode(ctx, code)
}
