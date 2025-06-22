package controller

import (
	"database/sql"
	"net/http"
	"salada/internal/blog/model"
	"salada/internal/blog/repositories"
	"time"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

// PostController handles blog post-related requests.
type PostController struct {
	Repo *repositories.PostRepository
}

// NewPostController creates a new PostController instance.
func NewPostController(repo *repositories.PostRepository) *PostController {
	return &PostController{Repo: repo}
}

// CreatePost handles POST /posts
func (pc *PostController) CreatePost(c *gin.Context) {
	var input struct {
		Title    string     `json:"title" binding:"required"`
		Slug     string     `json:"slug" binding:"required"`
		Content  string     `json:"content" binding:"required"`
		AuthorID *uuid.UUID `json:"author_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post := model.Post{
		Title:    input.Title,
		Slug:     input.Slug,
		Content:  input.Content,
		AuthorID: input.AuthorID,
		// PublishedAt will be set on publish, or remain nil
	}

	if err := pc.Repo.CreatePost(&post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, post)
}

// GetPosts handles GET /posts
func (pc *PostController) GetPosts(c *gin.Context) {
	posts, err := pc.Repo.GetPosts()
	if err != nil {
		c.HTML(http.StatusServiceUnavailable, "blog.html", gin.H{
			"title": "Blog Posts",
			"error": "Failed to retrieve posts",
		})
		return
	}
	c.HTML(http.StatusOK, "blog.html", gin.H{
		"title": "Blog Posts",
		"posts": posts,
	})
}

// GetPostBySlug handles GET /posts/:slug
func (pc *PostController) GetPostBySlug(c *gin.Context) {
	slug := c.Param("slug")
	post, err := pc.Repo.GetPostBySlug(slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve post"})
		return
	}
	if post == nil { // Check if no record was found by the repository
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}
	c.JSON(http.StatusOK, post)
}

// UpdatePost handles PUT /posts/:id
func (pc *PostController) UpdatePost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var input struct {
		Title       *string    `json:"title"`
		Slug        *string    `json:"slug"`
		Content     *string    `json:"content"`
		PublishedAt *time.Time `json:"published_at"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post, err := pc.Repo.GetPostByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find post"})
		return
	}
	if post == nil { // Check if no record was found
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Update fields if provided in the input
	if input.Title != nil {
		post.Title = *input.Title
	}
	if input.Slug != nil {
		post.Slug = *input.Slug
	}
	if input.Content != nil {
		post.Content = *input.Content
	}
	if input.PublishedAt != nil {
		post.PublishedAt = input.PublishedAt
	}

	if err := pc.Repo.UpdatePost(post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
}

// DeletePost handles DELETE /posts/:id
func (pc *PostController) DeletePost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	err = pc.Repo.DeletePost(id)
	if err != nil {
		if err == sql.ErrNoRows { // Check for no rows affected, indicating not found
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post", "details": err.Error()})
		return
	}

	c.Status(http.StatusNoContent) // 204 No Content for successful deletion
}

func (pc *PostController) GetAdmin(c *gin.Context) {
	posts, err := pc.Repo.GetAdmin()
	if err != nil {
		c.HTML(http.StatusServiceUnavailable, "blog.html", gin.H{
			"title": "Admin Interface",
			"error": "Failed to retrieve posts",
		})
		return
	}
	c.HTML(http.StatusOK, "admin.html", gin.H{
		"title": "Admin Interface",
		"posts": posts,
	})
}
