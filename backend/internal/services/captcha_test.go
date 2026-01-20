package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"formera/internal/models"
)

func TestCaptchaService_NoSecretsConfigured(t *testing.T) {
	service := NewCaptchaService("", "", "")

	// Should fail gracefully when no secrets are configured
	valid, err := service.VerifyToken(models.CaptchaProviderTurnstile, "test-token", "127.0.0.1", 0.5)
	if err == nil {
		t.Error("expected error when no secret is configured")
	}
	if valid {
		t.Error("expected validation to fail")
	}
}

func TestCaptchaService_Turnstile_Success(t *testing.T) {
	// Mock Turnstile API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"success": true,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewCaptchaService("test-turnstile-secret", "", "")
	service.turnstileURL = server.URL

	valid, err := service.VerifyToken(models.CaptchaProviderTurnstile, "test-token", "127.0.0.1", 0.5)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if !valid {
		t.Error("expected validation to succeed")
	}
}

func TestCaptchaService_Turnstile_Failure(t *testing.T) {
	// Mock Turnstile API server returning failure
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"success": false,
			"error-codes": []string{"invalid-input-response"},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewCaptchaService("test-turnstile-secret", "", "")
	service.turnstileURL = server.URL

	valid, err := service.VerifyToken(models.CaptchaProviderTurnstile, "invalid-token", "127.0.0.1", 0.5)
	if err != nil {
		t.Errorf("expected no error (just validation failure), got: %v", err)
	}
	if valid {
		t.Error("expected validation to fail")
	}
}

func TestCaptchaService_Recaptcha_Success(t *testing.T) {
	// Mock reCAPTCHA API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"success": true,
			"score":   0.9,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewCaptchaService("", "test-recaptcha-secret", "")
	service.recaptchaURL = server.URL

	valid, err := service.VerifyToken(models.CaptchaProviderRecaptcha, "test-token", "127.0.0.1", 0.5)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if !valid {
		t.Error("expected validation to succeed (score 0.9 > threshold 0.5)")
	}
}

func TestCaptchaService_Recaptcha_LowScore(t *testing.T) {
	// Mock reCAPTCHA API server with low score
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"success": true,
			"score":   0.3, // Below threshold
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewCaptchaService("", "test-recaptcha-secret", "")
	service.recaptchaURL = server.URL

	valid, err := service.VerifyToken(models.CaptchaProviderRecaptcha, "test-token", "127.0.0.1", 0.5)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if valid {
		t.Error("expected validation to fail (score 0.3 < threshold 0.5)")
	}
}

func TestCaptchaService_Recaptcha_Failure(t *testing.T) {
	// Mock reCAPTCHA API server returning failure
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"success":      false,
			"error-codes": []string{"invalid-input-response"},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewCaptchaService("", "test-recaptcha-secret", "")
	service.recaptchaURL = server.URL

	valid, err := service.VerifyToken(models.CaptchaProviderRecaptcha, "invalid-token", "127.0.0.1", 0.5)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if valid {
		t.Error("expected validation to fail")
	}
}

func TestCaptchaService_HCaptcha_Success(t *testing.T) {
	// Mock hCaptcha API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"success": true,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewCaptchaService("", "", "test-hcaptcha-secret")
	service.hcaptchaURL = server.URL

	valid, err := service.VerifyToken(models.CaptchaProviderHCaptcha, "test-token", "127.0.0.1", 0.5)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if !valid {
		t.Error("expected validation to succeed")
	}
}

func TestCaptchaService_HCaptcha_Failure(t *testing.T) {
	// Mock hCaptcha API server returning failure
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"success": false,
			"error-codes": []string{"invalid-input-response"},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	service := NewCaptchaService("", "", "test-hcaptcha-secret")
	service.hcaptchaURL = server.URL

	valid, err := service.VerifyToken(models.CaptchaProviderHCaptcha, "invalid-token", "127.0.0.1", 0.5)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if valid {
		t.Error("expected validation to fail")
	}
}

func TestCaptchaService_Timeout(t *testing.T) {
	t.Skip("Skipping timeout test as it takes 10+ seconds")

	// Mock slow server to test timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't respond - let it timeout
		select {}
	}))
	defer server.Close()

	service := NewCaptchaService("test-secret", "", "")
	service.turnstileURL = server.URL

	valid, err := service.VerifyToken(models.CaptchaProviderTurnstile, "test-token", "127.0.0.1", 0.5)
	if err == nil {
		t.Error("expected timeout error")
	}
	if valid {
		t.Error("expected validation to fail on timeout")
	}
}

func TestCaptchaService_InvalidProvider(t *testing.T) {
	service := NewCaptchaService("test", "test", "test")

	valid, err := service.VerifyToken("invalid-provider", "test-token", "127.0.0.1", 0.5)
	if err == nil {
		t.Error("expected error for invalid provider")
	}
	if valid {
		t.Error("expected validation to fail")
	}
}

func TestCaptchaService_EmptyToken(t *testing.T) {
	service := NewCaptchaService("test", "test", "test")

	providers := []models.CaptchaProvider{
		models.CaptchaProviderTurnstile,
		models.CaptchaProviderRecaptcha,
		models.CaptchaProviderHCaptcha,
	}

	for _, provider := range providers {
		t.Run(string(provider), func(t *testing.T) {
			valid, err := service.VerifyToken(provider, "", "127.0.0.1", 0.5)
			// Empty token should fail
			if valid {
				t.Error("expected validation to fail with empty token")
			}
			// Might or might not error depending on implementation
			_ = err
		})
	}
}
