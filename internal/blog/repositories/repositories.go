package repositories

import (
	"database/sql"
	"strconv"

	"salada/internal/blog"
	"salada/internal/blog/model"
	"time"
	"fmt"

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
func (r *PostRepository) CreatePost(postRequest *model.CreatePost) (uuid.UUID, error) {
	var post model.Post
	post.Title = postRequest.Title
	post.Content = postRequest.Content
	post.Tags = blog.SplitWithoutEmpty(postRequest.Tags, ",")
	post.Category = postRequest.Category
	post.CreatedAt = time.Now().UTC()
	post.UpdatedAt = post.CreatedAt
	post.Slug = blog.CreateSlug(postRequest.Title)


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

func (r *PostRepository) GetPublishedPosts(category string, q string) ([]model.Post, error) {
	// 1. Base Query with Lateral Join
	// The lateral join (LEFT JOIN LATERAL) finds the single best image (thumbnail) for each post.
	baseQueryTemplate := `
    SELECT 
        p.id, p.title, p.slug, p.content, p.author_id, p.author_name, p.published_at, p.created_at, p.updated_at, p.tags, p.category, 
        t.filepath AS thumbnail_url
    FROM 
        posts p
    LEFT JOIN LATERAL (
        SELECT filepath 
        FROM images i
        WHERE i.blog_post_id = p.id
        ORDER BY i.uploaded_at ASC -- Sort by upload time (earliest first)
        LIMIT 1                   -- Take only the first one
    ) t ON true
    WHERE 
        p.published_at IS NOT NULL`

	baseQuery := baseQueryTemplate

	// 2. Initialize slice and counter for dynamic WHERE clauses
	var args []interface{}
	paramIndex := 1

	// 3. Handle Category Filtering (Optional)
	if category != "" {
		baseQuery += ` AND p.category = $` + strconv.Itoa(paramIndex)
		args = append(args, category)
		paramIndex++
	}

	// 4. Handle Tag/Content Search (Optional)
	if q != "" {
		// Note: The $N placeholder must be used for the query 'q'.
		baseQuery += ` AND ($` + strconv.Itoa(paramIndex) + ` = ANY(p.tags) 
            OR to_tsvector('english', p.content) @@ websearch_to_tsquery('english', $` + strconv.Itoa(paramIndex) + `))`

		args = append(args, q)
		paramIndex++
	}

	// 5. Add the final ordering clause
	finalQuery := baseQuery + ` ORDER BY p.created_at DESC`

	// 6. Execute the query
	rows, err := r.db.Query(finalQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 7. Scanning logic
	var posts []model.Post
	for rows.Next() {
		var post model.Post
		var authorID sql.Null[uuid.UUID]
		var thumbnailURL sql.NullString // Use sql.NullString for the optional thumbnail URL

		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Slug,
			&post.Content,
			&authorID,
			&post.AuthorName,
			&post.PublishedAt,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Tags,
			&post.Category,
			&thumbnailURL, // Scan the new thumbnail_url column
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

		// Assign the optional Thumbnail URL
		if thumbnailURL.Valid {
			post.ThumbnailURL = thumbnailURL.String
		} else {
			post.ThumbnailURL = "" // Or nil, depending on your model.Post field type
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

	query := `UPDATE posts SET published_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, post.ID)
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

// DeletePost deletes a post by its ID and returns the filepaths of associated images.
func (r *PostRepository) DeletePost(id uuid.UUID) ([]string, error) {
    // The query attempts to delete the post and associated images, 
    // returning the filepaths of the deleted images.
	query := `DELETE FROM posts 
             USING images 
             WHERE posts.id = images.blog_post_id
             AND posts.id = $1
             RETURNING images.filepath;`

	rows, err := r.db.Query(query, id)
	if err != nil {
        // Return the error instead of calling log.Fatalf, which stops the program.
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()

    // 1. Initialize a slice to hold all the filepaths.
	var filepaths []string 
	
    // 2. Loop through all the returned rows.
	for rows.Next() {
		var filepath string
        // 3. Scan the filepath from the current row.
		err = rows.Scan(&filepath)
		if err != nil {
            // Return the error if scanning fails.
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
        // 4. Append the scanned filepath to the slice.
		filepaths = append(filepaths, filepath) 
	}
    
    // 5. Check for any error that occurred during iteration.
	if err = rows.Err(); err != nil {
        // Return the iteration error.
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
    
    // 6. Return the slice of filepaths and a nil error.
    // Note: The function's return signature has been changed to ([]string, error).
	return filepaths, nil
}
