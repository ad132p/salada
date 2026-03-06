package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Username  string `json:"username" form:"username"`
	Email     string `json:"email" form:"email"`
	Password  string `json:"password" form:"password"`
	CreatedAt time.Time
}

type LoginInput struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}
