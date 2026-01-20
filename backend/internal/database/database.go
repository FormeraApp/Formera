package database

import (
	"log"
	"os"
	"path/filepath"

	"formera/internal/models"
	"formera/internal/storage"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Initialize(dbPath string) error {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	// Auto-migrate the schema
	err = DB.AutoMigrate(&models.User{}, &models.Form{}, &models.Submission{}, &models.Settings{}, &storage.FileRecord{}, &models.AuditLog{}, &models.Webhook{}, &models.WebhookLog{})
	if err != nil {
		return err
	}

	// Initialize settings if not exists
	var settings models.Settings
	if result := DB.First(&settings); result.Error != nil {
		DB.Create(models.GetDefaultSettings())
	} else {
		// Migration: Initialize spam protection config if not set
		// Check if spam_protection is empty (for existing installations)
		if settings.SpamProtection.CaptchaProvider == "" {
			settings.SpamProtection = models.GetDefaultSpamProtectionConfig()
			DB.Save(&settings)
			log.Println("Migrated settings: initialized spam protection config")
		}
	}

	// Seed templates if none exist
	if err := seedTemplates(); err != nil {
		log.Printf("Warning: Failed to seed templates: %v", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// seedTemplates creates default form templates if they don't exist
func seedTemplates() error {
	var count int64
	if err := DB.Model(&models.Form{}).Where("is_template = ?", true).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		log.Printf("Templates already exist (%d templates found), skipping seeding", count)
		return nil
	}

	templates := getDefaultTemplates()
	for _, tmpl := range templates {
		// Use system user ID (you might want to create a dedicated system user)
		tmpl.UserID = "system"
		if err := DB.Create(&tmpl).Error; err != nil {
			log.Printf("Failed to create template '%s': %v", tmpl.Title, err)
			return err
		}
	}

	log.Printf("Successfully seeded %d form templates", len(templates))
	return nil
}

func GetDB() *gorm.DB {
	return DB
}
