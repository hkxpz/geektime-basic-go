package article

type Article struct {
	ID       int64  `gorm:"primaryKey,autoIncrement" bson:"id,omitempty"`
	Title    string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	Content  string `gorm:"type=BLOB" bson:"content,omitempty"`
	AuthorID int64  `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	CreateAt int64  `bson:"create_at,omitempty"`
	UpdateAt int64  `bson:"update_at,omitempty"`
}

type PublishedArticle Article
