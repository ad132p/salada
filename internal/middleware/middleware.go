package middleware

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"salada/internal/auth"
	"salada/internal/db"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func SetupMiddleware(router *gin.Engine) {
	// Logger middleware
	router.Use(gin.Logger())

	allowOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowOrigins == "" {
		allowOrigins = "http://localhost:8080" // Default for development
	}

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Split(allowOrigins, ","),
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

	// Check Auth Status
	router.Use(CheckAuthMiddleware)
}

// Function to verify JWT tokens
func AuthenticateMiddleware(c *gin.Context) {
	// Retrieve the token from the cookie

	intendedRoute := c.Request.URL.Path
	redirectURL := "/login?goto=" + url.QueryEscape(intendedRoute)
	isWebsocket := c.GetHeader("Upgrade") == "websocket"

	tokenString, err := c.Cookie("token")
	if err != nil {
		fmt.Println("Token missing in cookie")
		if isWebsocket {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
		} else {
			c.Abort()
			c.Redirect(http.StatusSeeOther, redirectURL)
		}
		return
	}

	// Verify the token
	token, err := auth.VerifyToken(tokenString)
	if err != nil {
		fmt.Printf("Token verification failed: %v\n", err)
		if isWebsocket {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
		} else {
			c.Abort()
			c.Redirect(http.StatusSeeOther, redirectURL)
		}
		return
	}

	// Get the claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println("Could not get claims from token")
		if isWebsocket {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
			c.Redirect(http.StatusSeeOther, redirectURL)
		}
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

// CheckAuthMiddleware checks if the user is logged in but does not enforce it.
// It sets "is_logged_in" to true and "username" if the token is valid.
func CheckAuthMiddleware(c *gin.Context) {
	tokenString, err := c.Cookie("token")
	if err != nil {
		// No token, just continue
		c.Next()
		return
	}

	// Verify the token
	token, err := auth.VerifyToken(tokenString)
	if err != nil {
		// Invalid token, just continue
		c.Next()
		return
	}

	// Get the claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		// Invalid claims, just continue
		c.Next()
		return
	}

	// Token is valid
	c.Set("is_logged_in", true)

	// Access the claims and set username if present
	if username, ok := claims["sub"].(string); ok {
		c.Set("username", username)
	}

	c.Next()
}

// ==============================================================================
// WebSocket Tracking Middleware
// ==============================================================================

// trackingConn wraps a net.Conn to track bytes read and written
type trackingConn struct {
	net.Conn
	bytesRead    int64
	bytesWritten int64
}

func (tc *trackingConn) Read(b []byte) (n int, err error) {
	n, err = tc.Conn.Read(b)
	if n > 0 {
		tc.bytesRead += int64(n)
	}
	return n, err
}

func (tc *trackingConn) Write(b []byte) (n int, err error) {
	n, err = tc.Conn.Write(b)
	if n > 0 {
		tc.bytesWritten += int64(n)
	}
	return n, err
}

// wsResponseWriter wraps gin.ResponseWriter to intercept the Hijack call
type wsResponseWriter struct {
	gin.ResponseWriter
	hijackedConn *trackingConn
}

// Hijack intercepts the hijacking to wrap the connection
func (w *wsResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying ResponseWriter does not support hijacking")
	}

	conn, _, err := hijacker.Hijack()
	if err != nil {
		return nil, nil, err
	}

	// Wrap the connection
	wrappedConn := &trackingConn{Conn: conn}
	w.hijackedConn = wrappedConn

	// Return the wrapped connection and a new ReadWriter that uses it
	return wrappedConn, bufio.NewReadWriter(bufio.NewReader(wrappedConn), bufio.NewWriter(wrappedConn)), nil
}

// WSTrackingMiddleware tracks metrics for WebSocket connections
func WSTrackingMiddleware(c *gin.Context) {
	// 1. Capture start time and request details
	start := time.Now()
	clientIP := c.ClientIP()
	path := c.Request.URL.Path

	// 2. Wrap the ResponseWriter to intercept Hijack()
	wrappedWriter := &wsResponseWriter{ResponseWriter: c.Writer}
	c.Writer = wrappedWriter

	// 3. Process the request
	c.Next()

	// 4. Check if the connection was hijacked (i.e. it became a WebSocket)
	if wrappedWriter.hijackedConn != nil {
		// Calculate duration
		duration := time.Since(start).Milliseconds()

		bytesRead := wrappedWriter.hijackedConn.bytesRead
		bytesWritten := wrappedWriter.hijackedConn.bytesWritten

		// Calculate watcher
		watcher := "anon"
		if u, ok := c.Get("username"); ok {
			if s, ok := u.(string); ok && s != "" {
				watcher = s
			}
		}

		// Calculate streamer
		streamer := watcher // Default to self if creating stream
		if roomParam := c.Query("room"); roomParam != "" {
			streamer = roomParam
		}

		// 5. Log to PostgreSQL
		_, err := db.DB.Exec(`
			INSERT INTO ws_metrics (client_ip, path, bytes_read, bytes_written, duration_ms, streamer, watcher)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			clientIP,
			path,
			bytesRead,
			bytesWritten,
			duration,
			streamer,
			watcher,
		)
		if err != nil {
			log.Printf("❌ ERROR: Failed to insert WS metric log into DB: %v", err)
		} else {
			log.Printf("WS Connection closed (IP: %s, Path: %s, Streamer: %s, Watcher: %s): %d bytes read, %d bytes written, duration: %d ms",
				clientIP, path, streamer, watcher, bytesRead, bytesWritten, duration)
		}
	}
}
