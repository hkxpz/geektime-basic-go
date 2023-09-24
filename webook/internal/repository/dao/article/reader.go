package article

type PublishedArticle struct {
	Article
	Statue uint8 `gorm:"default=1"`
}
