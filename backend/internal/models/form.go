package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FormStatus string

const (
	FormStatusDraft     FormStatus = "draft"
	FormStatusPublished FormStatus = "published"
	FormStatusClosed    FormStatus = "closed"
)

type FieldType string

const (
	// Input fields
	FieldTypeText     FieldType = "text"
	FieldTypeTextarea FieldType = "textarea"
	FieldTypeNumber   FieldType = "number"
	FieldTypeEmail    FieldType = "email"
	FieldTypePhone    FieldType = "phone"
	FieldTypeDate     FieldType = "date"
	FieldTypeTime     FieldType = "time"
	FieldTypeURL      FieldType = "url"
	FieldTypeRichtext FieldType = "richtext"
	// Choice fields
	FieldTypeSelect   FieldType = "select"
	FieldTypeRadio    FieldType = "radio"
	FieldTypeCheckbox FieldType = "checkbox"
	FieldTypeDropdown FieldType = "dropdown"
	// Special fields
	FieldTypeFile      FieldType = "file"
	FieldTypeRating    FieldType = "rating"
	FieldTypeScale     FieldType = "scale"
	FieldTypeSignature FieldType = "signature"
	// Layout fields
	FieldTypeSection   FieldType = "section"
	FieldTypePagebreak FieldType = "pagebreak"
	FieldTypeDivider   FieldType = "divider"
	FieldTypeHeading   FieldType = "heading"
	FieldTypeParagraph FieldType = "paragraph"
	FieldTypeImage     FieldType = "image"
)

type FormField struct {
	ID          string                 `json:"id"`
	Type        FieldType              `json:"type"`
	Label       string                 `json:"label"`
	Placeholder string                 `json:"placeholder,omitempty"`
	Required    bool                   `json:"required"`
	Options     []string               `json:"options,omitempty"`
	Validation  map[string]interface{} `json:"validation,omitempty"`
	Order       int                    `json:"order"`
	// Description/Help text
	Description string `json:"description,omitempty"`
	// Section-specific
	SectionTitle       string `json:"sectionTitle,omitempty"`
	SectionDescription string `json:"sectionDescription,omitempty"`
	Collapsible        bool   `json:"collapsible,omitempty"`
	Collapsed          bool   `json:"collapsed,omitempty"`
	// Layout-specific
	Content      string `json:"content,omitempty"`
	HeadingLevel int    `json:"headingLevel,omitempty"`
	ImageURL     string `json:"imageUrl,omitempty"`
	ImageAlt     string `json:"imageAlt,omitempty"`
	// Rich Text
	RichTextContent string `json:"richTextContent,omitempty"`
	// Rating/Scale
	MinValue int    `json:"minValue,omitempty"`
	MaxValue int    `json:"maxValue,omitempty"`
	MinLabel string `json:"minLabel,omitempty"`
	MaxLabel string `json:"maxLabel,omitempty"`
	// File Upload
	AllowedTypes []string `json:"allowedTypes,omitempty"`
	MaxFileSize  int      `json:"maxFileSize,omitempty"`
	Multiple     bool     `json:"multiple,omitempty"`
}

type FormFields []FormField

func (f FormFields) Value() (driver.Value, error) {
	return json.Marshal(f)
}

func (f *FormFields) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, f)
}

type FormDesign struct {
	PrimaryColor        string `json:"primaryColor,omitempty"`
	BackgroundColor     string `json:"backgroundColor,omitempty"`
	FormBackgroundColor string `json:"formBackgroundColor,omitempty"`
	TextColor           string `json:"textColor,omitempty"`
	BackgroundImage     string `json:"backgroundImage,omitempty"`
	BackgroundSize      string `json:"backgroundSize,omitempty"`
	BackgroundPosition  string `json:"backgroundPosition,omitempty"`
	MaxWidth            string `json:"maxWidth,omitempty"`
	BorderRadius        string `json:"borderRadius,omitempty"`
	HeaderStyle         string `json:"headerStyle,omitempty"`
	ButtonStyle         string `json:"buttonStyle,omitempty"`
	FontFamily          string `json:"fontFamily,omitempty"`
}

type FormSettings struct {
	SubmitButtonText    string      `json:"submit_button_text"`
	SuccessMessage      string      `json:"success_message"`
	AllowMultiple       bool        `json:"allow_multiple"`
	RequireLogin        bool        `json:"require_login"`
	NotifyOnSubmission  bool        `json:"notify_on_submission"`
	NotificationEmail   string      `json:"notification_email,omitempty"`
	MaxSubmissions      int         `json:"max_submissions,omitempty"`
	StartDate           string      `json:"start_date,omitempty"`
	EndDate             string      `json:"end_date,omitempty"`
	Design              *FormDesign `json:"design,omitempty"`
}

func (s FormSettings) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *FormSettings) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

type Form struct {
	ID          string       `json:"id" gorm:"primaryKey"`
	UserID      string       `json:"user_id" gorm:"index;not null"`
	Title       string       `json:"title" gorm:"not null"`
	Description string       `json:"description"`
	Slug        string       `json:"slug,omitempty" gorm:"uniqueIndex;size:100"`
	Fields      FormFields   `json:"fields" gorm:"type:json"`
	Settings    FormSettings `json:"settings" gorm:"type:json"`
	Status      FormStatus   `json:"status" gorm:"default:draft"`
	// Password protection
	PasswordProtected bool   `json:"password_protected" gorm:"default:false"`
	PasswordHash      string `json:"-" gorm:"size:255"` // Never expose hash in JSON
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
	Submissions       []Submission `json:"submissions,omitempty" gorm:"foreignKey:FormID"`
}

func (f *Form) BeforeCreate(tx *gorm.DB) error {
	f.ID = uuid.New().String()
	if f.Status == "" {
		f.Status = FormStatusDraft
	}
	if f.Settings.SubmitButtonText == "" {
		f.Settings.SubmitButtonText = "Absenden"
	}
	if f.Settings.SuccessMessage == "" {
		f.Settings.SuccessMessage = "Vielen Dank für Ihre Antwort!"
	}
	return nil
}
