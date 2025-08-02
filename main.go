package main

import (
	"fmt"
	"log"
	"net/http"
	admin_controller "salada/internal/admin/controller"
	"salada/internal/blog/controller"
	"salada/internal/blog/repositories"
	"salada/internal/db"
	salada_session "salada/internal/sessions"
	session_repo "salada/internal/sessions/repositories"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"

	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/crypto/acme/autocert"
)

func main() {

	// Connect to the database
	db.ConnectDatabase()
	// Ensure database connection is closed when main exits
	defer db.CloseDatabase()
	router := gin.Default()

	router.Use(gin.Logger())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Replace with your actual frontend origin(s)
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "Cookie"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, // How long preflight requests can be cached
	}))

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
	sessionRepo := session_repo.NewSessionRepository(db.DB)

	router.Use(sessions.Sessions("salada_session", sessionRepo.Store))

	// Initialize blog controller with the repository instance
	blogController := controller.NewBlogController(postRepo)

	adminController := admin_controller.NewAdminController(postRepo)

	// Define routes for blog posts
	postRoutes := router.Group("/blog/")
	{

		postRoutes.GET("/", blogController.GetPosts)
		postRoutes.GET("/:slug", blogController.GetPostBySlug) // Use slug for public access
		postRoutes.POST("/image", blogController.UploadImage)
		postRoutes.DELETE("/:id", blogController.DeletePost)
		postRoutes.PUT("/:id", blogController.UpdatePost)
		postRoutes.POST("/", blogController.CreatePost)
		postRoutes.GET("/new", blogController.GetNewPostForm)
		postRoutes.GET("/edit/:slug", blogController.EditPostForm)
	}

	//Define admin routes
	admin := router.Group("/admin", gin.BasicAuth(gin.Accounts{
		"foo": "bar",
	}))
	{
		admin.GET("/blog", adminController.GetPendingPosts)
		admin.GET("/blog/:slug", adminController.GetPostBySlug)
		admin.GET("/", adminController.GetAdminMain, salada_session.SetSessionValueMiddleware("role", "admin"))
	}

	//bindIp := fmt.Sprintf("%s:8080", os.Getenv("BIND_IP"))
	//router.Run(bindIp)
	m := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist("salada.dev"),
		Cache:      autocert.DirCache("/var/www/.cache"),
	}

	log.Fatal(autotls.RunWithManager(router, &m))
}
