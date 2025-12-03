package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type FooterLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type FooterLinks []FooterLink

func (f FooterLinks) Value() (driver.Value, error) {
	if f == nil {
		return "[]", nil
	}
	return json.Marshal(f)
}

func (f *FooterLinks) Scan(value interface{}) error {
	if value == nil {
		*f = FooterLinks{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			*f = FooterLinks{}
			return nil
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, f)
}

type Settings struct {
	ID                uint        `json:"id" gorm:"primaryKey"`
	AllowRegistration bool        `json:"allow_registration" gorm:"default:true"`
	SetupCompleted    bool        `json:"setup_completed" gorm:"default:false"`
	AppName           string      `json:"app_name" gorm:"default:Formera"`
	FooterLinks       FooterLinks `json:"footer_links" gorm:"type:text"`
	// Customization
	PrimaryColor       string `json:"primary_color" gorm:"default:#6366f1"`
	LogoURL            string `json:"logo_url"`
	LogoShowText       bool   `json:"logo_show_text" gorm:"default:true"`
	FaviconURL         string `json:"favicon_url"`
	LoginBackgroundURL string `json:"login_background_url"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func GetDefaultSettings() *Settings {
	return &Settings{
		ID:                 1,
		AllowRegistration:  true,
		SetupCompleted:     false,
		AppName:            "Formera",
		FooterLinks:        FooterLinks{},
		PrimaryColor:       "#6366f1",
		LogoURL:            "",
		LogoShowText:       true,
		FaviconURL:         "",
		LoginBackgroundURL: "",
	}
}
