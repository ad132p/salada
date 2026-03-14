package repositories

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"salada/internal/email/model"
)

// EmailRepository handles database operations for emails
type EmailRepository struct {
	db *sql.DB
}

// NewEmailRepository creates a new email repository
func NewEmailRepository(db *sql.DB) *EmailRepository {
	return &EmailRepository{db: db}
}

// CreateEmailLog creates a log entry for a sent email
func (r *EmailRepository) CreateEmailLog(emailID, recipient, status, errorMsg string) error {
	query := `
		INSERT INTO email_logs (id, email_id, recipient, status, error, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(query, uuid.New().String(), emailID, recipient, status, errorMsg, time.Now())
	return err
}

// GetEmailLogs retrieves email logs for a specific email ID
func (r *EmailRepository) GetEmailLogs(emailID string) ([]model.EmailLog, error) {
	query := `
		SELECT id, email_id, recipient, status, error, created_at
		FROM email_logs
		WHERE email_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, emailID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.EmailLog
	for rows.Next() {
		var log model.EmailLog
		err := rows.Scan(
			&log.ID,
			&log.EmailID,
			&log.Recipient,
			&log.Status,
			&log.Error,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

// CreateEmailTemplate creates a new email template
func (r *EmailRepository) CreateEmailTemplate(template *model.EmailTemplate) error {
	query := `
		INSERT INTO email_templates (id, name, subject, body, content_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	now := time.Now()
	template.ID = uuid.New().String()
	template.CreatedAt = now
	template.UpdatedAt = now

	_, err := r.db.Exec(query,
		template.ID,
		template.Name,
		template.Subject,
		template.Body,
		template.ContentType,
		template.CreatedAt,
		template.UpdatedAt,
	)
	return err
}

// GetEmailTemplateByName retrieves an email template by name
func (r *EmailRepository) GetEmailTemplateByName(name string) (*model.EmailTemplate, error) {
	query := `
		SELECT id, name, subject, body, content_type, created_at, updated_at
		FROM email_templates
		WHERE name = $1
	`
	var template model.EmailTemplate
	err := r.db.QueryRow(query, name).Scan(
		&template.ID,
		&template.Name,
		&template.Subject,
		&template.Body,
		&template.ContentType,
		&template.CreatedAt,
		&template.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// GetAllEmailTemplates retrieves all email templates
func (r *EmailRepository) GetAllEmailTemplates() ([]model.EmailTemplate, error) {
	query := `
		SELECT id, name, subject, body, content_type, created_at, updated_at
		FROM email_templates
		ORDER BY name
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []model.EmailTemplate
	for rows.Next() {
		var template model.EmailTemplate
		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Subject,
			&template.Body,
			&template.ContentType,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		templates = append(templates, template)
	}
	return templates, nil
}

// UpdateEmailTemplate updates an existing email template
func (r *EmailRepository) UpdateEmailTemplate(template *model.EmailTemplate) error {
	query := `
		UPDATE email_templates
		SET name = $1, subject = $2, body = $3, content_type = $4, updated_at = $5
		WHERE id = $6
	`
	template.UpdatedAt = time.Now()
	_, err := r.db.Exec(query,
		template.Name,
		template.Subject,
		template.Body,
		template.ContentType,
		template.UpdatedAt,
		template.ID,
	)
	return err
}

// DeleteEmailTemplate deletes an email template by ID
func (r *EmailRepository) DeleteEmailTemplate(id string) error {
	query := `DELETE FROM email_templates WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// CreateTables creates the necessary tables for email functionality
func (r *EmailRepository) CreateTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS email_logs (
			id UUID PRIMARY KEY,
			email_id TEXT NOT NULL,
			recipient TEXT NOT NULL,
			status TEXT NOT NULL,
			error TEXT,
			created_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS email_templates (
			id UUID PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			subject TEXT NOT NULL,
			body TEXT NOT NULL,
			content_type TEXT NOT NULL DEFAULT 'text/plain',
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,
	}

	for _, query := range queries {
		if _, err := r.db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}
