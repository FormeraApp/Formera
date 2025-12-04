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

func init() {
	gin.SetMode(gin.TestMode)
}

func TestFormHandler_Create(t *testing.T) {
	testutil.SetupTestDB(t)

	handler := NewFormHandler()
	router := gin.New()
	router.POST("/forms", func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		handler.Create(c)
	})

	body := CreateFormRequest{
		Title:       "Test Form",
		Description: "A test form description",
		Fields:      models.FormFields{},
		Settings:    models.FormSettings{},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/forms", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var response models.Form
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Title != "Test Form" {
		t.Errorf("expected title 'Test Form', got %s", response.Title)
	}
	if response.Status != models.FormStatusDraft {
		t.Errorf("expected status 'draft', got %s", response.Status)
	}
}

func TestFormHandler_Create_SanitizesXSS(t *testing.T) {
	testutil.SetupTestDB(t)

	handler := NewFormHandler()
	router := gin.New()
	router.POST("/forms", func(c *gin.Context) {
		c.Set("user_id", "test-user-id")
		handler.Create(c)
	})

	body := CreateFormRequest{
		Title:       "<script>alert('xss')</script>Test Form",
		Description: "<p>Valid HTML</p><script>alert('xss')</script>",
		Fields:      models.FormFields{},
		Settings:    models.FormSettings{},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/forms", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var response models.Form
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Title should have script tags stripped
	if response.Title == "<script>alert('xss')</script>Test Form" {
		t.Error("XSS script tag was not stripped from title")
	}
	// Description should have script stripped but valid HTML kept
	if response.Description == "<p>Valid HTML</p><script>alert('xss')</script>" {
		t.Error("XSS script tag was not stripped from description")
	}
}

func TestFormHandler_List(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := testutil.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	// Create test forms
	form1 := &models.Form{UserID: user.ID, Title: "Form 1", Status: models.FormStatusDraft}
	form2 := &models.Form{UserID: user.ID, Title: "Form 2", Status: models.FormStatusPublished}
	db.Create(form1)
	db.Create(form2)

	handler := NewFormHandler()
	router := gin.New()
	router.GET("/forms", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		handler.List(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/forms", nil)
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
		t.Errorf("expected 2 forms, got %d", response.TotalItems)
	}
}

func TestFormHandler_List_Pagination(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := testutil.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	// Create 25 test forms
	for i := 0; i < 25; i++ {
		form := &models.Form{UserID: user.ID, Title: "Form", Status: models.FormStatusDraft}
		db.Create(form)
	}

	handler := NewFormHandler()
	router := gin.New()
	router.GET("/forms", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		handler.List(c)
	})

	// Test first page
	req := httptest.NewRequest(http.MethodGet, "/forms?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response pagination.Result
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.TotalItems != 25 {
		t.Errorf("expected total 25 forms, got %d", response.TotalItems)
	}
	if response.TotalPages != 3 {
		t.Errorf("expected 3 pages, got %d", response.TotalPages)
	}
	if response.Page != 1 {
		t.Errorf("expected page 1, got %d", response.Page)
	}
}

func TestFormHandler_Get(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := testutil.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	form := &models.Form{UserID: user.ID, Title: "Test Form", Status: models.FormStatusDraft}
	db.Create(form)

	handler := NewFormHandler()
	router := gin.New()
	router.GET("/forms/:id", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		handler.Get(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/forms/"+form.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.Form
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Title != "Test Form" {
		t.Errorf("expected title 'Test Form', got %s", response.Title)
	}
}

func TestFormHandler_Get_NotFound(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := testutil.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	handler := NewFormHandler()
	router := gin.New()
	router.GET("/forms/:id", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		handler.Get(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/forms/non-existent-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestFormHandler_Get_WrongUser(t *testing.T) {
	db := testutil.SetupTestDB(t)
	owner := testutil.CreateTestUser(t, db, "owner@example.com", "password123", models.RoleUser)
	otherUser := testutil.CreateTestUser(t, db, "other@example.com", "password123", models.RoleUser)

	form := &models.Form{UserID: owner.ID, Title: "Test Form", Status: models.FormStatusDraft}
	db.Create(form)

	handler := NewFormHandler()
	router := gin.New()
	router.GET("/forms/:id", func(c *gin.Context) {
		c.Set("user_id", otherUser.ID) // Different user trying to access
		handler.Get(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/forms/"+form.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestFormHandler_Update(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := testutil.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	form := &models.Form{UserID: user.ID, Title: "Original Title", Status: models.FormStatusDraft}
	db.Create(form)

	handler := NewFormHandler()
	router := gin.New()
	router.PUT("/forms/:id", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		handler.Update(c)
	})

	body := UpdateFormRequest{
		Title:  "Updated Title",
		Status: models.FormStatusPublished,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/forms/"+form.ID, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response models.Form
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %s", response.Title)
	}
	if response.Status != models.FormStatusPublished {
		t.Errorf("expected status 'published', got %s", response.Status)
	}
}

func TestFormHandler_Delete(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := testutil.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	form := &models.Form{UserID: user.ID, Title: "Test Form", Status: models.FormStatusDraft}
	db.Create(form)

	// Create some submissions for this form
	submission := &models.Submission{FormID: form.ID, Data: map[string]interface{}{"field1": "value1"}}
	db.Create(submission)

	handler := NewFormHandler()
	router := gin.New()
	router.DELETE("/forms/:id", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/forms/"+form.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Verify form is deleted
	var deletedForm models.Form
	result := db.First(&deletedForm, "id = ?", form.ID)
	if result.Error == nil {
		t.Error("form should have been deleted")
	}

	// Verify submissions are deleted (transaction test)
	var deletedSubmission models.Submission
	result = db.First(&deletedSubmission, "form_id = ?", form.ID)
	if result.Error == nil {
		t.Error("submissions should have been deleted with form")
	}
}

func TestFormHandler_Duplicate(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := testutil.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	form := &models.Form{
		UserID:      user.ID,
		Title:       "Original Form",
		Description: "Original description",
		Status:      models.FormStatusPublished,
	}
	db.Create(form)

	handler := NewFormHandler()
	router := gin.New()
	router.POST("/forms/:id/duplicate", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		handler.Duplicate(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/forms/"+form.ID+"/duplicate", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var response models.Form
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Title != "Original Form (Kopie)" {
		t.Errorf("expected title 'Original Form (Kopie)', got %s", response.Title)
	}
	if response.Status != models.FormStatusDraft {
		t.Errorf("expected status 'draft', got %s", response.Status)
	}
	if response.ID == form.ID {
		t.Error("duplicated form should have a new ID")
	}
}

func TestFormHandler_GetPublic(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := testutil.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	form := &models.Form{
		UserID: user.ID,
		Title:  "Public Form",
		Slug:   "public-form",
		Status: models.FormStatusPublished,
	}
	db.Create(form)

	handler := NewFormHandler()
	router := gin.New()
	router.GET("/public/forms/:id", handler.GetPublic)

	// Test by ID
	req := httptest.NewRequest(http.MethodGet, "/public/forms/"+form.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Test by slug
	req = httptest.NewRequest(http.MethodGet, "/public/forms/public-form", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d for slug lookup, got %d", http.StatusOK, w.Code)
	}
}

func TestFormHandler_GetPublic_Draft(t *testing.T) {
	db := testutil.SetupTestDB(t)
	user := testutil.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	form := &models.Form{
		UserID: user.ID,
		Title:  "Draft Form",
		Status: models.FormStatusDraft, // Not published
	}
	db.Create(form)

	handler := NewFormHandler()
	router := gin.New()
	router.GET("/public/forms/:id", handler.GetPublic)

	req := httptest.NewRequest(http.MethodGet, "/public/forms/"+form.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d for draft form, got %d", http.StatusNotFound, w.Code)
	}
}
