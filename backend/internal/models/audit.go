package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditAction represents the type of auditable action
type AuditAction string

const (
	AuditActionLogin          AuditAction = "login"
	AuditActionLoginFailed    AuditAction = "login_failed"
	AuditActionLogout         AuditAction = "logout"
	AuditActionRegister       AuditAction = "register"
	AuditActionPasswordChange AuditAction = "password_change"
	AuditActionAccountLocked  AuditAction = "account_locked"
	AuditActionSetupComplete  AuditAction = "setup_complete"
)

// AuditLog stores security-relevant events
type AuditLog struct {
	ID        string      `json:"id" gorm:"primaryKey"`
	UserID    *string     `json:"user_id" gorm:"index"` // Nullable for failed logins
	Email     string      `json:"email" gorm:"index"`   // Email for identification
	Action    AuditAction `json:"action" gorm:"index"`
	IPAddress string      `json:"ip_address"`
	UserAgent string      `json:"user_agent"`
	Details   string      `json:"details"` // Additional details (JSON)
	CreatedAt time.Time   `json:"created_at" gorm:"index"`
}

func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	a.ID = uuid.New().String()
	return nil
}
