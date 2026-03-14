package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"salada/internal/email"
	"salada/internal/email/model"
)

// EmailController handles email-related HTTP requests
type EmailController struct {
	client *email.Client
}

// NewEmailController creates a new email controller with the default client
func NewEmailController() *EmailController {
	return &EmailController{
		client: email.NewClient(),
	}
}

// NewEmailControllerWithClient creates a new email controller with a custom client
func NewEmailControllerWithClient(client *email.Client) *EmailController {
	return &EmailController{
		client: client,
	}
}

// SendEmail handles sending a plain text or HTML email
func (ec *EmailController) SendEmail(c *gin.Context) {
	var req model.SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request: " + err.Error(),
		})
		return
	}

	// Default to text/plain if content type not specified
	if req.ContentType == "" {
		req.ContentType = "text/plain"
	}

	// Validate content type
	if req.ContentType != "text/plain" && req.ContentType != "text/html" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Content type must be 'text/plain' or 'text/html'",
		})
		return
	}

	var err error
	if req.ContentType == "text/html" {
		err = ec.client.SendHTMLEmail(req.To, req.Subject, req.Body)
	} else {
		err = ec.client.SendEmail(req.To, req.Subject, req.Body)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.SendEmailResponse{
		Success:   true,
		MessageID: uuid.New().String(),
		SentAt:    time.Now(),
	})
}

// SendEmailWithAttachments handles sending an email with file attachments
func (ec *EmailController) SendEmailWithAttachments(c *gin.Context) {
	// Parse multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32 MB max
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to parse form: " + err.Error(),
		})
		return
	}

	to := c.PostFormArray("to")
	subject := c.PostForm("subject")
	body := c.PostForm("body")
	contentType := c.PostForm("content_type")

	if len(to) == 0 || subject == "" || body == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Missing required fields: to, subject, body",
		})
		return
	}

	if contentType == "" {
		contentType = "text/plain"
	}

	// Process attachments
	form := c.Request.MultipartForm
	files := form.File["attachments"]
	var attachments []email.Attachment

	for _, file := range files {
		f, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to open attachment: " + err.Error(),
			})
			return
		}
		defer f.Close()

		data := make([]byte, file.Size)
		_, err = f.Read(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to read attachment: " + err.Error(),
			})
			return
		}

		attachments = append(attachments, email.Attachment{
			Filename:    file.Filename,
			ContentType: file.Header.Get("Content-Type"),
			Data:        data,
		})
	}

	// Send email with attachments
	err := ec.client.SendEmailWithAttachment(to, subject, body, contentType, attachments)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.SendEmailResponse{
		Success:   true,
		MessageID: uuid.New().String(),
		SentAt:    time.Now(),
	})
}

// GetSMTPConfig returns the current SMTP configuration (without sensitive data)
func (ec *EmailController) GetSMTPConfig(c *gin.Context) {
	config := ec.client.Config

	c.JSON(http.StatusOK, gin.H{
		"host":         config.SMTPHost,
		"port":         config.SMTPPort,
		"from_address": config.FromAddress,
		"from_name":    config.FromName,
		"use_tls":      config.UseTLS,
		"use_starttls": config.UseStartTLS,
		"authenticated": config.Username != "",
	})
}

// TestConnection tests the SMTP connection
func (ec *EmailController) TestConnection(c *gin.Context) {
	testEmail := model.SendEmailRequest{
		To:          []string{ec.client.Config.FromAddress},
		Subject:     "SMTP Test",
		Body:        "This is a test email to verify SMTP configuration.",
		ContentType: "text/plain",
	}

	err := ec.client.SendEmail(testEmail.To, testEmail.Subject, testEmail.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Test email sent successfully",
	})
}

// GetEmailStatus returns the status of an email by ID
func (ec *EmailController) GetEmailStatus(c *gin.Context) {
	emailID := c.Param("id")
	if emailID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Email ID is required",
		})
		return
	}

	// In a real implementation, this would query a database
	// For now, return a mock response
	c.JSON(http.StatusOK, gin.H{
		"id":     emailID,
		"status": "sent",
		"sent_at": time.Now(),
	})
}

// GetEmailPage renders the email sending page
func (ec *EmailController) GetEmailPage(c *gin.Context) {
	c.HTML(http.StatusOK, "pages/email.html", gin.H{
		"title":        "Send Email",
		"is_logged_in": c.GetBool("is_logged_in"),
	})
}
