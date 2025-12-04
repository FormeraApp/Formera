package pagination

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

// Params holds pagination parameters
type Params struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// Result holds paginated results
type Result struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalItems int64       `json:"total_items"`
	TotalPages int         `json:"total_pages"`
}

// GetParams extracts pagination parameters from request
func GetParams(c *gin.Context) Params {
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

	return Params{
		Page:     page,
		PageSize: pageSize,
	}
}

// Offset calculates the offset for database queries
func (p Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Paginate applies pagination to a GORM query
func Paginate(params Params) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(params.Offset()).Limit(params.PageSize)
	}
}

// CreateResult creates a paginated result
func CreateResult(data interface{}, params Params, totalItems int64) Result {
	totalPages := int(totalItems) / params.PageSize
	if int(totalItems)%params.PageSize > 0 {
		totalPages++
	}

	return Result{
		Data:       data,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}
