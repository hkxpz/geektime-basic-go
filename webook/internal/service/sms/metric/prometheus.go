package metric

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"geektime-basic-go/webook/internal/service/sms"
)

type PrometheusDecorator struct {
	svc    sms.Service
	vector *prometheus.SummaryVec
}

func (p *PrometheusDecorator) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	start := time.Now()
	defer func() {
		p.vector.WithLabelValues(tplId).Observe(float64(time.Since(start).Milliseconds()))
	}()
	return p.svc.Send(ctx, tplId, args, numbers...)
}

func NewPrometheusDecorator(svc sms.Service, namespace, subsystem, instanceID, name string) *PrometheusDecorator {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        name,
		ConstLabels: map[string]string{"instance_id": instanceID},
		Objectives: map[float64]float64{
			0.9:   0.01,
			0.95:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{"tpl"})
	prometheus.MustRegister(vector)
	return &PrometheusDecorator{svc: svc, vector: vector}
}
