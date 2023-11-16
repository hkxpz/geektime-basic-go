package domain

type Interactive struct {
	BizID      int64 `json:"biz_id"`
	ReadCnt    int64 `json:"read_cnt"`
	LikeCnt    int64 `json:"like_cnt"`
	CollectCnt int64 `json:"collect_cnt"`
	Liked      bool  `json:"liked"`
	Collected  bool  `json:"collected"`
}
