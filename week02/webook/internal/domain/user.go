package domain

import "time"

type User struct {
	Id       int64
	Email    string
	Password string
	CreateAt time.Time
}
