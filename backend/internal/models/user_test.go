package models

import (
	"testing"
)

func TestUser_SetPassword(t *testing.T) {
	user := &User{}

	err := user.SetPassword("testpassword123")
	if err != nil {
		t.Fatalf("SetPassword returned error: %v", err)
	}

	if user.Password == "" {
		t.Error("Password should not be empty after SetPassword")
	}

	if user.Password == "testpassword123" {
		t.Error("Password should be hashed, not stored in plain text")
	}
}

func TestUser_CheckPassword(t *testing.T) {
	user := &User{}
	password := "testpassword123"

	err := user.SetPassword(password)
	if err != nil {
		t.Fatalf("SetPassword returned error: %v", err)
	}

	if !user.CheckPassword(password) {
		t.Error("CheckPassword should return true for correct password")
	}

	if user.CheckPassword("wrongpassword") {
		t.Error("CheckPassword should return false for incorrect password")
	}

	if user.CheckPassword("") {
		t.Error("CheckPassword should return false for empty password")
	}
}

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		role     UserRole
		expected bool
	}{
		{"admin role", RoleAdmin, true},
		{"user role", RoleUser, false},
		{"empty role", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Role: tt.role}
			if got := user.IsAdmin(); got != tt.expected {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestUserRole_Constants(t *testing.T) {
	if RoleAdmin != "admin" {
		t.Errorf("RoleAdmin should be 'admin', got %s", RoleAdmin)
	}
	if RoleUser != "user" {
		t.Errorf("RoleUser should be 'user', got %s", RoleUser)
	}
}
