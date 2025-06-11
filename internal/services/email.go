package services

import (
	"fmt"
	"net/smtp"
	"strings"

	"market-watch-go/internal/config"
)

type EmailService struct {
	config *config.EmailConfig
}

type EmailMessage struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

// NewEmailService creates a new email service
func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{
		config: &cfg.Email,
	}
}

// SendEmail sends an email using Gmail SMTP
func (e *EmailService) SendEmail(message *EmailMessage) error {
	if !e.config.Enabled {
		return fmt.Errorf("email service is disabled in configuration")
	}

	if e.config.Username == "" || e.config.Password == "" {
		return fmt.Errorf("email credentials not configured")
	}

	// Set up authentication information
	auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.SMTPHost)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step
	to := strings.Join(message.To, ",")

	// Construct email headers
	contentType := "text/plain"
	if message.IsHTML {
		contentType = "text/html"
	}

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s <%s>\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: %s; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n",
		to,
		e.config.FromName,
		e.config.FromAddress,
		message.Subject,
		contentType,
		message.Body))

	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", e.config.SMTPHost, e.config.SMTPPort),
		auth,
		e.config.FromAddress,
		message.To,
		msg,
	)

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendTestEmail sends a test email to verify configuration
func (e *EmailService) SendTestEmail(recipient string) error {
	message := &EmailMessage{
		To:      []string{recipient},
		Subject: "Market Watch - Email Configuration Test",
		Body: `This is a test email from your Market Watch application.

If you received this email, your email notifications are configured correctly!

Configuration details:
- SMTP Host: ` + e.config.SMTPHost + `
- SMTP Port: ` + fmt.Sprintf("%d", e.config.SMTPPort) + `
- From: ` + e.config.FromName + ` <` + e.config.FromAddress + `>

You can now receive email notifications for:
- Trading setup alerts
- Price alerts
- System notifications
- Data collection status updates

Best regards,
Market Watch System`,
		IsHTML: false,
	}

	return e.SendEmail(message)
}

// SendTradingAlert sends a trading setup alert email
func (e *EmailService) SendTradingAlert(recipient, symbol, setupType string, score float64, details string) error {
	subject := fmt.Sprintf("🚨 Trading Alert: %s - %s Setup (Score: %.1f)", symbol, setupType, score)

	body := fmt.Sprintf(`Trading Setup Alert

Symbol: %s
Setup Type: %s
Quality Score: %.1f/100

Details:
%s

This is an automated alert from your Market Watch system.
Please verify all information before making any trading decisions.

Timestamp: %s
Dashboard: http://localhost:8080/

Best regards,
Market Watch System`,
		symbol,
		setupType,
		score,
		details,
		fmt.Sprintf("%v", "now")) // You can replace with actual timestamp

	message := &EmailMessage{
		To:      []string{recipient},
		Subject: subject,
		Body:    body,
		IsHTML:  false,
	}

	return e.SendEmail(message)
}

// IsConfigured checks if email service is properly configured
func (e *EmailService) IsConfigured() bool {
	return e.config.Enabled &&
		e.config.Username != "" &&
		e.config.Password != "" &&
		e.config.SMTPHost != "" &&
		e.config.SMTPPort > 0
}

// GetConfigStatus returns the configuration status for debugging
func (e *EmailService) GetConfigStatus() map[string]interface{} {
	return map[string]interface{}{
		"enabled":      e.config.Enabled,
		"smtp_host":    e.config.SMTPHost,
		"smtp_port":    e.config.SMTPPort,
		"from_name":    e.config.FromName,
		"from_address": e.config.FromAddress,
		"configured":   e.IsConfigured(),
		"username_set": e.config.Username != "",
		"password_set": e.config.Password != "",
	}
}
