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

	query := `INSERT INTO posts (title, slug, content, author_name, published_at, created_at, updated_at, tags, category, thumbnail_position)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, created_at, updated_at`

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
		post.ThumbnailPosition,
	).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt) // Scan the returned values

	return post.ID, err
}

// GetPosts fetches all posts from the database.
func (r *PostRepository) GetPosts() ([]model.Post, error) {
	query := `SELECT id, title, slug, content, author_id, published_at, created_at, updated_at, tags, category, thumbnail_position FROM posts ORDER BY created_at DESC`
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
			&post.ThumbnailPosition,
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
	query := `SELECT id, title, slug, content, author_name, published_at, created_at, updated_at, tags, category, thumbnail_position FROM posts WHERE slug = $1;`
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
		&post.ThumbnailPosition,
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

// GetCategories returns all available categories defined in the PostgreSQL database 'category' enum.
func (r *PostRepository) GetCategories() ([]string, error) {
	query := `SELECT enumlabel FROM pg_enum JOIN pg_type ON pg_enum.enumtypid = pg_type.oid WHERE pg_type.typname = 'category' ORDER BY enumsortorder;`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *PostRepository) GetTagCount() ([]model.TagCount, error) {
	query := `SELECT unnest(tags) as tag, count(*) FROM posts WHERE published_at IS NOT NULL GROUP BY tag`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tagCountSet []model.TagCount
	for rows.Next() {
		var tagCount model.TagCount
		err := rows.Scan(
			&tagCount.Tag,
			&tagCount.Count,
		)
		if err != nil {
			return nil, err
		}
		tagCountSet = append(tagCountSet, tagCount)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tagCountSet, nil
}

func (r *PostRepository) GetPublishedPosts(category string, q string, limit int, cursorPublishedAt *time.Time, cursorID *uuid.UUID) ([]model.Post, string, error) {
	// 1. Base Query (Removed Lateral Join for thumbnail)
	baseQueryTemplate := `
    SELECT 
        p.id, p.title, p.slug, p.content, p.author_id, p.author_name, p.published_at, p.created_at, p.updated_at, p.tags, p.category, p.thumbnail_position
    FROM 
        posts p
    WHERE 
        p.published_at IS NOT NULL`

	baseQuery := baseQueryTemplate

	// 2. Initialize slice and counter for dynamic WHERE clauses
	var args []interface{}
	paramIndex := 1

	// 3. Handle Category Filtering
	if category != "" {
		baseQuery += ` AND p.category = $` + strconv.Itoa(paramIndex)
		args = append(args, category)
		paramIndex++
	}

	// 4. Handle Tag/Content Search
	if q != "" {
		baseQuery += ` AND ($` + strconv.Itoa(paramIndex) + ` = ANY(p.tags) 
            OR to_tsvector('english', p.title || ' ' || p.content) @@ websearch_to_tsquery('english', $` + strconv.Itoa(paramIndex) + `))`
		args = append(args, q)
		paramIndex++
	}

	// 5. Handle Cursor (Keyset Pagination)
	if cursorPublishedAt != nil && cursorID != nil {
		baseQuery += ` AND (p.published_at < $` + strconv.Itoa(paramIndex) + ` 
			OR (p.published_at = $` + strconv.Itoa(paramIndex) + ` AND p.id < $` + strconv.Itoa(paramIndex+1) + `))`
		args = append(args, cursorPublishedAt, cursorID)
		paramIndex += 2
	}

	// 6. Add ordering
	// We order by published_at DESC, then id DESC for deterministic tie-breaking
	baseQuery += ` ORDER BY p.published_at DESC, p.id DESC`

	// 7. Handle Limit
	// Fetch one extra item to determine if there is a next page
	fetchLimit := limit + 1
	if limit > 0 {
		baseQuery += ` LIMIT $` + strconv.Itoa(paramIndex)
		args = append(args, fetchLimit)
	}

	// 8. Execute the query
	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	// 9. Scanning logic
	var posts []model.Post
	for rows.Next() {
		var post model.Post
		var authorID sql.Null[uuid.UUID]

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
			&post.ThumbnailPosition,
		)
		if err != nil {
			return nil, "", err
		}

		if authorID.Valid {
			post.AuthorID = &authorID.V
		}

		// Extract thumbnail from content
		post.ThumbnailURL = blog.GetFirstImage(post.Content)

		post.Content = blog.GetContentSummary(post.Content, 100)
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, "", err
	}

	// 10. Determine Next Cursor
	var nextCursor string
	if limit > 0 && len(posts) > limit {
		// Remove the extra item we fetched
		posts = posts[:limit]
		lastPost := posts[len(posts)-1]

		// Create cursor string: "published_at_timestamp,uuid"
		// We use UnixMicro for precision if needed, or RFC3339Nano
		if lastPost.PublishedAt != nil {
			nextCursor = fmt.Sprintf("%d,%s", lastPost.PublishedAt.UnixMicro(), lastPost.ID.String())
		}
	}

	return posts, nextCursor, nil
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
            p.created_at, p.updated_at, p.tags, p.category, p.seen, p.likes, p.thumbnail_position,
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
		&post.ThumbnailPosition,
		&commentsJSON, // Bind the JSON output
	)

	// Extract thumbnail from content
	post.ThumbnailURL = blog.GetFirstImage(post.Content)

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

	switch normalizedAction {
	case "like", "unlike":
		// Valid actions
	default:
		return fmt.Errorf("invalid action specified: %s. Must be 'like' or 'unlike'", req.Action)
	}

	query := `
		UPDATE posts 
		SET likes = CASE
			WHEN $2 = 'like' THEN likes + 1
			WHEN $2 = 'unlike' AND likes > 0 THEN likes - 1
			ELSE likes
		END,
		    updated_at = NOW() 
		WHERE id = $1
	`

	// Execute the update query.
	result, err := r.db.Exec(query, req.PostID, normalizedAction)
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
	query := `SELECT id, title, slug, content, author_name, published_at, created_at, updated_at, tags, category, thumbnail_position FROM posts WHERE id = $1`
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
		&post.ThumbnailPosition,
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
	query := `UPDATE posts SET title = $2, content = $3, updated_at = NOW(), tags = $4, category = $5, thumbnail_position = $6 WHERE id = $1`
	_, err := r.db.Exec(query,
		post.ID,
		post.Title,
		post.Content,
		pq.Array(blog.SplitWithoutEmpty(post.Tags, ",")),
		post.Category,
		post.ThumbnailPosition,
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
	// 1. Fetch filepaths of associated images
	queryImages := `SELECT filepath FROM images WHERE blog_post_id = $1`
	rows, err := r.db.Query(queryImages, id)
	if err != nil {
		return nil, fmt.Errorf("error querying images: %w", err)
	}
	defer rows.Close()

	var filepaths []string
	for rows.Next() {
		var filepath string
		if err := rows.Scan(&filepath); err != nil {
			return nil, fmt.Errorf("error scanning image filepath: %w", err)
		}
		filepaths = append(filepaths, filepath)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating image rows: %w", err)
	}

	// 2. Delete the post (images will be deleted via ON DELETE CASCADE)
	queryDelete := `DELETE FROM posts WHERE id = $1`
	result, err := r.db.Exec(queryDelete, id)
	if err != nil {
		return nil, fmt.Errorf("error deleting post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("error checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}

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
