package controller

import (
	"fmt"
	"log"
	"net/http"
	"salada/internal/blog/model"
	"salada/internal/blog/repositories"
	"time"

	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// BlogController handles blog post-related requests.
type BlogController struct {
	Repo *repositories.PostRepository
}

// NewPostController creates a new PostController instance.
func NewBlogController(repo *repositories.PostRepository) *BlogController {
	return &BlogController{Repo: repo}
}

// CreatePost handles POST /blog
func (pc *BlogController) CreatePost(c *gin.Context) {

	title := c.PostForm("title")
	content := c.PostForm("content")
	authorName := c.PostForm("author")
	tags := c.PostForm("tags")

	post := model.Post{
		Title:      title,
		Content:    content,
		AuthorName: authorName,
		Tags:       tags,
		// PublishedAt will be set on publish, or remain nil
	}

	if err := c.ShouldBind(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := pc.Repo.CreatePost(&post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post", "details": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, "/blog/")
}

// UpdatePost handles PUT /posts/:id
func (pc *BlogController) UpdatePost(c *gin.Context) {
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
		AuthorName  *string    `json:"author_name"`
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

	if input.AuthorName != nil {
		post.AuthorName = *input.AuthorName
	}

	if err := pc.Repo.UpdatePost(post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
}

// UpdatePost handles PUT /publish/:id
func (pc *BlogController) PublishPost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
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

	if post.PublishedAt != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Post not found"})
		return
	}
	if err := pc.Repo.PublishPost(post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
}

// UploadImage handles POST /blog
func (pc *BlogController) UploadImage(c *gin.Context) {
	image, _ := c.FormFile("image")
	log.Println(image.Filename)

	// Upload the file to specific dst.
	c.SaveUploadedFile(image, "./"+image.Filename)

	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", image.Filename))
}

// GetPosts handles GET /blog/
func (pc *BlogController) GetPosts(c *gin.Context) {
	posts, err := pc.Repo.GetPublishedPosts()
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

func (pc *BlogController) GetNewPostForm(c *gin.Context) {
	c.HTML(http.StatusOK, "post_form.html", gin.H{
		"title": "New Blog Entry",
	})
}

// GetPostBySlug handles GET /blog/:slug
func (pc *BlogController) GetPostBySlug(c *gin.Context) {
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
	c.HTML(http.StatusOK, "blog_post.html", gin.H{
		"title": post.Title,
		"post":  post,
	})
}

// DeletePost handles DELETE /blog/:id
func (pc *BlogController) DeletePost(c *gin.Context) {
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
	c.Next()
}

// EditPostForm handles GET /blog/edit/:slug/
func (pc *BlogController) EditPostForm(c *gin.Context) {
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
	c.HTML(http.StatusOK, "edit_post_form.html", gin.H{
		"title": post.Title,
		"post":  post,
	})
}
