package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"formera/internal/models"
	"formera/internal/pkg"
	"formera/internal/services"

	"github.com/gin-gonic/gin"
)

// Helper function to create a spam protection service for tests (no CAPTCHA configured)
func createTestSpamProtectionService() *services.SpamProtectionService {
	captchaService := services.NewCaptchaService("", "", "") // Empty secrets = no CAPTCHA
	return services.NewSpamProtectionService(captchaService)
}

func TestSubmissionHandler_Submit(t *testing.T) {
	db := pkg.SetupTestDB(t)
	user := pkg.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	form := &models.Form{
		UserID: user.ID,
		Title:  "Test Form",
		Status: models.FormStatusPublished,
		Fields: models.FormFields{
			{ID: "field1", Label: "Field 1", Type: "text", Required: true},
		},
		Settings: models.FormSettings{
			SuccessMessage: "Thank you!",
		},
	}
	db.Create(form)

	handler := NewSubmissionHandler(createTestSpamProtectionService())
	router := gin.New()
	router.POST("/public/forms/:id/submit", handler.Submit)

	body := SubmitRequest{
		Data: map[string]interface{}{
			"field1": "test value",
		},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/public/forms/"+form.ID+"/submit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["message"] != "Thank you!" {
		t.Errorf("expected success message 'Thank you!', got %v", response["message"])
	}
}

func TestSubmissionHandler_Submit_RequiredFieldMissing(t *testing.T) {
	db := pkg.SetupTestDB(t)
	user := pkg.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	form := &models.Form{
		UserID: user.ID,
		Title:  "Test Form",
		Status: models.FormStatusPublished,
		Fields: models.FormFields{
			{ID: "field1", Label: "Field 1", Type: "text", Required: true},
		},
	}
	db.Create(form)

	handler := NewSubmissionHandler(createTestSpamProtectionService())
	router := gin.New()
	router.POST("/public/forms/:id/submit", handler.Submit)

	body := SubmitRequest{
		Data: map[string]interface{}{}, // Missing required field
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/public/forms/"+form.ID+"/submit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSubmissionHandler_Submit_FormNotPublished(t *testing.T) {
	db := pkg.SetupTestDB(t)
	user := pkg.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	form := &models.Form{
		UserID: user.ID,
		Title:  "Test Form",
		Status: models.FormStatusDraft, // Not published
	}
	db.Create(form)

	handler := NewSubmissionHandler(createTestSpamProtectionService())
	router := gin.New()
	router.POST("/public/forms/:id/submit", handler.Submit)

	body := SubmitRequest{
		Data: map[string]interface{}{"field1": "value"},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/public/forms/"+form.ID+"/submit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestSubmissionHandler_Submit_MaxSubmissionsReached(t *testing.T) {
	db := pkg.SetupTestDB(t)
	user := pkg.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	form := &models.Form{
		UserID: user.ID,
		Title:  "Test Form",
		Status: models.FormStatusPublished,
		Settings: models.FormSettings{
			MaxSubmissions: 1,
		},
	}
	db.Create(form)

	// Create existing submission
	db.Create(&models.Submission{FormID: form.ID, Data: map[string]interface{}{}})

	handler := NewSubmissionHandler(createTestSpamProtectionService())
	router := gin.New()
	router.POST("/public/forms/:id/submit", handler.Submit)

	body := SubmitRequest{
		Data: map[string]interface{}{"field1": "value"},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/public/forms/"+form.ID+"/submit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestSubmissionHandler_Submit_SanitizesXSS(t *testing.T) {
	db := pkg.SetupTestDB(t)
	user := pkg.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	form := &models.Form{
		UserID: user.ID,
		Title:  "Test Form",
		Status: models.FormStatusPublished,
		Fields: models.FormFields{
			{ID: "field1", Label: "Field 1", Type: "text", Required: true},
		},
	}
	db.Create(form)

	handler := NewSubmissionHandler(createTestSpamProtectionService())
	router := gin.New()
	router.POST("/public/forms/:id/submit", handler.Submit)

	body := SubmitRequest{
		Data: map[string]interface{}{
			"field1": "<script>alert('xss')</script>Hello",
		},
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/public/forms/"+form.ID+"/submit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	// Check the stored submission
	var submission models.Submission
	db.First(&submission, "form_id = ?", form.ID)

	if val, ok := submission.Data["field1"].(string); ok {
		if val == "<script>alert('xss')</script>Hello" {
			t.Error("XSS script tag was not stripped from submission data")
		}
	}
}

func TestSubmissionHandler_List(t *testing.T) {
	db := pkg.SetupTestDB(t)
	user := pkg.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	form := &models.Form{UserID: user.ID, Title: "Test Form", Status: models.FormStatusPublished}
	db.Create(form)

	// Create submissions
	db.Create(&models.Submission{FormID: form.ID, Data: map[string]interface{}{"field1": "value1"}})
	db.Create(&models.Submission{FormID: form.ID, Data: map[string]interface{}{"field1": "value2"}})

	handler := NewSubmissionHandler(createTestSpamProtectionService())
	router := gin.New()
	router.GET("/forms/:id/submissions", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		handler.List(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/forms/"+form.ID+"/submissions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestSubmissionHandler_List_WrongUser(t *testing.T) {
	db := pkg.SetupTestDB(t)
	owner := pkg.CreateTestUser(t, db, "owner@example.com", "password123", models.RoleUser)
	otherUser := pkg.CreateTestUser(t, db, "other@example.com", "password123", models.RoleUser)

	form := &models.Form{UserID: owner.ID, Title: "Test Form", Status: models.FormStatusPublished}
	db.Create(form)

	handler := NewSubmissionHandler(createTestSpamProtectionService())
	router := gin.New()
	router.GET("/forms/:id/submissions", func(c *gin.Context) {
		c.Set("user_id", otherUser.ID) // Different user
		handler.List(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/forms/"+form.ID+"/submissions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestSubmissionHandler_Delete(t *testing.T) {
	db := pkg.SetupTestDB(t)
	user := pkg.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	form := &models.Form{UserID: user.ID, Title: "Test Form", Status: models.FormStatusPublished}
	db.Create(form)

	submission := &models.Submission{FormID: form.ID, Data: map[string]interface{}{"field1": "value1"}}
	db.Create(submission)

	handler := NewSubmissionHandler(createTestSpamProtectionService())
	router := gin.New()
	router.DELETE("/forms/:id/submissions/:submissionId", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		handler.Delete(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/forms/"+form.ID+"/submissions/"+submission.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	// Verify submission is deleted
	var deletedSubmission models.Submission
	result := db.First(&deletedSubmission, "id = ?", submission.ID)
	if result.Error == nil {
		t.Error("submission should have been deleted")
	}
}

func TestSubmissionHandler_Stats(t *testing.T) {
	db := pkg.SetupTestDB(t)
	user := pkg.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	form := &models.Form{
		UserID: user.ID,
		Title:  "Test Form",
		Status: models.FormStatusPublished,
		Fields: models.FormFields{
			{ID: "rating", Label: "Rating", Type: "select"},
		},
	}
	db.Create(form)

	// Create submissions with different values
	db.Create(&models.Submission{FormID: form.ID, Data: map[string]interface{}{"rating": "good"}})
	db.Create(&models.Submission{FormID: form.ID, Data: map[string]interface{}{"rating": "good"}})
	db.Create(&models.Submission{FormID: form.ID, Data: map[string]interface{}{"rating": "bad"}})

	handler := NewSubmissionHandler(createTestSpamProtectionService())
	router := gin.New()
	router.GET("/forms/:id/stats", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		handler.Stats(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/forms/"+form.ID+"/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["total_submissions"].(float64) != 3 {
		t.Errorf("expected 3 total submissions, got %v", response["total_submissions"])
	}
}

func TestSubmissionHandler_Stats_ConversionRate(t *testing.T) {
	db := pkg.SetupTestDB(t)
	user := pkg.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	form := &models.Form{
		UserID:    user.ID,
		Title:     "Test Form",
		Status:    models.FormStatusPublished,
		ViewCount: 100, // 100 views
	}
	db.Create(form)

	// Create 10 submissions (10% conversion rate)
	for i := 0; i < 10; i++ {
		db.Create(&models.Submission{FormID: form.ID, Data: map[string]interface{}{}})
	}

	handler := NewSubmissionHandler(createTestSpamProtectionService())
	router := gin.New()
	router.GET("/forms/:id/stats", func(c *gin.Context) {
		c.Set("user_id", user.ID)
		handler.Stats(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/forms/"+form.ID+"/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["total_submissions"].(float64) != 10 {
		t.Errorf("expected 10 total submissions, got %v", response["total_submissions"])
	}

	if response["total_views"].(float64) != 100 {
		t.Errorf("expected 100 total views, got %v", response["total_views"])
	}

	conversionRate := response["conversion_rate"].(float64)
	if conversionRate != 10.0 {
		t.Errorf("expected 10%% conversion rate, got %v%%", conversionRate)
	}
}

// SPAM PROTECTION INTEGRATION TESTS

func TestSubmissionHandler_SpamProtection_HoneypotFilled(t *testing.T) {
	db := pkg.SetupTestDB(t)
	user := pkg.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	// Settings are already initialized by SetupTestDB

	form := &models.Form{
		UserID: user.ID,
		Title:  "Test Form",
		Status: models.FormStatusPublished,
		Fields: models.FormFields{
			{ID: "field1", Label: "Field 1", Type: "text", Required: true},
		},
	}
	db.Create(form)

	handler := NewSubmissionHandler(createTestSpamProtectionService())
	router := gin.New()
	router.POST("/public/forms/:id/submit", handler.Submit)

	body := SubmitRequest{
		Data: map[string]interface{}{
			"field1": "test value",
		},
		HoneypotField: "bot-filled-this", // Honeypot filled = spam
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/public/forms/"+form.ID+"/submit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should be rejected as spam
	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d (Forbidden), got %d: %s", http.StatusForbidden, w.Code, w.Body.String())
	}

	// Verify no submission was created
	var count int64
	db.Model(&models.Submission{}).Where("form_id = ?", form.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 submissions, got %d", count)
	}
}

func TestSubmissionHandler_SpamProtection_HoneypotEmpty(t *testing.T) {
	db := pkg.SetupTestDB(t)
	user := pkg.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	// Settings are already initialized by SetupTestDB

	form := &models.Form{
		UserID: user.ID,
		Title:  "Test Form",
		Status: models.FormStatusPublished,
		Fields: models.FormFields{
			{ID: "field1", Label: "Field 1", Type: "text", Required: true},
		},
		Settings: models.FormSettings{
			SuccessMessage: "Thank you!",
		},
	}
	db.Create(form)

	handler := NewSubmissionHandler(createTestSpamProtectionService())
	router := gin.New()
	router.POST("/public/forms/:id/submit", handler.Submit)

	body := SubmitRequest{
		Data: map[string]interface{}{
			"field1": "test value",
		},
		HoneypotField: "", // Empty honeypot = legitimate
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/public/forms/"+form.ID+"/submit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should succeed
	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	// Verify submission was created
	var count int64
	db.Model(&models.Submission{}).Where("form_id = ?", form.ID).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 submission, got %d", count)
	}
}

func TestSubmissionHandler_SpamProtection_CaptchaRequired_NoToken(t *testing.T) {
	db := pkg.SetupTestDB(t)
	user := pkg.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	// Update existing settings to enable CAPTCHA
	var settings models.Settings
	db.First(&settings)
	settings.SpamProtection.CaptchaProvider = models.CaptchaProviderTurnstile
	settings.SpamProtection.TurnstileSiteKey = "test-site-key"
	db.Save(&settings)

	form := &models.Form{
		UserID: user.ID,
		Title:  "Test Form",
		Status: models.FormStatusPublished,
		Fields: models.FormFields{
			{ID: "field1", Label: "Field 1", Type: "text", Required: true},
		},
	}
	db.Create(form)

	handler := NewSubmissionHandler(createTestSpamProtectionService())
	router := gin.New()
	router.POST("/public/forms/:id/submit", handler.Submit)

	body := SubmitRequest{
		Data: map[string]interface{}{
			"field1": "test value",
		},
		HoneypotField: "",
		// Missing CaptchaToken
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/public/forms/"+form.ID+"/submit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should be rejected (CAPTCHA required but not provided)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d (Forbidden), got %d: %s", http.StatusForbidden, w.Code, w.Body.String())
	}

	// Verify no submission was created
	var count int64
	db.Model(&models.Submission{}).Where("form_id = ?", form.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 submissions, got %d", count)
	}
}

func TestSubmissionHandler_SpamProtection_NoSettings(t *testing.T) {
	db := pkg.SetupTestDB(t)
	user := pkg.CreateTestUser(t, db, "test@example.com", "password123", models.RoleUser)

	// Don't create settings - test graceful handling

	form := &models.Form{
		UserID: user.ID,
		Title:  "Test Form",
		Status: models.FormStatusPublished,
		Fields: models.FormFields{
			{ID: "field1", Label: "Field 1", Type: "text", Required: true},
		},
		Settings: models.FormSettings{
			SuccessMessage: "Thank you!",
		},
	}
	db.Create(form)

	handler := NewSubmissionHandler(createTestSpamProtectionService())
	router := gin.New()
	router.POST("/public/forms/:id/submit", handler.Submit)

	body := SubmitRequest{
		Data: map[string]interface{}{
			"field1": "test value",
		},
		HoneypotField: "",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/public/forms/"+form.ID+"/submit", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should allow submission (graceful fallback when settings don't exist)
	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
}
