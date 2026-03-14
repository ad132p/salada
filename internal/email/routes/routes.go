package routes

import (
	"salada/internal/email/controller"
	"salada/internal/middleware"

	"github.com/gin-gonic/gin"
)

// EmailRoutes registers all email-related routes
func EmailRoutes(router *gin.Engine, emailController *controller.EmailController) {
	emailRoutes := router.Group("/email")
	emailRoutes.Use(middleware.DBLoggerMiddleware())
	{
		// Send a simple email (JSON API)
		emailRoutes.POST("/send", middleware.AuthenticateMiddleware, emailController.SendEmail)

		// Send email with attachments (multipart/form-data)
		emailRoutes.POST("/send-with-attachments", middleware.AuthenticateMiddleware, emailController.SendEmailWithAttachments)

		// Get SMTP configuration
		emailRoutes.GET("/config", middleware.AuthenticateMiddleware, emailController.GetSMTPConfig)

		// Test SMTP connection
		emailRoutes.POST("/test", middleware.AuthenticateMiddleware, emailController.TestConnection)

		// Get email status by ID
		emailRoutes.GET("/status/:id", middleware.AuthenticateMiddleware, emailController.GetEmailStatus)
	}

	// Public email page (if needed)
	router.GET("/contact-email", middleware.DBLoggerMiddleware(), emailController.GetEmailPage)
}
