package email

import (
	"encoding/base64"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// Config holds the email client configuration
type Config struct {
	SMTPHost     string
	SMTPPort     string
	FromAddress  string
	FromName     string
	Username     string
	Password     string
	UseTLS       bool
	UseStartTLS  bool
}

// Client represents an email client that can send emails via SMTP
type Client struct {
	Config Config
}

// NewClient creates a new email client from environment variables
func NewClient() *Client {
	return &Client{
		Config: Config{
			SMTPHost:    getEnv("SMTP_HOST", "localhost"),
			SMTPPort:    getEnv("SMTP_PORT", "587"),
			FromAddress: getEnv("SMTP_FROM_ADDRESS", "noreply@localhost"),
			FromName:    getEnv("SMTP_FROM_NAME", "Salada"),
			Username:    getEnv("SMTP_USERNAME", ""),
			Password:    getEnv("SMTP_PASSWORD", ""),
			UseTLS:      getEnv("SMTP_USE_TLS", "false") == "true",
			UseStartTLS: getEnv("SMTP_USE_STARTTLS", "true") == "true",
		},
	}
}

// NewClientWithConfig creates a new email client with explicit configuration
func NewClientWithConfig(config Config) *Client {
	return &Client{Config: config}
}

// SendEmail sends a plain text email to the specified recipient
func (c *Client) SendEmail(to []string, subject, body string) error {
	from := c.formatFromAddress()
	msg := c.buildMessage(from, to, subject, body, "text/plain")

	addr := fmt.Sprintf("%s:%s", c.Config.SMTPHost, c.Config.SMTPPort)

	var auth smtp.Auth
	if c.Config.Username != "" && c.Config.Password != "" {
		auth = smtp.PlainAuth("", c.Config.Username, c.Config.Password, c.Config.SMTPHost)
	}

	return smtp.SendMail(addr, auth, c.Config.FromAddress, to, []byte(msg))
}

// SendHTMLEmail sends an HTML email to the specified recipient
func (c *Client) SendHTMLEmail(to []string, subject, htmlBody string) error {
	from := c.formatFromAddress()
	msg := c.buildMessage(from, to, subject, htmlBody, "text/html")

	addr := fmt.Sprintf("%s:%s", c.Config.SMTPHost, c.Config.SMTPPort)

	var auth smtp.Auth
	if c.Config.Username != "" && c.Config.Password != "" {
		auth = smtp.PlainAuth("", c.Config.Username, c.Config.Password, c.Config.SMTPHost)
	}

	return smtp.SendMail(addr, auth, c.Config.FromAddress, to, []byte(msg))
}

// SendEmailWithAttachment sends an email with file attachments
func (c *Client) SendEmailWithAttachment(to []string, subject, body, contentType string, attachments []Attachment) error {
	from := c.formatFromAddress()
	msg := c.buildMultipartMessage(from, to, subject, body, contentType, attachments)

	addr := fmt.Sprintf("%s:%s", c.Config.SMTPHost, c.Config.SMTPPort)

	var auth smtp.Auth
	if c.Config.Username != "" && c.Config.Password != "" {
		auth = smtp.PlainAuth("", c.Config.Username, c.Config.Password, c.Config.SMTPHost)
	}

	return smtp.SendMail(addr, auth, c.Config.FromAddress, to, []byte(msg))
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// formatFromAddress formats the From address with the sender name
func (c *Client) formatFromAddress() string {
	if c.Config.FromName != "" {
		return fmt.Sprintf("%s <%s>", c.Config.FromName, c.Config.FromAddress)
	}
	return c.Config.FromAddress
}

// buildMessage constructs a simple email message
func (c *Client) buildMessage(from string, to []string, subject, body, contentType string) string {
	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString(fmt.Sprintf("Content-Type: %s; charset=UTF-8\r\n", contentType))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)

	return msg.String()
}

// buildMultipartMessage constructs a multipart email message with attachments
func (c *Client) buildMultipartMessage(from string, to []string, subject, body, contentType string, attachments []Attachment) string {
	boundary := fmt.Sprintf("----=_Part_%d_%d", os.Getpid(), time.Now().UnixNano())

	var msg strings.Builder

	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary))
	msg.WriteString("\r\n")

	// Body part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString(fmt.Sprintf("Content-Type: %s; charset=UTF-8\r\n", contentType))
	msg.WriteString("Content-Transfer-Encoding: 7bit\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)
	msg.WriteString("\r\n")

	// Attachments
	for _, att := range attachments {
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString(fmt.Sprintf("Content-Type: %s\r\n", att.ContentType))
		msg.WriteString("Content-Transfer-Encoding: base64\r\n")
		msg.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", att.Filename))
		msg.WriteString("\r\n")
		msg.WriteString(encodeBase64(att.Data))
		msg.WriteString("\r\n")
	}

	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return msg.String()
}

// encodeBase64 encodes data to base64 string
func encodeBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
