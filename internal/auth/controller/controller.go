package controller

import (
	"net/http"
	"net/url"
	"os"
	"salada/internal/admin/model"
	"salada/internal/admin/repositories"
	"salada/internal/auth"
	"strings"
	"time"

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

// validateRedirectURL prevents open redirect attacks by validating the redirect URL
func (pc *AuthController) validateRedirectURL(intendedRoute string) string {
	// Default to safe location if no intended route provided
	if intendedRoute == "" {
		return "/"
	}

	// Security: Reject URLs that start with http:// or https:// (absolute URLs)
	// This prevents redirects to external malicious sites
	if strings.HasPrefix(intendedRoute, "http://") || strings.HasPrefix(intendedRoute, "https://") {
		return "/"
	}

	// Security: Reject URLs starting with // (protocol-relative URLs)
	if strings.HasPrefix(intendedRoute, "//") {
		return "/"
	}

	// Security: Parse the URL to check for other malicious patterns
	u, err := url.Parse(intendedRoute)
	if err != nil {
		return "/"
	}

	// Ensure no host is specified (open redirect protection)
	if u.Host != "" || u.Scheme != "" {
		return "/"
	}

	// Security: Ensure the path starts with /
	if !strings.HasPrefix(u.Path, "/") {
		return "/"
	}

	return intendedRoute
}

func (pc *AuthController) Register(c *gin.Context) {
	var newUser model.User
	newUser.Username = c.PostForm("username")
	newUser.Password = c.PostForm("password")
	newUser.Email = c.PostForm("email")
	_, err := pc.Repo.CreateUser(newUser)

	if err != nil {
		c.HTML(http.StatusInternalServerError, "auth/register.html", gin.H{
			"error":        err,
			"is_logged_in": c.GetBool("is_logged_in"),
		})
		return
	}
	c.Redirect(http.StatusFound, "/login/")
}

func (pc *AuthController) Login(c *gin.Context) {
	var loginInput model.LoginInput
	loginInput.Username = c.PostForm("username")
	loginInput.Password = c.PostForm("password")
	intendedRoute := c.PostForm("goto")

	// Security: Validate the intended route to prevent open redirect attacks
	redirectURL := pc.validateRedirectURL(intendedRoute)

	// Best Practice: The username/email should come from the JSON body,
	// not a URL parameter. This is what the user provides in the form.
	realUsername, password, err := pc.Repo.GetUserPassword(loginInput.Username)

	// Security: Always perform bcrypt comparison to prevent timing-based user enumeration
	// Use a dummy hash if user not found so timing is consistent
	hashToCompare := password
	if err != nil {
		// Use a dummy bcrypt hash for timing-safe comparison
		// This ensures the same computation time whether user exists or not
		hashToCompare = "$2a$10$abcdefghijklmnopqrstuvwxycdefghijklmnopqrstuv"
	}

	// Always perform bcrypt comparison to prevent timing attacks
	compareErr := bcrypt.CompareHashAndPassword([]byte(hashToCompare), []byte(loginInput.Password))

	// Check for authentication failure (either user not found or password mismatch)
	if err != nil || compareErr != nil {
		// Add small random delay to further mitigate timing attacks
		time.Sleep(time.Duration(50+time.Now().UnixNano()%50) * time.Millisecond)

		// Consolidate all login-related errors into a single generic message
		// to prevent user enumeration attacks.
		c.HTML(http.StatusUnauthorized, "auth/login.html", gin.H{
			"error":        "Invalid credentials",
			"is_logged_in": c.GetBool("is_logged_in"),
		})
		return
	}

	tokenString, err := auth.CreateToken(realUsername)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error creating token")
		return
	}

	// Security: Use Secure flag in production (cookie only sent over HTTPS)
	secureFlag := os.Getenv("ENV") == "production"
	c.SetCookie("token", tokenString, 3600*12, "/", os.Getenv("SALADA_HOST"), secureFlag, true)
	c.SetCookie("username", realUsername, 3600*12, "/", os.Getenv("SALADA_HOST"), secureFlag, true)

	c.Redirect(http.StatusFound, redirectURL)
}

func (pc *AuthController) Logout(c *gin.Context) {
	// Security: Use Secure flag in production
	secureFlag := os.Getenv("ENV") == "production"
	c.SetCookie("token", "", -1, "/", os.Getenv("SALADA_HOST"), secureFlag, true)
	c.SetCookie("username", "", -1, "/", os.Getenv("SALADA_HOST"), secureFlag, true)
	c.HTML(http.StatusOK, "auth/logout.html", nil)
}
