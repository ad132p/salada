package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupMiddleware(router *gin.Engine) {
	// Logger middleware
	router.Use(gin.Logger())

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Replace with your actual frontend origin(s)
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "Cookie"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, // How long preflight requests can be cached
	}))

	// Recovery middleware
	router.Use(gin.Recovery())

	// Filesize limits
	router.MaxMultipartMemory = 8 << 20

}
