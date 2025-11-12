package middleware

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"salada/internal/auth"
	"salada/internal/db"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

// Function to verify JWT tokens
func AuthenticateMiddleware(c *gin.Context) {
	// Retrieve the token from the cookie

	intendedRoute := c.Request.URL.Path
	redirectURL := "/login?goto=" + url.QueryEscape(intendedRoute)

	tokenString, err := c.Cookie("token")
	if err != nil {
		fmt.Println("Token missing in cookie")
		c.Abort()
		c.Redirect(http.StatusSeeOther, redirectURL)
		return
	}

	// Verify the token
	token, err := auth.VerifyToken(tokenString)
	if err != nil {
		fmt.Printf("Token verification failed: %v\\n", err)
		c.Abort()
		c.Redirect(http.StatusSeeOther, redirectURL)
		return
	}

	// Get the claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println("Could not get claims from token")
		c.AbortWithStatus(http.StatusUnauthorized)
		c.Redirect(http.StatusSeeOther, redirectURL)
		return
	}

	// Access the claims
	username, ok := claims["sub"].(string)
	if ok {
		c.Set("username", username)
	}

	// Continue with the next middleware or route handler
	c.Next()
}

// DBLogger is a Gin middleware that logs request data to the database
func DBLogger(c *gin.Context) {
	// 1. Capture data available BEFORE the request is handled and record start time
	start := time.Now() // Record the start time
	clientIP := c.ClientIP()
	method := c.Request.Method
	path := c.Request.URL.Path

	// 2. Process the rest of the request chain
	// Execution halts here until all subsequent handlers/middleware return.
	c.Next()

	// 3. Capture data available AFTER the request is handled
	statusCode := c.Writer.Status()

	// Calculate latency (duration)
	latency := time.Since(start)
	// Convert latency to milliseconds for storage
	latencyMs := int64(latency.Milliseconds())

	// --- Log to PostgreSQL ---
	_, err := db.DB.Exec(`
        INSERT INTO access_logs (client_ip, method, path, status_code, latency_ms)
        VALUES ($1, $2, $3, $4, $5)`,
		clientIP,
		method,
		path,
		statusCode,
		latencyMs, // Insert latency into the DB
	)
	if err != nil {
		log.Printf("❌ ERROR: Failed to insert access log into DB: %v", err)
	}
}

func DBLoggerMiddleware() gin.HandlerFunc {
	return DBLogger
}
