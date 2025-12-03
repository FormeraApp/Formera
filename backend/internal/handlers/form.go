package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"formera/internal/database"
	"formera/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type FormHandler struct{}

func NewFormHandler() *FormHandler {
	return &FormHandler{}
}

type CreateFormRequest struct {
	Title       string              `json:"title" binding:"required"`
	Description string              `json:"description"`
	Fields      models.FormFields   `json:"fields"`
	Settings    models.FormSettings `json:"settings"`
}

type UpdateFormRequest struct {
	Title             string              `json:"title"`
	Description       string              `json:"description"`
	Slug              *string             `json:"slug"`
	Fields            models.FormFields   `json:"fields"`
	Settings          models.FormSettings `json:"settings"`
	Status            models.FormStatus   `json:"status"`
	PasswordProtected *bool               `json:"password_protected"`
	Password          string              `json:"password,omitempty"` // Raw password, will be hashed
}

type VerifyPasswordRequest struct {
	Password string `json:"password" binding:"required"`
}

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func isValidSlug(slug string) bool {
	if len(slug) < 3 || len(slug) > 100 {
		return false
	}
	return slugRegex.MatchString(slug)
}

func normalizeSlug(slug string) string {
	slug = strings.ToLower(strings.TrimSpace(slug))
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	return slug
}

func (h *FormHandler) Create(c *gin.Context) {
	userID := c.GetString("user_id")

	var req CreateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	form := &models.Form{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Fields:      req.Fields,
		Settings:    req.Settings,
		Status:      models.FormStatusDraft,
	}

	if result := database.DB.Create(form); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create form"})
		return
	}

	c.JSON(http.StatusCreated, form)
}

func (h *FormHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")

	var forms []models.Form
	if result := database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&forms); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch forms"})
		return
	}

	c.JSON(http.StatusOK, forms)
}

func (h *FormHandler) Get(c *gin.Context) {
	userID := c.GetString("user_id")
	formID := c.Param("id")

	var form models.Form
	if result := database.DB.Where("id = ? AND user_id = ?", formID, userID).First(&form); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	c.JSON(http.StatusOK, form)
}

func (h *FormHandler) GetPublic(c *gin.Context) {
	identifier := c.Param("id")

	var form models.Form
	result := database.DB.Where("(id = ? OR slug = ?) AND status = ?", identifier, identifier, models.FormStatusPublished).First(&form)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found or not published"})
		return
	}

	if form.PasswordProtected {
		c.JSON(http.StatusOK, gin.H{
			"id":                 form.ID,
			"title":              form.Title,
			"description":        form.Description,
			"slug":               form.Slug,
			"password_protected": true,
			"status":             form.Status,
		})
		return
	}

	c.JSON(http.StatusOK, form)
}

func (h *FormHandler) VerifyPassword(c *gin.Context) {
	identifier := c.Param("id")

	var form models.Form
	result := database.DB.Where("(id = ? OR slug = ?) AND status = ?", identifier, identifier, models.FormStatusPublished).First(&form)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found or not published"})
		return
	}

	if !form.PasswordProtected {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Form is not password protected"})
		return
	}

	var req VerifyPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password required"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(form.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusOK, gin.H{"valid": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"form":  form,
	})
}

func (h *FormHandler) CheckSlugAvailability(c *gin.Context) {
	userID := c.GetString("user_id")
	slug := c.Query("slug")
	formID := c.Query("form_id") // Exclude this form from check

	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug required"})
		return
	}

	slug = normalizeSlug(slug)
	if !isValidSlug(slug) {
		c.JSON(http.StatusOK, gin.H{
			"available": false,
			"slug":      slug,
			"reason":    "invalid",
		})
		return
	}

	var existingForm models.Form
	query := database.DB.Where("slug = ?", slug)
	if formID != "" {
		query = query.Where("id != ?", formID)
	}
	if result := query.First(&existingForm); result.Error == nil {
		c.JSON(http.StatusOK, gin.H{
			"available": false,
			"slug":      slug,
			"reason":    "taken",
		})
		return
	}

	_ = userID // Used for potential future per-user slug namespacing

	c.JSON(http.StatusOK, gin.H{
		"available": true,
		"slug":      slug,
	})
}

func (h *FormHandler) Update(c *gin.Context) {
	userID := c.GetString("user_id")
	formID := c.Param("id")

	var form models.Form
	if result := database.DB.Where("id = ? AND user_id = ?", formID, userID).First(&form); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	var req UpdateFormRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title != "" {
		form.Title = req.Title
	}
	if req.Description != "" {
		form.Description = req.Description
	}
	if req.Fields != nil {
		form.Fields = req.Fields
	}
	if req.Status != "" {
		form.Status = req.Status
	}
	form.Settings = req.Settings

	if req.Slug != nil {
		slug := *req.Slug
		if slug == "" {
			// When slug is cleared, use first 8 chars of ID (form accessible via ID)
			form.Slug = form.ID[:8]
		} else {
			slug = normalizeSlug(slug)
			if !isValidSlug(slug) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Ungültiger Slug. Verwende nur Kleinbuchstaben, Zahlen und Bindestriche (min. 3 Zeichen)."})
				return
			}
			var existingForm models.Form
			if result := database.DB.Where("slug = ? AND id != ?", slug, formID).First(&existingForm); result.Error == nil {
				c.JSON(http.StatusConflict, gin.H{"error": "Dieser Slug ist bereits vergeben."})
				return
			}
			form.Slug = slug
		}
	}

	if req.PasswordProtected != nil {
		form.PasswordProtected = *req.PasswordProtected
		if !*req.PasswordProtected {
			form.PasswordHash = ""
		}
	}

	if req.Password != "" && (req.PasswordProtected == nil || *req.PasswordProtected) {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		form.PasswordHash = string(hash)
		form.PasswordProtected = true
	}

	if result := database.DB.Save(&form); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update form"})
		return
	}

	c.JSON(http.StatusOK, form)
}

func (h *FormHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	formID := c.Param("id")

	var form models.Form
	if result := database.DB.Where("id = ? AND user_id = ?", formID, userID).First(&form); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	database.DB.Where("form_id = ?", formID).Delete(&models.Submission{})

	if result := database.DB.Delete(&form); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete form"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Form deleted successfully"})
}

func (h *FormHandler) Duplicate(c *gin.Context) {
	userID := c.GetString("user_id")
	formID := c.Param("id")

	var originalForm models.Form
	if result := database.DB.Where("id = ? AND user_id = ?", formID, userID).First(&originalForm); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	newForm := &models.Form{
		UserID:      userID,
		Title:       originalForm.Title + " (Kopie)",
		Description: originalForm.Description,
		Fields:      originalForm.Fields,
		Settings:    originalForm.Settings,
		Status:      models.FormStatusDraft,
	}

	if result := database.DB.Create(newForm); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to duplicate form"})
		return
	}

	c.JSON(http.StatusCreated, newForm)
}
