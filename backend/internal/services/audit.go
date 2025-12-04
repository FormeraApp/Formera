package services

import (
	"encoding/json"

	"formera/internal/database"
	"formera/internal/models"
	"formera/internal/pkg"

	"github.com/gin-gonic/gin"
)

// LogAuthEvent logs an authentication-related event
func LogAuthEvent(c *gin.Context, action models.AuditAction, userID *string, email string, details map[string]interface{}) {
	detailsJSON := ""
	if details != nil {
		if jsonBytes, err := json.Marshal(details); err == nil {
			detailsJSON = string(jsonBytes)
		}
	}

	audit := &models.AuditLog{
		UserID:    userID,
		Email:     email,
		Action:    action,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Details:   detailsJSON,
	}

	if err := database.DB.Create(audit).Error; err != nil {
		pkg.LogError().Err(err).Str("action", string(action)).Msg("Failed to create audit log")
	}
}

// LogLogin logs a successful login
func LogLogin(c *gin.Context, userID string, email string) {
	LogAuthEvent(c, models.AuditActionLogin, &userID, email, nil)
}

// LogLoginFailed logs a failed login attempt
func LogLoginFailed(c *gin.Context, email string, reason string) {
	LogAuthEvent(c, models.AuditActionLoginFailed, nil, email, map[string]interface{}{
		"reason": reason,
	})
}

// LogAccountLocked logs when an account gets locked
func LogAccountLocked(c *gin.Context, userID string, email string) {
	LogAuthEvent(c, models.AuditActionAccountLocked, &userID, email, nil)
}

// LogRegister logs a new user registration
func LogRegister(c *gin.Context, userID string, email string) {
	LogAuthEvent(c, models.AuditActionRegister, &userID, email, nil)
}

// LogSetupComplete logs initial setup completion
func LogSetupComplete(c *gin.Context, userID string, email string) {
	LogAuthEvent(c, models.AuditActionSetupComplete, &userID, email, nil)
}
