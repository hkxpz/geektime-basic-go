package domain

import "time"

type User struct {
	Id       int64     `json:"id,omitempty"`
	Email    string    `json:"email,omitempty"`
	Password string    `json:"password,omitempty"`
	Nickname string    `json:"nickname,omitempty"`
	AboutMe  string    `json:"introduction,omitempty"`
	Birthday time.Time `json:"birthday,omitempty"`
	CreateAt time.Time `json:"createAt,omitempty"`
}
