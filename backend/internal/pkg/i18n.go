package pkg

// Translations holds all translatable strings
var translations = map[string]map[string]string{
	"en": {
		"webhook.event.submission.created": "New Form Submission",
		"webhook.event.submission.deleted": "Submission Deleted",
		"webhook.event.test":               "Test Webhook",
		"webhook.field.form":               "Form",
		"webhook.field.slug":               "Slug",
		"webhook.footer":                   "Formera",
		"webhook.test.field.name":          "Name",
		"webhook.test.field.email":         "Email",
		"webhook.test.field.message":       "Message",
	},
	"de": {
		"webhook.event.submission.created": "Neue Formular-Einreichung",
		"webhook.event.submission.deleted": "Einreichung gelöscht",
		"webhook.event.test":               "Test Webhook",
		"webhook.field.form":               "Formular",
		"webhook.field.slug":               "Slug",
		"webhook.footer":                   "Formera",
		"webhook.test.field.name":          "Name",
		"webhook.test.field.email":         "E-Mail",
		"webhook.test.field.message":       "Nachricht",
	},
}

// T returns the translated string for the given key and language
// Falls back to English if the key is not found in the requested language
func T(lang, key string) string {
	if lang == "" {
		lang = "en"
	}

	// Try requested language
	if langMap, ok := translations[lang]; ok {
		if val, ok := langMap[key]; ok {
			return val
		}
	}

	// Fallback to English
	if langMap, ok := translations["en"]; ok {
		if val, ok := langMap[key]; ok {
			return val
		}
	}

	// Return key if nothing found
	return key
}
