package handlers

import (
	"net/http"
	"strings"

	"formera/internal/database"
	"formera/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TemplateHandler struct{}

func NewTemplateHandler() *TemplateHandler {
	return &TemplateHandler{}
}

// TemplateCategoryInfo represents category metadata
type TemplateCategoryInfo struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
	Label    string `json:"label"`
	Icon     string `json:"icon"`
}

// List returns all available templates
// GET /api/templates?category=contact&search=feedback&language=de
func (h *TemplateHandler) List(c *gin.Context) {
	category := c.Query("category")
	search := c.Query("search")
	language := c.Query("language")

	// Default to German if no language specified
	if language == "" {
		language = "de"
	}

	db := database.DB.Where("is_template = ? AND language = ?", true, language)

	// Filter by category if provided
	if category != "" && category != "all" {
		db = db.Where("template_category = ?", category)
	}

	// Filter by search query (title or description)
	if search != "" {
		searchPattern := "%" + strings.ToLower(search) + "%"
		db = db.Where("LOWER(title) LIKE ? OR LOWER(description) LIKE ?", searchPattern, searchPattern)
	}

	var templates []models.Form
	if err := db.Order("template_category, title").Find(&templates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch templates"})
		return
	}

	c.JSON(http.StatusOK, templates)
}

// Categories returns all template categories with counts
// GET /api/templates/categories?language=de
func (h *TemplateHandler) Categories(c *gin.Context) {
	language := c.Query("language")

	// Default to German if no language specified
	if language == "" {
		language = "de"
	}

	categories := []models.TemplateCategory{
		models.TemplateCategoryContact,
		models.TemplateCategoryFeedback,
		models.TemplateCategorySurvey,
		models.TemplateCategoryRegistration,
		models.TemplateCategoryApplication,
		models.TemplateCategoryNewsletter,
		models.TemplateCategoryQuote,
		models.TemplateCategorySupport,
		models.TemplateCategoryBooking,
		models.TemplateCategoryOrder,
		models.TemplateCategoryComplaint,
		models.TemplateCategorySuggestion,
	}

	categoryLabels := map[models.TemplateCategory]string{
		models.TemplateCategoryContact:      "Kontakt",
		models.TemplateCategoryFeedback:     "Feedback",
		models.TemplateCategorySurvey:       "Umfrage",
		models.TemplateCategoryRegistration: "Registrierung",
		models.TemplateCategoryApplication:  "Bewerbung",
		models.TemplateCategoryNewsletter:   "Newsletter",
		models.TemplateCategoryQuote:        "Angebot",
		models.TemplateCategorySupport:      "Support",
		models.TemplateCategoryBooking:      "Buchung",
		models.TemplateCategoryOrder:        "Bestellung",
		models.TemplateCategoryComplaint:    "Beschwerde",
		models.TemplateCategorySuggestion:   "Vorschlag",
	}

	categoryIcons := map[models.TemplateCategory]string{
		models.TemplateCategoryContact:      "envelope",
		models.TemplateCategoryFeedback:     "comment",
		models.TemplateCategorySurvey:       "chart-bar",
		models.TemplateCategoryRegistration: "calendar",
		models.TemplateCategoryApplication:  "briefcase",
		models.TemplateCategoryNewsletter:   "newspaper",
		models.TemplateCategoryQuote:        "file-invoice",
		models.TemplateCategorySupport:      "headset",
		models.TemplateCategoryBooking:      "calendar-check",
		models.TemplateCategoryOrder:        "shopping-cart",
		models.TemplateCategoryComplaint:    "exclamation-triangle",
		models.TemplateCategorySuggestion:   "lightbulb",
	}

	var result []TemplateCategoryInfo

	for _, cat := range categories {
		var count int64
		database.DB.Model(&models.Form{}).
			Where("is_template = ? AND template_category = ? AND language = ?", true, string(cat), language).
			Count(&count)

		result = append(result, TemplateCategoryInfo{
			Category: string(cat),
			Count:    count,
			Label:    categoryLabels[cat],
			Icon:     categoryIcons[cat],
		})
	}

	c.JSON(http.StatusOK, result)
}

// Get returns a single template by ID
// GET /api/templates/:id
func (h *TemplateHandler) Get(c *gin.Context) {
	templateID := c.Param("id")

	var template models.Form
	if err := database.DB.Where("id = ? AND is_template = ?", templateID, true).First(&template).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// UseTemplateRequest represents the request body for using a template
type UseTemplateRequest struct {
	Title string `json:"title,omitempty"`
}

// UseTemplate creates a new form from a template
// POST /api/templates/:id/use
func (h *TemplateHandler) UseTemplate(c *gin.Context) {
	templateID := c.Param("id")
	userID, _ := c.Get("user_id")

	// Get the template
	var template models.Form
	if err := database.DB.Where("id = ? AND is_template = ?", templateID, true).First(&template).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	// Parse optional title override
	var req UseTemplateRequest
	_ = c.ShouldBindJSON(&req)

	// Create new form from template
	newForm := models.Form{
		ID:                uuid.New().String(),
		UserID:            userID.(string),
		Title:             template.Title,
		Description:       template.Description,
		Fields:            template.Fields,
		Settings:          template.Settings,
		Status:            models.FormStatusDraft,
		IsTemplate:        false, // Important: not a template
		TemplateCategory:  "",    // Clear template category
		PasswordProtected: false,
		ViewCount:         0,
	}

	// Override title if provided
	if req.Title != "" {
		newForm.Title = req.Title
	}

	// Create the new form
	if err := database.DB.Create(&newForm).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create form from template"})
		return
	}

	c.JSON(http.StatusCreated, newForm)
}

// Create creates a new template (admin only)
// POST /api/admin/templates
func (h *TemplateHandler) Create(c *gin.Context) {
	var template models.Form
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Force template fields
	template.ID = uuid.New().String()
	template.UserID = "system"
	template.IsTemplate = true
	template.Status = models.FormStatusDraft

	if err := database.DB.Create(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// Update updates an existing template (admin only)
// PUT /api/admin/templates/:id
func (h *TemplateHandler) Update(c *gin.Context) {
	templateID := c.Param("id")

	var template models.Form
	if err := database.DB.Where("id = ? AND is_template = ?", templateID, true).First(&template).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	var updates models.Form
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields (prevent changing ID, UserID, IsTemplate)
	template.Title = updates.Title
	template.Description = updates.Description
	template.TemplateCategory = updates.TemplateCategory
	template.Fields = updates.Fields
	template.Settings = updates.Settings

	if err := database.DB.Save(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// Delete deletes a template (admin only)
// DELETE /api/admin/templates/:id
func (h *TemplateHandler) Delete(c *gin.Context) {
	templateID := c.Param("id")

	var template models.Form
	if err := database.DB.Where("id = ? AND is_template = ?", templateID, true).First(&template).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	if err := database.DB.Delete(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully"})
}
