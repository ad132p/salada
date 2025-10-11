package routes

import (
	"salada/internal/blog/controller"
	"salada/internal/middleware"

	"github.com/gin-gonic/gin"
)

// BlogRoutes initialises and defines all blog-related routes.
func BlogRoutes(router *gin.Engine, blogController *controller.BlogController) {
	// Note: I've kept the middleware logic here since it's intrinsic to the route group.
	postRoutes := router.Group("/blog", middleware.AuthenticateMiddleware)
	{
		postRoutes.GET("/", blogController.GetPosts)
		postRoutes.POST("/", blogController.CreatePost)
		postRoutes.GET("/:slug", blogController.GetPostBySlug)
		postRoutes.POST("/uploads", blogController.UploadImage)
		postRoutes.DELETE("/:id", blogController.DeletePost)
		postRoutes.PUT("/:id", blogController.UpdatePost)
		postRoutes.GET("/new", blogController.GetNewPostForm)
		postRoutes.GET("/edit/:slug", blogController.EditPostForm)
		postRoutes.PATCH("/publish/:id", blogController.PublishPost)
		postRoutes.GET("/category/:name", blogController.GetCategory)
		postRoutes.GET("/tags/:name", blogController.GetTag)
		postRoutes.GET("/search", blogController.GetTagOrContent)
	}
}
