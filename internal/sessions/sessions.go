package sessions

import (
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func AdminAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		// 1. Check if a session exists (user is logged in)
		role := session.Get("role")
		if role == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required. Please log in."})
			c.Abort() // Stop processing further handlers
			return
		}

		// 2. Check the user's role
		if role.(string) != "admin" { // Check if role exists and is "admin"
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: Admin privileges required."})
			c.Abort() // Stop processing further handlers
			return
		}

		// If both checks pass, the user is an authenticated admin.
		c.Next() // Proceed to the next handler in the chain
	}
}

// AuthRequired is a middleware to check if a user is authenticated.
func AuthRequired(c *gin.Context) {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}
	c.Next()
}

func SetSessionValueMiddleware(key string, value any) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		if session.Get(key) == nil {
			log.Printf("Setting session value: %s = %v\n", key, value)
			session.Set(key, value)
			// It's crucial to save the session after modifying it
			if err := session.Save(); err != nil {
				log.Printf("Error saving session in middleware: %v\n", err)
			}
		}
		c.Next()
	}
}
