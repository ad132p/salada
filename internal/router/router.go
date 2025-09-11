package router

import (
	"net/http"
	admin "salada/internal/admin"
	admin_controller "salada/internal/admin/controller"
	admin_repositories "salada/internal/admin/repositories"
	auth_controller "salada/internal/auth/controller"
	"salada/internal/blog/controller"
	"salada/internal/blog/repositories"
	"salada/internal/db"
	"salada/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {

	// Setup static routes and assets
	router.Static("/assets/", "./web/assets")
	router.Static("/images/", "./web/images")
	router.Static("/uploads/", "./uploads")
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

	router.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"title": "Register",
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
	// Initialize repositories
	postRepo := repositories.NewPostRepository(db.DB)
	adminRepo := admin_repositories.NewAdminRepository(db.DB)

	// Initialize controllers
	blogController := controller.NewBlogController(postRepo)
	authController := auth_controller.NewAuthController(adminRepo)
	adminController := admin_controller.NewAdminController(adminRepo)

	// Define routes for blog posts
	postRoutes := router.Group("/blog", middleware.AuthenticateMiddleware)
	{
		postRoutes.GET("/", blogController.GetPosts)
		postRoutes.GET("/:slug", blogController.GetPostBySlug) // Use slug for public access
		postRoutes.POST("/uploads", blogController.UploadImage)
		postRoutes.DELETE("/:id", blogController.DeletePost)
		postRoutes.PUT("/:id", blogController.UpdatePost)
		postRoutes.POST("/", blogController.CreatePost)
		postRoutes.GET("/new", blogController.GetNewPostForm)
		postRoutes.GET("/edit/:slug", blogController.EditPostForm)
		postRoutes.PATCH("/publish/:id", blogController.PublishPost)
	}

	//Define auth routes:
	router.POST("/register", authController.Register)
	router.POST("/login", authController.Login)
	router.GET("/logout", authController.Logout)

	//Define admin routes
	admin := router.Group("/admin", admin.AdminRoleRequired())
	{
		admin.GET("/blog", adminController.GetPendingPosts)
		admin.GET("/blog/:slug", adminController.GetPostBySlug)
		admin.GET("/", adminController.GetAdminMain)
	}
}
