package controller

import (
	"net/http"
	"salada/internal/admin/model"
	"salada/internal/admin/repositories"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AuthController handles auth related requests.
type AuthController struct {
	Repo *repositories.AdminRepository
}

// NewPostController creates a new PostController instance.
func NewAuthController(repo *repositories.AdminRepository) *AuthController {
	return &AuthController{Repo: repo}
}

func (pc *AuthController) Register(c *gin.Context) {
	var newUser model.User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	username := c.Param("username")
	id, err := pc.Repo.CreateUser(username)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user_id": id})
}

func (pc *AuthController) Login(c *gin.Context) {
	var loginInput model.LoginInput
	if err := c.ShouldBindJSON(&loginInput); err != nil {
		// Provide more detailed error message for better debugging
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Best Practice: The username/email should come from the JSON body,
	// not a URL parameter. This is what the user provides in the form.
	user, err := pc.Repo.GetUserCredentials(loginInput.Username)
	if err != nil {
		// Consolidate all login-related errors into a single "Invalid credentials" message
		// to prevent user enumeration attacks.
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Compare the provided password with the stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginInput.Password)); err != nil {
		// Return the same generic error for password mismatch as for user not found
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

}
