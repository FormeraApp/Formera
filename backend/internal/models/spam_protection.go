package models

import (
	"database/sql/driver"
	"encoding/json"
)

type CaptchaProvider string

const (
	CaptchaProviderNone      CaptchaProvider = "none"
	CaptchaProviderTurnstile CaptchaProvider = "turnstile"
	CaptchaProviderRecaptcha CaptchaProvider = "recaptcha_v3"
	CaptchaProviderHCaptcha  CaptchaProvider = "hcaptcha"
)

type SpamProtectionConfig struct {
	// Honeypot is always enabled
	HoneypotEnabled bool `json:"honeypot_enabled"`

	// CAPTCHA configuration
	CaptchaProvider CaptchaProvider `json:"captcha_provider"`

	// Provider-specific settings (stored in database for UI configuration)
	// These are NOT secrets - they are public site keys
	TurnstileSiteKey string `json:"turnstile_site_key,omitempty"`
	RecaptchaSiteKey string `json:"recaptcha_site_key,omitempty"`
	HCaptchaSiteKey  string `json:"hcaptcha_site_key,omitempty"`

	// Recaptcha v3 specific
	RecaptchaMinScore float64 `json:"recaptcha_min_score"`
}

func (s SpamProtectionConfig) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *SpamProtectionConfig) Scan(value interface{}) error {
	if value == nil {
		*s = GetDefaultSpamProtectionConfig()
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		str, ok := value.(string)
		if !ok {
			*s = GetDefaultSpamProtectionConfig()
			return nil
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, s)
}

func GetDefaultSpamProtectionConfig() SpamProtectionConfig {
	return SpamProtectionConfig{
		HoneypotEnabled:   true,
		CaptchaProvider:   CaptchaProviderNone,
		RecaptchaMinScore: 0.5,
	}
}
