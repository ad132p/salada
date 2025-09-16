package model

import (
	"time"
)

type User struct {
	ID        int
	Username  string
	Email     string
	Password  string
	CreatedAt time.Time
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
