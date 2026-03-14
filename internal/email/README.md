# Email Module

This module provides email sending capabilities for the Salada application using SMTP. It is designed to work with [maddy](https://maddy.email/), a modern, all-in-one mail server written in Go, but can work with any SMTP server.

## Features

- Send plain text emails
- Send HTML emails
- Send emails with file attachments
- SMTP authentication support
- TLS/STARTTLS support
- Environment-based configuration

## Configuration

The email module is configured via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `SMTP_HOST` | SMTP server hostname | `localhost` |
| `SMTP_PORT` | SMTP server port | `587` |
| `SMTP_FROM_ADDRESS` | Default sender email address | `noreply@localhost` |
| `SMTP_FROM_NAME` | Default sender name | `Salada` |
| `SMTP_USERNAME` | SMTP authentication username | (empty) |
| `SMTP_PASSWORD` | SMTP authentication password | (empty) |
| `SMTP_USE_TLS` | Use implicit TLS (port 465) | `false` |
| `SMTP_USE_STARTTLS` | Use STARTTLS (port 587) | `true` |

## API Endpoints

### Send Email
```
POST /email/send
Content-Type: application/json

{
  "to": ["recipient@example.com"],
  "subject": "Hello",
  "body": "Email body content",
  "content_type": "text/plain"  // or "text/html"
}
```

### Send Email with Attachments
```
POST /email/send-with-attachments
Content-Type: multipart/form-data

Fields:
- to: recipient email (can be multiple)
- subject: email subject
- body: email body
- content_type: text/plain or text/html
- attachments: file attachments
```

### Get SMTP Configuration
```
GET /email/config
```

### Test SMTP Connection
```
POST /email/test
```

### Get Email Status
```
GET /email/status/:id
```

## Using with Maddy

1. Install maddy:
   ```bash
   # Download from https://github.com/foxcpp/maddy/releases
   # Or use your package manager
   ```

2. Copy the example configuration:
   ```bash
   sudo mkdir -p /etc/maddy
   sudo cp internal/email/maddy.conf.example /etc/maddy/maddy.conf
   ```

3. Generate TLS certificates (using Let's Encrypt or self-signed):
   ```bash
   # For local development, generate self-signed certificates
   openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
   sudo mkdir -p /etc/maddy/certs
   sudo cp cert.pem /etc/maddy/certs/fullchain.pem
   sudo cp key.pem /etc/maddy/certs/privkey.pem
   ```

4. Create a user account:
   ```bash
   sudo maddyctl creds create username@localhost
   sudo maddyctl creds password username@localhost
   ```

5. Start maddy:
   ```bash
   sudo maddy run
   ```

6. Configure Salada to use maddy:
   ```bash
   # In your .env file
   SMTP_HOST=localhost
   SMTP_PORT=587
   SMTP_FROM_ADDRESS=salada@localhost
   SMTP_FROM_NAME=Salada
   SMTP_USERNAME=your_username@localhost
   SMTP_PASSWORD=your_password
   SMTP_USE_STARTTLS=true
   ```

## Testing

Send a test email using curl:

```bash
curl -X POST http://localhost:8080/email/send \
  -H "Content-Type: application/json" \
  -H "Cookie: your_auth_cookie" \
  -d '{
    "to": ["test@example.com"],
    "subject": "Test Email",
    "body": "This is a test email from Salada",
    "content_type": "text/plain"
  }'
```

## Architecture

The email module follows the same pattern as other Salada modules:

- `email.go` - Core email client and SMTP logic
- `model/model.go` - Data structures for emails and requests
- `controller/controller.go` - HTTP handlers for email endpoints
- `routes/routes.go` - Route definitions

## Security Considerations

- All email endpoints require authentication (except where noted)
- SMTP credentials should be stored securely in environment variables
- Use TLS or STARTTLS for production deployments
- Consider rate limiting to prevent email abuse
