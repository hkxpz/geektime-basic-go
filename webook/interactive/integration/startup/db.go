package startup

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"

	"geektime-basic-go/webook/interactive/repository/dao"
)

func InitDB() *gorm.DB {
	cfg := struct {
		DSN string `yaml:"dsn"`
	}{}
	if err := viper.UnmarshalKey("db.mysql", &cfg); err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN))
	if err != nil {
		panic(err)
	}

	if err = db.Use(tracing.NewPlugin(tracing.WithoutMetrics())); err != nil {
		panic(err)
	}

	if err = dao.InitTables(db); err != nil {
		panic(err)
	}

	return db
}
