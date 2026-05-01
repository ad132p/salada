package main

import (
	_ "github.com/joho/godotenv/autoload"
	salada "salada/internal"
	"salada/internal/db"
	"salada/internal/middleware"
	"salada/internal/server"

	"github.com/gin-gonic/gin"
)

func main() {

	// Connect to the database
	db.ConnectDatabase()
	// Ensure database connection is closed when main exits
	defer db.CloseDatabase()
	router := gin.New()
	gin.SetMode(gin.DebugMode)

	//Setup middleware configuration
	middleware.SetupMiddleware(router)

	// Setup routes
	salada.SetupRoutes(router)

	server.Run(router)
}
