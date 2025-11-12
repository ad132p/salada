package routes

import (
	admin "salada/internal/admin"
	admin_controller "salada/internal/admin/controller"
	"salada/internal/blog/controller" // Need blog controller for EditPostForm
	"salada/internal/middleware"

	"github.com/gin-gonic/gin"
)

// AdminRoutes initialises and defines all admin-only routes.
func AdminRoutes(router *gin.Engine, adminController *admin_controller.AdminController, blogController *controller.BlogController) {
	adminGroup := router.Group("/admin", middleware.AuthenticateMiddleware, admin.AdminRoleRequired)
	adminGroup.Use(middleware.DBLoggerMiddleware())
	{
		adminGroup.GET("/blog", adminController.GetPendingPosts)
		adminGroup.GET("/blog/:slug", adminController.GetPostBySlug)
		adminGroup.GET("/blog/edit/:slug", blogController.EditPostForm)
		adminGroup.GET("/", adminController.GetAdminMain)
	}
}
