package admin

import (
	"fmt"
	"net/http"
	"salada/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AdminRoleRequired checks if the logged-in user has the "admin" role
// Function to verify JWT tokens
func AdminRoleRequired(c *gin.Context) {
	// Retrieve the token from the cookie
	tokenString, err := c.Cookie("token")
	if err != nil {
		fmt.Println("Token missing in cookie")
		c.Redirect(http.StatusSeeOther, "/login")
		c.Abort()
		return
	}

	// Verify the token
	token, err := auth.VerifyToken(tokenString)
	if err != nil {
		fmt.Printf("Token verification failed: %v\\n", err)
		c.Redirect(http.StatusSeeOther, "/login")
		c.Abort()
		return
	}

	// Get the claims from the token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println("Could not get claims from token")
		return
	}

	// Access the claims
	role, ok := claims["aud"].(string)
	if !ok {
		fmt.Println("Could not get role from token")
	}

	if role != "super" {
		c.HTML(http.StatusForbidden, "pages/denied.html", gin.H{
			"message": "Access Denied: You must be an administrator.",
		})
		// Abort the request to prevent the next handler from running.
		c.Abort()
		return
	}
	c.Next()
}
