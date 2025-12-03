package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"formera/internal/models"
	"formera/internal/testutil"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	handler := NewAuthHandler("test-secret")
	router := gin.New()
	router.POST("/login", handler.Login)

	body := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Token == "" {
		t.Error("expected token in response")
	}
	if response.User == nil {
		t.Error("expected user in response")
	}
	if response.User.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", response.User.Email)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	handler := NewAuthHandler("test-secret")
	router := gin.New()
	router.POST("/login", handler.Login)

	tests := []struct {
		name     string
		email    string
		password string
	}{
		{"wrong password", "test@example.com", "wrongpassword"},
		{"wrong email", "wrong@example.com", "password123"},
		{"both wrong", "wrong@example.com", "wrongpassword"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := LoginRequest{
				Email:    tt.email,
				Password: tt.password,
			}
			jsonBody, _ := json.Marshal(body)

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
			}
		})
	}
}

func TestAuthHandler_Login_InvalidRequest(t *testing.T) {
	testutil.SetupTestDB(t)

	handler := NewAuthHandler("test-secret")
	router := gin.New()
	router.POST("/login", handler.Login)

	tests := []struct {
		name string
		body string
	}{
		{"empty body", "{}"},
		{"missing email", `{"password": "test123"}`},
		{"missing password", `{"email": "test@example.com"}`},
		{"invalid email format", `{"email": "notanemail", "password": "test123"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestAuthHandler_Register_Disabled(t *testing.T) {
	db := testutil.SetupTestDB(t)
	db.Model(&models.Settings{}).Where("id = ?", 1).Update("allow_registration", false)

	handler := NewAuthHandler("test-secret")
	router := gin.New()
	router.POST("/register", handler.Register)

	body := RegisterRequest{
		Email:    "new@example.com",
		Password: "password123",
		Name:     "New User",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestAuthHandler_Register_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	db.Model(&models.Settings{}).Where("id = ?", 1).Update("allow_registration", true)

	handler := NewAuthHandler("test-secret")
	router := gin.New()
	router.POST("/register", handler.Register)

	body := RegisterRequest{
		Email:    "new@example.com",
		Password: "password123",
		Name:     "New User",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var response AuthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Token == "" {
		t.Error("expected token in response")
	}
	if response.User.Email != "new@example.com" {
		t.Errorf("expected email new@example.com, got %s", response.User.Email)
	}
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	db := testutil.SetupTestDB(t)
	db.Model(&models.Settings{}).Where("id = ?", 1).Update("allow_registration", true)
	testutil.CreateTestUser(t, db, "existing@example.com", "password123", models.RoleUser)

	handler := NewAuthHandler("test-secret")
	router := gin.New()
	router.POST("/register", handler.Register)

	body := RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
		Name:     "New User",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestAuthHandler_Me_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := testutil.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	handler := NewAuthHandler("test-secret")
	router := gin.New()
	router.GET("/me", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		handler.Me(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.User
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", response.Email)
	}
}

func TestAuthHandler_Me_NotFound(t *testing.T) {
	testutil.SetupTestDB(t)

	handler := NewAuthHandler("test-secret")
	router := gin.New()
	router.GET("/me", func(c *gin.Context) {
		c.Set("user_id", "non-existent-id")
		handler.Me(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
