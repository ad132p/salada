package repositories

import (
	"database/sql"
	"encoding/json"
	"log"
	"strconv"

	"fmt"
	"salada/internal/blog"
	"salada/internal/blog/model"
	"strings"
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
func (r *PostRepository) CreatePost(postRequest model.CreatePost) (uuid.UUID, error) {
	var post model.Post
	post.Title = postRequest.Title
	post.AuthorName = postRequest.AuthorName
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

		post.Content = blog.GetContentSummary(post.Content, 100)

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *PostRepository) GetPostAndCommentsBySlug(slug string) (*model.Post, error) {
	// 1. Separate 'seen' count increment (optional but cleaner than mixing with SELECT)
	// You can keep the original UPDATE...RETURNING logic separate if you prefer
	// to track the 'seen' count atomically upon post access.
	// For simplicity, we'll assume the seen count is updated elsewhere or removed here.

	// 2. New SQL Query: SELECT the post details and aggregate comments into a JSON array.
	query := `
        SELECT
            p.id, p.title, p.slug, p.content, p.author_name, p.published_at, 
            p.created_at, p.updated_at, p.tags, p.category, p.seen, p.likes,
            COALESCE(
                json_agg(
                    json_build_object(
                        'id', c.id,
                        'content', c.content,
                        'created_at', c.created_at,
                        'author_name', c.author_name
                    ) ORDER BY c.created_at ASC
                ) FILTER (WHERE c.id IS NOT NULL), 
                '[]'
            ) AS comments_json
        FROM posts p
        LEFT JOIN comments c ON p.id = c.blog_post_id
        WHERE p.slug = $1
        GROUP BY p.id;
    `
	var post model.Post
	var commentsJSON []byte // To store the JSON array from the database

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
		&post.Seen,
		&post.Likes,
		&commentsJSON, // Bind the JSON output
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Post not found
		}
		return nil, fmt.Errorf("error executing SELECT or scanning row: %w", err)
	}

	// 3. Unmarshal the comments JSON into the Post struct's Comments field
	if len(commentsJSON) > 0 {
		err = json.Unmarshal(commentsJSON, &post.Comments)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling comments JSON: %w", err)
		}
	}

	_, err = r.db.Exec(`UPDATE posts SET seen = seen + 1, updated_at = NOW() WHERE slug = $1;`, slug)
	if err != nil {
		log.Printf("Warning: Failed to increment seen count for slug %s: %v", slug, err)
	}

	return &post, nil
}

func (r *PostRepository) LikePostByID(req model.LikeRequest) error {
	// Normalize action to lowercase for reliable comparison
	normalizedAction := strings.ToLower(req.Action)

	var updateOperator string

	switch normalizedAction {
	case "like":
		// Increment the likes count
		updateOperator = "+ 1"
	case "unlike":
		// Decrement the likes count (and ensure it never goes below zero)
		updateOperator = "- 1"
	default:
		return fmt.Errorf("invalid action specified: %s. Must be 'like' or 'unlike'", req.Action)
	}

	// SQL command to update the 'likes' count
	// Using a CASE statement or WHERE clause (e.g., likes > 0) is safer for decrementing,
	// but the simple update is used here for direct control via the application:
	query := fmt.Sprintf(`
		UPDATE posts 
		SET likes = likes %s, 
		    updated_at = NOW() 
		WHERE id = $1
	`, updateOperator)

	// Execute the update query.
	result, err := r.db.Exec(query, req.PostID)
	if err != nil {
		return fmt.Errorf("error executing like update for PostID %s: %w", req.PostID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected after like update: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("post with ID '%s' not found or no rows updated", req.PostID)
	}

	return nil
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
	fmt.Println(image)
	query := `UPDATE images SET blog_post_id = $2, status = 'owned' WHERE id = ANY($1)`
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

// CreateComment inserts a new comment into the database.
func (r *PostRepository) CreateComment(req model.CreateCommentRequest) (uuid.UUID, error) {
	// 1. Prepare data model structure for scanning the returned values
	var commentID uuid.UUID
	var createdAt time.Time
	var updatedAt time.Time

	// 2. Define the SQL query
	query := `
        INSERT INTO comments (blog_post_id, content, author_name, created_at, updated_at)
        VALUES ($1, $2, $3, NOW(), NOW()) 
        RETURNING id, created_at, updated_at`

	// 3. Execute the query and scan the returned values
	err := r.db.QueryRow(query,
		req.PostID,
		req.Content,
		req.AuthorName,
	).Scan(
		&commentID,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		// Check if the error is due to a foreign key violation (e.g., PostID doesn't exist)
		// Specific error checking depends on your database driver/wrapper,
		// but a general check is often sufficient.
		return uuid.Nil, fmt.Errorf("failed to insert comment: %w", err)
	}

	// 4. Return the newly generated comment ID
	return commentID, nil
}
