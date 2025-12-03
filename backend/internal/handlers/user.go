package handlers

import (
	"net/http"

	"formera/internal/database"
	"formera/internal/models"

	"github.com/gin-gonic/gin"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

type CreateUserRequest struct {
	Email    string          `json:"email" binding:"required,email"`
	Password string          `json:"password" binding:"required,min=8"`
	Name     string          `json:"name" binding:"required"`
	Role     models.UserRole `json:"role"`
}

type UpdateUserRequest struct {
	Email    string          `json:"email"`
	Password string          `json:"password"`
	Name     string          `json:"name"`
	Role     models.UserRole `json:"role"`
}

func (h *UserHandler) List(c *gin.Context) {
	var users []models.User
	if result := database.DB.Find(&users); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) Get(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if result := database.DB.First(&user, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	if result := database.DB.Where("email = ?", req.Email).First(&existingUser); result.Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	role := req.Role
	if role == "" {
		role = models.RoleUser
	}

	user := &models.User{
		Email: req.Email,
		Name:  req.Name,
		Role:  role,
	}

	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if result := database.DB.Create(user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if result := database.DB.First(&user, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Email != "" && req.Email != user.Email {
		var existingUser models.User
		if result := database.DB.Where("email = ?", req.Email).First(&existingUser); result.Error == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}
		user.Email = req.Email
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	if req.Password != "" {
		if err := user.SetPassword(req.Password); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
	}

	if req.Role != "" && req.Role != user.Role {
		if user.Role == models.RoleAdmin && req.Role != models.RoleAdmin {
			var adminCount int64
			database.DB.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&adminCount)
			if adminCount <= 1 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove the last admin"})
				return
			}
		}
		user.Role = req.Role
	}

	if result := database.DB.Save(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	currentUserID := c.GetString("user_id")

	if id == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	var user models.User
	if result := database.DB.First(&user, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var userCount int64
	database.DB.Model(&models.User{}).Count(&userCount)
	if userCount <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete the last user"})
		return
	}

	if user.Role == models.RoleAdmin {
		var adminCount int64
		database.DB.Model(&models.User{}).Where("role = ?", models.RoleAdmin).Count(&adminCount)
		if adminCount <= 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete the last admin"})
			return
		}
	}

	if result := database.DB.Delete(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
