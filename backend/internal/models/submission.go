package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubmissionData map[string]interface{}

func (s SubmissionData) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *SubmissionData) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

type SubmissionMetadata struct {
	IP        string `json:"ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	Referrer  string `json:"referrer,omitempty"`
	// UTM/Tracking parameters
	UTMSource   string            `json:"utm_source,omitempty"`
	UTMMedium   string            `json:"utm_medium,omitempty"`
	UTMCampaign string            `json:"utm_campaign,omitempty"`
	UTMTerm     string            `json:"utm_term,omitempty"`
	UTMContent  string            `json:"utm_content,omitempty"`
	Tracking    map[string]string `json:"tracking,omitempty"`
}

func (s SubmissionMetadata) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *SubmissionMetadata) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

type Submission struct {
	ID        string             `json:"id" gorm:"primaryKey"`
	FormID    string             `json:"form_id" gorm:"index;not null"`
	Data      SubmissionData     `json:"data" gorm:"type:json"`
	Metadata  SubmissionMetadata `json:"metadata" gorm:"type:json"`
	CreatedAt time.Time          `json:"created_at"`
}

func (s *Submission) BeforeCreate(tx *gorm.DB) error {
	s.ID = uuid.New().String()
	return nil
}
