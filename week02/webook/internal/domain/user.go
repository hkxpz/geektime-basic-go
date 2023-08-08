package domain

import "time"

type User struct {
	Id           int64     `json:"id,omitempty"`
	Email        string    `json:"email,omitempty"`
	Password     string    `json:"password,omitempty"`
	Nickname     string    `json:"nickname,omitempty"`
	Birthday     string    `json:"birthday,omitempty"`
	Introduction string    `json:"introduction,omitempty"`
	CreateAt     time.Time `json:"createAt,omitempty"`
}
