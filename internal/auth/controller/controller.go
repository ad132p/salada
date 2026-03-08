package controller

import (
	"log"
	"net/http"
	"net/url"
	"salada/internal/admin/model"
	"salada/internal/admin/repositories"
	"salada/internal/auth"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AuthConfig holds configuration for the auth controller
type AuthConfig struct {
	CookieDomain   string
	CookieSecure   bool
	CookieSameSite http.SameSite
}

// AuthController handles auth related requests.
type AuthController struct {
	Repo   *repositories.AdminRepository
	Config AuthConfig
}

// dummyHash is a valid bcrypt hash used for timing-safe comparison
// when a user is not found. This prevents user enumeration attacks.
//nolint:gochecknoglobals
var dummyHash string

func init() {
	// Generate a real bcrypt hash once at startup for timing-safe comparisons.
	// Using a valid hash ensures CompareHashAndPassword takes consistent time
	// regardless of whether the user exists.
	hash, err := bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing-mitigation"), bcrypt.DefaultCost)
	if err != nil {
		// This should never happen, but if it does, we need to know immediately
		panic("failed to generate dummy bcrypt hash: " + err.Error())
	}
	dummyHash = string(hash)
}

// NewAuthController creates a new AuthController instance.
func NewAuthController(repo *repositories.AdminRepository, config AuthConfig) *AuthController {
	return &AuthController{
		Repo:   repo,
		Config: config,
	}
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
	if err := c.ShouldBind(&newUser); err != nil {
		c.HTML(http.StatusBadRequest, "auth/register.html", gin.H{
			"error":        "Invalid input",
			"is_logged_in": c.GetBool("is_logged_in"),
		})
		return
	}

	_, err := pc.Repo.CreateUser(newUser)
	if err != nil {
		// Security: Log the actual error server-side for debugging
		// but return a generic message to prevent information leakage
		log.Printf("[SECURITY] User registration failed: %v", err)
		c.HTML(http.StatusInternalServerError, "auth/register.html", gin.H{
			"error":        "Registration failed. Please try a different username or email.",
			"is_logged_in": c.GetBool("is_logged_in"),
		})
		return
	}

	c.Redirect(http.StatusFound, "/login")
}

func (pc *AuthController) Login(c *gin.Context) {
	var loginInput model.LoginInput
	if err := c.ShouldBind(&loginInput); err != nil {
		c.HTML(http.StatusBadRequest, "auth/login.html", gin.H{
			"error":        "Invalid input",
			"is_logged_in": c.GetBool("is_logged_in"),
		})
		return
	}

	// Get the intended redirect route (not part of LoginInput model)
	intendedRoute := c.PostForm("goto")

	// Security: Validate the intended route to prevent open redirect attacks
	redirectURL := pc.validateRedirectURL(intendedRoute)

	// Best Practice: The username/email should come from the form body,
	// not a URL parameter. This is what the user provides in the form.
	realUsername, password, err := pc.Repo.GetUserPassword(loginInput.Username)

	// Security: Always perform bcrypt comparison to prevent timing-based user enumeration
	// Use the pre-generated dummy hash if user not found so timing is consistent
	hashToCompare := password
	if err != nil {
		// Use the valid bcrypt dummy hash for timing-safe comparison
		// This ensures the same computation time whether user exists or not
		hashToCompare = dummyHash
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

	// Security: Set SameSite attribute to prevent CSRF attacks
	c.SetSameSite(pc.Config.CookieSameSite)
	// Set cookies with Secure flag (HTTPS only in production)
	c.SetCookie("token", tokenString, 3600*12, "/", pc.Config.CookieDomain, pc.Config.CookieSecure, true)
	c.SetCookie("username", realUsername, 3600*12, "/", pc.Config.CookieDomain, pc.Config.CookieSecure, true)

	c.Redirect(http.StatusFound, redirectURL)
}

func (pc *AuthController) Logout(c *gin.Context) {
	// Security: Set SameSite attribute to prevent CSRF attacks
	c.SetSameSite(pc.Config.CookieSameSite)
	// Clear cookies with Secure flag
	c.SetCookie("token", "", -1, "/", pc.Config.CookieDomain, pc.Config.CookieSecure, true)
	c.SetCookie("username", "", -1, "/", pc.Config.CookieDomain, pc.Config.CookieSecure, true)
	// Redirect to login page instead of rendering logout page
	// This prevents users from bookmarking or refreshing a stale logout page
	c.Redirect(http.StatusFound, "/login")
}
