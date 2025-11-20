package controller

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"salada/internal/blog/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPostRepository is a mock implementation of the PostRepository interface
type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) CreatePost(postRequest model.CreatePost) (uuid.UUID, error) {
	args := m.Called(postRequest)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockPostRepository) UpdateImagesWithPostID(image *model.UpdateImages) error {
	args := m.Called(image)
	return args.Error(0)
}

func (m *MockPostRepository) CreateComment(req model.CreateCommentRequest) (uuid.UUID, error) {
	args := m.Called(req)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockPostRepository) GetPostByID(id uuid.UUID) (*model.Post, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Post), args.Error(1)
}

func (m *MockPostRepository) UpdatePost(post *model.UpdatePost) error {
	args := m.Called(post)
	return args.Error(0)
}

func (m *MockPostRepository) PublishPost(post *model.Post) error {
	args := m.Called(post)
	return args.Error(0)
}

func (m *MockPostRepository) AddImage(image model.Image) (uuid.UUID, error) {
	args := m.Called(image)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockPostRepository) GetPublishedPosts(category string, q string, limit int, cursorPublishedAt *time.Time, cursorID *uuid.UUID) ([]model.Post, string, error) {
	args := m.Called(category, q, limit, cursorPublishedAt, cursorID)
	return args.Get(0).([]model.Post), args.String(1), args.Error(2)
}

func (m *MockPostRepository) GetCategoryCount() ([]model.CategoryCount, error) {
	args := m.Called()
	return args.Get(0).([]model.CategoryCount), args.Error(1)
}

func (m *MockPostRepository) GetPostAndCommentsBySlug(slug string) (*model.Post, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Post), args.Error(1)
}

func (m *MockPostRepository) DeletePost(id uuid.UUID) ([]string, error) {
	args := m.Called(id)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockPostRepository) GetPostBySlug(slug string) (*model.Post, error) {
	args := m.Called(slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Post), args.Error(1)
}

func (m *MockPostRepository) LikePostByID(req model.LikeRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func TestGetRecentPosts(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockPostRepository)
	controller := NewBlogController(mockRepo)
	router := gin.Default()

	// Load HTML templates (mocking this is tricky without actual files,
	// so we might need to skip template rendering or use a simple string response for API tests.
	// However, since the controller renders HTML, we need to handle that.
	// For this test, we'll rely on the fact that if the template is missing, Gin might panic or error.
	// A better approach for unit testing controllers that return HTML is to test the data passed to the template
	// or refactor to separate data fetching from rendering.
	// For now, let's try to mock the template loading or just test the error path if templates aren't found,
	// OR we can point Gin to the actual templates if they exist relative to the test execution.
	// Let's assume we are running from the root or can adjust the path.
	// Actually, simpler: we can check if the controller calls the repo correctly.
	// If template rendering fails, it returns 500 or panics.

	// To make this robust without depending on file system for templates,
	// we can use a custom engine or just accept that we are testing the logic up to rendering.
	// But wait, the controller calls c.HTML.
	// Let's try to load the templates from the actual path.
	router.LoadHTMLGlob("../../../web/templates/html/*")

	router.GET("/blog", controller.GetRecentPosts)

	// Mock Data
	mockPosts := []model.Post{
		{Title: "Test Post 1", Slug: "test-post-1"},
		{Title: "Test Post 2", Slug: "test-post-2"},
	}
	mockCategories := []model.CategoryCount{
		{Category: "tech", Count: 5},
	}

	// Expectations
	mockRepo.On("GetPublishedPosts", "", "", 3, (*time.Time)(nil), (*uuid.UUID)(nil)).Return(mockPosts, "", nil)
	mockRepo.On("GetCategoryCount").Return(mockCategories, nil)

	// Request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/blog", nil)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Test Post 1")
	assert.Contains(t, w.Body.String(), "Test Post 2")

	mockRepo.AssertExpectations(t)
}

func TestLikePost(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockPostRepository)
	controller := NewBlogController(mockRepo)
	router := gin.Default()

	router.POST("/blog/like", controller.LikePost)

	// Mock Data
	postID := uuid.New()
	reqBody := model.LikeRequest{
		PostID: postID.String(),
		Action: "like",
	}

	// Expectations
	mockRepo.On("LikePostByID", reqBody).Return(nil)

	// Request
	// Create JSON body
	jsonBody := `{"post_id": "` + postID.String() + `", "action": "like"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/blog/like", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")

	mockRepo.AssertExpectations(t)
}
