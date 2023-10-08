package domain

type Article struct {
	ID      int64
	Title   string
	Status  ArticleStatus
	Content string
	Author  Author
}

type ArticleStatus uint8

//go:inline
func (as ArticleStatus) ToUint8() uint8 {
	return uint8(as)
}

const (
	// ArticleStatusUnknown 未知状态
	ArticleStatusUnknown ArticleStatus = iota
	// ArticleStatusUnpublished 未发表
	ArticleStatusUnpublished
	// ArticleStatusPublished 已发表
	ArticleStatusPublished
	// ArticleStatusPrivate 仅自己可见
	ArticleStatusPrivate
)

type Author struct {
	ID   int64
	Name string
}
