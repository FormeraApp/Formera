package handlers

import (
	"net/http"

	"formera/internal/database"
	"formera/internal/models"
	"formera/internal/services"

	"github.com/gin-gonic/gin"
)

type WebhookHandler struct {
	service *services.WebhookService
}

func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{
		service: services.NewWebhookService(database.DB),
	}
}

// GetWebhookService returns the webhook service for use by other handlers
func (h *WebhookHandler) GetWebhookService() *services.WebhookService {
	return h.service
}

type CreateWebhookRequest struct {
	URL           string                 `json:"url" binding:"required,url"`
	Secret        string                 `json:"secret,omitempty"`
	Events        []models.WebhookEvent  `json:"events" binding:"required,min=1"`
	Headers       map[string]string      `json:"headers,omitempty"`
	DiscordConfig *models.DiscordConfig  `json:"discord_config,omitempty"`
	Enabled       bool                   `json:"enabled"`
}

type UpdateWebhookRequest struct {
	URL           string                 `json:"url,omitempty"`
	Secret        string                 `json:"secret,omitempty"`
	Events        []models.WebhookEvent  `json:"events,omitempty"`
	Headers       map[string]string      `json:"headers,omitempty"`
	DiscordConfig *models.DiscordConfig  `json:"discord_config,omitempty"`
	Enabled       *bool                  `json:"enabled,omitempty"`
}

// ListGlobal godoc
// @Summary      List global webhooks
// @Description  Get all global webhooks (admin only)
// @Tags         Webhooks
// @Produce      json
// @Success      200 {array} models.Webhook
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /admin/webhooks [get]
func (h *WebhookHandler) ListGlobal(c *gin.Context) {
	var webhooks []models.Webhook
	if err := database.DB.Where("form_id IS NULL").Order("created_at DESC").Find(&webhooks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch webhooks"})
		return
	}
	c.JSON(http.StatusOK, webhooks)
}

// CreateGlobal godoc
// @Summary      Create global webhook
// @Description  Create a new global webhook (admin only)
// @Tags         Webhooks
// @Accept       json
// @Produce      json
// @Param        request body CreateWebhookRequest true "Webhook data"
// @Success      201 {object} models.Webhook
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /admin/webhooks [post]
func (h *WebhookHandler) CreateGlobal(c *gin.Context) {
	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate secret if not provided
	secret := req.Secret
	if secret == "" {
		var err error
		secret, err = services.GenerateSecret()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate secret"})
			return
		}
	}

	webhook := &models.Webhook{
		FormID:        nil, // Global webhook
		URL:           req.URL,
		Secret:        secret,
		Events:        req.Events,
		Headers:       req.Headers,
		DiscordConfig: req.DiscordConfig,
		Enabled:       req.Enabled,
	}

	if err := database.DB.Create(webhook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create webhook"})
		return
	}

	// Return the webhook with the secret (only on creation)
	c.JSON(http.StatusCreated, gin.H{
		"webhook": webhook,
		"secret":  secret, // Include secret only on creation
	})
}

// GetGlobal godoc
// @Summary      Get global webhook
// @Description  Get a specific global webhook (admin only)
// @Tags         Webhooks
// @Produce      json
// @Param        id path string true "Webhook ID"
// @Success      200 {object} models.Webhook
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /admin/webhooks/{id} [get]
func (h *WebhookHandler) GetGlobal(c *gin.Context) {
	webhookID := c.Param("id")

	var webhook models.Webhook
	if err := database.DB.Where("id = ? AND form_id IS NULL", webhookID).First(&webhook).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	c.JSON(http.StatusOK, webhook)
}

// UpdateGlobal godoc
// @Summary      Update global webhook
// @Description  Update a global webhook (admin only)
// @Tags         Webhooks
// @Accept       json
// @Produce      json
// @Param        id path string true "Webhook ID"
// @Param        request body UpdateWebhookRequest true "Webhook data"
// @Success      200 {object} models.Webhook
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /admin/webhooks/{id} [put]
func (h *WebhookHandler) UpdateGlobal(c *gin.Context) {
	webhookID := c.Param("id")

	var webhook models.Webhook
	if err := database.DB.Where("id = ? AND form_id IS NULL", webhookID).First(&webhook).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	var req UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	if req.URL != "" {
		webhook.URL = req.URL
	}
	if req.Secret != "" {
		webhook.Secret = req.Secret
	}
	if len(req.Events) > 0 {
		webhook.Events = req.Events
	}
	if req.Headers != nil {
		webhook.Headers = req.Headers
	}
	if req.DiscordConfig != nil {
		webhook.DiscordConfig = req.DiscordConfig
	}
	if req.Enabled != nil {
		webhook.Enabled = *req.Enabled
	}

	if err := database.DB.Save(&webhook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update webhook"})
		return
	}

	c.JSON(http.StatusOK, webhook)
}

// DeleteGlobal godoc
// @Summary      Delete global webhook
// @Description  Delete a global webhook (admin only)
// @Tags         Webhooks
// @Produce      json
// @Param        id path string true "Webhook ID"
// @Success      200 {object} MessageResponse
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /admin/webhooks/{id} [delete]
func (h *WebhookHandler) DeleteGlobal(c *gin.Context) {
	webhookID := c.Param("id")

	var webhook models.Webhook
	if err := database.DB.Where("id = ? AND form_id IS NULL", webhookID).First(&webhook).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	// Delete associated logs
	database.DB.Where("webhook_id = ?", webhookID).Delete(&models.WebhookLog{})

	if err := database.DB.Delete(&webhook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook deleted successfully"})
}

// TestGlobal godoc
// @Summary      Test global webhook
// @Description  Send a test payload to a global webhook (admin only)
// @Tags         Webhooks
// @Produce      json
// @Param        id path string true "Webhook ID"
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} ErrorResponse
// @Failure      403 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /admin/webhooks/{id}/test [post]
func (h *WebhookHandler) TestGlobal(c *gin.Context) {
	webhookID := c.Param("id")

	var webhook models.Webhook
	if err := database.DB.Where("id = ? AND form_id IS NULL", webhookID).First(&webhook).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	success, statusCode, responseBody, err := h.service.SendTestWebhook(&webhook)

	result := gin.H{
		"success":       success,
		"status_code":   statusCode,
		"response_body": responseBody,
	}
	if err != nil {
		result["error"] = err.Error()
	}

	c.JSON(http.StatusOK, result)
}

// ListForForm godoc
// @Summary      List form webhooks
// @Description  Get all webhooks for a specific form
// @Tags         Webhooks
// @Produce      json
// @Param        id path string true "Form ID"
// @Success      200 {array} models.Webhook
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /forms/{id}/webhooks [get]
func (h *WebhookHandler) ListForForm(c *gin.Context) {
	userID := c.GetString("user_id")
	formID := c.Param("id")

	// Verify form ownership
	var form models.Form
	if err := database.DB.Where("id = ? AND user_id = ?", formID, userID).First(&form).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	var webhooks []models.Webhook
	if err := database.DB.Where("form_id = ?", formID).Order("created_at DESC").Find(&webhooks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch webhooks"})
		return
	}

	c.JSON(http.StatusOK, webhooks)
}

// CreateForForm godoc
// @Summary      Create form webhook
// @Description  Create a new webhook for a specific form
// @Tags         Webhooks
// @Accept       json
// @Produce      json
// @Param        id path string true "Form ID"
// @Param        request body CreateWebhookRequest true "Webhook data"
// @Success      201 {object} models.Webhook
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /forms/{id}/webhooks [post]
func (h *WebhookHandler) CreateForForm(c *gin.Context) {
	userID := c.GetString("user_id")
	formID := c.Param("id")

	// Verify form ownership
	var form models.Form
	if err := database.DB.Where("id = ? AND user_id = ?", formID, userID).First(&form).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate secret if not provided
	secret := req.Secret
	if secret == "" {
		var err error
		secret, err = services.GenerateSecret()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate secret"})
			return
		}
	}

	webhook := &models.Webhook{
		FormID:        &formID,
		URL:           req.URL,
		Secret:        secret,
		Events:        req.Events,
		Headers:       req.Headers,
		DiscordConfig: req.DiscordConfig,
		Enabled:       req.Enabled,
	}

	if err := database.DB.Create(webhook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create webhook"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"webhook": webhook,
		"secret":  secret,
	})
}

// GetForForm godoc
// @Summary      Get form webhook
// @Description  Get a specific webhook for a form
// @Tags         Webhooks
// @Produce      json
// @Param        id path string true "Form ID"
// @Param        webhookId path string true "Webhook ID"
// @Success      200 {object} models.Webhook
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /forms/{id}/webhooks/{webhookId} [get]
func (h *WebhookHandler) GetForForm(c *gin.Context) {
	userID := c.GetString("user_id")
	formID := c.Param("id")
	webhookID := c.Param("webhookId")

	// Verify form ownership
	var form models.Form
	if err := database.DB.Where("id = ? AND user_id = ?", formID, userID).First(&form).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	var webhook models.Webhook
	if err := database.DB.Where("id = ? AND form_id = ?", webhookID, formID).First(&webhook).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	c.JSON(http.StatusOK, webhook)
}

// UpdateForForm godoc
// @Summary      Update form webhook
// @Description  Update a webhook for a specific form
// @Tags         Webhooks
// @Accept       json
// @Produce      json
// @Param        id path string true "Form ID"
// @Param        webhookId path string true "Webhook ID"
// @Param        request body UpdateWebhookRequest true "Webhook data"
// @Success      200 {object} models.Webhook
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /forms/{id}/webhooks/{webhookId} [put]
func (h *WebhookHandler) UpdateForForm(c *gin.Context) {
	userID := c.GetString("user_id")
	formID := c.Param("id")
	webhookID := c.Param("webhookId")

	// Verify form ownership
	var form models.Form
	if err := database.DB.Where("id = ? AND user_id = ?", formID, userID).First(&form).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	var webhook models.Webhook
	if err := database.DB.Where("id = ? AND form_id = ?", webhookID, formID).First(&webhook).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	var req UpdateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	if req.URL != "" {
		webhook.URL = req.URL
	}
	if req.Secret != "" {
		webhook.Secret = req.Secret
	}
	if len(req.Events) > 0 {
		webhook.Events = req.Events
	}
	if req.Headers != nil {
		webhook.Headers = req.Headers
	}
	if req.DiscordConfig != nil {
		webhook.DiscordConfig = req.DiscordConfig
	}
	if req.Enabled != nil {
		webhook.Enabled = *req.Enabled
	}

	if err := database.DB.Save(&webhook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update webhook"})
		return
	}

	c.JSON(http.StatusOK, webhook)
}

// DeleteForForm godoc
// @Summary      Delete form webhook
// @Description  Delete a webhook from a specific form
// @Tags         Webhooks
// @Produce      json
// @Param        id path string true "Form ID"
// @Param        webhookId path string true "Webhook ID"
// @Success      200 {object} MessageResponse
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /forms/{id}/webhooks/{webhookId} [delete]
func (h *WebhookHandler) DeleteForForm(c *gin.Context) {
	userID := c.GetString("user_id")
	formID := c.Param("id")
	webhookID := c.Param("webhookId")

	// Verify form ownership
	var form models.Form
	if err := database.DB.Where("id = ? AND user_id = ?", formID, userID).First(&form).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	var webhook models.Webhook
	if err := database.DB.Where("id = ? AND form_id = ?", webhookID, formID).First(&webhook).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	// Delete associated logs
	database.DB.Where("webhook_id = ?", webhookID).Delete(&models.WebhookLog{})

	if err := database.DB.Delete(&webhook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook deleted successfully"})
}

// TestForForm godoc
// @Summary      Test form webhook
// @Description  Send a test payload to a form webhook
// @Tags         Webhooks
// @Produce      json
// @Param        id path string true "Form ID"
// @Param        webhookId path string true "Webhook ID"
// @Success      200 {object} map[string]interface{}
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /forms/{id}/webhooks/{webhookId}/test [post]
func (h *WebhookHandler) TestForForm(c *gin.Context) {
	userID := c.GetString("user_id")
	formID := c.Param("id")
	webhookID := c.Param("webhookId")

	// Verify form ownership
	var form models.Form
	if err := database.DB.Where("id = ? AND user_id = ?", formID, userID).First(&form).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	var webhook models.Webhook
	if err := database.DB.Where("id = ? AND form_id = ?", webhookID, formID).First(&webhook).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	success, statusCode, responseBody, err := h.service.SendTestWebhook(&webhook)

	result := gin.H{
		"success":       success,
		"status_code":   statusCode,
		"response_body": responseBody,
	}
	if err != nil {
		result["error"] = err.Error()
	}

	c.JSON(http.StatusOK, result)
}

// GetLogs godoc
// @Summary      Get webhook logs
// @Description  Get recent delivery logs for a webhook
// @Tags         Webhooks
// @Produce      json
// @Param        id path string true "Form ID"
// @Param        webhookId path string true "Webhook ID"
// @Success      200 {array} models.WebhookLog
// @Failure      401 {object} ErrorResponse
// @Failure      404 {object} ErrorResponse
// @Security     BearerAuth
// @Router       /forms/{id}/webhooks/{webhookId}/logs [get]
func (h *WebhookHandler) GetLogs(c *gin.Context) {
	userID := c.GetString("user_id")
	formID := c.Param("id")
	webhookID := c.Param("webhookId")

	// Verify form ownership
	var form models.Form
	if err := database.DB.Where("id = ? AND user_id = ?", formID, userID).First(&form).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	var webhook models.Webhook
	if err := database.DB.Where("id = ? AND form_id = ?", webhookID, formID).First(&webhook).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	logs, err := h.service.GetWebhookLogs(webhookID, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}

	c.JSON(http.StatusOK, logs)
}
