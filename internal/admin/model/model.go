package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Username  string
	Email     string
	Password  string
	CreatedAt time.Time
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
