package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Post represents a blog post.
// Fields are mapped directly to database columns.
// Pointers are used for nullable fields in the database.
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
	Tags        pq.StringArray
	Category    string
}

type UpdatePost struct {
	ID       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	Content  string    `json:"content"`
	Tags     string    `json:"tags"`
	Category string    `json:"category"`
}

type CategoryCount struct {
	Category string
	Count    int
}
