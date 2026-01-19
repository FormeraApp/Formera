package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ShareToken errors
var (
	ErrTokenExpired  = errors.New("share token has expired")
	ErrTokenInvalid  = errors.New("invalid share token")
	ErrTokenMismatch = errors.New("token does not match path")
)

// DefaultShareTokenDuration is the default expiration time for share tokens
const DefaultShareTokenDuration = 1 * time.Hour

// ShareTokenService handles generation and validation of share tokens
type ShareTokenService struct {
	secret []byte
}

// NewShareTokenService creates a new share token service
func NewShareTokenService(jwtSecret string) *ShareTokenService {
	return &ShareTokenService{
		secret: []byte(jwtSecret),
	}
}

// GenerateShareToken creates a time-limited token for a file path
// Format: base64(expires:signature)
// Where signature = HMAC-SHA256(path + ":" + expires)
func (s *ShareTokenService) GenerateShareToken(filePath string, duration time.Duration) string {
	if duration == 0 {
		duration = DefaultShareTokenDuration
	}

	expires := time.Now().Add(duration).Unix()
	expiresStr := strconv.FormatInt(expires, 10)

	// Create signature: HMAC-SHA256(path:expires)
	message := filePath + ":" + expiresStr
	signature := s.sign(message)

	// Combine expires and signature
	token := fmt.Sprintf("%s:%s", expiresStr, signature)
	return base64.URLEncoding.EncodeToString([]byte(token))
}

// ValidateShareToken validates a token for a given file path
// Returns nil if valid, error otherwise
func (s *ShareTokenService) ValidateShareToken(filePath string, token string) error {
	// Decode base64
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return ErrTokenInvalid
	}

	// Split into expires:signature
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return ErrTokenInvalid
	}

	expiresStr := parts[0]
	providedSig := parts[1]

	// Check expiration
	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil {
		return ErrTokenInvalid
	}

	if time.Now().Unix() > expires {
		return ErrTokenExpired
	}

	// Verify signature
	message := filePath + ":" + expiresStr
	expectedSig := s.sign(message)

	if !hmac.Equal([]byte(providedSig), []byte(expectedSig)) {
		return ErrTokenMismatch
	}

	return nil
}

// sign creates an HMAC-SHA256 signature
func (s *ShareTokenService) sign(message string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// GetExpirationTime extracts expiration time from a token (for display purposes)
func (s *ShareTokenService) GetExpirationTime(token string) (time.Time, error) {
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return time.Time{}, ErrTokenInvalid
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return time.Time{}, ErrTokenInvalid
	}

	expires, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, ErrTokenInvalid
	}

	return time.Unix(expires, 0), nil
}
