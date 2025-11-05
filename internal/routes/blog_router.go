package routes

import (
	"salada/internal/blog/controller"
	"salada/internal/middleware"

	"github.com/gin-gonic/gin"
)

func BlogRoutes(router *gin.Engine, blogController *controller.BlogController) {
	postRoutes := router.Group("/blog", middleware.AuthenticateMiddleware)
	postRoutes.Use(middleware.DBLoggerMiddleware())
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
