package routes

import (
	"salada/internal/middleware"
	"salada/internal/stream/controller"

	"github.com/gin-gonic/gin"
)

func StreamRoutes(router *gin.Engine, streamController *controller.StreamController) {

	streamRoutes := router.Group("/stream")
	streamRoutes.Use(middleware.DBLoggerMiddleware())
	{

		// Streamer dashboard page (must be logged in)
		streamRoutes.GET("", middleware.AuthenticateMiddleware, streamController.GetStreamPage)

		// WebSocket signaling endpoint
		// Auth is optional: streamers are authenticated, viewers are public
		streamRoutes.GET("/ws", streamController.HandleWebSocket)

		// API: connected clients
		streamRoutes.GET("/clients", streamController.GetConnectedClients)

		// API: active rooms
		streamRoutes.GET("/rooms", streamController.GetActiveRooms)
	}

	// Public page listing all live streams
	router.GET(
		"/rooms",
		middleware.DBLoggerMiddleware(),
		streamController.GetRoomsPage,
	)

	// Public watch page
	router.GET(
		"/rooms/:username",
		middleware.DBLoggerMiddleware(),
		streamController.GetWatchRoomPage,
	)
}
