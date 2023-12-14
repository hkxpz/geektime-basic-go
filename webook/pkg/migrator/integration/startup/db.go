package startup

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitSrcDB() *gorm.DB {
	return initDB("db.mysql")
}

func InitIntrDB() *gorm.DB {
	return initDB("db.mysql.intr")
}

func initDB(key string) *gorm.DB {
	cfg := struct {
		DSN string `yaml:"dsn"`
	}{}
	if err := viper.UnmarshalKey(key, &cfg); err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN))
	if err != nil {
		panic(err)
	}
	return db
}
