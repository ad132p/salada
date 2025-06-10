package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type BlogController struct {
}

func NewBlogController() *BlogController {
	return &BlogController{}
}

func RegisterBlogRoutes(router *gin.Engine) {
	bc := NewBlogController()

	blogGroup := router.Group("/blog/")
	{
		blogGroup.GET("/", bc.GetPosts)
		blogGroup.GET("/:id", bc.GetPostByID)
		blogGroup.POST("/", bc.CreatePost)
		// Add more user routes
	}

}

// GetUsers handles GET /users
func (bc *BlogController) GetPosts(c *gin.Context) {
	c.HTML(http.StatusOK, "blog.html", gin.H{
		"title": "Contact",
	})
	//c.JSON(http.StatusOK, gin.H{"message": "Get all posts"})
}

// GetUserByID handles GET /users/:id
func (bc *BlogController) GetPostByID(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{"message": "Get user by ID: " + id})
}

// CreateUser handles POST /users
func (bc *BlogController) CreatePost(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "Create a new user"})
}
