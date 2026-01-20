package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"formera/internal/models"
	"formera/internal/pkg"

	"gorm.io/gorm"
)

// DiscordEmbed represents a Discord embed object
type DiscordEmbed struct {
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	Color       int            `json:"color,omitempty"`
	Fields      []DiscordField `json:"fields,omitempty"`
	Timestamp   string         `json:"timestamp,omitempty"`
	Footer      *DiscordFooter `json:"footer,omitempty"`
}

// DiscordField represents a field in a Discord embed
type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// DiscordFooter represents footer in a Discord embed
type DiscordFooter struct {
	Text string `json:"text"`
}

// DiscordWebhookPayload represents the Discord webhook format
type DiscordWebhookPayload struct {
	Content string         `json:"content,omitempty"`
	Embeds  []DiscordEmbed `json:"embeds,omitempty"`
}

// WebhookService handles webhook delivery
type WebhookService struct {
	db     *gorm.DB
	client *http.Client
}

// NewWebhookService creates a new webhook service
func NewWebhookService(db *gorm.DB) *WebhookService {
	return &WebhookService{
		db: db,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// getLanguage retrieves the language setting from the database
func (s *WebhookService) getLanguage() string {
	var settings models.Settings
	if err := s.db.First(&settings).Error; err != nil {
		return "en" // Default to English
	}
	if settings.Language == "" {
		return "en"
	}
	return settings.Language
}

// GenerateSecret generates a random secret for webhook signing
func GenerateSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// TriggerEvent sends webhooks for a specific event
// This should be called asynchronously (in a goroutine)
func (s *WebhookService) TriggerEvent(event models.WebhookEvent, form *models.Form, data map[string]interface{}) {
	// Get all webhooks for this event (both global and form-specific)
	var webhooks []models.Webhook

	// Query: (form_id IS NULL OR form_id = ?) AND enabled = true
	err := s.db.Where("enabled = ? AND (form_id IS NULL OR form_id = ?)", true, form.ID).Find(&webhooks).Error
	if err != nil {
		pkg.LogError().Err(err).Str("form_id", form.ID).Msg("Failed to fetch webhooks")
		return
	}

	// Filter webhooks by event
	for _, webhook := range webhooks {
		if webhook.Events.Contains(event) {
			go s.deliverWebhook(&webhook, event, form, data)
		}
	}
}

// deliverWebhook sends a webhook with retry logic
func (s *WebhookService) deliverWebhook(webhook *models.Webhook, event models.WebhookEvent, form *models.Form, data map[string]interface{}) {
	payload := models.WebhookPayload{
		Event:     event,
		Timestamp: time.Now().UTC(),
		WebhookID: webhook.ID,
		Form: models.WebhookFormData{
			ID:    form.ID,
			Title: form.Title,
			Slug:  form.Slug,
		},
		Data: data,
	}

	var jsonBody []byte
	var err error

	// Transform payload for Discord webhooks
	if isDiscordWebhook(webhook.URL) {
		lang := s.getLanguage()
		jsonBody, err = transformToDiscordPayload(payload, webhook.DiscordConfig, lang)
	} else {
		jsonBody, err = json.Marshal(payload)
	}

	if err != nil {
		pkg.LogError().Err(err).Str("webhook_id", webhook.ID).Msg("Failed to marshal webhook payload")
		return
	}

	// Retry logic: 3 attempts with exponential backoff (1s, 5s, 30s)
	retryDelays := []time.Duration{0, 1 * time.Second, 5 * time.Second, 30 * time.Second}

	for attempt := 1; attempt <= 3; attempt++ {
		if attempt > 1 {
			time.Sleep(retryDelays[attempt])
		}

		success, statusCode, responseBody, sendErr := s.sendRequest(webhook, jsonBody)

		// Log the attempt
		logEntry := &models.WebhookLog{
			WebhookID:    webhook.ID,
			Event:        string(event),
			RequestBody:  string(jsonBody),
			ResponseCode: statusCode,
			ResponseBody: responseBody,
			Success:      success,
			Attempt:      attempt,
		}
		if sendErr != nil {
			logEntry.Error = sendErr.Error()
		}

		if err := s.db.Create(logEntry).Error; err != nil {
			pkg.LogError().Err(err).Str("webhook_id", webhook.ID).Msg("Failed to save webhook log")
		}

		if success {
			pkg.LogInfo().
				Str("webhook_id", webhook.ID).
				Str("event", string(event)).
				Int("attempt", attempt).
				Msg("Webhook delivered successfully")
			return
		}

		pkg.LogWarn().
			Str("webhook_id", webhook.ID).
			Str("event", string(event)).
			Int("attempt", attempt).
			Int("status_code", statusCode).
			Str("error", fmt.Sprintf("%v", sendErr)).
			Msg("Webhook delivery failed")
	}

	pkg.LogError().
		Str("webhook_id", webhook.ID).
		Str("event", string(event)).
		Msg("Webhook delivery failed after all retries")
}

// isDiscordWebhook checks if the URL is a Discord webhook
func isDiscordWebhook(url string) bool {
	return strings.Contains(url, "discord.com/api/webhooks/") ||
		strings.Contains(url, "discordapp.com/api/webhooks/")
}

// transformToDiscordPayload converts our webhook payload to Discord format
func transformToDiscordPayload(payload models.WebhookPayload, config *models.DiscordConfig, lang string) ([]byte, error) {
	// Debug log to track Discord config
	if config != nil {
		fmt.Printf("[DEBUG] Discord config received: show_form_title=%v, show_slug=%v, fields=%v, custom_title=%s\n",
			config.ShowFormTitle, config.ShowSlug, config.Fields, config.CustomTitle)
	} else {
		fmt.Println("[DEBUG] Discord config is nil")
	}

	// Color based on event type (Discord uses decimal color values)
	color := 5814783 // Default blue (#58b9bf)
	eventTitle := "Webhook Event"

	switch payload.Event {
	case models.WebhookEventSubmissionCreated:
		color = 5763719 // Green (#57F287)
		eventTitle = pkg.T(lang, "webhook.event.submission.created")
	case models.WebhookEventSubmissionDeleted:
		color = 15548997 // Red (#ED4245)
		eventTitle = pkg.T(lang, "webhook.event.submission.deleted")
	case models.WebhookEvent("test"):
		color = 5793266 // Blurple (#5865F2)
		eventTitle = pkg.T(lang, "webhook.event.test")
	}

	// Use custom title if configured
	if config != nil && config.CustomTitle != "" {
		eventTitle = config.CustomTitle
	}

	// Build fields from submission data
	var fields []DiscordField

	// Add form info based on config (default: show both)
	showFormTitle := config == nil || config.ShowFormTitle
	showSlug := config == nil || config.ShowSlug

	if showFormTitle {
		fields = append(fields, DiscordField{
			Name:   pkg.T(lang, "webhook.field.form"),
			Value:  payload.Form.Title,
			Inline: true,
		})
	}

	if showSlug && payload.Form.Slug != "" {
		fields = append(fields, DiscordField{
			Name:   pkg.T(lang, "webhook.field.slug"),
			Value:  payload.Form.Slug,
			Inline: true,
		})
	}

	// Get configured fields filter (empty = show all)
	var allowedFields map[string]bool
	if config != nil && len(config.Fields) > 0 {
		allowedFields = make(map[string]bool)
		for _, f := range config.Fields {
			allowedFields[f] = true
		}
		fmt.Printf("[DEBUG] Field filtering enabled: allowed_fields=%v\n", allowedFields)
	} else {
		fmt.Println("[DEBUG] Field filtering disabled - showing all fields")
	}

	// Get field labels map (ID -> Label) if available
	fieldLabels := make(map[string]string)
	if payload.Data != nil {
		if labels, ok := payload.Data["field_labels"].(map[string]interface{}); ok {
			for k, v := range labels {
				if label, ok := v.(string); ok {
					fieldLabels[k] = label
				}
			}
		}
	}

	// Add submission data fields (limit to first 10 to avoid Discord limits)
	fieldCount := 0
	fmt.Printf("[DEBUG] payload.Data is nil: %v\n", payload.Data == nil)
	if payload.Data != nil {
		fmt.Printf("[DEBUG] payload.Data keys: %v\n", getMapKeys(payload.Data))
		// Check if there's a submission object with data
		submission, ok := payload.Data["submission"].(map[string]interface{})
		fmt.Printf("[DEBUG] submission cast ok: %v\n", ok)
		if ok {
			fmt.Printf("[DEBUG] submission keys: %v\n", getMapKeys(submission))
			// Use reflection to handle any map type (models.SubmissionData or map[string]interface{})
			var submissionData map[string]interface{}
			var ok2 bool
			if rawData := submission["data"]; rawData != nil {
				rv := reflect.ValueOf(rawData)
				fmt.Printf("[DEBUG] rawData reflect kind: %v, type: %T\n", rv.Kind(), rawData)
				if rv.Kind() == reflect.Map {
					submissionData = make(map[string]interface{})
					for _, key := range rv.MapKeys() {
						submissionData[key.String()] = rv.MapIndex(key).Interface()
					}
					ok2 = true
				}
			}
			fmt.Printf("[DEBUG] submissionData cast ok: %v, type: %T\n", ok2, submission["data"])
			if ok2 {
				fmt.Printf("[DEBUG] Processing submission data keys: %v\n", getMapKeys(submissionData))
				for key, value := range submissionData {
					if fieldCount >= 10 {
						break
					}
					// Skip if field filtering is enabled and this field is not in the list
					if allowedFields != nil && !allowedFields[key] {
						fmt.Printf("[DEBUG] Skipping field '%s' - not in allowed list\n", key)
						continue
					}
					valueStr := formatFieldValue(value)
					if len(valueStr) > 200 {
						valueStr = valueStr[:197] + "..."
					}
					// Use field label if available, otherwise use field ID
					fieldName := key
					if label, ok := fieldLabels[key]; ok && label != "" {
						fieldName = label
					}
					fields = append(fields, DiscordField{
						Name:   fieldName,
						Value:  valueStr,
						Inline: len(valueStr) < 50,
					})
					fieldCount++
				}
			}
		}
	}

	discordPayload := DiscordWebhookPayload{
		Embeds: []DiscordEmbed{
			{
				Title:     eventTitle,
				Color:     color,
				Fields:    fields,
				Timestamp: payload.Timestamp.Format(time.RFC3339),
				Footer: &DiscordFooter{
					Text: pkg.T(lang, "webhook.footer"),
				},
			},
		},
	}

	return json.Marshal(discordPayload)
}

// sendRequest sends the actual HTTP request
func (s *WebhookService) sendRequest(webhook *models.Webhook, body []byte) (success bool, statusCode int, responseBody string, err error) {
	req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(body))
	if err != nil {
		return false, 0, "", err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Formera-Webhook/1.0")

	// Add HMAC signature if secret is set (not for Discord webhooks)
	if webhook.Secret != "" && !isDiscordWebhook(webhook.URL) {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		signature := s.computeSignature(webhook.Secret, timestamp, body)

		req.Header.Set("X-Webhook-Timestamp", timestamp)
		req.Header.Set("X-Webhook-Signature", "sha256="+signature)
	}

	// Add custom headers (not for Discord webhooks as they don't need them)
	if !isDiscordWebhook(webhook.URL) {
		for key, value := range webhook.Headers {
			req.Header.Set(key, value)
		}
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return false, 0, "", err
	}
	defer resp.Body.Close()

	// Read response body (limited to 1KB for logging)
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	responseBody = string(respBody)

	// Success if status code is 2xx
	success = resp.StatusCode >= 200 && resp.StatusCode < 300
	return success, resp.StatusCode, responseBody, nil
}

// computeSignature computes HMAC-SHA256 signature
// Format: HMAC-SHA256(timestamp + "." + body)
func (s *WebhookService) computeSignature(secret, timestamp string, body []byte) string {
	message := timestamp + "." + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// SendTestWebhook sends a test payload to a webhook
func (s *WebhookService) SendTestWebhook(webhook *models.Webhook) (success bool, statusCode int, responseBody string, err error) {
	lang := s.getLanguage()

	// Build test submission data - use configured field IDs if available
	submissionData := make(map[string]interface{})
	fieldLabels := make(map[string]interface{})

	testValues := []string{
		"Max Mustermann",
		"test@example.com",
		"This is a test message from Formera",
	}
	testLabels := []string{
		pkg.T(lang, "webhook.test.field.name"),
		pkg.T(lang, "webhook.test.field.email"),
		pkg.T(lang, "webhook.test.field.message"),
	}

	// If Discord config has specific fields configured, use those IDs
	if webhook.DiscordConfig != nil && len(webhook.DiscordConfig.Fields) > 0 {
		for i, fieldID := range webhook.DiscordConfig.Fields {
			if i < len(testValues) {
				submissionData[fieldID] = testValues[i]
				fieldLabels[fieldID] = testLabels[i]
			}
		}
	} else {
		// Default test fields
		submissionData["field_name"] = testValues[0]
		submissionData["field_email"] = testValues[1]
		submissionData["field_msg"] = testValues[2]
		fieldLabels["field_name"] = testLabels[0]
		fieldLabels["field_email"] = testLabels[1]
		fieldLabels["field_msg"] = testLabels[2]
	}

	testPayload := models.WebhookPayload{
		Event:     models.WebhookEvent("test"),
		Timestamp: time.Now().UTC(),
		WebhookID: webhook.ID,
		Form: models.WebhookFormData{
			ID:    "test-form-id",
			Title: "Test Form",
			Slug:  "test-form",
		},
		Data: map[string]interface{}{
			"submission": map[string]interface{}{
				"id":         "test-submission-id",
				"created_at": time.Now().UTC(),
				"data":       submissionData,
			},
			"field_labels": fieldLabels,
		},
	}

	var jsonBody []byte

	// Transform payload for Discord webhooks
	if isDiscordWebhook(webhook.URL) {
		jsonBody, err = transformToDiscordPayload(testPayload, webhook.DiscordConfig, lang)
	} else {
		jsonBody, err = json.Marshal(testPayload)
	}

	if err != nil {
		return false, 0, "", err
	}

	return s.sendRequest(webhook, jsonBody)
}

// GetWebhookLogs returns recent logs for a webhook
func (s *WebhookService) GetWebhookLogs(webhookID string, limit int) ([]models.WebhookLog, error) {
	var logs []models.WebhookLog
	err := s.db.Where("webhook_id = ?", webhookID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// CleanupOldLogs removes webhook logs older than the specified duration
func (s *WebhookService) CleanupOldLogs(maxAge time.Duration) error {
	cutoff := time.Now().Add(-maxAge)
	return s.db.Where("created_at < ?", cutoff).Delete(&models.WebhookLog{}).Error
}

// getMapKeys returns all keys from a map for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// formatFieldValue formats a field value for Discord display
// Handles file paths, arrays, ratings, signatures, rich text, and regular values
func formatFieldValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return formatSingleValue(v)
	case float64:
		// Handle numbers (JSON unmarshals numbers as float64)
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%.2f", v)
	case int, int64, int32:
		return fmt.Sprintf("%d", v)
	case bool:
		if v {
			return "Ja"
		}
		return "Nein"
	case []interface{}:
		// Handle arrays (e.g., multiple file uploads, multi-select, checkboxes)
		if len(v) == 0 {
			return ""
		}
		var parts []string
		for _, item := range v {
			if str, ok := item.(string); ok {
				parts = append(parts, formatSingleValue(str))
			} else {
				parts = append(parts, fmt.Sprintf("%v", item))
			}
		}
		return strings.Join(parts, ", ")
	case map[string]interface{}:
		// Handle complex objects (might be rating or other structured data)
		if rating, ok := v["value"].(float64); ok {
			return formatRating(int(rating))
		}
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", value)
	}
}

// formatSingleValue formats a single string value
// Handles file paths, signatures, rich text, and data URLs
func formatSingleValue(s string) string {
	// Check if it's a signature (data URL)
	if strings.HasPrefix(s, "data:image/") {
		return "[Unterschrift]"
	}

	// Check if it's a file path
	if strings.HasPrefix(s, "files/") || strings.HasPrefix(s, "images/") {
		// Extract just the filename (remove UUID prefix if present)
		parts := strings.Split(s, "/")
		if len(parts) > 0 {
			filename := parts[len(parts)-1]
			// Remove UUID prefix (format: uuid_originalname.ext)
			if idx := strings.Index(filename, "_"); idx > 0 && idx < 40 {
				filename = filename[idx+1:]
			}
			return filename
		}
	}

	// Check if it contains HTML (Rich Text)
	if strings.Contains(s, "<") && strings.Contains(s, ">") {
		return stripHTMLTags(s)
	}

	return s
}

// formatRating converts a rating number to stars
func formatRating(rating int) string {
	if rating < 1 {
		rating = 1
	}
	if rating > 5 {
		rating = 5
	}
	filled := strings.Repeat("★", rating)
	empty := strings.Repeat("☆", 5-rating)
	return filled + empty
}

// stripHTMLTags removes HTML tags from a string
func stripHTMLTags(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			// Add space after block elements
			result.WriteRune(' ')
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}
	// Clean up multiple spaces
	cleaned := strings.Join(strings.Fields(result.String()), " ")
	return strings.TrimSpace(cleaned)
}
