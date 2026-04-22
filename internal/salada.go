package salada

import (
	"net/http"
	"os"
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
	router.LoadHTMLGlob("web/templates/html/*/*")
}

// setupPublicPages defines simple GET routes for static HTML pages.
func setupPublicPages(router *gin.Engine) {
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "pages/index.html", gin.H{
			"title":        "Home",
			"is_logged_in": c.GetBool("is_logged_in"),
		})
	})

	router.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "pages/about.html", gin.H{
			"title":        "About",
			"is_logged_in": c.GetBool("is_logged_in"),
		})
	})

	router.GET("/login", func(c *gin.Context) {
		// Already authenticated — no need to show the login page
		if c.GetBool("is_logged_in") {
			intendedRoute := c.Query("goto")
			if intendedRoute == "" {
				intendedRoute = "/"
			}
			c.Redirect(http.StatusSeeOther, intendedRoute)
			return
		}

		intendedRoute := c.Query("goto")

		if intendedRoute == "" {
			intendedRoute = "/blog"
		}

		c.HTML(http.StatusOK, "auth/login.html", gin.H{
			"title": "Login",
			// Pass the variable so the template can access it as {{.IntendedRoute}}
			"IntendedRoute": intendedRoute,
			"is_logged_in":  false,
		})
	})

	router.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "auth/register.html", gin.H{
			"title":        "Register",
			"is_logged_in": c.GetBool("is_logged_in"),
		})
	})

	router.GET("/contact", func(c *gin.Context) {
		c.HTML(http.StatusOK, "pages/contact.html", gin.H{
			"title":        "Contact",
			"is_logged_in": c.GetBool("is_logged_in"),
		})
	})

	router.GET("/thankyou", func(c *gin.Context) {
		c.HTML(http.StatusOK, "pages/thankyou.html", gin.H{
			"title":        "Thank you",
			"is_logged_in": c.GetBool("is_logged_in"),
		})
	})
}

// setupErrorHandlers configures the 404 and 405 responses.
func setupErrorHandlers(router *gin.Engine) {
	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{"title": "404"})
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

	// 2. Load Auth Configuration
	mode := os.Getenv("MODE")
	authConfig := auth_controller.AuthConfig{
		CookieDomain:   os.Getenv("SALADA_HOST"),
		CookieSecure:   mode == "prod", // Default
		CookieSameSite: http.SameSiteLaxMode,
	}

	// In dev mode, if we are running on port 443 (TLS), we should allow secure cookies.
	// Also, if we are on localhost, we might want to unset CookieDomain.
	if mode == "dev" {
		authConfig.CookieSecure = true // We are using RunTLS in dev now
		authConfig.CookieDomain = ""   // Allow localhost/IP
	}

	// 3. Initialize Controllers
	blogController := blog_controller.NewBlogController(postRepo)
	authController := auth_controller.NewAuthController(adminRepo, authConfig)
	adminController := admin_controller.NewAdminController(adminRepo)

	// 4. Register Modular Routes
	blog_routes.BlogRoutes(router, blogController)
	auth_routes.AuthRoutes(router, authController)
	admin_routes.AdminRoutes(router, adminController, blogController)
}
