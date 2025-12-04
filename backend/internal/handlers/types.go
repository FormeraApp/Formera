package handlers

// Swagger API Types - used for documentation only
// Types that are already defined in other files are not duplicated here

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message" example:"Operation successful"`
}

// SlugCheckResponse represents slug availability check
type SlugCheckResponse struct {
	Available bool   `json:"available" example:"true"`
	Slug      string `json:"slug" example:"contact-form"`
	Reason    string `json:"reason,omitempty" example:"taken"`
}

// SubmissionListResponse represents paginated submissions
type SubmissionListResponse struct {
	Form        interface{} `json:"form"`
	Submissions interface{} `json:"submissions"`
}

// FormStatsResponse represents form statistics
type FormStatsResponse struct {
	TotalSubmissions int                    `json:"total_submissions" example:"150"`
	FieldStats       map[string]interface{} `json:"field_stats"`
}

// SubmissionsByDateResponse represents submissions grouped by date
type SubmissionsByDateResponse struct {
	Date  string `json:"date" example:"2025-01-15"`
	Count int    `json:"count" example:"25"`
}
