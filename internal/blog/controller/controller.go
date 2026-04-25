package controller

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"salada/internal/blog"
	"salada/internal/blog/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PostRepository defines the interface for post-related data operations.
// This allows for mocking in tests.
type PostRepository interface {
	CreatePost(postRequest model.CreatePost) (uuid.UUID, error)
	UpdateImagesWithPostID(image *model.UpdateImages) error
	CreateComment(req model.CreateCommentRequest) (uuid.UUID, error)
	GetPostByID(id uuid.UUID) (*model.Post, error)
	UpdatePost(post *model.UpdatePost) error
	PublishPost(post *model.Post) error
	AddImage(image model.Image) (uuid.UUID, error)
	GetPublishedPosts(category string, q string, limit int, cursorPublishedAt *time.Time, cursorID *uuid.UUID) ([]model.Post, string, error)
	GetCategoryCount() ([]model.CategoryCount, error)
	GetTagCount() ([]model.TagCount, error)
	GetPostAndCommentsBySlug(slug string) (*model.Post, error)
	DeletePost(id uuid.UUID) ([]string, error)
	GetPostBySlug(slug string) (*model.Post, error)
	LikePostByID(req model.LikeRequest) error
}

// BlogController handles blog post-related requests.
type BlogController struct {
	Repo PostRepository
}

// NewBlogController creates a new BlogController instance.
func NewBlogController(repo PostRepository) *BlogController {
	return &BlogController{Repo: repo}
}

func (pc *BlogController) CreatePost(c *gin.Context) {
	username, ok := c.Get("username")
	if !ok {
		// If authentication is required, block anonymous access here
		c.JSON(http.StatusUnauthorized, gin.H{"message": "Authentication required"})
		return
	}

	postRequest := model.CreatePost{AuthorName: username.(string)}
	if err := c.ShouldBindJSON(&postRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error binding JSON": err.Error()})
		return
	}

	postID, err := pc.Repo.CreatePost(postRequest)

	if err != nil {
		log.Printf("Failed to create post: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	updateRequest := model.UpdateImages{PostID: postID, ImageIDs: postRequest.ImageIDs}

	err = pc.Repo.UpdateImagesWithPostID(&updateRequest)

	if err != nil {
		log.Printf("Failed to update images of post: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update images of post"})
		return
	}

	c.JSON(http.StatusFound, gin.H{"msg": "Success", "ID": postID})
}

func (pc *BlogController) CreateComment(c *gin.Context) {

	commentRequest := model.CreateCommentRequest{}
	if err := c.ShouldBindJSON(&commentRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error binding JSON": err.Error()})
		return
	}

	commentID, err := pc.Repo.CreateComment(commentRequest)

	if err != nil {
		log.Printf("Failed to create comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "Success", "ID": commentID})
}

// UpdatePost handles PUT on /blog/:id
func (pc *BlogController) UpdatePost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var updatePost model.UpdatePost
	if err := c.ShouldBindJSON(&updatePost); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	post, err := pc.Repo.GetPostByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find post"})
		return
	}
	if post == nil { // Check if no record was found
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	updatePost.ID = id

	if err := pc.Repo.UpdatePost(&updatePost); err != nil {
		log.Printf("Failed to update post: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}

	updateRequest := model.UpdateImages{PostID: updatePost.ID, ImageIDs: updatePost.ImageIDs}

	err = pc.Repo.UpdateImagesWithPostID(&updateRequest)

	if err != nil {
		log.Printf("Failed to update images of post: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update images of post"})
		return
	}

	c.JSON(http.StatusOK, updatePost)
}

func (pc *BlogController) PublishPost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}
	post, err := pc.Repo.GetPostByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find post"})
		return
	}
	if post == nil { // Check if no record was found
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	if post.PublishedAt != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Post not found"})
		return
	}
	if err := pc.Repo.PublishPost(post); err != nil {
		log.Printf("Failed to publish post: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}

	c.JSON(http.StatusOK, post)
}

func (pc *BlogController) UploadImage(c *gin.Context) {
	// Single file
	file, err := c.FormFile("image") // The name "image" must match the form field name in the client request

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve the file"})
		return
	}

	// Extract the extension and validate it
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[strings.ToLower(ext)] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type"})
		return
	}

	// Generate a secure UUID filename to prevent traversal and conflicts
	filename := uuid.New().String() + ext

	// Save the file to the specified path
	dst := filepath.Join("uploads", filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save the file"})
		return
	}

	// Construct the URL for the saved image
	imageURL := fmt.Sprintf("/uploads/%s", filename)
	image := model.Image{Filepath: imageURL, Status: "pending"}

	imageUUID, err := pc.Repo.AddImage(image)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add image", "details": err})
		return
	}

	// Respond to the client with the image URL in the expected format
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"filepath": imageURL,
			"image_id": imageUUID,
		},
	})
}

// GetPosts handles GET /blog/
func (pc *BlogController) GetAllPosts(c *gin.Context) {
	var posts []model.Post
	category := c.Query("category")
	cursorPublishedAt, cursorID := getCursorFromContext(c)

	posts, nextCursor, err := pc.Repo.GetPublishedPosts(category, "", 10, cursorPublishedAt, cursorID)

	if err != nil {
		c.HTML(http.StatusServiceUnavailable, "blog/blog_post_list.html", gin.H{
			"title":        "Blog Posts",
			"error":        err,
			"is_logged_in": c.GetBool("is_logged_in"),
		})
		return
	}

	categories, err := pc.Repo.GetCategoryCount()

	if err != nil {
		c.HTML(http.StatusServiceUnavailable, "blog/blog_post_list.html", gin.H{
			"title":        "Blog Posts",
			"error":        err,
			"is_logged_in": c.GetBool("is_logged_in"),
		})
		return
	}

	encodedNextCursor := ""
	if nextCursor != "" {
		encodedNextCursor = base64.StdEncoding.EncodeToString([]byte(nextCursor))
	}

	c.HTML(http.StatusOK, "blog/blog_post_list.html", gin.H{
		"title":         "Blog Posts",
		"posts":         posts,
		"categories":    categories,
		"next_cursor":   encodedNextCursor,
		"is_logged_in":  c.GetBool("is_logged_in"),
		"is_first_page": c.Query("cursor") == "",
	})
}

// GetPosts handles GET /blog/
func (pc *BlogController) GetRecentPosts(c *gin.Context) {
	var posts []model.Post
	category := c.Query("category")
	posts, _, err := pc.Repo.GetPublishedPosts(category, "", 3, nil, nil)

	if err != nil {
		c.HTML(http.StatusServiceUnavailable, "blog/blog.html", gin.H{
			"title":        "Blog Posts",
			"error":        err,
			"is_logged_in": c.GetBool("is_logged_in"),
		})
		return
	}

	categories, err := pc.Repo.GetCategoryCount()

	if err != nil {
		c.HTML(http.StatusServiceUnavailable, "blog/blog.html", gin.H{
			"title":        "Blog Posts",
			"error":        err,
			"is_logged_in": c.GetBool("is_logged_in"),
		})
		return
	}
	c.HTML(http.StatusOK, "blog/blog.html", gin.H{
		"title":        "Blog Posts",
		"posts":        posts,
		"categories":   categories,
		"is_logged_in": c.GetBool("is_logged_in"),
	})
}

func (pc *BlogController) GetNewPostForm(c *gin.Context) {
	username, ok := c.MustGet("username").(string)
	if !ok {
		// This should not happen if the middleware is correctly set up.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Username not found in context"})
		return
	}

	categoriesJSON, _ := json.Marshal(blog.Categories)

	c.HTML(http.StatusOK, "blog/blog_new.html", gin.H{
		"title":          "New Blog Entry",
		"username":       username,
		"categoriesJSON": string(categoriesJSON),
		"is_logged_in":   c.GetBool("is_logged_in"),
	})
}

// GetPostAndCommentsBySlug handles GET /blog/:slug
func (pc *BlogController) GetPostAndCommentsBySlug(c *gin.Context) {
	slug := c.Param("slug")
	post, err := pc.Repo.GetPostAndCommentsBySlug(slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // Use err.Error() for proper JSON output
		return
	}
	if post == nil { // Check if no record was found by the repository
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// 1. Attempt to retrieve the username from Gin context (c.Get)
	// The username is typically stored in the context by a preceding middleware.
	username, err := c.Cookie("username")

	if err != nil {
		// let anon comment
		username = "anon"
	}

	commentsJSONBytes, err := json.Marshal(post.Comments)
	if err != nil {
		// Handle the error if marshalling fails (e.g., if data is invalid)
		log.Printf("Error marshalling comments: %v", err)
		commentsJSONBytes = []byte("[]") // Default to an empty array string on failure
	}

	// 3. Include the username in the HTML response data
	// Render markdown and extract table of contents
	renderedHTML, tocItems := blog.RenderMarkdownToHTMLWithIDs(post.Content)

	c.HTML(http.StatusOK, "blog/blog_post.html", gin.H{
		"title":        post.Title,
		"post":         post,
		"comments":     string(commentsJSONBytes),
		"content":      template.HTML(renderedHTML),
		"toc":          tocItems,
		"username":     username,
		"is_logged_in": c.GetBool("is_logged_in"),
	})
}

// GetPostBySlug handles GET /blog/comments/:slug
func (pc *BlogController) GetCommentsBySlug(c *gin.Context) {
	slug := c.Param("slug")
	post, err := pc.Repo.GetPostAndCommentsBySlug(slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	if post == nil { // Check if no record was found by the repository
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	commentsJSONBytes, err := json.Marshal(post.Comments)
	if err != nil {
		// Handle the error if marshalling fails (e.g., if data is invalid)
		log.Printf("Error marshalling comments: %v", err)
		commentsJSONBytes = []byte("[]") // Default to an empty array string on failure
	}

	c.JSON(http.StatusOK, gin.H{
		"comments": string(commentsJSONBytes),
	})
}

func (pc *BlogController) GetCategory(c *gin.Context) {
	category := c.Param("name")
	cursorPublishedAt, cursorID := getCursorFromContext(c)

	posts, nextCursor, err := pc.Repo.GetPublishedPosts(category, "", 10, cursorPublishedAt, cursorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	if len(posts) == 0 && c.Query("cursor") == "" { // Check if no record was found by the repository (only on first page)
		c.JSON(http.StatusNotFound, gin.H{"error": "Category does not have any posts yet,"})
		return
	}

	categories, err := pc.Repo.GetCategoryCount()
	if err != nil {
		c.HTML(http.StatusServiceUnavailable, "blog/blog_post_list.html", gin.H{
			"title":        "Posts by Category",
			"error":        err,
			"is_logged_in": c.GetBool("is_logged_in"),
		})
		return
	}

	encodedNextCursor := ""
	if nextCursor != "" {
		encodedNextCursor = base64.StdEncoding.EncodeToString([]byte(nextCursor))
	}

	c.HTML(http.StatusOK, "blog/blog_post_list.html", gin.H{
		"title":         "Posts by Category",
		"posts":         posts,
		"category":      category,
		"categories":    categories,
		"next_cursor":   encodedNextCursor,
		"is_logged_in":  c.GetBool("is_logged_in"),
		"is_first_page": c.Query("cursor") == "",
	})
}

func (pc *BlogController) GetTag(c *gin.Context) {
	tag := c.Param("name")
	posts, _, err := pc.Repo.GetPublishedPosts("", tag, 0, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	if len(posts) == 0 { // Check if no record was found by the repository
		c.JSON(http.StatusNotFound, gin.H{"error": "No post has such tag."})
		return
	}

	tags, err := pc.Repo.GetTagCount()
	if err != nil {
		c.HTML(http.StatusServiceUnavailable, "blog/blog_post_list.html", gin.H{
			"title":        "Posts by Tag",
			"error":        err,
			"is_logged_in": c.GetBool("is_logged_in"),
		})
		return
	}

	c.HTML(http.StatusOK, "blog/blog_post_list.html", gin.H{
		"title":         "Posts by Tag",
		"posts":         posts,
		"tag":           tag,
		"tags":          tags,
		"is_logged_in":  c.GetBool("is_logged_in"),
		"is_first_page": true,
	})
}

func (pc *BlogController) GetTagOrContent(c *gin.Context) {
	query := c.Query("q")
	posts, _, err := pc.Repo.GetPublishedPosts("", query, 0, nil, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	if len(posts) == 0 { // Check if no record was found by the repository
		c.HTML(http.StatusNotFound, "blog/blog.html", gin.H{
			"error":        "No post has such tag.",
			"is_logged_in": c.GetBool("is_logged_in"),
		})
		return
	}

	tags, err := pc.Repo.GetTagCount()
	if err != nil {
		c.HTML(http.StatusServiceUnavailable, "blog/blog.html", gin.H{
			"title":        "Posts by Tag",
			"error":        err,
			"is_logged_in": c.GetBool("is_logged_in"),
		})
		return
	}

	c.HTML(http.StatusOK, "blog/blog_post_list.html", gin.H{
		"title":         "Posts by Tag",
		"posts":         posts,
		"tags":          tags,
		"is_logged_in":  c.GetBool("is_logged_in"),
		"is_first_page": true,
	})
}

// DeletePost handles DELETE /blog/:id
func (pc *BlogController) DeletePost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	images, err := pc.Repo.DeletePost(id)
	if err != nil {
		if err == sql.ErrNoRows { // Check for no rows affected, indicating not found
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		log.Printf("Failed to delete post: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}
	fmt.Println(images)
	blog.DeleteFiles(images)       // Ignoring error for now
	c.Status(http.StatusNoContent) // 204 No Content for successful deletion
	c.Next()
}

// EditPostForm handles GET /blog/edit/:slug/
func (pc *BlogController) EditPostForm(c *gin.Context) {
	slug := c.Param("slug")
	post, err := pc.Repo.GetPostBySlug(slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve post"})
		return
	}
	if post == nil { // Check if no record was found by the repository
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	username, _ := c.Get("username")
	postJSON, _ := json.Marshal(post)
	categoriesJSON, _ := json.Marshal(blog.Categories)

	c.HTML(http.StatusOK, "blog/edit_post_form.html", gin.H{
		"title":          post.Title,
		"postJSON":       template.JS(postJSON),
		"categoriesJSON": template.JS(categoriesJSON),
		"username":       username,
		"is_logged_in":   c.GetBool("is_logged_in"),
	})
}

func (pc *BlogController) LikePost(c *gin.Context) {
	var req model.LikeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	err := pc.Repo.LikePostByID(req)
	if err != nil {
		// Use a more specific status if the post wasn't found
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update like status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func getCursorFromContext(c *gin.Context) (*time.Time, *uuid.UUID) {
	cursor := c.Query("cursor")
	if cursor == "" {
		return nil, nil
	}

	decodedCursor, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, nil
	}

	parts := strings.Split(string(decodedCursor), ",")
	if len(parts) != 2 {
		return nil, nil
	}

	ts, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, nil
	}

	t := time.UnixMicro(ts).UTC()
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, nil
	}

	return &t, &id
}
