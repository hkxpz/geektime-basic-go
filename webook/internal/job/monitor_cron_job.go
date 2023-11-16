package job

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron/v3"

	"geektime-basic-go/webook/pkg/logger"
)

type CronJobBuilder struct {
	vector *prometheus.SummaryVec
	l      logger.Logger
}

func NewCronJobBuilder(l logger.Logger, opt prometheus.SummaryOpts) *CronJobBuilder {
	vector := prometheus.NewSummaryVec(opt, []string{"name", "success"})
	prometheus.MustRegister(vector)
	return &CronJobBuilder{vector: vector, l: l}
}

func (c *CronJobBuilder) Build(job Job) cron.Job {
	name := job.Name()
	return cronJobAdapterFunc(func() {
		start := time.Now()
		c.l.Debug("任务开始", logger.String("name", name), logger.String("time", start.String()))
		err := job.Run()
		duration := time.Since(start)
		if err != nil {
			c.l.Error("执行任务失败", logger.String("name", name), logger.Error(err))
		}
		c.l.Debug("任务结束", logger.String("name", name))
		c.vector.WithLabelValues(name, strconv.FormatBool(err == nil)).Observe(float64(duration.Milliseconds()))
	})
}

var _ cron.Job = (*cronJobAdapterFunc)(nil)

type cronJobAdapterFunc func()

func (c cronJobAdapterFunc) Run() { c() }
