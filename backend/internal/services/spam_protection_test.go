package services

import (
	"testing"

	"formera/internal/models"
)

func TestSpamProtectionService_CheckHoneypot_Empty(t *testing.T) {
	captchaService := NewCaptchaService("", "", "")
	service := NewSpamProtectionService(captchaService)

	config := models.SpamProtectionConfig{
		HoneypotEnabled: true,
		CaptchaProvider: models.CaptchaProviderNone,
	}

	// Empty honeypot should pass
	err := service.CheckSubmission(config, "", "", "127.0.0.1")
	if err != nil {
		t.Errorf("expected no error for empty honeypot, got: %v", err)
	}
}

func TestSpamProtectionService_CheckHoneypot_Filled(t *testing.T) {
	captchaService := NewCaptchaService("", "", "")
	service := NewSpamProtectionService(captchaService)

	config := models.SpamProtectionConfig{
		HoneypotEnabled: true,
		CaptchaProvider: models.CaptchaProviderNone,
	}

	// Filled honeypot should fail
	err := service.CheckSubmission(config, "bot-filled-this", "", "127.0.0.1")
	if err == nil {
		t.Error("expected error for filled honeypot, got nil")
	}
	if err.Error() != "spam detected" {
		t.Errorf("expected 'spam detected' error, got: %v", err)
	}
}

func TestSpamProtectionService_NoCaptcha(t *testing.T) {
	captchaService := NewCaptchaService("", "", "")
	service := NewSpamProtectionService(captchaService)

	config := models.SpamProtectionConfig{
		HoneypotEnabled: true,
		CaptchaProvider: models.CaptchaProviderNone,
	}

	// With no CAPTCHA provider, should pass with empty honeypot
	err := service.CheckSubmission(config, "", "", "127.0.0.1")
	if err != nil {
		t.Errorf("expected no error with no CAPTCHA, got: %v", err)
	}
}

func TestSpamProtectionService_CaptchaRequired_NoToken(t *testing.T) {
	captchaService := NewCaptchaService("test-secret", "", "")
	service := NewSpamProtectionService(captchaService)

	testCases := []struct {
		name     string
		provider models.CaptchaProvider
	}{
		{"Turnstile", models.CaptchaProviderTurnstile},
		{"reCAPTCHA", models.CaptchaProviderRecaptcha},
		{"hCaptcha", models.CaptchaProviderHCaptcha},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := models.SpamProtectionConfig{
				HoneypotEnabled: true,
				CaptchaProvider: tc.provider,
			}

			// Missing CAPTCHA token should fail
			err := service.CheckSubmission(config, "", "", "127.0.0.1")
			if err == nil {
				t.Errorf("expected error for missing CAPTCHA token with provider %s, got nil", tc.provider)
			}
			if err.Error() != "CAPTCHA verification required" {
				t.Errorf("expected 'CAPTCHA verification required' error, got: %v", err)
			}
		})
	}
}

func TestSpamProtectionService_HoneypotPriority(t *testing.T) {
	captchaService := NewCaptchaService("test-secret", "", "")
	service := NewSpamProtectionService(captchaService)

	config := models.SpamProtectionConfig{
		HoneypotEnabled: true,
		CaptchaProvider: models.CaptchaProviderTurnstile,
	}

	// Filled honeypot should fail before CAPTCHA check
	err := service.CheckSubmission(config, "bot-value", "valid-token", "127.0.0.1")
	if err == nil {
		t.Error("expected error for filled honeypot, got nil")
	}
	if err.Error() != "spam detected" {
		t.Errorf("expected 'spam detected' error (honeypot should be checked first), got: %v", err)
	}
}

func TestSpamProtectionService_DefaultMinScore(t *testing.T) {
	captchaService := NewCaptchaService("", "test-secret", "")
	service := NewSpamProtectionService(captchaService)

	config := models.SpamProtectionConfig{
		HoneypotEnabled:   true,
		CaptchaProvider:   models.CaptchaProviderRecaptcha,
		RecaptchaMinScore: 0, // Should default to 0.5
	}

	// This will fail because we don't have a real token, but it tests the default score logic
	err := service.CheckSubmission(config, "", "fake-token", "127.0.0.1")
	if err == nil {
		t.Error("expected error due to invalid token")
	}
	// We expect an error but not a panic due to zero score
}

func TestSpamProtectionService_ConfigValidation(t *testing.T) {
	captchaService := NewCaptchaService("", "", "")
	service := NewSpamProtectionService(captchaService)

	testCases := []struct {
		name        string
		config      models.SpamProtectionConfig
		honeypot    string
		token       string
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid: No CAPTCHA, empty honeypot",
			config: models.SpamProtectionConfig{
				HoneypotEnabled: true,
				CaptchaProvider: models.CaptchaProviderNone,
			},
			honeypot:    "",
			token:       "",
			expectError: false,
		},
		{
			name: "Invalid: Honeypot filled",
			config: models.SpamProtectionConfig{
				HoneypotEnabled: true,
				CaptchaProvider: models.CaptchaProviderNone,
			},
			honeypot:    "spam",
			token:       "",
			expectError: true,
			errorMsg:    "spam detected",
		},
		{
			name: "Invalid: CAPTCHA required but no token",
			config: models.SpamProtectionConfig{
				HoneypotEnabled: true,
				CaptchaProvider: models.CaptchaProviderTurnstile,
			},
			honeypot:    "",
			token:       "",
			expectError: true,
			errorMsg:    "CAPTCHA verification required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.CheckSubmission(tc.config, tc.honeypot, tc.token, "127.0.0.1")

			if tc.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
			if tc.expectError && err != nil && tc.errorMsg != "" {
				if err.Error() != tc.errorMsg {
					t.Errorf("expected error '%s', got: %v", tc.errorMsg, err)
				}
			}
		})
	}
}
