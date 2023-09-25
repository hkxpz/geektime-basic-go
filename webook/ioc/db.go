package ioc

import (
	"time"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"

	"geektime-basic-go/webook/internal/repository/dao"
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
		Logger: glogger.New(gormLoggerFunc(l.Info), glogger.Config{
			SlowThreshold:        50 * time.Millisecond,
			LogLevel:             glogger.Info,
			ParameterizedQueries: true,
		}),
	})
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
	g(msg, logger.Field{Key: "args", Value: args})
}
