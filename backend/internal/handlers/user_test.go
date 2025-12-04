package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"formera/internal/models"
	"formera/internal/pagination"
	"formera/internal/testutil"

	"github.com/gin-gonic/gin"
)

func TestUserHandler_List(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.CreateTestUser(t, db, "user1@example.com", "password123", models.RoleUser)
	testutil.CreateTestUser(t, db, "user2@example.com", "password123", models.RoleAdmin)

	handler := NewUserHandler()
	router := gin.New()
	router.GET("/users", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response pagination.Result
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.TotalItems != 2 {
		t.Errorf("expected 2 users, got %d", response.TotalItems)
	}
}

func TestUserHandler_Get(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := testutil.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	handler := NewUserHandler()
	router := gin.New()
	router.GET("/users/:id", handler.Get)

	req := httptest.NewRequest(http.MethodGet, "/users/"+user.ID, nil)
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

func TestUserHandler_Get_NotFound(t *testing.T) {
	testutil.SetupTestDB(t)

	handler := NewUserHandler()
	router := gin.New()
	router.GET("/users/:id", handler.Get)

	req := httptest.NewRequest(http.MethodGet, "/users/non-existent-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUserHandler_Create(t *testing.T) {
	testutil.SetupTestDB(t)

	handler := NewUserHandler()
	router := gin.New()
	router.POST("/users", handler.Create)

	body := CreateUserRequest{
		Email:    "new@example.com",
		Password: "password123",
		Name:     "New User",
		Role:     models.RoleUser,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var response models.User
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Email != "new@example.com" {
		t.Errorf("expected email new@example.com, got %s", response.Email)
	}
}

func TestUserHandler_Create_DuplicateEmail(t *testing.T) {
	db := testutil.SetupTestDB(t)
	testutil.CreateTestUser(t, db, "existing@example.com", "password123", models.RoleUser)

	handler := NewUserHandler()
	router := gin.New()
	router.POST("/users", handler.Create)

	body := CreateUserRequest{
		Email:    "existing@example.com",
		Password: "password123",
		Name:     "New User",
		Role:     models.RoleUser,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestUserHandler_Update(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := testutil.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	handler := NewUserHandler()
	router := gin.New()
	router.PUT("/users/:id", handler.Update)

	body := UpdateUserRequest{
		Name: "Updated Name",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/users/"+user.ID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response models.User
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %s", response.Name)
	}
}

func TestUserHandler_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	admin := testutil.CreateTestUser(t, db, "admin@example.com", "password123", models.RoleAdmin)
	user := testutil.CreateTestUser(t, db, "user@example.com", "password123", models.RoleUser)

	handler := NewUserHandler()
	router := gin.New()
	router.DELETE("/users/:id", func(c *gin.Context) {
		c.Set("user_id", admin.ID)
		handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/users/"+user.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestUserHandler_Delete_SelfDelete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := testutil.CreateTestUser(t, db, "user@example.com", "password123", models.RoleUser)

	handler := NewUserHandler()
	router := gin.New()
	router.DELETE("/users/:id", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/users/"+user.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUserHandler_Delete_LastAdmin(t *testing.T) {
	db := testutil.SetupTestDB(t)
	admin := testutil.CreateTestUser(t, db, "admin@example.com", "password123", models.RoleAdmin)
	user := testutil.CreateTestUser(t, db, "user@example.com", "password123", models.RoleUser)

	handler := NewUserHandler()
	router := gin.New()
	router.DELETE("/users/:id", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/users/"+admin.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUserHandler_Delete_LastUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := testutil.CreateTestUser(t, db, "user@example.com", "password123", models.RoleUser)

	handler := NewUserHandler()
	router := gin.New()
	router.DELETE("/users/:id", func(c *gin.Context) {
		c.Set("user_id", "some-other-id")
		handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/users/"+user.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
