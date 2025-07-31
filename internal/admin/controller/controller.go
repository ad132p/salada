package admin_controller

import (
	"fmt"
	"log"
	"net/http"
	"salada/internal/blog/repositories"

	"github.com/gin-gonic/gin"
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
