package handlers

import (
	"net/http"
	"time"

	"formera/internal/database"
	"formera/internal/middleware"
	"formera/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	JWTSecret string
}

func NewAuthHandler(jwtSecret string) *AuthHandler {
	return &AuthHandler{JWTSecret: jwtSecret}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

// Register godoc
// @Summary      Register new user
// @Description  Create a new user account (if registration is enabled)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "Registration data"
// @Success      201 {object} AuthResponse
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      403 {object} ErrorResponse "Registration disabled"
// @Failure      409 {object} ErrorResponse "Email already registered"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var settings models.Settings
	database.DB.First(&settings)

	if !settings.AllowRegistration {
		c.JSON(http.StatusForbidden, gin.H{"error": "Registration is disabled"})
		return
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	if result := database.DB.Where("email = ?", req.Email).First(&existingUser); result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	user := &models.User{
		Email: req.Email,
		Name:  req.Name,
	}

	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if result := database.DB.Create(user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	token, err := h.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Token: token,
		User:  user,
	})
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate with email and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login credentials"
// @Success      200 {object} AuthResponse
// @Failure      400 {object} ErrorResponse "Invalid request"
// @Failure      401 {object} ErrorResponse "Invalid credentials"
// @Failure      429 {object} ErrorResponse "Rate limit exceeded"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if result := database.DB.Where("email = ?", req.Email).First(&user); result.Error != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token, err := h.generateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User:  &user,
	})
}

// Me godoc
// @Summary      Get current user
// @Description  Get the authenticated user's profile
// @Tags         Auth
// @Produce      json
// @Success      200 {object} models.User
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "User not found"
// @Security     BearerAuth
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("user_id")

	var user models.User
	if result := database.DB.First(&user, "id = ?", userID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) generateToken(user *models.User) (string, error) {
	claims := &middleware.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 7 * time.Hour)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.JWTSecret))
}
