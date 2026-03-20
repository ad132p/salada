package routes

import (
	"salada/internal/auth/controller"
	"salada/internal/middleware"

	"github.com/gin-gonic/gin"
)

// AuthRoutes initialises and defines all authentication routes.
func AuthRoutes(router *gin.Engine, authController *controller.AuthController) {
	router.POST("/register", middleware.AuthRateLimitMiddleware(), authController.Register)
	router.POST("/login", middleware.AuthRateLimitMiddleware(), authController.Login)
	router.POST("/logout", middleware.AuthenticateMiddleware, authController.Logout)
}
