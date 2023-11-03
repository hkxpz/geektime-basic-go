package domain

type HistoryRecord struct {
	// 考虑到历史记录可以支持不同的类型，例如视频之类的，这里也沿用 biz 和 bizId 的设计
	Biz   string
	BizID int64
	Uid   int64
}
