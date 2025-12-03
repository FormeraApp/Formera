package testutil

import (
	"formera/internal/database"
	"formera/internal/models"
	"formera/internal/storage"
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Form{}, &models.Submission{}, &models.Settings{}, &storage.FileRecord{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	db.Create(models.GetDefaultSettings())

	database.DB = db
	return db
}

func CreateTestUser(t *testing.T, db *gorm.DB, email, password string, role models.UserRole) *models.User {
	t.Helper()

	user := &models.User{
		Email: email,
		Name:  "Test User",
		Role:  role,
	}
	if err := user.SetPassword(password); err != nil {
		t.Fatalf("failed to set password: %v", err)
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

func SetupTestEnv(t *testing.T) {
	t.Helper()
	os.Setenv("JWT_SECRET", "test-secret-key")
}

func CleanupTestEnv(t *testing.T) {
	t.Helper()
	os.Unsetenv("JWT_SECRET")
}
