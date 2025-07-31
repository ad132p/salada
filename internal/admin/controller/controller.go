package admin_controller

import (
	"fmt"
	"log"
	"net/http"
	"salada/internal/blog/model"
	"salada/internal/blog/repositories"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminController handles blog post-related requests.
type AdminController struct {
	Repo    *repositories.PostRepository
	Secrets *gin.H
}

var secrets = gin.H{
	"admin": gin.H{"email": "foo@bar.com", "phone": "123433"},
}

// NewAdminController creates a new AdminController instance.
func NewAdminController(repo *repositories.PostRepository) *AdminController {
	return &AdminController{Repo: repo}
}

// CreatePost handles POST /blog
func (pc *AdminController) CreatePost(c *gin.Context) {

	title := c.PostForm("title")
	content := c.PostForm("content")
	authorName := c.PostForm("author")

	post := model.Post{
		Title:      title,
		Content:    content,
		AuthorName: authorName,
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

// UploadImage handles POST /admin/blog
func (pc *AdminController) UploadImage(c *gin.Context) {
	file, _ := c.FormFile("file")
	log.Println(file.Filename)

	// Upload the file to specific dst.
	c.SaveUploadedFile(file, "./"+file.Filename)

	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
}

// GetPosts handles GET /blog/
func (pc *AdminController) GetPendingPosts(c *gin.Context) {
	posts, err := pc.Repo.GetPosts()
	if err != nil {
		c.HTML(http.StatusServiceUnavailable, "blog_admin.html", gin.H{
			"title": "Blog Posts - Admin Page",
			"error": "Failed to retrieve posts",
		})
		return
	}
	c.HTML(http.StatusOK, "blog_admin.html", gin.H{
		"title": "Blog Posts - Admin Page         ",
		"posts": posts,
	})
}

func (pc *AdminController) GetAdminMain(c *gin.Context) {
	// get user, it was set by the BasicAuth middleware
	user := c.MustGet(gin.AuthUserKey).(string)
	if secret, ok := secrets[user]; ok {
		c.JSON(http.StatusOK, gin.H{"user": user, "secret": secret})
	} else {
		c.HTML(http.StatusOK, "admin.html", gin.H{
			"title": "New Blog Entry",
		})
	}
}

func (pc *AdminController) GetNewPostForm(c *gin.Context) {
	// get user, it was set by the BasicAuth middleware
	user := c.MustGet(gin.AuthUserKey).(string)
	if secret, ok := secrets[user]; ok {
		c.JSON(http.StatusOK, gin.H{"user": user, "secret": secret})
	} else {
		c.HTML(http.StatusOK, "post_form.html", gin.H{
			"title": "New Blog Entry",
		})
	}
}

// GetPostBySlug handles GET /blog/:slug
func (pc *AdminController) GetPostBySlug(c *gin.Context) {
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
	c.HTML(http.StatusOK, "blog_post_admin.html", gin.H{
		"title": post.Title,
		"post":  post,
	})
}

// UpdatePost handles PUT /posts/:id
func (pc *AdminController) UpdatePost(c *gin.Context) {
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
