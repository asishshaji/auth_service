package models

import "time"

type User struct {
	Username  string
	Password  string
	Company   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
