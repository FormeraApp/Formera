package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WebhookEvent string

const (
	WebhookEventSubmissionCreated WebhookEvent = "submission.created"
	WebhookEventSubmissionDeleted WebhookEvent = "submission.deleted"
)

type WebhookEvents []WebhookEvent

func (e WebhookEvents) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (e *WebhookEvents) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, e)
}

func (e WebhookEvents) Contains(event WebhookEvent) bool {
	for _, ev := range e {
		if ev == event {
			return true
		}
	}
	return false
}

type WebhookHeaders map[string]string

func (h WebhookHeaders) Value() (driver.Value, error) {
	return json.Marshal(h)
}

func (h *WebhookHeaders) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, h)
}

// DiscordConfig holds Discord-specific webhook configuration
type DiscordConfig struct {
	ShowFormTitle bool     `json:"show_form_title"` // Show form title in embed
	ShowSlug      bool     `json:"show_slug"`       // Show form slug in embed
	Fields        []string `json:"fields"`          // Field names to include (empty = all)
	CustomTitle   string   `json:"custom_title"`    // Custom embed title (optional)
}

func (d DiscordConfig) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *DiscordConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, d)
}

type Webhook struct {
	ID            string         `json:"id" gorm:"primaryKey"`
	FormID        *string        `json:"form_id,omitempty" gorm:"index"` // NULL = global webhook
	URL           string         `json:"url" gorm:"not null"`
	Secret        string         `json:"-" gorm:"size:255"` // HMAC secret, excluded from JSON
	Events        WebhookEvents  `json:"events" gorm:"type:json"`
	Headers       WebhookHeaders `json:"headers,omitempty" gorm:"type:json"`
	DiscordConfig *DiscordConfig `json:"discord_config,omitempty" gorm:"type:json"`
	Enabled       bool           `json:"enabled" gorm:"default:true"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func (w *Webhook) BeforeCreate(tx *gorm.DB) error {
	w.ID = uuid.New().String()
	return nil
}

// WebhookPayload represents the payload sent to webhook endpoints
type WebhookPayload struct {
	Event     WebhookEvent           `json:"event"`
	Timestamp time.Time              `json:"timestamp"`
	WebhookID string                 `json:"webhook_id"`
	Form      WebhookFormData        `json:"form"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

type WebhookFormData struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Slug  string `json:"slug,omitempty"`
}

// WebhookLog tracks webhook delivery attempts
type WebhookLog struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	WebhookID    string    `json:"webhook_id" gorm:"index;not null"`
	Event        string    `json:"event"`
	RequestBody  string    `json:"-" gorm:"type:text"`
	ResponseCode int       `json:"response_code"`
	ResponseBody string    `json:"-" gorm:"type:text"`
	Success      bool      `json:"success"`
	Attempt      int       `json:"attempt"`
	Error        string    `json:"error,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

func (l *WebhookLog) BeforeCreate(tx *gorm.DB) error {
	l.ID = uuid.New().String()
	return nil
}
