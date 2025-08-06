package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        int
	Username  string
	Email     string
	Password  string
	CreatedAt time.Time
}

type Post struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Content     string     `json:"content"`
	AuthorID    *uuid.UUID `json:"author_id,omitempty"`    // Nullable in DB
	AuthorName  string     `json:"author_name,omitempty"`  // Nullable in DB
	PublishedAt *time.Time `json:"published_at,omitempty"` // Nullable in DB
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Tags        string
}

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
