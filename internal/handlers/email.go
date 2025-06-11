package handlers

import (
	"net/http"

	"market-watch-go/internal/services"

	"github.com/gin-gonic/gin"
)

type EmailHandler struct {
	emailService *services.EmailService
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(emailService *services.EmailService) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
	}
}

// GetEmailStatus returns the email service configuration status
func (eh *EmailHandler) GetEmailStatus(c *gin.Context) {
	status := eh.emailService.GetConfigStatus()

	if eh.emailService.IsConfigured() {
		c.JSON(http.StatusOK, gin.H{
			"status":  "configured",
			"config":  status,
			"message": "Email service is properly configured and ready to use",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status":  "not_configured",
			"config":  status,
			"message": "Email service requires configuration. Please update configs/config.yaml with your Gmail credentials",
			"setup_instructions": gin.H{
				"gmail_setup": []string{
					"1. Enable 2-Factor Authentication on your Gmail account",
					"2. Generate an App Password: Google Account > Security > App passwords",
					"3. Update configs/config.yaml with your Gmail address and App Password",
					"4. Restart the application",
				},
				"config_example": gin.H{
					"username": "your-email@gmail.com",
					"password": "your-16-char-app-password",
					"file":     "configs/config.yaml",
				},
			},
		})
	}
}

// SendTestEmail sends a test email to verify configuration
func (eh *EmailHandler) SendTestEmail(c *gin.Context) {
	var request struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if !eh.emailService.IsConfigured() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Email service not configured",
			"message": "Please update configs/config.yaml with your Gmail credentials",
		})
		return
	}

	if err := eh.emailService.SendTestEmail(request.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send test email",
			"details": err.Error(),
			"troubleshooting": gin.H{
				"common_issues": []string{
					"Invalid Gmail credentials",
					"App Password not generated or incorrect",
					"2-Factor Authentication not enabled",
					"SMTP settings incorrect",
					"Firewall blocking SMTP connection",
				},
				"gmail_app_password": "Visit: https://myaccount.google.com/apppasswords",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Test email sent successfully",
		"recipient": request.Email,
		"status":    "Email notifications are working correctly",
	})
}

// SendTradingAlert sends a trading setup alert email
func (eh *EmailHandler) SendTradingAlert(c *gin.Context) {
	var request struct {
		Email     string  `json:"email" binding:"required,email"`
		Symbol    string  `json:"symbol" binding:"required"`
		SetupType string  `json:"setup_type" binding:"required"`
		Score     float64 `json:"score" binding:"required"`
		Details   string  `json:"details"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if !eh.emailService.IsConfigured() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email service not configured",
		})
		return
	}

	if err := eh.emailService.SendTradingAlert(
		request.Email,
		request.Symbol,
		request.SetupType,
		request.Score,
		request.Details,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send trading alert",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Trading alert sent successfully",
		"recipient":  request.Email,
		"symbol":     request.Symbol,
		"setup_type": request.SetupType,
		"score":      request.Score,
	})
}
