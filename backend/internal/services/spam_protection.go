package services

import (
	"errors"
	"formera/internal/models"
	"formera/internal/pkg"
)

type SpamProtectionService struct {
	captchaService *CaptchaService
}

func NewSpamProtectionService(captchaService *CaptchaService) *SpamProtectionService {
	return &SpamProtectionService{
		captchaService: captchaService,
	}
}

// CheckSubmission performs all spam checks
func (s *SpamProtectionService) CheckSubmission(
	config models.SpamProtectionConfig,
	honeypotValue string,
	captchaToken string,
	clientIP string,
) error {
	// 1. Honeypot check (always active)
	if err := s.checkHoneypot(honeypotValue); err != nil {
		return err
	}

	// 2. CAPTCHA check (if enabled)
	if config.CaptchaProvider != models.CaptchaProviderNone {
		if err := s.checkCaptcha(config, captchaToken, clientIP); err != nil {
			return err
		}
	}

	return nil
}

func (s *SpamProtectionService) checkHoneypot(value string) error {
	if value != "" {
		pkg.LogWarn().Msg("Honeypot triggered - likely spam")
		return errors.New("spam detected")
	}
	return nil
}

func (s *SpamProtectionService) checkCaptcha(
	config models.SpamProtectionConfig,
	token string,
	clientIP string,
) error {
	if token == "" {
		pkg.LogWarn().Msg("CAPTCHA token missing")
		return errors.New("CAPTCHA verification required")
	}

	minScore := config.RecaptchaMinScore
	if minScore == 0 {
		minScore = 0.5 // default
	}

	valid, err := s.captchaService.VerifyToken(
		config.CaptchaProvider,
		token,
		clientIP,
		minScore,
	)

	if err != nil {
		pkg.LogError().Err(err).Msg("CAPTCHA verification failed")
		// Graceful degradation: if CAPTCHA service is down, allow submission
		// but log it for monitoring
		pkg.LogWarn().Msg("CAPTCHA service unavailable - allowing submission")
		return nil
	}

	if !valid {
		pkg.LogWarn().
			Str("provider", string(config.CaptchaProvider)).
			Str("ip", clientIP).
			Msg("CAPTCHA verification failed")
		return errors.New("CAPTCHA verification failed")
	}

	return nil
}
