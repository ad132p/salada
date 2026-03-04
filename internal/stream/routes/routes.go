package routes

import (
	"salada/internal/middleware"
	"salada/internal/stream/controller"

	"github.com/gin-gonic/gin"
)

// StreamRoutes registers all stream-related routes
func StreamRoutes(router *gin.Engine, streamController *controller.StreamController) {
	streamRoutes := router.Group("/stream")
	streamRoutes.Use(middleware.DBLoggerMiddleware())
	{
		// Page route - serves the React UI
		streamRoutes.GET("", streamController.GetStreamPage)

		// WebSocket endpoint for video streaming
		streamRoutes.GET("/ws", streamController.HandleWebSocket)

		// API endpoint to get connected client count
		streamRoutes.GET("/clients", streamController.GetConnectedClients)
	}
}
