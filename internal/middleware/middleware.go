package middleware

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"salada/internal/auth"
	"salada/internal/db"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
)

func SetupMiddleware(router *gin.Engine) {
	// Trust proxies defined in environment, otherwise trust all (nil)
	// to allow X-Forwarded-For to work in containerized/proxied environments.
	tp := os.Getenv("TRUSTED_PROXIES")
	if tp != "" {
		router.SetTrustedProxies(strings.Split(tp, ","))
	} else {
		router.SetTrustedProxies(nil)
	}

	// Logger middleware
	router.Use(gin.Logger())

	allowOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowOrigins == "" {
		allowOrigins = "https://localhost" // Default for development
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
		fmt.Printf("Token verification failed: %v\n", err)
		c.Abort()
		c.Redirect(http.StatusSeeOther, redirectURL)
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

// GetClientIP returns the real client IP by checking common proxy headers first.
func GetClientIP(c *gin.Context) string {
	// 1. Check standard proxy headers manually first to bypass Gin's trust check if needed,
	// but still prioritizing them in a standard order.
	headers := []string{"X-Forwarded-For", "X-Real-IP", "Forwarded"}
	for _, h := range headers {
		if val := c.GetHeader(h); val != "" {
			if h == "Forwarded" {
				// Basic RFC 7239 parsing for 'for='
				for _, part := range strings.Split(val, ",") {
					for _, pair := range strings.Split(part, ";") {
						pair = strings.TrimSpace(pair)
						if strings.HasPrefix(strings.ToLower(pair), "for=") {
							ip := strings.Trim(pair[4:], "\"")
							if strings.HasPrefix(ip, "[") {
								// IPv6 with brackets, possibly followed by port: [addr]:port
								end := strings.Index(ip, "]")
								if end != -1 {
									return ip[1:end]
								}
							}
							// IPv4 or IPv6 without brackets
							if lastColon := strings.LastIndex(ip, ":"); lastColon != -1 {
								// Only strip if it looks like a port (exactly one colon or IPv4:port)
								if !strings.Contains(ip, "]") && (strings.Count(ip, ":") == 1 || strings.HasPrefix(ip, "::")) {
									return ip[:lastColon]
								}
							}
							return ip
						}
					}
				}
			} else {
				// For X-Forwarded-For, take the first (original client) IP
				parts := strings.Split(val, ",")
				return strings.TrimSpace(parts[0])
			}
		}
	}

	// 2. Fallback to Gin's ClientIP, which uses RemoteAddr if no trusted headers are found.
	return c.ClientIP()
}

// DBLogger is a Gin middleware that logs request data to the database
func DBLogger(c *gin.Context) {
	// 1. Capture data available BEFORE the request is handled and record start time
	start := time.Now() // Record the start time
	clientIP := GetClientIP(c)
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
// Auth Rate Limiting Middleware
// ==============================================================================

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

func init() {
	go cleanupVisitors()
}

func getVisitor(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[ip]
	if !exists {
		// 1 request per second, burst of 5
		limiter := rate.NewLimiter(rate.Limit(1), 5)
		visitors[ip] = &visitor{limiter, time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}

// AuthRateLimitMiddleware rate limits authentication attempts to prevent brute force
func AuthRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := GetClientIP(c)
		limiter := getVisitor(ip)
		if !limiter.Allow() {
			// Log the rate limit hit into access_logs
			_, err := db.DB.Exec(`
				INSERT INTO access_logs (client_ip, method, path, status_code, latency_ms)
				VALUES ($1, $2, $3, $4, $5)`,
				ip,
				c.Request.Method,
				c.Request.URL.Path,
				http.StatusTooManyRequests,
				0,
			)
			if err != nil {
				log.Printf("❌ ERROR: Failed to insert rate-limit access log into DB: %v", err)
			}

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}

