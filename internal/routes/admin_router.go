package routes

import (
	admin "salada/internal/admin"
	admin_controller "salada/internal/admin/controller"
	"salada/internal/blog/controller" // Need blog controller for EditPostForm

	"github.com/gin-gonic/gin"
)

// AdminRoutes initialises and defines all admin-only routes.
func AdminRoutes(router *gin.Engine, adminController *admin_controller.AdminController, blogController *controller.BlogController) {
	adminGroup := router.Group("/admin", admin.AdminRoleRequired)
	{
		adminGroup.GET("/blog", adminController.GetPendingPosts)
		adminGroup.GET("/blog/:slug", adminController.GetPostBySlug)
		// You might consider moving the blogController dependency to its own handler in the admin module
		adminGroup.GET("/blog/edit/:slug", blogController.EditPostForm)
		adminGroup.GET("/", adminController.GetAdminMain)
	}
}
