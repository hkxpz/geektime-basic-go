package article

import "geektime-basic-go/webook/internal/domain"

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

type Req struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type LimitReq struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type LikeReq struct {
	ID   int64 `json:"id"`
	Like bool  `json:"like"`
}

type CollectReq struct {
	ID  int64 `json:"id"`
	Cid int64 `json:"cid"`
}

func (req *Req) toDomain(uid int64) domain.Article {
	return domain.Article{
		ID:      req.ID,
		Title:   req.Title,
		Content: req.Content,
		Author:  domain.Author{ID: uid},
	}
}
