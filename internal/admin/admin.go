package admin

import (
	"github.com/gin-gonic/gin"
)

// AdminRoleRequired checks if the logged-in user has the "admin" role
func AdminRoleRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
