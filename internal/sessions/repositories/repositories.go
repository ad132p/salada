package repositories

import (
	"database/sql"
	"os"

	"github.com/gin-contrib/sessions/postgres"
)

// SessionRepository defines methods for interacting with post data.
type SessionRepository struct {
	db    *sql.DB
	Store postgres.Store
}

// NewPostRepository creates a new PostRepository.
func NewSessionRepository(db *sql.DB) *SessionRepository {
	store, err := postgres.NewStore(db, []byte(os.Getenv("SESSION_SECRET")))
	if err != nil {
		// handle err
	}

	return &SessionRepository{db: db, Store: store}
}
