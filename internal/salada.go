package salada

import (
	"net/http"
	admin_controller "salada/internal/admin/controller"
	admin_repositories "salada/internal/admin/repositories"
	admin_routes "salada/internal/admin/routes"
	auth_controller "salada/internal/auth/controller"
	auth_routes "salada/internal/auth/routes"
	blog_controller "salada/internal/blog/controller"
	blog_repositories "salada/internal/blog/repositories"
	blog_routes "salada/internal/blog/routes"
	"salada/internal/db"

	// NEW: Import the routes package
	"github.com/gin-gonic/gin"
)

// SetupRoutes handles all router configuration and dependency injection.
func SetupRoutes(router *gin.Engine) {
	setupConfiguration(router)
	setupPublicPages(router)
	setupErrorHandlers(router)
	setupDependenciesAndGroups(router)
}

// setupConfiguration sets up static files, templates, etc.
func setupConfiguration(router *gin.Engine) {
	router.Static("/assets/", "./web/assets")
	router.Static("/images/", "./web/images")
	router.Static("/uploads/", "./uploads")
	router.LoadHTMLGlob("web/templates/html/*")
}

// setupPublicPages defines simple GET routes for static HTML pages.
func setupPublicPages(router *gin.Engine) {
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "Home"})
	})

	router.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "about.html", gin.H{"title": "About"})
	})

	router.GET("/login", func(c *gin.Context) {
		intendedRoute := c.Query("goto")

		if intendedRoute == "" {
			intendedRoute = "/blog"
		}

		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "Login",
			// Pass the variable so the template can access it as {{.IntendedRoute}}
			"IntendedRoute": intendedRoute,
		})
	})

	router.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", gin.H{"title": "Register"})
	})

	router.GET("/contact", func(c *gin.Context) {
		c.HTML(http.StatusOK, "contact.html", gin.H{"title": "Contact"})
	})

	router.GET("/thankyou", func(c *gin.Context) {
		c.HTML(http.StatusOK, "thankyou.html", gin.H{"title": "Thank you"})
	})
}

// setupErrorHandlers configures the 404 and 405 responses.
func setupErrorHandlers(router *gin.Engine) {
	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", gin.H{"title": "404"})
	})

	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"code": "METHOD_NOT_ALLOWED", "message": "405 method not allowed"})
	})
}

// setupDependenciesAndGroups initializes all dependencies and registers the modular route groups.
func setupDependenciesAndGroups(router *gin.Engine) {
	// 1. Initialize Repositories
	postRepo := blog_repositories.NewPostRepository(db.DB)
	adminRepo := admin_repositories.NewAdminRepository(db.DB)

	// 2. Initialize Controllers
	blogController := blog_controller.NewBlogController(postRepo)
	authController := auth_controller.NewAuthController(adminRepo)
	adminController := admin_controller.NewAdminController(adminRepo)

	// 3. Register Modular Routes
	blog_routes.BlogRoutes(router, blogController)
	auth_routes.AuthRoutes(router, authController)
	admin_routes.AdminRoutes(router, adminController, blogController)
}
