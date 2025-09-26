package main

import (
	salada "salada/internal"
	"salada/internal/db"
	"salada/internal/middleware"
	"salada/internal/server"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

func main() {

	// Connect to the database
	db.ConnectDatabase()
	// Ensure database connection is closed when main exits
	defer db.CloseDatabase()
	router := gin.Default()
	gin.SetMode(gin.DebugMode)

	// Setup routes
	salada.SetupRoutes(router)

	//Setup middleware configuration
	middleware.SetupMiddleware(router)

	server.Run(router)

}
