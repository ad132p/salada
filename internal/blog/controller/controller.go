package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"salada/internal/blog"
	"salada/internal/blog/model"
	"salada/internal/blog/repositories"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// BlogController handles blog post-related requests.
type BlogController struct {
	Repo *repositories.PostRepository
}

// NewPostController creates a new PostController instance.
func NewBlogController(repo *repositories.PostRepository) *BlogController {
	return &BlogController{Repo: repo}
}

func (pc *BlogController) CreatePost(c *gin.Context) {
	username, _ := c.MustGet("username").(string) //Skipin for now
	postRequest := model.CreatePost{AuthorName: username}
	if err := c.ShouldBindJSON(&postRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error binding JSON": err.Error()})
		return
	}

	postID, err := pc.Repo.CreatePost(postRequest)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post", "details": err.Error()})
		return
	}

	updateRequest := model.UpdateImages{PostID: postID, ImageIDs: postRequest.ImageIDs}

	err = pc.Repo.UpdateImagesWithPostID(&updateRequest)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update images of post", "details": err.Error()})
		return
	}

	c.JSON(http.StatusFound, gin.H{"msg": "Success", "ID": postID})
}

func (pc *BlogController) CreateComment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}
	username, _ := c.MustGet("username").(string) //Skipin for now
	commentRequest := model.CreateCommentRequest{AuthorName: username, PostID: id}
	if err := c.ShouldBindJSON(&commentRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error binding JSON": err.Error()})
		return
	}

	commentID, err := pc.Repo.CreateComment(commentRequest)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post", "details": err.Error()})
		return
	}

	c.JSON(http.StatusFound, gin.H{"msg": "Success", "ID": commentID})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post", "details": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post", "details": err.Error()})
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

	// Generate a unique filename to prevent conflicts
	filename := filepath.Base(file.Filename)

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
func (pc *BlogController) GetPosts(c *gin.Context) {
	username, ok := c.MustGet("username").(string)
	var err error
	var posts []model.Post

	if !ok {
		// This should not happen if the middleware is correctly set up.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Username not found in context"})
		return
	}

	category := c.Query("category")
	posts, err = pc.Repo.GetPublishedPosts(category, "")

	if err != nil {
		c.HTML(http.StatusServiceUnavailable, "blog.html", gin.H{
			"title": "Blog Posts",
			"error": err,
		})
		return
	}

	categories, err := pc.Repo.GetCategoryCount()

	if err != nil {
		c.HTML(http.StatusServiceUnavailable, "blog.html", gin.H{
			"title": "Blog Posts",
			"error": err,
		})
		return
	}
	c.HTML(http.StatusOK, "blog.html", gin.H{
		"title":      "Blog Posts",
		"posts":      posts,
		"username":   username,
		"categories": categories,
	})
}

func (pc *BlogController) GetNewPostForm(c *gin.Context) {
	username, ok := c.MustGet("username").(string)
	if !ok {
		// This should not happen if the middleware is correctly set up.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Username not found in context"})
		return
	}
	c.HTML(http.StatusOK, "blog_new.html", gin.H{
		"title":    "New Blog Entry",
		"username": username,
	})
}

// GetPostBySlug handles GET /blog/:slug
func (pc *BlogController) GetPostBySlug(c *gin.Context) {
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

	c.HTML(http.StatusOK, "blog_post.html", gin.H{
		"title":    post.Title,
		"post":     post,
		"comments": string(commentsJSONBytes),
		"content":  template.HTML(blog.RenderMarkdownToHTML(post.Content)),
	})
}

func (pc *BlogController) GetCategory(c *gin.Context) {
	category := c.Param("name")
	posts, err := pc.Repo.GetPublishedPosts(category, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	if len(posts) == 0 { // Check if no record was found by the repository
		c.JSON(http.StatusNotFound, gin.H{"error": "Category does not have any posts yet,"})
		return
	}

	categories, err := pc.Repo.GetCategoryCount()
	if err != nil {
		c.HTML(http.StatusServiceUnavailable, "blog.html", gin.H{
			"title": "Posts by Category",
			"error": err,
		})
		return
	}

	c.HTML(http.StatusOK, "blog.html", gin.H{
		"title":      "Posts by Category",
		"posts":      posts,
		"category":   category,
		"categories": categories,
	})
}

func (pc *BlogController) GetTag(c *gin.Context) {
	tag := c.Param("name")
	posts, err := pc.Repo.GetPublishedPosts("", tag)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	if len(posts) == 0 { // Check if no record was found by the repository
		c.JSON(http.StatusNotFound, gin.H{"error": "No post has such tag."})
		return
	}

	categories, err := pc.Repo.GetCategoryCount()
	if err != nil {
		c.HTML(http.StatusServiceUnavailable, "blog.html", gin.H{
			"title": "Posts by Tag",
			"error": err,
		})
		return
	}

	c.HTML(http.StatusOK, "blog.html", gin.H{
		"title":      "Posts by Tag",
		"posts":      posts,
		"tag":        tag,
		"categories": categories,
	})
}

func (pc *BlogController) GetTagOrContent(c *gin.Context) {
	query := c.Query("q")
	fmt.Println(query)
	posts, err := pc.Repo.GetPublishedPosts("", query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	if len(posts) == 0 { // Check if no record was found by the repository
		c.HTML(http.StatusNotFound, "blog.html", gin.H{"error": "No post has such tag."})
		return
	}

	categories, err := pc.Repo.GetCategoryCount()
	if err != nil {
		c.HTML(http.StatusServiceUnavailable, "blog.html", gin.H{
			"title": "Posts by Tag",
			"error": err,
		})
		return
	}

	c.HTML(http.StatusOK, "blog.html", gin.H{
		"title":      "Posts by Tag",
		"posts":      posts,
		"categories": categories,
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post", "details": err.Error()})
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
	c.HTML(http.StatusOK, "edit_post_form.html", gin.H{
		"title":      post.Title,
		"post":       post,
		"categories": blog.Categories,
	})
}
