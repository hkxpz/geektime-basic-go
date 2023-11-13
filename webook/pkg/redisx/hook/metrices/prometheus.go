package metrices

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type PrometheusHook struct {
	vector *prometheus.SummaryVec
}

func NewPrometheusHook(namespace, subsystem, instanceID, name string) *PrometheusHook {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        name,
		ConstLabels: map[string]string{"instance_id": instanceID},
	}, []string{"cmd", "key_exist"})
	prometheus.MustRegister(vector)
	return &PrometheusHook{vector: vector}
}

func (p *PrometheusHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (p *PrometheusHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) (err error) {
		start := time.Now()
		defer func() {
			p.vector.WithLabelValues(cmd.Name(), strconv.FormatBool(errors.Is(err, redis.Nil))).
				Observe(float64(time.Since(start).Milliseconds()))
		}()
		return next(ctx, cmd)
	}
}

func (p *PrometheusHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		return next(ctx, cmds)
	}
}
