package ioc

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"

	"geektime-basic-go/webook/interactive/repository/dao"
	intrdao "geektime-basic-go/webook/interactive/repository/dao"
	prometheus2 "geektime-basic-go/webook/pkg/gormx/callbacks/prometheus"
	"geektime-basic-go/webook/pkg/logger"
)

func InitDB(l logger.Logger) *gorm.DB {
	cfg := struct {
		DSN string `yaml:"dsn"`
	}{}
	if err := viper.UnmarshalKey("db.mysql", &cfg); err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		// 慢查询日志
		//Logger: glogger.New(gormLoggerFunc(l.Info), glogger.Config{
		//	SlowThreshold:        100 * time.Millisecond,
		//	LogLevel:             glogger.Error,
		//	ParameterizedQueries: true,
		//}),
	})
	if err != nil {
		panic(err)
	}

	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook",
		RefreshInterval: 15,
		StartServer:     false,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{VariableNames: []string{"thread_running"}},
		},
	}))
	if err != nil {
		panic(err)
	}

	if err = db.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
		panic(err)
	}

	err = (&prometheus2.Callbacks{
		NameSpace:  "hkxpz",
		Subsystem:  "webook",
		Name:       "gorm",
		InstanceID: "instance-1",
		Help:       "gorm DB 查询",
	}).Register(db)
	if err != nil {
		panic(err)
	}

	if err = dao.InitTables(db); err != nil {
		panic(err)
	}
	if err = intrdao.InitTables(db); err != nil {
		panic(err)
	}

	return db
}

type gormLoggerFunc func(msg string, fields ...any)

func (g gormLoggerFunc) Printf(msg string, args ...any) {
	g(msg, logger.Field{Key: "args", Value: args})
}
