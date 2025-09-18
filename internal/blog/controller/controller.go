package controller

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"salada/internal/blog"
	"salada/internal/blog/model"
	"salada/internal/blog/repositories"

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

func (pc *BlogController) CreatePost(c *gin.Context) {

	title := c.PostForm("title")
	content := c.PostForm("content")
	author := c.PostForm("author")

	category := c.PostForm("category")

	// The rest of your code remains the same...
	post := model.Post{
		Title:      title,
		Content:    content,
		AuthorName: author, // Use the safely retrieved variable
		//Tags:       tags,
		Category: category,
	}

	// Bind and save the post...
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

// UpdatePost handles PUT on /posts/:id
func (pc *BlogController) UpdatePost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	updatePayload := model.Post{}

	if err := c.ShouldBindJSON(&updatePayload); err != nil {
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
	if updatePayload.Title != "" {
		post.Title = updatePayload.Title
	}
	if updatePayload.Slug != "" {
		post.Slug = updatePayload.Slug
	}
	if updatePayload.Content != "" {
		post.Content = updatePayload.Content
	}
	if updatePayload.PublishedAt != nil {
		post.PublishedAt = updatePayload.PublishedAt
	}

	if updatePayload.Category != "" {
		post.Category = updatePayload.Category
	}

	if len(updatePayload.Tags) != 0 {
		post.Tags = updatePayload.Tags
	}

	if err := pc.Repo.UpdatePost(post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, post)
}

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

func (pc *BlogController) UploadImage(c *gin.Context) {
	// Single file
	file, err := c.FormFile("image") // The name "image" must match the form field name in the client request
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve the file"})
		return
	}

	// Generate a unique filename to prevent conflicts
	filename := filepath.Base(file.Filename)

	// Save the file to the specified path
	dst := filepath.Join("uploads", filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the file"})
		return
	}

	// Construct the URL for the saved image
	imageURL := fmt.Sprintf("uploads/%s", filename)

	// Respond to the client with the image URL in the expected format
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"filePath": imageURL,
		},
	})
}

// GetPosts handles GET /blog/
func (pc *BlogController) GetPosts(c *gin.Context) {
	username, ok := c.MustGet("username").(string)
	if !ok {
		// This should not happen if the middleware is correctly set up.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Username not found in context"})
		return
	}

	posts, err := pc.Repo.GetPublishedPosts()

	if err != nil {
		c.HTML(http.StatusServiceUnavailable, "blog.html", gin.H{
			"title": "Blog Posts",
			"error": err,
		})
		return
	}
	c.HTML(http.StatusOK, "blog.html", gin.H{
		"title":    "Blog Posts",
		"posts":    posts,
		"username": username,
	})
}

func (pc *BlogController) GetNewPostForm(c *gin.Context) {
	username, ok := c.MustGet("username").(string)
	if !ok {
		// This should not happen if the middleware is correctly set up.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Username not found in context"})
		return
	}
	c.HTML(http.StatusOK, "blog_new.html", gin.H{
		"title":    "New Blog Entry",
		"username": username,
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
		"title":   post.Title,
		"post":    post,
		"content": template.HTML(blog.RenderMarkdownToHTML(post.Content)),
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
