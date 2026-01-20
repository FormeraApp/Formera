package services

import (
	"encoding/json"
	"errors"
	"formera/internal/models"
	"formera/internal/pkg"
	"io"
	"net/http"
	"net/url"
	"time"
)

type CaptchaService struct {
	turnstileSecret string
	recaptchaSecret string
	hcaptchaSecret  string
	client          *http.Client
	// URLs are exposed for testing
	turnstileURL string
	recaptchaURL string
	hcaptchaURL  string
}

func NewCaptchaService(turnstileSecret, recaptchaSecret, hcaptchaSecret string) *CaptchaService {
	return &CaptchaService{
		turnstileSecret: turnstileSecret,
		recaptchaSecret: recaptchaSecret,
		hcaptchaSecret:  hcaptchaSecret,
		turnstileURL:    "https://challenges.cloudflare.com/turnstile/v0/siteverify",
		recaptchaURL:    "https://www.google.com/recaptcha/api/siteverify",
		hcaptchaURL:     "https://hcaptcha.com/siteverify",
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// VerifyToken verifies a CAPTCHA token based on the provider
func (s *CaptchaService) VerifyToken(provider models.CaptchaProvider, token string, clientIP string, minScore float64) (bool, error) {
	switch provider {
	case models.CaptchaProviderNone:
		return true, nil
	case models.CaptchaProviderTurnstile:
		return s.verifyTurnstile(token, clientIP)
	case models.CaptchaProviderRecaptcha:
		return s.verifyRecaptchaV3(token, clientIP, minScore)
	case models.CaptchaProviderHCaptcha:
		return s.verifyHCaptcha(token, clientIP)
	default:
		return false, errors.New("unknown CAPTCHA provider")
	}
}

func (s *CaptchaService) verifyTurnstile(token string, clientIP string) (bool, error) {
	if s.turnstileSecret == "" {
		return false, errors.New("Turnstile secret key not configured")
	}

	data := url.Values{
		"secret":   {s.turnstileSecret},
		"response": {token},
		"remoteip": {clientIP},
	}

	resp, err := s.client.PostForm(s.turnstileURL, data)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, err
	}

	return result.Success, nil
}

func (s *CaptchaService) verifyRecaptchaV3(token string, clientIP string, minScore float64) (bool, error) {
	if s.recaptchaSecret == "" {
		return false, errors.New("reCAPTCHA secret key not configured")
	}

	data := url.Values{
		"secret":   {s.recaptchaSecret},
		"response": {token},
		"remoteip": {clientIP},
	}

	resp, err := s.client.PostForm(s.recaptchaURL, data)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var result struct {
		Success bool    `json:"success"`
		Score   float64 `json:"score"`
		Action  string  `json:"action"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, err
	}

	if !result.Success {
		return false, nil
	}

	// Check score threshold (v3 specific)
	if result.Score < minScore {
		pkg.LogWarn().
			Float64("score", result.Score).
			Float64("min_score", minScore).
			Msg("reCAPTCHA score below threshold")
		return false, nil
	}

	return true, nil
}

func (s *CaptchaService) verifyHCaptcha(token string, clientIP string) (bool, error) {
	if s.hcaptchaSecret == "" {
		return false, errors.New("hCaptcha secret key not configured")
	}

	data := url.Values{
		"secret":   {s.hcaptchaSecret},
		"response": {token},
		"remoteip": {clientIP},
	}

	resp, err := s.client.PostForm(s.hcaptchaURL, data)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, err
	}

	return result.Success, nil
}
