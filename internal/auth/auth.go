package auth

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Add a new global variable for the secret key
var secretKey []byte

func init() {
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		log.Fatal("SESSION_SECRET environment variable is required")
	}
	if len(secret) < 32 {
		log.Fatal("SESSION_SECRET must be at least 32 characters long")
	}
	secretKey = []byte(secret)
}

// Function to create JWT tokens with claims
func CreateToken(username string) (string, error) {
	// Create a new JWT token with claims
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,                         // Subject (user identifier)
		"iss": "salada",                         // Issuer
		"aud": getRole(username),                // Audience (user role)
		"exp": time.Now().Add(time.Hour).Unix(), // Expiration time
		"iat": time.Now().Unix(),                // Issued at
	})

	tokenString, err := claims.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func getRole(username string) string {
	if username == os.Getenv("SUPERUSER") {
		return "super"
	}
	return "rw_only"
}

// VerifyToken verifies the JWT token's signature and its claims.
func VerifyToken(tokenString string) (*jwt.Token, error) {
	// Parse the token with a callback function to provide the key.
	// This callback is crucial for security.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 1. Validate the signing algorithm
		// Use a type switch to ensure the token's signing method is what you expect.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// 2. Return your secret key
		return secretKey, nil
	})

	// Handle parsing errors
	if err != nil {
		return nil, fmt.Errorf("token parsing failed: %w", err)
	}

	// 3. Check if the token is valid after parsing
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// 4. Return the validated token
	return token, nil
}
