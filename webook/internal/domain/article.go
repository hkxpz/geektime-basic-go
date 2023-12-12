package domain

import "time"

type Article struct {
	ID      int64
	Title   string
	Status  ArticleStatus
	Content string
	// 作者
	Author   Author
	CreateAt time.Time
	UpdateAt time.Time
}

func (a *Article) Abstract() string {
	cs := []rune(a.Content)
	if len(cs) < 100 {
		return a.Content
	}
	return string(cs[:100])
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

type Vo struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	// 摘要
	Abstract string `json:"abstract"`
	// 内容
	Content  string `json:"content"`
	Status   uint8  `json:"status"`
	Author   string `json:"author"`
	CreateAt string `json:"create_at"`
	UpdateAt string `json:"update_at"`

	// 点赞之类的信息
	LikeCnt    int64 `json:"likeCnt"`
	CollectCnt int64 `json:"collectCnt"`
	ReadCnt    int64 `json:"readCnt"`

	// 个人是否点赞的信息
	Liked     bool `json:"liked"`
	Collected bool `json:"collected"`
}
