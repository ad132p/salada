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
		// Page route - serves the React UI (protected - requires login)
		streamRoutes.GET("", middleware.AuthenticateMiddleware, streamController.GetStreamPage)

		// WebSocket endpoint for video streaming (protected - requires login)
		streamRoutes.GET("/ws", middleware.AuthenticateMiddleware, streamController.HandleWebSocket)

		// API endpoint to get connected client count
		streamRoutes.GET("/clients", streamController.GetConnectedClients)

		// API endpoint to get active rooms
		streamRoutes.GET("/rooms", streamController.GetActiveRooms)
	}

	// Public rooms page - shows all active streams
	router.GET("/rooms", middleware.DBLoggerMiddleware(), streamController.GetRoomsPage)

	// Public room watch page - watch a specific stream by username
	router.GET("/rooms/:username", middleware.DBLoggerMiddleware(), streamController.GetWatchRoomPage)
}
