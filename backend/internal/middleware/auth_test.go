package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"formera/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func generateTestToken(secret, userID, email, role string, expired bool) string {
	expiresAt := time.Now().Add(time.Hour)
	if expired {
		expiresAt = time.Now().Add(-time.Hour)
	}

	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	secret := "test-secret"
	token := generateTestToken(secret, "user-123", "test@example.com", "user", false)

	router := gin.New()
	router.Use(AuthMiddleware(secret))
	router.GET("/protected", func(c *gin.Context) {
		userID := c.GetString("user_id")
		email := c.GetString("email")
		c.JSON(http.StatusOK, gin.H{"user_id": userID, "email": email})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware("test-secret"))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware("test-secret"))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	tests := []struct {
		name   string
		header string
	}{
		{"no Bearer prefix", "token-only"},
		{"wrong prefix", "Basic token"},
		{"empty token", "Bearer "},
		{"too many parts", "Bearer token extra"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", tt.header)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
			}
		})
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	token := generateTestToken(secret, "user-123", "test@example.com", "user", true)

	router := gin.New()
	router.Use(AuthMiddleware(secret))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuthMiddleware_InvalidSecret(t *testing.T) {
	token := generateTestToken("correct-secret", "user-123", "test@example.com", "user", false)

	router := gin.New()
	router.Use(AuthMiddleware("wrong-secret"))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAdminMiddleware_AdminUser(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "admin-123")
		c.Set("user_role", string(models.RoleAdmin))
		c.Next()
	})
	router.Use(AdminMiddleware())
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestAdminMiddleware_NonAdminUser(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("user_role", string(models.RoleUser))
		c.Next()
	})
	router.Use(AdminMiddleware())
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestAdminMiddleware_NoRole(t *testing.T) {
	router := gin.New()
	// No user_role set in context
	router.Use(AdminMiddleware())
	router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
