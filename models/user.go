package models

import "time"

type User struct {
	Id        int
	Username  string
	Password  string
	Salt      string
	CreatedAt time.Time `db:"created_at"`
}
