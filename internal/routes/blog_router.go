package routes

import (
	"salada/internal/blog/controller"
	"salada/internal/middleware"

	"github.com/gin-gonic/gin"
)

func BlogRoutes(router *gin.Engine, blogController *controller.BlogController) {
	postRoutes := router.Group("/blog")
	postRoutes.Use(middleware.DBLoggerMiddleware())
	{
		postRoutes.GET("/", blogController.GetPosts)
		postRoutes.POST("/", middleware.AuthenticateMiddleware, blogController.CreatePost)
		postRoutes.POST("/comment/:id", blogController.CreateComment)
		postRoutes.GET("/:slug", blogController.GetPostBySlug)
		postRoutes.GET("/comment/:slug", blogController.GetCommentsBySlug)
		postRoutes.POST("/uploads", middleware.AuthenticateMiddleware, blogController.UploadImage)
		postRoutes.POST("/like", blogController.LikePost)
		postRoutes.DELETE("/:id", middleware.AuthenticateMiddleware, blogController.DeletePost)
		postRoutes.PUT("/:id", middleware.AuthenticateMiddleware, blogController.UpdatePost)
		postRoutes.GET("/new", middleware.AuthenticateMiddleware, blogController.GetNewPostForm)
		postRoutes.PATCH("/publish/:id", blogController.PublishPost)
		postRoutes.GET("/category/:name", blogController.GetCategory)
		postRoutes.GET("/tags/:name", blogController.GetTag)
		postRoutes.GET("/search", blogController.GetTagOrContent)
	}
}
