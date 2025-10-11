package repositories

import (
	"database/sql"
	"fmt"

	"salada/internal/blog"
	"salada/internal/blog/model"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// PostRepository defines methods for interacting with post data.
type PostRepository struct {
	db *sql.DB
}

// NewPostRepository creates a new PostRepository.
func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

type AdminRepository struct {
	db *sql.DB
}

// NewAdminRepository creates a new AdminRepository.
func NewAdminRepository(db *sql.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

// CreatePost inserts a new post into the database.
func (r *PostRepository) CreatePost(post *model.Post) (uuid.UUID, error) {
	// Set UUID if not already set (e.g., if client provides it)
	fmt.Println(post)
	if post.ID == uuid.Nil {
		post.ID = uuid.New()
	}
	// Set creation/update timestamps
	post.CreatedAt = time.Now().UTC()
	post.UpdatedAt = post.CreatedAt
	post.Slug = blog.CreateSlug(post.Title)
	post.Tags = blog.SplitWithoutEmpty(post.TagsString, ",")

	query := `INSERT INTO posts (title, slug, content, author_name, published_at, created_at, updated_at, tags, category)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, created_at, updated_at`

	// Use QueryRow to get back the generated ID and timestamps (if DB generates)
	// Or use Exec if you're setting ID in Go and don't need returns
	err := r.db.QueryRow(query,
		post.Title,
		post.Slug,
		post.Content,
		post.AuthorName,  // Will be NULL if *uuid.UUID is nil
		post.PublishedAt, // Will be NULL if *time.Time is nil
		post.CreatedAt,
		post.UpdatedAt,
		pq.Array(post.Tags),
		post.Category,
	).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt) // Scan the returned values

	return post.ID, err
}

// GetPosts fetches all posts from the database.
func (r *PostRepository) GetPosts() ([]model.Post, error) {
	query := `SELECT id, title, slug, content, author_id, published_at, created_at, updated_at, tags, category FROM posts ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var post model.Post
		// Scan into post fields. Use sql.Null* types for nullable columns.
		var authorID sql.Null[uuid.UUID]
		var publishedAt sql.NullTime

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Slug,
			&post.Content,
			&post.AuthorName,
			&post.PublishedAt,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Tags,
			&post.Category,
		)
		if err != nil {
			return nil, err
		}

		// Assign nullable fields
		if authorID.Valid {
			post.AuthorID = &authorID.V
		} else {
			post.AuthorID = nil
		}
		if publishedAt.Valid {
			post.PublishedAt = &publishedAt.Time
		} else {
			post.PublishedAt = nil
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// GetPublishedPosts fetches all published posts from the database.
func (r *PostRepository) GetPublishedPosts() ([]model.Post, error) {
	query := `SELECT id, title, slug, content, author_id, author_name, published_at, created_at, updated_at, tags, category FROM posts
	WHERE published_at IS NOT NULL ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var post model.Post
		// Scan into post fields. Use sql.Null* types for nullable columns.
		var authorID sql.Null[uuid.UUID]

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Slug,
			&post.Content,
			&post.AuthorID,
			&post.AuthorName,
			&post.PublishedAt,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Tags,
			&post.Category,
		)
		if err != nil {
			return nil, err
		}

		// Assign nullable fields
		if authorID.Valid {
			post.AuthorID = &authorID.V
		} else {
			post.AuthorID = nil
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// GetPublishedPosts fetches all published posts from the database.
func (r *PostRepository) GetCategoryCount() ([]model.CategoryCount, error) {
	query := `SELECT category, count(category) FROM posts WHERE published_at IS NOT NULL GROUP BY category`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categoryCountSet []model.CategoryCount
	for rows.Next() {
		var categoryCount model.CategoryCount
		err := rows.Scan(
			&categoryCount.Category,
			&categoryCount.Count,
		)
		if err != nil {
			return nil, err
		}
		categoryCountSet = append(categoryCountSet, categoryCount)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categoryCountSet, nil
}

func (r *PostRepository) GetPublishedPostsByCategory(category string) ([]model.Post, error) {
	query := `SELECT id, title, slug, content, author_id, author_name, published_at, created_at, updated_at, tags, category FROM posts
	WHERE published_at IS NOT NULL 
	AND category = $1
	ORDER BY created_at DESC`
	rows, err := r.db.Query(query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var post model.Post
		// Scan into post fields. Use sql.Null* types for nullable columns.
		var authorID sql.Null[uuid.UUID]

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Slug,
			&post.Content,
			&post.AuthorID,
			&post.AuthorName,
			&post.PublishedAt,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Tags,
			&post.Category,
		)
		if err != nil {
			return nil, err
		}

		// Assign nullable fields
		if authorID.Valid {
			post.AuthorID = &authorID.V
		} else {
			post.AuthorID = nil
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *PostRepository) GetPublishedPostsByTag(tag string) ([]model.Post, error) {
	query := `SELECT id, title, slug, content, author_id, author_name, published_at, created_at, updated_at, tags, category FROM posts
	WHERE published_at IS NOT NULL 
	AND $1 = ANY(tags)
	ORDER BY created_at DESC`
	rows, err := r.db.Query(query, tag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var post model.Post
		// Scan into post fields. Use sql.Null* types for nullable columns.
		var authorID sql.Null[uuid.UUID]

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Slug,
			&post.Content,
			&post.AuthorID,
			&post.AuthorName,
			&post.PublishedAt,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Tags,
			&post.Category,
		)
		if err != nil {
			return nil, err
		}

		// Assign nullable fields
		if authorID.Valid {
			post.AuthorID = &authorID.V
		} else {
			post.AuthorID = nil
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *PostRepository) GetPublishedPostsByTagOrContent(q string) ([]model.Post, error) {
	query := `SELECT id, title, slug, content, author_id, author_name, published_at, created_at, updated_at, tags, category FROM posts
	WHERE published_at IS NOT NULL 
	AND $1 = ANY(tags)
	OR to_tsvector(content) @@ to_tsquery($1)
	ORDER BY created_at DESC`
	rows, err := r.db.Query(query, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var post model.Post
		// Scan into post fields. Use sql.Null* types for nullable columns.
		var authorID sql.Null[uuid.UUID]

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Slug,
			&post.Content,
			&post.AuthorID,
			&post.AuthorName,
			&post.PublishedAt,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Tags,
			&post.Category,
		)
		if err != nil {
			return nil, err
		}

		// Assign nullable fields
		if authorID.Valid {
			post.AuthorID = &authorID.V
		} else {
			post.AuthorID = nil
		}

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// GetPostBySlug fetches a single post by its slug.
func (r *PostRepository) GetPostBySlug(slug string) (*model.Post, error) {
	query := `SELECT id, title, slug, content, author_name, published_at, created_at, updated_at, tags, category FROM posts WHERE slug = $1;`
	var post model.Post

	err := r.db.QueryRow(query, slug).Scan(
		&post.ID,
		&post.Title,
		&post.Slug,
		&post.Content,
		&post.AuthorName,
		&post.PublishedAt,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Tags,
		&post.Category,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return &post, nil // Return nil, nil if no row is found
		}
		return &post, err
	}
	return &post, nil
}

// GetPostByID fetches a single post by its ID.
func (r *PostRepository) GetPostByID(id uuid.UUID) (*model.Post, error) {
	query := `SELECT id, title, slug, content, author_name, published_at, created_at, updated_at, tags, category FROM posts WHERE id = $1`
	var post model.Post
	var publishedAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Slug,
		&post.Content,
		&post.AuthorName,
		&publishedAt,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Tags,
		&post.Category,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil, nil if no row is found
		}
		return nil, err
	}

	if publishedAt.Valid {
		post.PublishedAt = &publishedAt.Time
	} else {
		post.PublishedAt = nil
	}

	return &post, nil
}

// UpdatePost updates an existing post in the database.
func (r *PostRepository) UpdatePost(post *model.UpdatePost) error {
	query := `UPDATE posts SET title = $2, content = $3, updated_at = NOW(), tags = $4, category = $5 WHERE id = $1`
	_, err := r.db.Exec(query,
		post.ID,
		post.Title,
		post.Content,
		pq.Array(blog.SplitWithoutEmpty(post.Tags, ",")),
		post.Category,
	)
	return err
}

// UpdateImage updates an existing image in the database.
func (r *PostRepository) UpdateImagesWithPostID(image *model.UpdateImages) error {
	query := `UPDATE images SET blog_post_id = $2 WHERE id = ANY($1)`
	_, err := r.db.Exec(query,
		pq.Array(image.ImageIDs),
		image.PostID,
	)
	return err
}

// UpdatePost updates an existing post in the database.
func (r *PostRepository) PublishPost(post *model.Post) error {
	post.UpdatedAt = time.Now().UTC() // Update the timestamp

	query := `UPDATE posts SET published_at = NOW()`
	_, err := r.db.Exec(query)
	return err
}

// UpdatePost updates an existing post in the database.
func (r *PostRepository) AddImage(image model.Image) (uuid.UUID, error) {
	// Set UUID if not already set (e.g., if client provides it)
	if image.ID == uuid.Nil {
		image.ID = uuid.New()
	}
	image.UploadedAt = time.Now().UTC()

	query := `INSERT INTO images (id, filepath, status, blog_post_id, uploaded_at)
              VALUES ($1, $2, $3, NULL, $4) RETURNING id, filepath, uploaded_at`

	// Use QueryRow to get back the generated ID and timestamps (if DB generates)
	// Or use Exec if you're setting ID in Go and don't need returns
	err := r.db.QueryRow(query,
		image.ID,
		image.Filepath,
		image.Status,
		image.UploadedAt,
	).Scan(&image.ID, &image.Filepath, &image.UploadedAt) // Scan the returned values
	return image.ID, err
}

// DeletePost deletes a post by its ID.
func (r *PostRepository) DeletePost(id uuid.UUID) error {
	query := `DELETE FROM posts WHERE id = $1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows // Indicate that no row was deleted
	}
	return nil
}
