package ioc

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"

	"geektime-basic-go/webook/interactive/repository/dao"
	prometheus2 "geektime-basic-go/webook/pkg/gormx/callbacks/prometheus"
	"geektime-basic-go/webook/pkg/gormx/connpool"
	"geektime-basic-go/webook/pkg/logger"
)

// SrcDB 纯粹是为了 wire 而准备的
type SrcDB *gorm.DB

// DstDB 纯粹是为了 wire 而准备的
type DstDB *gorm.DB

func InitSRC(l logger.Logger) SrcDB {
	return initDB("db.mysql", l)
}

func InitDST(l logger.Logger) DstDB {
	return initDB("db.mysql.intr", l)
}

func InitDoubleWritePool(src SrcDB, dst DstDB, l logger.Logger) *connpool.DoubleWritePool {
	return connpool.NewDoubleWritePool(src, dst, l)
}

func InitBizDB(pool *connpool.DoubleWritePool) *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: pool}))
	if err != nil {
		panic(err)
	}
	return db
}

func initDB(key string, l logger.Logger) *gorm.DB {
	cfg := struct {
		DSN string `yaml:"dsn"`
	}{}
	if err := viper.UnmarshalKey(key, &cfg); err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		//慢查询日志
		Logger: glogger.New(gormLoggerFunc(l.Warn), glogger.Config{
			SlowThreshold:        50 * time.Millisecond,
			LogLevel:             glogger.Warn,
			ParameterizedQueries: true,
		}),
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
		Name:       "gorm_" + strings.ReplaceAll(key, ".", "_"),
		InstanceID: "instance-1",
		Help:       "gorm DB 查询",
	}).Register(db)
	if err != nil {
		panic(err)
	}

	if err = dao.InitTables(db); err != nil {
		panic(err)
	}

	return db
}

type gormLoggerFunc func(msg string, fields ...any)

func (g gormLoggerFunc) Printf(msg string, args ...any) {
	g("GORM LOG", logger.String("args", fmt.Sprintf(msg, args...)))
}
