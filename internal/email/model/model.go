package model

import (
	"time"
)

// Email represents an email message to be sent
type Email struct {
	ID          string    `json:"id"`
	To          []string  `json:"to"`
	From        string    `json:"from"`
	Subject     string    `json:"subject"`
	Body        string    `json:"body"`
	ContentType string    `json:"content_type"`
	Attachments []Attachment `json:"attachments,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
	Status      string    `json:"status"`
	Error       string    `json:"error,omitempty"`
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	Data        []byte `json:"-"` // Exclude from JSON
}

// EmailTemplate represents a reusable email template
type EmailTemplate struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Subject     string    `json:"subject"`
	Body        string    `json:"body"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SendEmailRequest represents a request to send an email
type SendEmailRequest struct {
	To          []string     `json:"to" binding:"required"`
	Subject     string       `json:"subject" binding:"required"`
	Body        string       `json:"body" binding:"required"`
	ContentType string       `json:"content_type"` // "text/plain" or "text/html"
	Attachments []Attachment `json:"attachments,omitempty"`
}

// SendEmailResponse represents the response after sending an email
type SendEmailResponse struct {
	Success   bool      `json:"success"`
	MessageID string    `json:"message_id,omitempty"`
	Error     string    `json:"error,omitempty"`
	SentAt    time.Time `json:"sent_at"`
}

// EmailLog represents a log entry for sent emails
type EmailLog struct {
	ID        string    `json:"id"`
	EmailID   string    `json:"email_id"`
	Recipient string    `json:"recipient"`
	Status    string    `json:"status"` // "sent", "failed", "pending"
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// SMTPConfig represents SMTP server configuration
type SMTPConfig struct {
	Host        string `json:"host"`
	Port        string `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	FromAddress string `json:"from_address"`
	FromName    string `json:"from_name"`
	UseTLS      bool   `json:"use_tls"`
	UseStartTLS bool   `json:"use_starttls"`
}
