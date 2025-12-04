package pkg

import (
	"github.com/microcosm-cc/bluemonday"
)

var (
	// strictPolicy strips all HTML - use for text-only fields
	strictPolicy *bluemonday.Policy

	// ugcPolicy allows safe HTML for user-generated content
	ugcPolicy *bluemonday.Policy
)

func init() {
	// Strict policy: no HTML allowed at all
	strictPolicy = bluemonday.StrictPolicy()

	// UGC policy: allows safe HTML tags
	ugcPolicy = bluemonday.UGCPolicy()
}

// StripHTML removes all HTML tags from the input
func StripHTML(input string) string {
	return strictPolicy.Sanitize(input)
}

// SanitizeHTML allows safe HTML while removing dangerous elements
func SanitizeHTML(input string) string {
	return ugcPolicy.Sanitize(input)
}

// SanitizeFormField sanitizes a form field value based on its type
func SanitizeFormField(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return StripHTML(v)
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = SanitizeFormField(item)
		}
		return result
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			result[StripHTML(key)] = SanitizeFormField(val)
		}
		return result
	default:
		return v
	}
}

// SanitizeSubmissionData sanitizes all values in a submission data map
func SanitizeSubmissionData(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		result[key] = SanitizeFormField(value)
	}
	return result
}
