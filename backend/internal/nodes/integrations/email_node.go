// backend/internal/nodes/integrations/email_node.go
package integrations

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"citadel-agent/backend/internal/workflow/core/engine"
)

// EmailNodeConfig represents the configuration for an email node
type EmailNodeConfig struct {
	SMTPServer   string   `json:"smtp_server"`
	SMTPPort     int      `json:"smtp_port"`
	SMTPUsername string   `json:"smtp_username"`
	SMTPPassword string   `json:"smtp_password"`
	FromAddress  string   `json:"from_address"`
	FromName     string   `json:"from_name"`
	Subject      string   `json:"subject"`
	Body         string   `json:"body"`
	ContentType  string   `json:"content_type"` // "text/plain" or "text/html"
	Recipients   []string `json:"recipients"`
	CCRecipients []string `json:"cc_recipients"`
	BCCRecipients []string `json:"bcc_recipients"`
	Attachments  []string `json:"attachments"` // file paths
	EnableTLS    bool     `json:"enable_tls"`
}

// EmailNode represents an email integration node
type EmailNode struct {
	config *EmailNodeConfig
}

// NewEmailNode creates a new email node
func NewEmailNode(config *EmailNodeConfig) *EmailNode {
	if config.ContentType == "" {
		config.ContentType = "text/plain"
	}
	if config.SMTPPort == 0 {
		config.SMTPPort = 587 // Default SMTP port
	}
	if config.FromName == "" {
		config.FromName = "Citadel Agent"
	}
	if config.EnableTLS {
		config.EnableTLS = true
	}

	return &EmailNode{
		config: config,
	}
}

// Execute executes the email operation
func (en *EmailNode) Execute(ctx context.Context, inputs map[string]interface{}) (map[string]interface{}, error) {
	// Override config values with inputs if provided
	fromAddress := en.config.FromAddress
	if addr, exists := inputs["from_address"]; exists {
		if addrStr, ok := addr.(string); ok {
			fromAddress = addrStr
		}
	}

	fromName := en.config.FromName
	if name, exists := inputs["from_name"]; exists {
		if nameStr, ok := name.(string); ok {
			fromName = nameStr
		}
	}

	subject := en.config.Subject
	if subj, exists := inputs["subject"]; exists {
		if subjStr, ok := subj.(string); ok {
			subject = subjStr
		}
	}

	body := en.config.Body
	if b, exists := inputs["body"]; exists {
		if bStr, ok := b.(string); ok {
			body = bStr
		}
	}

	contentType := en.config.ContentType
	if ct, exists := inputs["content_type"]; exists {
		if ctStr, ok := ct.(string); ok {
			contentType = ctStr
		}
	}

	recipients := en.config.Recipients
	if rec, exists := inputs["recipients"]; exists {
		if recList, ok := rec.([]interface{}); ok {
			newRecipients := make([]string, len(recList))
			for i, r := range recList {
				if rStr, ok := r.(string); ok {
					newRecipients[i] = rStr
				}
			}
			recipients = newRecipients
		}
	}

	ccRecipients := en.config.CCRecipients
	if cc, exists := inputs["cc_recipients"]; exists {
		if ccList, ok := cc.([]interface{}); ok {
			newRecipients := make([]string, len(ccList))
			for i, r := range ccList {
				if rStr, ok := r.(string); ok {
					newRecipients[i] = rStr
				}
			}
			ccRecipients = newRecipients
		}
	}

	bccRecipients := en.config.BCCRecipients
	if bcc, exists := inputs["bcc_recipients"]; exists {
		if bccList, ok := bcc.([]interface{}); ok {
			newRecipients := make([]string, len(bccList))
			for i, r := range bccList {
				if rStr, ok := r.(string); ok {
					newRecipients[i] = rStr
				}
			}
			bccRecipients = newRecipients
		}
	}

	// Validate required fields
	if fromAddress == "" {
		return nil, fmt.Errorf("from address is required")
	}
	if subject == "" {
		return nil, fmt.Errorf("subject is required")
	}
	if len(recipients) == 0 {
		return nil, fmt.Errorf("at least one recipient is required")
	}

	// Create the email message
	message := en.buildEmailMessage(fromAddress, fromName, recipients, ccRecipients, bccRecipients, subject, body, contentType)

	// Configure SMTP authentication
	auth := smtp.PlainAuth("", en.config.SMTPUsername, en.config.SMTPPassword, en.config.SMTPServer)

	// Prepare addresses
	allRecipients := append(recipients, ccRecipients...)
	allRecipients = append(allRecipients, bccRecipients...)

	// Send the email
	var err error
	if en.config.EnableTLS {
		err = en.sendTLSEmail(auth, allRecipients, []byte(message))
	} else {
		err = smtp.SendMail(
			fmt.Sprintf("%s:%d", en.config.SMTPServer, en.config.SMTPPort),
			auth,
			fromAddress,
			allRecipients,
			[]byte(message),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	result := map[string]interface{}{
		"success":       true,
		"message":       "Email sent successfully",
		"recipients":    recipients,
		"cc_recipients": ccRecipients,
		"bcc_recipients": bccRecipients,
		"subject":       subject,
		"timestamp":     time.Now().Unix(),
		"delivery_info": map[string]interface{}{
			"from":    fromAddress,
			"to":      recipients,
			"server":  en.config.SMTPServer,
			"port":    en.config.SMTPPort,
			"tls":     en.config.EnableTLS,
		},
	}

	return result, nil
}

// buildEmailMessage constructs the email message with proper headers
func (en *EmailNode) buildEmailMessage(
	fromAddress, fromName string,
	toAddresses, ccAddresses, bccAddresses []string,
	subject, body, contentType string,
) string {
	// Format the from header
	var fromHeader string
	if fromName != "" {
		fromHeader = fmt.Sprintf("\"%s\" <%s>", fromName, fromAddress)
	} else {
		fromHeader = fromAddress
	}

	// Prepare headers
	headers := make(map[string]string)
	headers["From"] = fromHeader
	headers["To"] = strings.Join(toAddresses, ", ")
	if len(ccAddresses) > 0 {
		headers["Cc"] = strings.Join(ccAddresses, ", ")
	}
	headers["Subject"] = subject
	headers["Content-Type"] = contentType + "; charset=utf-8"
	headers["MIME-Version"] = "1.0"
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	// Build the message string
	var message string
	for key, value := range headers {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" // Separates headers from body
	message += body

	return message
}

// sendTLSEmail sends an email using TLS
func (en *EmailNode) sendTLSEmail(auth smtp.Auth, recipients []string, message []byte) error {
	client, err := smtp.Dial(fmt.Sprintf("%s:%d", en.config.SMTPServer, en.config.SMTPPort))
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.Hello("localhost"); err != nil {
		return err
	}

	if en.config.EnableTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err = client.StartTLS(nil); err != nil {
				return err
			}
		}
	}

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(en.config.FromAddress); err != nil {
		return err
	}

	for _, rcpt := range recipients {
		if err = client.Rcpt(rcpt); err != nil {
			return err
		}
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}

	_, err = writer.Write(message)
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	return client.Quit()
}

// RegisterEmailNode registers the Email node type with the engine
func RegisterEmailNode(registry *engine.NodeRegistry) {
	registry.RegisterNodeType("email_integration", func(config map[string]interface{}) (engine.NodeInstance, error) {
		var smtpServer string
		if server, exists := config["smtp_server"]; exists {
			if serverStr, ok := server.(string); ok {
				smtpServer = serverStr
			}
		}

		var smtpPort float64
		if port, exists := config["smtp_port"]; exists {
			if portNum, ok := port.(float64); ok {
				smtpPort = portNum
			}
		}

		var smtpUsername string
		if username, exists := config["smtp_username"]; exists {
			if usernameStr, ok := username.(string); ok {
				smtpUsername = usernameStr
			}
		}

		var smtpPassword string
		if password, exists := config["smtp_password"]; exists {
			if passwordStr, ok := password.(string); ok {
				smtpPassword = passwordStr
			}
		}

		var fromAddress string
		if addr, exists := config["from_address"]; exists {
			if addrStr, ok := addr.(string); ok {
				fromAddress = addrStr
			}
		}

		var fromName string
		if name, exists := config["from_name"]; exists {
			if nameStr, ok := name.(string); ok {
				fromName = nameStr
			}
		}

		var subject string
		if subj, exists := config["subject"]; exists {
			if subjStr, ok := subj.(string); ok {
				subject = subjStr
			}
		}

		var body string
		if b, exists := config["body"]; exists {
			if bStr, ok := b.(string); ok {
				body = bStr
			}
		}

		var contentType string
		if ct, exists := config["content_type"]; exists {
			if ctStr, ok := ct.(string); ok {
				contentType = ctStr
			}
		}

		var recipients []string
		if recList, exists := config["recipients"]; exists {
			if recSlice, ok := recList.([]interface{}); ok {
				for _, r := range recSlice {
					if rStr, ok := r.(string); ok {
						recipients = append(recipients, rStr)
					}
				}
			}
		}

		var ccRecipients []string
		if ccList, exists := config["cc_recipients"]; exists {
			if ccSlice, ok := ccList.([]interface{}); ok {
				for _, r := range ccSlice {
					if rStr, ok := r.(string); ok {
						ccRecipients = append(ccRecipients, rStr)
					}
				}
			}
		}

		var bccRecipients []string
		if bccList, exists := config["bcc_recipients"]; exists {
			if bccSlice, ok := bccList.([]interface{}); ok {
				for _, r := range bccSlice {
					if rStr, ok := r.(string); ok {
						bccRecipients = append(bccRecipients, rStr)
					}
				}
			}
		}

		var attachments []string
		if attachList, exists := config["attachments"]; exists {
			if attachSlice, ok := attachList.([]interface{}); ok {
				for _, a := range attachSlice {
					if aStr, ok := a.(string); ok {
						attachments = append(attachments, aStr)
					}
				}
			}
		}

		var enableTLS bool
		if tls, exists := config["enable_tls"]; exists {
			if tlsBool, ok := tls.(bool); ok {
				enableTLS = tlsBool
			}
		}

		nodeConfig := &EmailNodeConfig{
			SMTPServer:   smtpServer,
			SMTPPort:     int(smtpPort),
			SMTPUsername: smtpUsername,
			SMTPPassword: smtpPassword,
			FromAddress:  fromAddress,
			FromName:     fromName,
			Subject:      subject,
			Body:         body,
			ContentType:  contentType,
			Recipients:   recipients,
			CCRecipients: ccRecipients,
			BCCRecipients: bccRecipients,
			Attachments:  attachments,
			EnableTLS:    enableTLS,
		}

		return NewEmailNode(nodeConfig), nil
	})
}