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
	err = DB.AutoMigrate(&models.User{}, &models.Form{}, &models.Submission{}, &models.Settings{}, &storage.FileRecord{})
	if err != nil {
		return err
	}

	// Initialize settings if not exists
	var settings models.Settings
	if result := DB.First(&settings); result.Error != nil {
		DB.Create(models.GetDefaultSettings())
	}

	log.Println("Database initialized successfully")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}
