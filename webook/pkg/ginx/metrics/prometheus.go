package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusBuilder struct {
	NameSpace  string
	Subsystem  string
	Name       string
	Help       string
	InstanceID string
}

func (p *PrometheusBuilder) BuildResponseTime() gin.HandlerFunc {
	labels := []string{"method", "pattern", "status"}
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:   p.NameSpace,
		Subsystem:   p.Subsystem,
		Name:        p.Name + "_resp_time",
		Help:        p.Help,
		ConstLabels: map[string]string{"instance_id": p.InstanceID},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.01,
			0.999: 0.001,
		},
	}, labels)
	prometheus.MustRegister(vector)
	return func(ctx *gin.Context) {
		start := time.Now()
		defer func() {
			vector.WithLabelValues(
				ctx.Request.Method,
				ctx.FullPath(),
				strconv.Itoa(ctx.Writer.Status()),
			).Observe(float64(time.Since(start).Milliseconds()))
		}()
		ctx.Next()
	}
}

func (p *PrometheusBuilder) BuildActiveRequest() gin.HandlerFunc {
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   p.NameSpace,
		Subsystem:   p.Subsystem,
		Name:        p.Name + "_active_req",
		Help:        p.Help,
		ConstLabels: map[string]string{"instance_id": p.InstanceID},
	})
	prometheus.MustRegister(gauge)
	return func(ctx *gin.Context) {
		gauge.Inc()
		defer gauge.Dec()
		ctx.Next()
	}
}
