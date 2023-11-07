package prometheus

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

type Callbacks struct {
	NameSpace  string
	Subsystem  string
	Name       string
	InstanceID string
	Help       string
	vector     *prometheus.SummaryVec
}

func (c *Callbacks) Register(db *gorm.DB) error {
	c.vector = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:      c.NameSpace,
		Subsystem: c.Subsystem,
		Namespace: c.Name,
		Help:      c.Help,
		ConstLabels: map[string]string{
			"db_name":     db.Name(),
			"instance_id": c.InstanceID,
		},
		Objectives: map[float64]float64{
			0.9:  0.01,
			0.99: 0.01,
		},
	}, []string{"type", "table"})
	prometheus.MustRegister(c.vector)

	if err := db.Callback().Query().Before("*").Register("prometheus_query_before", c.before("query")); err != nil {
		return err
	}
	if err := db.Callback().Query().After("*").Register("prometheus_query_after", c.after("query")); err != nil {
		return err
	}

	if err := db.Callback().Row().Before("*").Register("prometheus_raw_before", c.before("raw")); err != nil {
		return err
	}
	if err := db.Callback().Row().After("*").Register("prometheus_raw_after", c.after("raw")); err != nil {
		return err
	}

	if err := db.Callback().Create().Before("*").Register("prometheus_create_before", c.before("create")); err != nil {
		return err
	}
	if err := db.Callback().Create().After("*").Register("prometheus_create_after", c.after("create")); err != nil {
		return err
	}

	if err := db.Callback().Update().Before("*").Register("prometheus_update_before", c.before("update")); err != nil {
		return err
	}
	if err := db.Callback().Update().After("*").Register("prometheus_update_after", c.after("update")); err != nil {
		return err
	}

	if err := db.Callback().Delete().Before("*").Register("prometheus_delete_before", c.before("delete")); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("*").Register("prometheus_delete_after", c.after("delete")); err != nil {
		return err
	}
	return nil
}

func (c *Callbacks) before(typ string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		start := time.Now()
		db.Set("start_time", start)
	}
}

func (c *Callbacks) after(typ string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("start_time")
		start, ok := val.(time.Time)
		if !ok {
			return
		}
		c.vector.WithLabelValues(typ, db.Statement.Table).Observe(float64(time.Since(start).Milliseconds()))
	}
}
