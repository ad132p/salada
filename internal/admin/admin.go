package admin

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// AdminRoleRequired checks if the logged-in user has the "admin" role
func AdminRoleRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		role := session.Get("role") // Assuming you store user role in the session

		if role == nil || role.(string) != "admin" { // Type assertion `.(string)`
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. Insufficient privileges."})
			c.Abort()
			return
		}
		c.Next()
	}
}
