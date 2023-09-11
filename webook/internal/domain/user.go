package domain

import "time"

type User struct {
	ID       int64
	Email    string
	Nickname string
	Password string
	Phone    string
	AboutMe  string
	Birthday time.Time
	CreateAt time.Time

	WechatInfo WechatInfo
}

type WechatInfo struct {
	OpenID  string
	UnionID string
}
