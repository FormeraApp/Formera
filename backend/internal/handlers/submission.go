package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"sort"
	"time"

	"formera/internal/database"
	"formera/internal/models"

	"github.com/gin-gonic/gin"
)

type SubmissionHandler struct{}

func NewSubmissionHandler() *SubmissionHandler {
	return &SubmissionHandler{}
}

type SubmitRequest struct {
	Data     models.SubmissionData `json:"data" binding:"required"`
	Metadata map[string]string     `json:"metadata,omitempty"`
}

func (h *SubmissionHandler) Submit(c *gin.Context) {
	formID := c.Param("id")

	var form models.Form
	if result := database.DB.Where("id = ? AND status = ?", formID, models.FormStatusPublished).First(&form); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found or not accepting submissions"})
		return
	}

	if form.Settings.MaxSubmissions > 0 {
		var count int64
		database.DB.Model(&models.Submission{}).Where("form_id = ?", formID).Count(&count)
		if int(count) >= form.Settings.MaxSubmissions {
			c.JSON(http.StatusForbidden, gin.H{"error": "Maximale Anzahl an Einreichungen erreicht"})
			return
		}
	}

	now := time.Now()
	if form.Settings.StartDate != "" {
		startDate, err := time.Parse("2006-01-02T15:04", form.Settings.StartDate)
		if err != nil {
			startDate, err = time.Parse("2006-01-02", form.Settings.StartDate)
		}
		if err == nil && now.Before(startDate) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Formular ist noch nicht für Einreichungen geöffnet"})
			return
		}
	}
	if form.Settings.EndDate != "" {
		endDate, err := time.Parse("2006-01-02T15:04", form.Settings.EndDate)
		if err != nil {
			endDate, err = time.Parse("2006-01-02", form.Settings.EndDate)
			if err == nil {
				endDate = endDate.Add(24 * time.Hour)
			}
		}
		if err == nil && now.After(endDate) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Formular akzeptiert keine Einreichungen mehr"})
			return
		}
	}

	var req SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, field := range form.Fields {
		if field.Required {
			if val, ok := req.Data[field.ID]; !ok || val == nil || val == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Feld '%s' ist erforderlich", field.Label)})
				return
			}
		}
	}

	metadata := models.SubmissionMetadata{
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		Referrer:  c.GetHeader("Referer"),
	}

	if req.Metadata != nil {
		if v, ok := req.Metadata["utm_source"]; ok {
			metadata.UTMSource = v
		}
		if v, ok := req.Metadata["utm_medium"]; ok {
			metadata.UTMMedium = v
		}
		if v, ok := req.Metadata["utm_campaign"]; ok {
			metadata.UTMCampaign = v
		}
		if v, ok := req.Metadata["utm_term"]; ok {
			metadata.UTMTerm = v
		}
		if v, ok := req.Metadata["utm_content"]; ok {
			metadata.UTMContent = v
		}

		customTracking := make(map[string]string)
		for k, v := range req.Metadata {
			if k != "utm_source" && k != "utm_medium" && k != "utm_campaign" && k != "utm_term" && k != "utm_content" {
				customTracking[k] = v
			}
		}
		if len(customTracking) > 0 {
			metadata.Tracking = customTracking
		}
	}

	submission := &models.Submission{
		FormID:   formID,
		Data:     req.Data,
		Metadata: metadata,
	}

	if result := database.DB.Create(submission); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save submission"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    form.Settings.SuccessMessage,
		"submission": submission,
	})
}

func (h *SubmissionHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	formID := c.Param("id")

	var form models.Form
	if result := database.DB.Where("id = ? AND user_id = ?", formID, userID).First(&form); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	var submissions []models.Submission
	if result := database.DB.Where("form_id = ?", formID).Order("created_at DESC").Find(&submissions); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch submissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"form":        form,
		"submissions": submissions,
		"count":       len(submissions),
	})
}

func (h *SubmissionHandler) Get(c *gin.Context) {
	userID := c.GetString("user_id")
	formID := c.Param("id")
	submissionID := c.Param("submissionId")

	var form models.Form
	if result := database.DB.Where("id = ? AND user_id = ?", formID, userID).First(&form); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	var submission models.Submission
	if result := database.DB.Where("id = ? AND form_id = ?", submissionID, formID).First(&submission); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Submission not found"})
		return
	}

	c.JSON(http.StatusOK, submission)
}

func (h *SubmissionHandler) Delete(c *gin.Context) {
	userID := c.GetString("user_id")
	formID := c.Param("id")
	submissionID := c.Param("submissionId")

	var form models.Form
	if result := database.DB.Where("id = ? AND user_id = ?", formID, userID).First(&form); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	if result := database.DB.Where("id = ? AND form_id = ?", submissionID, formID).Delete(&models.Submission{}); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete submission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Submission deleted successfully"})
}

func (h *SubmissionHandler) Stats(c *gin.Context) {
	userID := c.GetString("user_id")
	formID := c.Param("id")

	var form models.Form
	if result := database.DB.Where("id = ? AND user_id = ?", formID, userID).First(&form); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	var submissions []models.Submission
	database.DB.Where("form_id = ?", formID).Find(&submissions)

	fieldStats := make(map[string]interface{})
	for _, field := range form.Fields {
		stats := make(map[string]int)
		for _, sub := range submissions {
			if val, ok := sub.Data[field.ID]; ok {
				switch v := val.(type) {
				case string:
					stats[v]++
				case []interface{}:
					for _, item := range v {
						if str, ok := item.(string); ok {
							stats[str]++
						}
					}
				}
			}
		}
		if len(stats) > 0 {
			fieldStats[field.ID] = stats
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_submissions": len(submissions),
		"field_stats":       fieldStats,
	})
}

func (h *SubmissionHandler) ExportCSV(c *gin.Context) {
	userID := c.GetString("user_id")
	formID := c.Param("id")

	var form models.Form
	if result := database.DB.Where("id = ? AND user_id = ?", formID, userID).First(&form); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	var submissions []models.Submission
	database.DB.Where("form_id = ?", formID).Order("created_at ASC").Find(&submissions)

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s-submissions.csv", form.ID))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	headers := []string{"ID", "Submitted At"}
	for _, field := range form.Fields {
		headers = append(headers, field.Label)
	}
	_ = writer.Write(headers)

	for _, sub := range submissions {
		row := []string{sub.ID, sub.CreatedAt.Format(time.RFC3339)}
		for _, field := range form.Fields {
			val := ""
			if v, ok := sub.Data[field.ID]; ok {
				switch typed := v.(type) {
				case string:
					val = typed
				case []interface{}:
					strs := make([]string, len(typed))
					for i, item := range typed {
						strs[i] = fmt.Sprintf("%v", item)
					}
					val = fmt.Sprintf("%v", strs)
				default:
					val = fmt.Sprintf("%v", typed)
				}
			}
			row = append(row, val)
		}
		_ = writer.Write(row)
	}
}

func (h *SubmissionHandler) ExportJSON(c *gin.Context) {
	userID := c.GetString("user_id")
	formID := c.Param("id")

	var form models.Form
	if result := database.DB.Where("id = ? AND user_id = ?", formID, userID).First(&form); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	var submissions []models.Submission
	database.DB.Where("form_id = ?", formID).Order("created_at ASC").Find(&submissions)

	exportData := make([]map[string]interface{}, len(submissions))
	for i, sub := range submissions {
		record := map[string]interface{}{
			"id":           sub.ID,
			"submitted_at": sub.CreatedAt,
		}
		for _, field := range form.Fields {
			if val, ok := sub.Data[field.ID]; ok {
				record[field.Label] = val
			} else {
				record[field.Label] = nil
			}
		}
		exportData[i] = record
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s-submissions.json", form.ID))
	c.JSON(http.StatusOK, exportData)
}

func (h *SubmissionHandler) SubmissionsByDate(c *gin.Context) {
	userID := c.GetString("user_id")
	formID := c.Param("id")

	var form models.Form
	if result := database.DB.Where("id = ? AND user_id = ?", formID, userID).First(&form); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Form not found"})
		return
	}

	var submissions []models.Submission
	database.DB.Where("form_id = ?", formID).Find(&submissions)

	byDate := make(map[string]int)
	for _, sub := range submissions {
		date := sub.CreatedAt.Format("2006-01-02")
		byDate[date]++
	}

	type DateCount struct {
		Date  string `json:"date"`
		Count int    `json:"count"`
	}
	result := make([]DateCount, 0, len(byDate))
	for date, count := range byDate {
		result = append(result, DateCount{Date: date, Count: count})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Date < result[j].Date
	})

	c.JSON(http.StatusOK, result)
}
