package communication

import (
	"context"
	"fmt"
	"net/smtp"
)

// EmailClient handles email sending
type EmailClient struct {
	smtpHost string
	smtpPort string
	username string
	password string
}

// NewEmailClient creates a new email client
func NewEmailClient(smtpHost, smtpPort, username, password string) *EmailClient {
	return &EmailClient{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		username: username,
		password: password,
	}
}

// SendEmail sends an email
func (e *EmailClient) SendEmail(ctx context.Context, to, subject, body string) error {
	from := e.username
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", to, subject, body))

	auth := smtp.PlainAuth("", e.username, e.password, e.smtpHost)
	addr := fmt.Sprintf("%s:%s", e.smtpHost, e.smtpPort)

	err := smtp.SendMail(addr, auth, from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
