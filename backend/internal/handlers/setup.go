package handlers

import (
	"net/http"

	"formera/internal/database"
	"formera/internal/models"

	"github.com/gin-gonic/gin"
)

type SetupHandler struct {
	JWTSecret string
}

func NewSetupHandler(jwtSecret string) *SetupHandler {
	return &SetupHandler{JWTSecret: jwtSecret}
}

type SetupStatusResponse struct {
	SetupRequired      bool               `json:"setup_required"`
	AllowRegistration  bool               `json:"allow_registration"`
	AppName            string             `json:"app_name"`
	FooterLinks        models.FooterLinks `json:"footer_links"`
	PrimaryColor       string             `json:"primary_color"`
	LogoURL            string             `json:"logo_url"`
	LogoShowText       bool               `json:"logo_show_text"`
	FaviconURL         string             `json:"favicon_url"`
	LoginBackgroundURL string             `json:"login_background_url"`
	Language           string             `json:"language"`
	Theme              string             `json:"theme"`
}

type SetupRequest struct {
	Email             string `json:"email" binding:"required,email"`
	Password          string `json:"password" binding:"required,min=8"`
	Name              string `json:"name" binding:"required"`
	AppName           string `json:"app_name"`
	AllowRegistration bool   `json:"allow_registration"`
}

func (h *SetupHandler) GetStatus(c *gin.Context) {
	var settings models.Settings
	database.DB.First(&settings)

	var userCount int64
	database.DB.Model(&models.User{}).Count(&userCount)

	setupRequired := userCount == 0

	c.JSON(http.StatusOK, SetupStatusResponse{
		SetupRequired:      setupRequired,
		AllowRegistration:  settings.AllowRegistration,
		AppName:            settings.AppName,
		FooterLinks:        settings.FooterLinks,
		PrimaryColor:       settings.PrimaryColor,
		LogoURL:            settings.LogoURL,
		LogoShowText:       settings.LogoShowText,
		FaviconURL:         settings.FaviconURL,
		LoginBackgroundURL: settings.LoginBackgroundURL,
		Language:           settings.Language,
		Theme:              settings.Theme,
	})
}

func (h *SetupHandler) CompleteSetup(c *gin.Context) {
	var userCount int64
	database.DB.Model(&models.User{}).Count(&userCount)

	if userCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setup already completed"})
		return
	}

	var req SetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &models.User{
		Email: req.Email,
		Name:  req.Name,
		Role:  models.RoleAdmin,
	}

	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if result := database.DB.Create(user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	var settings models.Settings
	database.DB.First(&settings)

	settings.SetupCompleted = true
	settings.AllowRegistration = req.AllowRegistration
	if req.AppName != "" {
		settings.AppName = req.AppName
	}

	database.DB.Save(&settings)

	authHandler := NewAuthHandler(h.JWTSecret)
	token, err := authHandler.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User:  user,
	})
}

func (h *SetupHandler) GetSettings(c *gin.Context) {
	var settings models.Settings
	database.DB.First(&settings)

	c.JSON(http.StatusOK, settings)
}

type UpdateSettingsRequest struct {
	AllowRegistration  *bool               `json:"allow_registration"`
	AppName            string              `json:"app_name"`
	FooterLinks        *models.FooterLinks `json:"footer_links"`
	PrimaryColor       string              `json:"primary_color"`
	LogoURL            *string             `json:"logo_url"`
	LogoShowText       *bool               `json:"logo_show_text"`
	FaviconURL         *string             `json:"favicon_url"`
	LoginBackgroundURL *string             `json:"login_background_url"`
	Language           string              `json:"language"`
	Theme              string              `json:"theme"`
}

func (h *SetupHandler) UpdateSettings(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var settings models.Settings
	database.DB.First(&settings)

	if req.AllowRegistration != nil {
		settings.AllowRegistration = *req.AllowRegistration
	}
	if req.AppName != "" {
		settings.AppName = req.AppName
	}
	if req.FooterLinks != nil {
		settings.FooterLinks = *req.FooterLinks
	}
	if req.PrimaryColor != "" {
		settings.PrimaryColor = req.PrimaryColor
	}
	if req.LogoURL != nil {
		settings.LogoURL = *req.LogoURL
	}
	if req.LogoShowText != nil {
		settings.LogoShowText = *req.LogoShowText
	}
	if req.FaviconURL != nil {
		settings.FaviconURL = *req.FaviconURL
	}
	if req.LoginBackgroundURL != nil {
		settings.LoginBackgroundURL = *req.LoginBackgroundURL
	}
	if req.Language != "" {
		settings.Language = req.Language
	}
	if req.Theme != "" {
		settings.Theme = req.Theme
	}

	database.DB.Save(&settings)

	c.JSON(http.StatusOK, settings)
}
