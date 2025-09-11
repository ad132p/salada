package model

// UserData holds the data extracted from the token claims.
type UserData struct {
	Username string
	Role     string
}
