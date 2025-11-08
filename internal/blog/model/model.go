package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Comment struct definition
type Comment struct {
	ID         string    `json:"id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	AuthorName string    `json:"author_name"`
}

// Post represents a blog post.
// Fields are mapped directly to database columns.
// Pointers are used for nullable fields in the database.
type Post struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	Slug         string     `json:"slug"`
	Content      string     `json:"content"`
	AuthorID     *uuid.UUID `json:"author_id,omitempty"`    // Nullable in DB
	AuthorName   string     `json:"author,omitempty"`       // Nullable in DB
	PublishedAt  *time.Time `json:"published_at,omitempty"` // Nullable in DB
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Tags         pq.StringArray
	ThumbnailURL string    `json:"thumbnail"`
	Category     string    `json:"category"`
	ImageIDs     []string  `json:"image_ids"`
	Seen         int       `json:"seen"`
	Comments     []Comment `json:"comments"`
}

type CreateCommentRequest struct {
	PostID     uuid.UUID `json:"postID"` // The UUID of the blog post this comment belongs to
	Content    string    `json:"comment"`
	AuthorName string    `json:"author_name"`
}

type LikeRequest struct {
	PostID string `json:"post_id"` // Assuming this holds the slug
	Action string `json:"action"`
}

type UpdatePost struct {
	ID         uuid.UUID `json:"id"`
	AuthorName string
	Title      string `json:"title"`
	Content    string `json:"content"`
	Tags       string `json:"tags"`
	Category   string `json:"category"`
}

type CreatePost struct {
	Title      string         `json:"title"`
	Content    string         `json:"content"`
	Tags       string         `json:"tags"`
	Category   string         `json:"category"`
	ImageIDs   pq.StringArray `json:"image_ids"`
	AuthorName string
}

type UpdateImages struct {
	ImageIDs pq.StringArray `json:"image_ids"`
	PostID   uuid.UUID      `json:"post_id"`
}

type Image struct {
	ID         uuid.UUID `json:"id"`
	Filepath   string    `json:"title"`
	Status     string    `json:"content"`
	BlogPostID uuid.UUID `json:"tags"`
	UploadedAt time.Time
}

type CategoryCount struct {
	Category string
	Count    int
}
