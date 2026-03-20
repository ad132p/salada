package routes

import (
	"salada/internal/auth/controller"
	"salada/internal/middleware"

	"github.com/gin-gonic/gin"
)

// AuthRoutes initialises and defines all authentication routes.
func AuthRoutes(router *gin.Engine, authController *controller.AuthController) {
	router.POST("/register", authController.Register)
	router.POST("/login", authController.Login)
	router.POST("/logout", middleware.AuthenticateMiddleware, authController.Logout)
}
