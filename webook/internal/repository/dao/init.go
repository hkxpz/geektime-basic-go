package dao

import (
	"gorm.io/gorm"

	"geektime-basic-go/webook/internal/repository/dao/article"
)

func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},

		&SMS{},

		&article.Article{}, &article.PublishedArticle{},
	)
}
