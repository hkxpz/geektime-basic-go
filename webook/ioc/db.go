package ioc

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"geektime-basic-go/webook/config"
	"geektime-basic-go/webook/internal/repository/dao"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open(config.Config.DB.DSN))
	if err != nil {
		panic(err)
	}

	if err = dao.InitTables(db); err != nil {
		panic(err)
	}

	return db
}
