package migrator

type Entity interface {
	// Id 要求返回 Id
	Id() int64
	// CompareTo dst 必然也是 Entity 正常来说类型是一样的
	CompareTo(dst Entity) bool
}
