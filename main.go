package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"salada/internal/blog/controller"
	"salada/internal/blog/repositories"
	"salada/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, assuming environment variables are set.")
	}

	// Connect to the database
	db.ConnectDatabase()
	// Ensure database connection is closed when main exits
	defer db.CloseDatabase()
	router := gin.Default()

	router.Static("/assets/", "./web/assets")
	router.Static("/images/", "./web/images")
	router.LoadHTMLGlob("web/templates/html/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Home",
		})
	})
	router.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "about.html", gin.H{
			"title": "About",
		})
	})

	router.GET("/contact", func(c *gin.Context) {
		c.HTML(http.StatusOK, "contact.html", gin.H{
			"title": "Contact",
		})
	})

	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", gin.H{
			"title": "404",
		})
	})

	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"code": "METHOD_NOT_ALLOWED", "message": "405 method not allowed"})
	})

	// Initialize repository with the *sql.DB instance
	postRepo := repositories.NewPostRepository(db.DB)

	// Initialize controller with the repository instance
	postController := controller.NewPostController(postRepo)

	// Define routes for blog posts
	postRoutes := router.Group("/blog/")
	{
		postRoutes.POST("/", postController.CreatePost)
		postRoutes.GET("/", postController.GetPosts)
		postRoutes.GET("/:slug", postController.GetPostBySlug) // Use slug for public access
		postRoutes.PUT("/:id", postController.UpdatePost)
		postRoutes.DELETE("/:id", postController.DeletePost)
	}

	bindIp := fmt.Sprintf("%s:8080", os.Getenv("BIND_IP"))
	router.Run(bindIp)
}
