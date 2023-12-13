package migrator

type Entity interface {
	// ID 要求返回 ID
	ID() int64
	// TableName 返回表名
	TableName() string
	// CompareTo dst 必然也是 Entity 正常来说类型是一样的
	CompareTo(dst Entity) bool
}
