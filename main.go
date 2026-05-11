package main

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
	salada "salada/internal"
	"salada/internal/db"
	"salada/internal/middleware"
	"salada/internal/server"

	"github.com/gin-gonic/gin"
)

func main() {
	mode := os.Getenv("MODE")
	if mode == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Connect to the database
	db.ConnectDatabase()
	// Ensure database connection is closed when main exits
	defer db.CloseDatabase()

	router := gin.New()

	//Setup middleware configuration
	middleware.SetupMiddleware(router)

	// Setup routes
	salada.SetupRoutes(router)

	server.Run(router)
}
