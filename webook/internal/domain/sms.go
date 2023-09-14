package domain

import "time"

type SMS struct {
	Biz      string
	Args     string
	Numbers  string
	Status   int64
	CreateAt time.Time
}
