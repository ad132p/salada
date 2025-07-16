package main

import (
	"fmt"
	"net/http"
	"os"
	"salada/internal/blog/controller"
	"salada/internal/blog/repositories"
	"salada/internal/db"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

var secrets = gin.H{
	"admin": gin.H{"email": "foo@bar.com", "phone": "123433"},
}

func main() {

	// Connect to the database
	db.ConnectDatabase()
	// Ensure database connection is closed when main exits
	defer db.CloseDatabase()
	router := gin.Default()

	router.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	router.Use(gin.Recovery())

	// Filesize limits global
	router.MaxMultipartMemory = 8 << 20

	router.Static("/assets/", "./web/assets")
	router.Static("/images/", "./web/images")
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.POST("/upload", func(c *gin.Context) {
		file, _ := c.FormFile("file")
		c.SaveUploadedFile(file, "./files/"+file.Filename)
		c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
	})
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

	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "Login",
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

	// Initialize blog controller with the repository instance
	blogController := controller.NewBlogController(postRepo)

	// Define routes for blog posts
	postRoutes := router.Group("/blog/")
	{
		postRoutes.POST("/", blogController.CreatePost)
		postRoutes.GET("/", blogController.GetPosts)
		postRoutes.GET("/:slug", blogController.GetPostBySlug) // Use slug for public access
		postRoutes.PUT("/:id", blogController.UpdatePost)
		postRoutes.DELETE("/:id", blogController.DeletePost)
		postRoutes.POST("/image", blogController.UploadImage)

	}

	//Define admin routes
	authorized := router.Group("/admin", gin.BasicAuth(gin.Accounts{
		"foo": "bar",
	}))

	authorized.GET("/blog", func(c *gin.Context) {
		// get user, it was set by the BasicAuth middleware
		user := c.MustGet(gin.AuthUserKey).(string)
		if secret, ok := secrets[user]; ok {
			c.JSON(http.StatusOK, gin.H{"user": user, "secret": secret})
		} else {
			c.HTML(http.StatusOK, "post_form.html", gin.H{
				"title": "New Blog Entry",
			})
		}
	})

	bindIp := fmt.Sprintf("%s:8080", os.Getenv("BIND_IP"))
	router.Run(bindIp)
}
