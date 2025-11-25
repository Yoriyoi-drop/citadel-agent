package communication

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"time"

	"citadel-agent/backend/internal/nodes/base"
)

// EmailNode implements email sending
type EmailNode struct {
	*base.BaseNode
}

// EmailConfig holds email configuration
type EmailConfig struct {
	SMTPHost string   `json:"smtp_host"`
	SMTPPort int      `json:"smtp_port"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	From     string   `json:"from"`
	To       []string `json:"to"`
	Subject  string   `json:"subject"`
	Body     string   `json:"body"`
	UseTLS   bool     `json:"use_tls"`
}

// NewEmailNode creates email sending node
func NewEmailNode() base.Node {
	metadata := base.NodeMetadata{
		ID:          "send_email",
		Name:        "Send Email",
		Category:    "communication",
		Description: "Send email via SMTP",
		Version:     "1.0.0",
		Author:      "Citadel Agent",
		Icon:        "mail",
		Color:       "#ec4899",
		Inputs: []base.NodeInput{
			{
				ID:          "trigger",
				Name:        "Trigger",
				Type:        "any",
				Required:    false,
				Description: "Trigger email send",
			},
		},
		Outputs: []base.NodeOutput{
			{
				ID:          "success",
				Name:        "Success",
				Type:        "boolean",
				Description: "Email sent successfully",
			},
		},
		Config: []base.NodeConfig{
			{
				Name:        "smtp_host",
				Label:       "SMTP Host",
				Description: "SMTP server host",
				Type:        "string",
				Required:    true,
			},
			{
				Name:        "smtp_port",
				Label:       "SMTP Port",
				Description: "SMTP server port",
				Type:        "number",
				Required:    true,
				Default:     587,
			},
			{
				Name:        "username",
				Label:       "Username",
				Description: "SMTP username",
				Type:        "string",
				Required:    true,
			},
			{
				Name:        "password",
				Label:       "Password",
				Description: "SMTP password",
				Type:        "password",
				Required:    true,
			},
			{
				Name:        "from",
				Label:       "From",
				Description: "Sender email",
				Type:        "string",
				Required:    true,
			},
			{
				Name:        "to",
				Label:       "To",
				Description: "Recipient emails (comma-separated)",
				Type:        "string",
				Required:    true,
			},
			{
				Name:        "subject",
				Label:       "Subject",
				Description: "Email subject",
				Type:        "string",
				Required:    true,
			},
			{
				Name:        "body",
				Label:       "Body",
				Description: "Email body",
				Type:        "textarea",
				Required:    true,
			},
			{
				Name:        "use_tls",
				Label:       "Use TLS",
				Description: "Use TLS encryption",
				Type:        "boolean",
				Required:    false,
				Default:     true,
			},
		},
		Tags: []string{"email", "smtp", "communication"},
	}

	return &EmailNode{
		BaseNode: base.NewBaseNode(metadata),
	}
}

// Execute sends email
func (n *EmailNode) Execute(ctx *base.ExecutionContext, inputs map[string]interface{}) (*base.ExecutionResult, error) {
	startTime := time.Now()

	// Parse configuration
	var config EmailConfig
	if err := base.UnmarshalConfig(ctx.Variables, &config); err != nil {
		return base.CreateErrorResult(err, time.Since(startTime)), err
	}

	// Build email message
	message := fmt.Sprintf("From: %s\r\n", config.From)
	message += fmt.Sprintf("To: %s\r\n", config.To[0])
	message += fmt.Sprintf("Subject: %s\r\n", config.Subject)
	message += "\r\n" + config.Body

	// Setup authentication
	auth := smtp.PlainAuth("", config.Username, config.Password, config.SMTPHost)

	// Send email
	addr := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)

	var err error
	if config.UseTLS {
		// TLS connection
		tlsConfig := &tls.Config{
			ServerName: config.SMTPHost,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return base.CreateErrorResult(err, time.Since(startTime)), err
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, config.SMTPHost)
		if err != nil {
			return base.CreateErrorResult(err, time.Since(startTime)), err
		}
		defer client.Quit()

		if err = client.Auth(auth); err != nil {
			return base.CreateErrorResult(err, time.Since(startTime)), err
		}

		if err = client.Mail(config.From); err != nil {
			return base.CreateErrorResult(err, time.Since(startTime)), err
		}

		for _, to := range config.To {
			if err = client.Rcpt(to); err != nil {
				return base.CreateErrorResult(err, time.Since(startTime)), err
			}
		}

		w, err := client.Data()
		if err != nil {
			return base.CreateErrorResult(err, time.Since(startTime)), err
		}

		_, err = w.Write([]byte(message))
		if err != nil {
			return base.CreateErrorResult(err, time.Since(startTime)), err
		}

		err = w.Close()
		if err != nil {
			return base.CreateErrorResult(err, time.Since(startTime)), err
		}
	} else {
		// Plain SMTP
		err = smtp.SendMail(addr, auth, config.From, config.To, []byte(message))
		if err != nil {
			return base.CreateErrorResult(err, time.Since(startTime)), err
		}
	}

	result := map[string]interface{}{
		"success": true,
		"to":      config.To,
		"subject": config.Subject,
	}

	ctx.Logger.Info("Email sent successfully", map[string]interface{}{
		"to":      config.To,
		"subject": config.Subject,
	})

	return base.CreateSuccessResult(result, time.Since(startTime)), nil
}
