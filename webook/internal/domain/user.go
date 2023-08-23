package domain

import "time"

type User struct {
	Id       int64
	Email    string
	Nickname string
	Password string
	Phone    string
	AboutMe  string
	Birthday time.Time
	CreateAt time.Time
}
