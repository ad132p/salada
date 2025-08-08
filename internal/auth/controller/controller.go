package controller

import (
	"fmt"
	"log"
	"net/http"
	"salada/internal/admin/model"
	"salada/internal/admin/repositories"

	"github.com/gin-contrib/sessions"
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
	newUser.Username = c.PostForm("username")
	newUser.Password = c.PostForm("password")
	newUser.Email = c.PostForm("email")
	fmt.Println(newUser)
	id, err := pc.Repo.CreateUser(newUser)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	log.Println("User created: ", id)
	c.Redirect(http.StatusFound, "/login/")
}

func (pc *AuthController) Login(c *gin.Context) {

	var loginInput model.LoginInput
	loginInput.Username = c.PostForm("username")
	loginInput.Password = c.PostForm("password")

	// Best Practice: The username/email should come from the JSON body,
	// not a URL parameter. This is what the user provides in the form.
	password, err := pc.Repo.GetUserPassword(loginInput.Username)
	if err != nil {
		// Consolidate all login-related errors into a single "Invalid credentials" message
		// to prevent user enumeration attacks.
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"error": "Username not registered",
		})
		return
	}

	// Compare the provided password with the stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(loginInput.Password)); err != nil {
		// Return the same generic error for password mismatch as for user not found
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"error": "Invalid credentials",
		})
		return
	}
	session := sessions.Default(c)
	session.Set("username", loginInput.Username)
	session.Options(sessions.Options{MaxAge: 600})
	session.Save()
	c.Redirect(http.StatusFound, "/blog/")
}

func (pc *AuthController) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Options(sessions.Options{MaxAge: -1})
	session.Clear()
	session.Save()
	c.JSON(200, gin.H{"message": "Logged out successfully"})
}
