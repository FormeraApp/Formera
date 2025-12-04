package pkg

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// PaginationResult holds paginated results
type PaginationResult struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int64       `json:"total_items"`
	TotalPages int         `json:"total_pages"`
}

// GetPaginationParams extracts pagination parameters from request
func GetPaginationParams(c *gin.Context) PaginationParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = DefaultPage
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
}

// Offset calculates the offset for database queries
func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Paginate applies pagination to a GORM query
func Paginate(params PaginationParams) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(params.Offset()).Limit(params.PageSize)
	}
}

// CreatePaginationResult creates a paginated result
func CreatePaginationResult(data interface{}, params PaginationParams, totalItems int64) PaginationResult {
	totalPages := int(totalItems) / params.PageSize
	if int(totalItems)%params.PageSize > 0 {
		totalPages++
	}

	return PaginationResult{
		Data:       data,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}
