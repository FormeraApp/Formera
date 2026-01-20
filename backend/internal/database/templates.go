package database

import (
	"formera/internal/models"
)

// getDefaultTemplates returns all default form templates
func getDefaultTemplates() []models.Form {
	return []models.Form{
		// German templates
		getContactFormTemplate(),
		getFeedbackFormTemplate(),
		getSurveyTemplate(),
		getEventRegistrationTemplate(),
		getJobApplicationTemplate(),
		getNewsletterSignupTemplate(),
		getQuoteRequestTemplate(),
		getSupportTicketTemplate(),
		getBookingTemplate(),
		getOrderFormTemplate(),
		getComplaintFormTemplate(),
		getSuggestionFormTemplate(),
		// English templates
		getContactFormTemplateEN(),
		getFeedbackFormTemplateEN(),
		getSurveyTemplateEN(),
		getEventRegistrationTemplateEN(),
		getJobApplicationTemplateEN(),
		getNewsletterSignupTemplateEN(),
		getQuoteRequestTemplateEN(),
		getSupportTicketTemplateEN(),
		getBookingTemplateEN(),
		getOrderFormTemplateEN(),
		getComplaintFormTemplateEN(),
		getSuggestionFormTemplateEN(),
	}
}

// getContactFormTemplate creates a contact form template (German)
func getContactFormTemplate() models.Form {
	return models.Form{
		Title:            "Kontaktformular",
		Description:      "Ein einfaches Kontaktformular für Ihre Website-Besucher",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategoryContact),
		Language:         "de",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "name",
				Type:        models.FieldTypeText,
				Label:       "Ihr Name",
				Placeholder: "Max Mustermann",
				Required:    true,
				Order:       0,
			},
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "E-Mail-Adresse",
				Placeholder: "max@beispiel.de",
				Required:    true,
				Order:       1,
			},
			{
				ID:          "subject",
				Type:        models.FieldTypeText,
				Label:       "Betreff",
				Placeholder: "Worum geht es?",
				Required:    true,
				Order:       2,
			},
			{
				ID:          "message",
				Type:        models.FieldTypeTextarea,
				Label:       "Nachricht",
				Placeholder: "Ihre Nachricht hier...",
				Required:    true,
				Order:       3,
				Validation: map[string]interface{}{
					"minLength": 10,
				},
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Nachricht senden",
			SuccessMessage:     "Vielen Dank für Ihre Nachricht! Wir werden uns bald bei Ihnen melden.",
			AllowMultiple:      true,
			NotifyOnSubmission: true,
		},
	}
}

// getFeedbackFormTemplate creates a feedback form template (German)
func getFeedbackFormTemplate() models.Form {
	return models.Form{
		Title:            "Feedback-Formular",
		Description:      "Sammeln Sie Feedback von Ihren Nutzern",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategoryFeedback),
		Language:         "de",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:       "rating",
				Type:     models.FieldTypeRating,
				Label:    "Wie würden Sie Ihre Erfahrung bewerten?",
				Required: true,
				Order:    0,
				MinValue: 1,
				MaxValue: 5,
				MinLabel: "Schlecht",
				MaxLabel: "Ausgezeichnet",
			},
			{
				ID:          "feedback",
				Type:        models.FieldTypeTextarea,
				Label:       "Bitte teilen Sie uns Ihr Feedback mit",
				Placeholder: "Was hat Ihnen gefallen oder nicht gefallen?",
				Required:    true,
				Order:       1,
			},
			{
				ID:       "improvements",
				Type:     models.FieldTypeTextarea,
				Label:    "Was könnten wir verbessern?",
				Required: false,
				Order:    2,
			},
			{
				ID:       "recommend",
				Type:     models.FieldTypeScale,
				Label:    "Wie wahrscheinlich würden Sie uns weiterempfehlen? (0-10)",
				Required: true,
				Order:    3,
				MinValue: 0,
				MaxValue: 10,
				MinLabel: "Unwahrscheinlich",
				MaxLabel: "Sehr wahrscheinlich",
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Feedback absenden",
			SuccessMessage:     "Vielen Dank für Ihr wertvolles Feedback!",
			AllowMultiple:      false,
			NotifyOnSubmission: true,
		},
	}
}

// getSurveyTemplate creates a survey template (German)
func getSurveyTemplate() models.Form {
	return models.Form{
		Title:            "Kunden-Umfrage",
		Description:      "Gewinnen Sie Einblicke von Ihren Kunden",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategorySurvey),
		Language:         "de",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:                 "section1",
				Type:               models.FieldTypeSection,
				Label:              "Über Sie",
				SectionTitle:       "Persönliche Informationen",
				SectionDescription: "Helfen Sie uns, Sie besser zu verstehen",
				Order:              0,
			},
			{
				ID:       "age_group",
				Type:     models.FieldTypeRadio,
				Label:    "Altersgruppe",
				Required: true,
				Order:    1,
				Options:  []string{"18-24", "25-34", "35-44", "45-54", "55+"},
			},
			{
				ID:       "occupation",
				Type:     models.FieldTypeSelect,
				Label:    "Beruf",
				Required: false,
				Order:    2,
				Options:  []string{"Student/in", "Angestellt", "Selbstständig", "Arbeitslos", "Rentner/in"},
			},
			{
				ID:                 "section2",
				Type:               models.FieldTypeSection,
				Label:              "Produkt-Feedback",
				SectionTitle:       "Ihre Erfahrung",
				SectionDescription: "Teilen Sie uns Ihre Meinung mit",
				Order:              3,
			},
			{
				ID:       "usage_frequency",
				Type:     models.FieldTypeRadio,
				Label:    "Wie oft nutzen Sie unser Produkt?",
				Required: true,
				Order:    4,
				Options:  []string{"Täglich", "Wöchentlich", "Monatlich", "Selten"},
			},
			{
				ID:       "satisfaction",
				Type:     models.FieldTypeRating,
				Label:    "Gesamtzufriedenheit",
				Required: true,
				Order:    5,
				MinValue: 1,
				MaxValue: 5,
			},
			{
				ID:          "suggestions",
				Type:        models.FieldTypeTextarea,
				Label:       "Weitere Kommentare oder Vorschläge",
				Placeholder: "Teilen Sie Ihre Gedanken...",
				Required:    false,
				Order:       6,
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Umfrage absenden",
			SuccessMessage:     "Vielen Dank für die Teilnahme an unserer Umfrage!",
			AllowMultiple:      false,
			NotifyOnSubmission: true,
		},
	}
}

// getEventRegistrationTemplate creates an event registration template (German)
func getEventRegistrationTemplate() models.Form {
	return models.Form{
		Title:            "Event-Registrierung",
		Description:      "Registrieren Sie Teilnehmer für Ihre Veranstaltung",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategoryRegistration),
		Language:         "de",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "full_name",
				Type:        models.FieldTypeText,
				Label:       "Vollständiger Name",
				Placeholder: "Max Mustermann",
				Required:    true,
				Order:       0,
			},
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "E-Mail-Adresse",
				Placeholder: "max@beispiel.de",
				Required:    true,
				Order:       1,
			},
			{
				ID:          "phone",
				Type:        models.FieldTypePhone,
				Label:       "Telefonnummer",
				Placeholder: "+49 123 456789",
				Required:    false,
				Order:       2,
			},
			{
				ID:          "attendees",
				Type:        models.FieldTypeNumber,
				Label:       "Anzahl der Teilnehmer",
				Placeholder: "1",
				Required:    true,
				Order:       3,
				Validation: map[string]interface{}{
					"min": 1,
					"max": 10,
				},
			},
			{
				ID:       "dietary",
				Type:     models.FieldTypeCheckbox,
				Label:    "Ernährungspräferenzen (falls vorhanden)",
				Required: false,
				Order:    4,
				Options:  []string{"Vegetarisch", "Vegan", "Glutenfrei", "Halal", "Koscher", "Keine Präferenzen"},
			},
			{
				ID:          "special_requirements",
				Type:        models.FieldTypeTextarea,
				Label:       "Besondere Anforderungen",
				Placeholder: "Gibt es besondere Bedürfnisse oder andere Anforderungen?",
				Required:    false,
				Order:       5,
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Jetzt registrieren",
			SuccessMessage:     "Registrierung erfolgreich! Sie erhalten in Kürze eine Bestätigungs-E-Mail.",
			AllowMultiple:      false,
			NotifyOnSubmission: true,
		},
	}
}

// getJobApplicationTemplate creates a job application template (German)
func getJobApplicationTemplate() models.Form {
	return models.Form{
		Title:            "Bewerbungsformular",
		Description:      "Nehmen Sie Bewerbungen für offene Stellen entgegen",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategoryApplication),
		Language:         "de",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "full_name",
				Type:        models.FieldTypeText,
				Label:       "Vollständiger Name",
				Placeholder: "Max Mustermann",
				Required:    true,
				Order:       0,
			},
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "E-Mail-Adresse",
				Placeholder: "max@beispiel.de",
				Required:    true,
				Order:       1,
			},
			{
				ID:          "phone",
				Type:        models.FieldTypePhone,
				Label:       "Telefonnummer",
				Placeholder: "+49 123 456789",
				Required:    true,
				Order:       2,
			},
			{
				ID:       "position",
				Type:     models.FieldTypeSelect,
				Label:    "Position, für die Sie sich bewerben",
				Required: true,
				Order:    3,
				Options:  []string{"Software-Entwickler", "Produktmanager", "Designer", "Marketing-Spezialist", "Vertriebsmitarbeiter"},
			},
			{
				ID:       "experience",
				Type:     models.FieldTypeRadio,
				Label:    "Jahre Berufserfahrung",
				Required: true,
				Order:    4,
				Options:  []string{"0-2 Jahre", "3-5 Jahre", "6-10 Jahre", "10+ Jahre"},
			},
			{
				ID:           "resume",
				Type:         models.FieldTypeFile,
				Label:        "Lebenslauf",
				Description:  "Laden Sie Ihren Lebenslauf hoch (PDF, DOC, DOCX)",
				Required:     true,
				Order:        5,
				AllowedTypes: []string{".pdf", ".doc", ".docx"},
				MaxFileSize:  5,
			},
			{
				ID:          "cover_letter",
				Type:        models.FieldTypeTextarea,
				Label:       "Anschreiben",
				Placeholder: "Erzählen Sie uns, warum Sie perfekt zu uns passen...",
				Required:    false,
				Order:       6,
			},
			{
				ID:          "linkedin",
				Type:        models.FieldTypeURL,
				Label:       "LinkedIn-Profil (optional)",
				Placeholder: "https://linkedin.com/in/ihrprofil",
				Required:    false,
				Order:       7,
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Bewerbung absenden",
			SuccessMessage:     "Ihre Bewerbung wurde erfolgreich eingereicht! Wir werden sie prüfen und uns bald bei Ihnen melden.",
			AllowMultiple:      false,
			NotifyOnSubmission: true,
		},
	}
}

// getNewsletterSignupTemplate creates a newsletter signup template (German)
func getNewsletterSignupTemplate() models.Form {
	return models.Form{
		Title:            "Newsletter-Anmeldung",
		Description:      "Erweitern Sie Ihre E-Mail-Liste mit diesem einfachen Anmeldeformular",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategoryNewsletter),
		Language:         "de",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "E-Mail-Adresse",
				Placeholder: "ihre@email.de",
				Required:    true,
				Order:       0,
			},
			{
				ID:          "first_name",
				Type:        models.FieldTypeText,
				Label:       "Vorname",
				Placeholder: "Max",
				Required:    false,
				Order:       1,
			},
			{
				ID:       "interests",
				Type:     models.FieldTypeCheckbox,
				Label:    "Ich interessiere mich für:",
				Required: false,
				Order:    2,
				Options:  []string{"Produkt-Updates", "Unternehmensneuigkeiten", "Branchen-Insights", "Spezielle Angebote", "Veranstaltungen"},
			},
			{
				ID:       "frequency",
				Type:     models.FieldTypeRadio,
				Label:    "Wie oft möchten Sie von uns hören?",
				Required: false,
				Order:    3,
				Options:  []string{"Täglich", "Wöchentlich", "Monatlich"},
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Abonnieren",
			SuccessMessage:     "Willkommen! Sie haben sich erfolgreich für unseren Newsletter angemeldet.",
			AllowMultiple:      false,
			NotifyOnSubmission: true,
		},
	}
}

// ============================================================================
// ENGLISH TEMPLATES
// ============================================================================

// getContactFormTemplateEN creates a contact form template (English)
func getContactFormTemplateEN() models.Form {
	return models.Form{
		Title:            "Contact Form",
		Description:      "A simple contact form for your website visitors",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategoryContact),
		Language:         "en",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "name",
				Type:        models.FieldTypeText,
				Label:       "Your Name",
				Placeholder: "John Doe",
				Required:    true,
				Order:       0,
			},
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "Email Address",
				Placeholder: "john@example.com",
				Required:    true,
				Order:       1,
			},
			{
				ID:          "subject",
				Type:        models.FieldTypeText,
				Label:       "Subject",
				Placeholder: "What is this about?",
				Required:    true,
				Order:       2,
			},
			{
				ID:          "message",
				Type:        models.FieldTypeTextarea,
				Label:       "Message",
				Placeholder: "Your message here...",
				Required:    true,
				Order:       3,
				Validation: map[string]interface{}{
					"minLength": 10,
				},
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Send Message",
			SuccessMessage:     "Thank you for your message! We'll get back to you soon.",
			AllowMultiple:      true,
			NotifyOnSubmission: true,
		},
	}
}

// getFeedbackFormTemplateEN creates a feedback form template (English)
func getFeedbackFormTemplateEN() models.Form {
	return models.Form{
		Title:            "Feedback Form",
		Description:      "Collect feedback from your users",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategoryFeedback),
		Language:         "en",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:       "rating",
				Type:     models.FieldTypeRating,
				Label:    "How would you rate your experience?",
				Required: true,
				Order:    0,
				MinValue: 1,
				MaxValue: 5,
				MinLabel: "Poor",
				MaxLabel: "Excellent",
			},
			{
				ID:          "feedback",
				Type:        models.FieldTypeTextarea,
				Label:       "Please share your feedback",
				Placeholder: "What did you like or dislike?",
				Required:    true,
				Order:       1,
			},
			{
				ID:       "improvements",
				Type:     models.FieldTypeTextarea,
				Label:    "What could we improve?",
				Required: false,
				Order:    2,
			},
			{
				ID:       "recommend",
				Type:     models.FieldTypeScale,
				Label:    "How likely are you to recommend us? (0-10)",
				Required: true,
				Order:    3,
				MinValue: 0,
				MaxValue: 10,
				MinLabel: "Not likely",
				MaxLabel: "Very likely",
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Submit Feedback",
			SuccessMessage:     "Thank you for your valuable feedback!",
			AllowMultiple:      false,
			NotifyOnSubmission: true,
		},
	}
}

// getSurveyTemplateEN creates a survey template (English)
func getSurveyTemplateEN() models.Form {
	return models.Form{
		Title:            "Customer Survey",
		Description:      "Gather insights from your customers",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategorySurvey),
		Language:         "en",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:                 "section1",
				Type:               models.FieldTypeSection,
				Label:              "About You",
				SectionTitle:       "Personal Information",
				SectionDescription: "Help us understand who you are",
				Order:              0,
			},
			{
				ID:       "age_group",
				Type:     models.FieldTypeRadio,
				Label:    "Age Group",
				Required: true,
				Order:    1,
				Options:  []string{"18-24", "25-34", "35-44", "45-54", "55+"},
			},
			{
				ID:       "occupation",
				Type:     models.FieldTypeSelect,
				Label:    "Occupation",
				Required: false,
				Order:    2,
				Options:  []string{"Student", "Employed", "Self-employed", "Unemployed", "Retired"},
			},
			{
				ID:                 "section2",
				Type:               models.FieldTypeSection,
				Label:              "Product Feedback",
				SectionTitle:       "Your Experience",
				SectionDescription: "Share your thoughts with us",
				Order:              3,
			},
			{
				ID:       "usage_frequency",
				Type:     models.FieldTypeRadio,
				Label:    "How often do you use our product?",
				Required: true,
				Order:    4,
				Options:  []string{"Daily", "Weekly", "Monthly", "Rarely"},
			},
			{
				ID:       "satisfaction",
				Type:     models.FieldTypeRating,
				Label:    "Overall satisfaction",
				Required: true,
				Order:    5,
				MinValue: 1,
				MaxValue: 5,
			},
			{
				ID:          "suggestions",
				Type:        models.FieldTypeTextarea,
				Label:       "Additional comments or suggestions",
				Placeholder: "Share your thoughts...",
				Required:    false,
				Order:       6,
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Submit Survey",
			SuccessMessage:     "Thank you for completing our survey!",
			AllowMultiple:      false,
			NotifyOnSubmission: true,
		},
	}
}

// getEventRegistrationTemplateEN creates an event registration template (English)
func getEventRegistrationTemplateEN() models.Form {
	return models.Form{
		Title:            "Event Registration",
		Description:      "Register attendees for your event",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategoryRegistration),
		Language:         "en",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "full_name",
				Type:        models.FieldTypeText,
				Label:       "Full Name",
				Placeholder: "John Doe",
				Required:    true,
				Order:       0,
			},
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "Email Address",
				Placeholder: "john@example.com",
				Required:    true,
				Order:       1,
			},
			{
				ID:          "phone",
				Type:        models.FieldTypePhone,
				Label:       "Phone Number",
				Placeholder: "+1 (555) 123-4567",
				Required:    false,
				Order:       2,
			},
			{
				ID:          "attendees",
				Type:        models.FieldTypeNumber,
				Label:       "Number of Attendees",
				Placeholder: "1",
				Required:    true,
				Order:       3,
				Validation: map[string]interface{}{
					"min": 1,
					"max": 10,
				},
			},
			{
				ID:       "dietary",
				Type:     models.FieldTypeCheckbox,
				Label:    "Dietary Preferences (if any)",
				Required: false,
				Order:    4,
				Options:  []string{"Vegetarian", "Vegan", "Gluten-free", "Halal", "Kosher", "No preferences"},
			},
			{
				ID:          "special_requirements",
				Type:        models.FieldTypeTextarea,
				Label:       "Special Requirements",
				Placeholder: "Any accessibility needs or other requirements?",
				Required:    false,
				Order:       5,
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Register Now",
			SuccessMessage:     "Registration successful! You'll receive a confirmation email shortly.",
			AllowMultiple:      false,
			NotifyOnSubmission: true,
		},
	}
}

// getJobApplicationTemplateEN creates a job application template (English)
func getJobApplicationTemplateEN() models.Form {
	return models.Form{
		Title:            "Job Application Form",
		Description:      "Accept applications for job positions",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategoryApplication),
		Language:         "en",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "full_name",
				Type:        models.FieldTypeText,
				Label:       "Full Name",
				Placeholder: "John Doe",
				Required:    true,
				Order:       0,
			},
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "Email Address",
				Placeholder: "john@example.com",
				Required:    true,
				Order:       1,
			},
			{
				ID:          "phone",
				Type:        models.FieldTypePhone,
				Label:       "Phone Number",
				Placeholder: "+1 (555) 123-4567",
				Required:    true,
				Order:       2,
			},
			{
				ID:       "position",
				Type:     models.FieldTypeSelect,
				Label:    "Position Applied For",
				Required: true,
				Order:    3,
				Options:  []string{"Software Engineer", "Product Manager", "Designer", "Marketing Specialist", "Sales Representative"},
			},
			{
				ID:       "experience",
				Type:     models.FieldTypeRadio,
				Label:    "Years of Experience",
				Required: true,
				Order:    4,
				Options:  []string{"0-2 years", "3-5 years", "6-10 years", "10+ years"},
			},
			{
				ID:           "resume",
				Type:         models.FieldTypeFile,
				Label:        "Resume/CV",
				Description:  "Upload your resume (PDF, DOC, DOCX)",
				Required:     true,
				Order:        5,
				AllowedTypes: []string{".pdf", ".doc", ".docx"},
				MaxFileSize:  5,
			},
			{
				ID:          "cover_letter",
				Type:        models.FieldTypeTextarea,
				Label:       "Cover Letter",
				Placeholder: "Tell us why you're a great fit...",
				Required:    false,
				Order:       6,
			},
			{
				ID:          "linkedin",
				Type:        models.FieldTypeURL,
				Label:       "LinkedIn Profile (optional)",
				Placeholder: "https://linkedin.com/in/yourprofile",
				Required:    false,
				Order:       7,
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Submit Application",
			SuccessMessage:     "Your application has been submitted successfully! We'll review it and get back to you soon.",
			AllowMultiple:      false,
			NotifyOnSubmission: true,
		},
	}
}

// getNewsletterSignupTemplateEN creates a newsletter signup template (English)
func getNewsletterSignupTemplateEN() models.Form {
	return models.Form{
		Title:            "Newsletter Signup",
		Description:      "Grow your email list with this simple signup form",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategoryNewsletter),
		Language:         "en",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "Email Address",
				Placeholder: "your@email.com",
				Required:    true,
				Order:       0,
			},
			{
				ID:          "first_name",
				Type:        models.FieldTypeText,
				Label:       "First Name",
				Placeholder: "John",
				Required:    false,
				Order:       1,
			},
			{
				ID:       "interests",
				Type:     models.FieldTypeCheckbox,
				Label:    "I'm interested in:",
				Required: false,
				Order:    2,
				Options:  []string{"Product Updates", "Company News", "Industry Insights", "Special Offers", "Events"},
			},
			{
				ID:       "frequency",
				Type:     models.FieldTypeRadio,
				Label:    "How often would you like to hear from us?",
				Required: false,
				Order:    3,
				Options:  []string{"Daily", "Weekly", "Monthly"},
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Subscribe",
			SuccessMessage:     "Welcome! You've successfully subscribed to our newsletter.",
			AllowMultiple:      false,
			NotifyOnSubmission: true,
		},
	}
}

// ==================== Quote/Anfrage Templates ====================

func getQuoteRequestTemplate() models.Form {
	return models.Form{
		Title:            "Angebotsanfrage",
		Description:      "Sammeln Sie Anfragen für Angebote und Kostenvoranschläge",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategoryQuote),
		Language:         "de",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "company_name",
				Type:        models.FieldTypeText,
				Label:       "Firmenname",
				Placeholder: "Ihre Firma GmbH",
				Required:    true,
				Order:       0,
			},
			{
				ID:          "contact_person",
				Type:        models.FieldTypeText,
				Label:       "Ansprechpartner",
				Placeholder: "Max Mustermann",
				Required:    true,
				Order:       1,
			},
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "E-Mail",
				Placeholder: "max@firma.de",
				Required:    true,
				Order:       2,
			},
			{
				ID:          "phone",
				Type:        models.FieldTypePhone,
				Label:       "Telefon",
				Placeholder: "+49 123 456789",
				Required:    false,
				Order:       3,
			},
			{
				ID:       "service_type",
				Type:     models.FieldTypeSelect,
				Label:    "Art der Dienstleistung",
				Required: true,
				Order:    4,
				Options:  []string{"Beratung", "Entwicklung", "Design", "Support", "Sonstiges"},
			},
			{
				ID:          "project_description",
				Type:        models.FieldTypeTextarea,
				Label:       "Projektbeschreibung",
				Placeholder: "Beschreiben Sie Ihr Projekt...",
				Required:    true,
				Order:       5,
			},
			{
				ID:       "budget",
				Type:     models.FieldTypeSelect,
				Label:    "Geschätztes Budget",
				Required: false,
				Order:    6,
				Options:  []string{"< 5.000 €", "5.000 - 10.000 €", "10.000 - 25.000 €", "25.000 - 50.000 €", "> 50.000 €"},
			},
			{
				ID:       "timeline",
				Type:     models.FieldTypeSelect,
				Label:    "Gewünschter Zeitrahmen",
				Required: false,
				Order:    7,
				Options:  []string{"Dringend (< 1 Monat)", "Kurzfristig (1-3 Monate)", "Mittelfristig (3-6 Monate)", "Langfristig (> 6 Monate)"},
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Anfrage senden",
			SuccessMessage:     "Vielen Dank! Wir werden uns zeitnah bei Ihnen melden.",
			AllowMultiple:      false,
			NotifyOnSubmission: true,
		},
	}
}

func getQuoteRequestTemplateEN() models.Form {
	return models.Form{
		Title:            "Quote Request",
		Description:      "Collect requests for quotes and cost estimates",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategoryQuote),
		Language:         "en",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "company_name",
				Type:        models.FieldTypeText,
				Label:       "Company Name",
				Placeholder: "Your Company Ltd",
				Required:    true,
				Order:       0,
			},
			{
				ID:          "contact_person",
				Type:        models.FieldTypeText,
				Label:       "Contact Person",
				Placeholder: "John Doe",
				Required:    true,
				Order:       1,
			},
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "Email",
				Placeholder: "john@company.com",
				Required:    true,
				Order:       2,
			},
			{
				ID:          "phone",
				Type:        models.FieldTypePhone,
				Label:       "Phone",
				Placeholder: "+1 234 567 8900",
				Required:    false,
				Order:       3,
			},
			{
				ID:       "service_type",
				Type:     models.FieldTypeSelect,
				Label:    "Type of Service",
				Required: true,
				Order:    4,
				Options:  []string{"Consulting", "Development", "Design", "Support", "Other"},
			},
			{
				ID:          "project_description",
				Type:        models.FieldTypeTextarea,
				Label:       "Project Description",
				Placeholder: "Describe your project...",
				Required:    true,
				Order:       5,
			},
			{
				ID:       "budget",
				Type:     models.FieldTypeSelect,
				Label:    "Estimated Budget",
				Required: false,
				Order:    6,
				Options:  []string{"< $5,000", "$5,000 - $10,000", "$10,000 - $25,000", "$25,000 - $50,000", "> $50,000"},
			},
			{
				ID:       "timeline",
				Type:     models.FieldTypeSelect,
				Label:    "Desired Timeline",
				Required: false,
				Order:    7,
				Options:  []string{"Urgent (< 1 month)", "Short-term (1-3 months)", "Medium-term (3-6 months)", "Long-term (> 6 months)"},
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Send Request",
			SuccessMessage:     "Thank you! We'll get back to you shortly.",
			AllowMultiple:      false,
			NotifyOnSubmission: true,
		},
	}
}

// ==================== Support Templates ====================

func getSupportTicketTemplate() models.Form {
	return models.Form{
		Title:            "Support-Anfrage",
		Description:      "Kunden-Support-Formular für technische Probleme und Fragen",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategorySupport),
		Language:         "de",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "name",
				Type:        models.FieldTypeText,
				Label:       "Ihr Name",
				Placeholder: "Max Mustermann",
				Required:    true,
				Order:       0,
			},
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "E-Mail",
				Placeholder: "ihre@email.de",
				Required:    true,
				Order:       1,
			},
			{
				ID:       "priority",
				Type:     models.FieldTypeRadio,
				Label:    "Priorität",
				Required: true,
				Order:    2,
				Options:  []string{"Niedrig", "Mittel", "Hoch", "Kritisch"},
			},
			{
				ID:       "category",
				Type:     models.FieldTypeSelect,
				Label:    "Kategorie",
				Required: true,
				Order:    3,
				Options:  []string{"Technisches Problem", "Frage zur Nutzung", "Rechnungsfrage", "Feature-Anfrage", "Sonstiges"},
			},
			{
				ID:          "subject",
				Type:        models.FieldTypeText,
				Label:       "Betreff",
				Placeholder: "Kurze Beschreibung des Problems",
				Required:    true,
				Order:       4,
			},
			{
				ID:          "description",
				Type:        models.FieldTypeTextarea,
				Label:       "Detaillierte Beschreibung",
				Placeholder: "Bitte beschreiben Sie Ihr Anliegen so detailliert wie möglich...",
				Required:    true,
				Order:       5,
			},
			{
				ID:       "attachment",
				Type:     models.FieldTypeFile,
				Label:    "Anhang (optional)",
				Required: false,
				Order:    6,
				AllowedTypes: []string{".pdf", ".jpg", ".png", ".doc", ".docx"},
				MaxFileSize:  10,
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Ticket erstellen",
			SuccessMessage:     "Ihre Support-Anfrage wurde erfasst. Wir melden uns schnellstmöglich bei Ihnen.",
			AllowMultiple:      true,
			NotifyOnSubmission: true,
		},
	}
}

func getSupportTicketTemplateEN() models.Form {
	return models.Form{
		Title:            "Support Request",
		Description:      "Customer support form for technical issues and questions",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategorySupport),
		Language:         "en",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "name",
				Type:        models.FieldTypeText,
				Label:       "Your Name",
				Placeholder: "John Doe",
				Required:    true,
				Order:       0,
			},
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "Email",
				Placeholder: "your@email.com",
				Required:    true,
				Order:       1,
			},
			{
				ID:       "priority",
				Type:     models.FieldTypeRadio,
				Label:    "Priority",
				Required: true,
				Order:    2,
				Options:  []string{"Low", "Medium", "High", "Critical"},
			},
			{
				ID:       "category",
				Type:     models.FieldTypeSelect,
				Label:    "Category",
				Required: true,
				Order:    3,
				Options:  []string{"Technical Issue", "Usage Question", "Billing Question", "Feature Request", "Other"},
			},
			{
				ID:          "subject",
				Type:        models.FieldTypeText,
				Label:       "Subject",
				Placeholder: "Brief description of the issue",
				Required:    true,
				Order:       4,
			},
			{
				ID:          "description",
				Type:        models.FieldTypeTextarea,
				Label:       "Detailed Description",
				Placeholder: "Please describe your issue in as much detail as possible...",
				Required:    true,
				Order:       5,
			},
			{
				ID:       "attachment",
				Type:     models.FieldTypeFile,
				Label:    "Attachment (optional)",
				Required: false,
				Order:    6,
				AllowedTypes: []string{".pdf", ".jpg", ".png", ".doc", ".docx"},
				MaxFileSize:  10,
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Submit Ticket",
			SuccessMessage:     "Your support request has been received. We'll get back to you as soon as possible.",
			AllowMultiple:      true,
			NotifyOnSubmission: true,
		},
	}
}

// ==================== Booking Templates ====================

func getBookingTemplate() models.Form {
	return models.Form{
		Title:            "Termin-Buchung",
		Description:      "Einfaches Buchungsformular für Termine und Reservierungen",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategoryBooking),
		Language:         "de",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "name",
				Type:        models.FieldTypeText,
				Label:       "Vollständiger Name",
				Placeholder: "Max Mustermann",
				Required:    true,
				Order:       0,
			},
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "E-Mail",
				Placeholder: "ihre@email.de",
				Required:    true,
				Order:       1,
			},
			{
				ID:          "phone",
				Type:        models.FieldTypePhone,
				Label:       "Telefonnummer",
				Placeholder: "+49 123 456789",
				Required:    true,
				Order:       2,
			},
			{
				ID:       "service",
				Type:     models.FieldTypeSelect,
				Label:    "Gewünschte Dienstleistung",
				Required: true,
				Order:    3,
				Options:  []string{"Beratungsgespräch", "Behandlung", "Workshop", "Training", "Sonstiges"},
			},
			{
				ID:       "preferred_date",
				Type:     models.FieldTypeDate,
				Label:    "Wunschtermin",
				Required: true,
				Order:    4,
			},
			{
				ID:       "preferred_time",
				Type:     models.FieldTypeRadio,
				Label:    "Bevorzugte Uhrzeit",
				Required: true,
				Order:    5,
				Options:  []string{"09:00 - 12:00", "12:00 - 15:00", "15:00 - 18:00", "Flexible"},
			},
			{
				ID:       "participants",
				Type:     models.FieldTypeNumber,
				Label:    "Anzahl Teilnehmer",
				Required: true,
				Order:    6,
				Validation: map[string]interface{}{
					"min": 1,
					"max": 50,
				},
			},
			{
				ID:          "notes",
				Type:        models.FieldTypeTextarea,
				Label:       "Zusätzliche Anmerkungen",
				Placeholder: "Besondere Wünsche oder Anforderungen...",
				Required:    false,
				Order:       7,
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Termin anfragen",
			SuccessMessage:     "Ihre Buchungsanfrage wurde erfasst. Wir werden den Termin bestätigen.",
			AllowMultiple:      false,
			NotifyOnSubmission: true,
		},
	}
}

func getBookingTemplateEN() models.Form {
	return models.Form{
		Title:            "Appointment Booking",
		Description:      "Simple booking form for appointments and reservations",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategoryBooking),
		Language:         "en",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "name",
				Type:        models.FieldTypeText,
				Label:       "Full Name",
				Placeholder: "John Doe",
				Required:    true,
				Order:       0,
			},
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "Email",
				Placeholder: "your@email.com",
				Required:    true,
				Order:       1,
			},
			{
				ID:          "phone",
				Type:        models.FieldTypePhone,
				Label:       "Phone Number",
				Placeholder: "+1 234 567 8900",
				Required:    true,
				Order:       2,
			},
			{
				ID:       "service",
				Type:     models.FieldTypeSelect,
				Label:    "Desired Service",
				Required: true,
				Order:    3,
				Options:  []string{"Consultation", "Treatment", "Workshop", "Training", "Other"},
			},
			{
				ID:       "preferred_date",
				Type:     models.FieldTypeDate,
				Label:    "Preferred Date",
				Required: true,
				Order:    4,
			},
			{
				ID:       "preferred_time",
				Type:     models.FieldTypeRadio,
				Label:    "Preferred Time",
				Required: true,
				Order:    5,
				Options:  []string{"09:00 - 12:00", "12:00 - 15:00", "15:00 - 18:00", "Flexible"},
			},
			{
				ID:       "participants",
				Type:     models.FieldTypeNumber,
				Label:    "Number of Participants",
				Required: true,
				Order:    6,
				Validation: map[string]interface{}{
					"min": 1,
					"max": 50,
				},
			},
			{
				ID:          "notes",
				Type:        models.FieldTypeTextarea,
				Label:       "Additional Notes",
				Placeholder: "Special requests or requirements...",
				Required:    false,
				Order:       7,
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Request Appointment",
			SuccessMessage:     "Your booking request has been received. We'll confirm your appointment shortly.",
			AllowMultiple:      false,
			NotifyOnSubmission: true,
		},
	}
}

// ==================== Order Templates ====================

func getOrderFormTemplate() models.Form {
	return models.Form{
		Title:            "Bestellformular",
		Description:      "Produktbestellungen und Anfragen erfassen",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategoryOrder),
		Language:         "de",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "customer_name",
				Type:        models.FieldTypeText,
				Label:       "Name",
				Placeholder: "Max Mustermann",
				Required:    true,
				Order:       0,
			},
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "E-Mail",
				Placeholder: "ihre@email.de",
				Required:    true,
				Order:       1,
			},
			{
				ID:          "phone",
				Type:        models.FieldTypePhone,
				Label:       "Telefon",
				Placeholder: "+49 123 456789",
				Required:    true,
				Order:       2,
			},
			{
				ID:       "product",
				Type:     models.FieldTypeSelect,
				Label:    "Produkt",
				Required: true,
				Order:    3,
				Options:  []string{"Produkt A", "Produkt B", "Produkt C", "Produkt D", "Sonstiges"},
			},
			{
				ID:       "quantity",
				Type:     models.FieldTypeNumber,
				Label:    "Menge",
				Required: true,
				Order:    4,
				Validation: map[string]interface{}{
					"min": 1,
					"max": 1000,
				},
			},
			{
				ID:          "delivery_address",
				Type:        models.FieldTypeTextarea,
				Label:       "Lieferadresse",
				Placeholder: "Straße, PLZ, Stadt",
				Required:    true,
				Order:       5,
			},
			{
				ID:       "payment_method",
				Type:     models.FieldTypeRadio,
				Label:    "Zahlungsmethode",
				Required: true,
				Order:    6,
				Options:  []string{"Rechnung", "Vorkasse", "PayPal", "Kreditkarte"},
			},
			{
				ID:          "notes",
				Type:        models.FieldTypeTextarea,
				Label:       "Anmerkungen",
				Placeholder: "Besondere Wünsche zur Bestellung...",
				Required:    false,
				Order:       7,
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Bestellung absenden",
			SuccessMessage:     "Ihre Bestellung wurde erfasst. Sie erhalten eine Bestätigung per E-Mail.",
			AllowMultiple:      false,
			NotifyOnSubmission: true,
		},
	}
}

func getOrderFormTemplateEN() models.Form {
	return models.Form{
		Title:            "Order Form",
		Description:      "Collect product orders and requests",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategoryOrder),
		Language:         "en",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "customer_name",
				Type:        models.FieldTypeText,
				Label:       "Name",
				Placeholder: "John Doe",
				Required:    true,
				Order:       0,
			},
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "Email",
				Placeholder: "your@email.com",
				Required:    true,
				Order:       1,
			},
			{
				ID:          "phone",
				Type:        models.FieldTypePhone,
				Label:       "Phone",
				Placeholder: "+1 234 567 8900",
				Required:    true,
				Order:       2,
			},
			{
				ID:       "product",
				Type:     models.FieldTypeSelect,
				Label:    "Product",
				Required: true,
				Order:    3,
				Options:  []string{"Product A", "Product B", "Product C", "Product D", "Other"},
			},
			{
				ID:       "quantity",
				Type:     models.FieldTypeNumber,
				Label:    "Quantity",
				Required: true,
				Order:    4,
				Validation: map[string]interface{}{
					"min": 1,
					"max": 1000,
				},
			},
			{
				ID:          "delivery_address",
				Type:        models.FieldTypeTextarea,
				Label:       "Delivery Address",
				Placeholder: "Street, ZIP, City",
				Required:    true,
				Order:       5,
			},
			{
				ID:       "payment_method",
				Type:     models.FieldTypeRadio,
				Label:    "Payment Method",
				Required: true,
				Order:    6,
				Options:  []string{"Invoice", "Prepayment", "PayPal", "Credit Card"},
			},
			{
				ID:          "notes",
				Type:        models.FieldTypeTextarea,
				Label:       "Notes",
				Placeholder: "Special requests for your order...",
				Required:    false,
				Order:       7,
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Submit Order",
			SuccessMessage:     "Your order has been received. You'll receive a confirmation email shortly.",
			AllowMultiple:      false,
			NotifyOnSubmission: true,
		},
	}
}

// ==================== Complaint Templates ====================

func getComplaintFormTemplate() models.Form {
	return models.Form{
		Title:            "Beschwerde-Formular",
		Description:      "Kundenbeschwerden strukturiert erfassen und bearbeiten",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategoryComplaint),
		Language:         "de",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "name",
				Type:        models.FieldTypeText,
				Label:       "Ihr Name",
				Placeholder: "Max Mustermann",
				Required:    true,
				Order:       0,
			},
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "E-Mail",
				Placeholder: "ihre@email.de",
				Required:    true,
				Order:       1,
			},
			{
				ID:          "phone",
				Type:        models.FieldTypePhone,
				Label:       "Telefon (optional)",
				Placeholder: "+49 123 456789",
				Required:    false,
				Order:       2,
			},
			{
				ID:          "order_number",
				Type:        models.FieldTypeText,
				Label:       "Bestellnummer (falls vorhanden)",
				Placeholder: "ORD-12345",
				Required:    false,
				Order:       3,
			},
			{
				ID:       "complaint_type",
				Type:     models.FieldTypeSelect,
				Label:    "Art der Beschwerde",
				Required: true,
				Order:    4,
				Options:  []string{"Produktqualität", "Lieferprobleme", "Kundenservice", "Rechnung", "Sonstiges"},
			},
			{
				ID:       "severity",
				Type:     models.FieldTypeRadio,
				Label:    "Schweregrad",
				Required: true,
				Order:    5,
				Options:  []string{"Geringfügig", "Mittel", "Schwerwiegend"},
			},
			{
				ID:          "description",
				Type:        models.FieldTypeTextarea,
				Label:       "Beschreibung der Beschwerde",
				Placeholder: "Bitte beschreiben Sie Ihr Anliegen detailliert...",
				Required:    true,
				Order:       6,
			},
			{
				ID:       "date_of_incident",
				Type:     models.FieldTypeDate,
				Label:    "Datum des Vorfalls",
				Required: false,
				Order:    7,
			},
			{
				ID:          "desired_resolution",
				Type:        models.FieldTypeTextarea,
				Label:       "Gewünschte Lösung",
				Placeholder: "Wie können wir das Problem beheben?",
				Required:    false,
				Order:       8,
			},
			{
				ID:       "attachment",
				Type:     models.FieldTypeFile,
				Label:    "Anhang (z.B. Fotos, Belege)",
				Required: false,
				Order:    9,
				AllowedTypes: []string{".pdf", ".jpg", ".png", ".doc", ".docx"},
				MaxFileSize:  10,
				Multiple:     true,
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Beschwerde einreichen",
			SuccessMessage:     "Ihre Beschwerde wurde erfasst. Wir werden uns umgehend darum kümmern.",
			AllowMultiple:      false,
			NotifyOnSubmission: true,
		},
	}
}

func getComplaintFormTemplateEN() models.Form {
	return models.Form{
		Title:            "Complaint Form",
		Description:      "Capture and process customer complaints systematically",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategoryComplaint),
		Language:         "en",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "name",
				Type:        models.FieldTypeText,
				Label:       "Your Name",
				Placeholder: "John Doe",
				Required:    true,
				Order:       0,
			},
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "Email",
				Placeholder: "your@email.com",
				Required:    true,
				Order:       1,
			},
			{
				ID:          "phone",
				Type:        models.FieldTypePhone,
				Label:       "Phone (optional)",
				Placeholder: "+1 234 567 8900",
				Required:    false,
				Order:       2,
			},
			{
				ID:          "order_number",
				Type:        models.FieldTypeText,
				Label:       "Order Number (if applicable)",
				Placeholder: "ORD-12345",
				Required:    false,
				Order:       3,
			},
			{
				ID:       "complaint_type",
				Type:     models.FieldTypeSelect,
				Label:    "Type of Complaint",
				Required: true,
				Order:    4,
				Options:  []string{"Product Quality", "Delivery Issues", "Customer Service", "Billing", "Other"},
			},
			{
				ID:       "severity",
				Type:     models.FieldTypeRadio,
				Label:    "Severity",
				Required: true,
				Order:    5,
				Options:  []string{"Minor", "Moderate", "Severe"},
			},
			{
				ID:          "description",
				Type:        models.FieldTypeTextarea,
				Label:       "Description of Complaint",
				Placeholder: "Please describe your issue in detail...",
				Required:    true,
				Order:       6,
			},
			{
				ID:       "date_of_incident",
				Type:     models.FieldTypeDate,
				Label:    "Date of Incident",
				Required: false,
				Order:    7,
			},
			{
				ID:          "desired_resolution",
				Type:        models.FieldTypeTextarea,
				Label:       "Desired Resolution",
				Placeholder: "How can we resolve this issue?",
				Required:    false,
				Order:       8,
			},
			{
				ID:       "attachment",
				Type:     models.FieldTypeFile,
				Label:    "Attachment (e.g. photos, receipts)",
				Required: false,
				Order:    9,
				AllowedTypes: []string{".pdf", ".jpg", ".png", ".doc", ".docx"},
				MaxFileSize:  10,
				Multiple:     true,
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Submit Complaint",
			SuccessMessage:     "Your complaint has been received. We'll address it promptly.",
			AllowMultiple:      false,
			NotifyOnSubmission: true,
		},
	}
}

// ==================== Suggestion Templates ====================

func getSuggestionFormTemplate() models.Form {
	return models.Form{
		Title:            "Verbesserungsvorschläge",
		Description:      "Sammeln Sie Ideen und Vorschläge von Kunden und Mitarbeitern",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategorySuggestion),
		Language:         "de",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "name",
				Type:        models.FieldTypeText,
				Label:       "Ihr Name (optional)",
				Placeholder: "Max Mustermann",
				Required:    false,
				Order:       0,
			},
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "E-Mail (optional)",
				Placeholder: "ihre@email.de",
				Required:    false,
				Order:       1,
			},
			{
				ID:       "category",
				Type:     models.FieldTypeSelect,
				Label:    "Kategorie",
				Required: true,
				Order:    2,
				Options:  []string{"Produkt", "Service", "Website", "Prozesse", "Sonstiges"},
			},
			{
				ID:          "suggestion_title",
				Type:        models.FieldTypeText,
				Label:       "Titel des Vorschlags",
				Placeholder: "Kurze Zusammenfassung",
				Required:    true,
				Order:       3,
			},
			{
				ID:          "suggestion_description",
				Type:        models.FieldTypeTextarea,
				Label:       "Detaillierte Beschreibung",
				Placeholder: "Beschreiben Sie Ihren Vorschlag ausführlich...",
				Required:    true,
				Order:       4,
			},
			{
				ID:          "current_situation",
				Type:        models.FieldTypeTextarea,
				Label:       "Aktuelle Situation",
				Placeholder: "Was möchten Sie verbessern?",
				Required:    false,
				Order:       5,
			},
			{
				ID:          "expected_benefit",
				Type:        models.FieldTypeTextarea,
				Label:       "Erwarteter Nutzen",
				Placeholder: "Welche Verbesserungen erwarten Sie?",
				Required:    false,
				Order:       6,
			},
			{
				ID:       "priority",
				Type:     models.FieldTypeRadio,
				Label:    "Priorität",
				Required: false,
				Order:    7,
				Options:  []string{"Niedrig", "Mittel", "Hoch"},
			},
			{
				ID:       "willing_to_collaborate",
				Type:     models.FieldTypeCheckbox,
				Label:    "Ich möchte bei der Umsetzung mithelfen",
				Required: false,
				Order:    8,
				Options:  []string{"Ja, ich möchte kontaktiert werden"},
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Vorschlag einreichen",
			SuccessMessage:     "Vielen Dank für Ihren Vorschlag! Wir prüfen alle Eingaben sorgfältig.",
			AllowMultiple:      true,
			NotifyOnSubmission: true,
		},
	}
}

func getSuggestionFormTemplateEN() models.Form {
	return models.Form{
		Title:            "Suggestion Form",
		Description:      "Collect ideas and suggestions from customers and employees",
		IsTemplate:       true,
		TemplateCategory: string(models.TemplateCategorySuggestion),
		Language:         "en",
		Status:           models.FormStatusDraft,
		Fields: models.FormFields{
			{
				ID:          "name",
				Type:        models.FieldTypeText,
				Label:       "Your Name (optional)",
				Placeholder: "John Doe",
				Required:    false,
				Order:       0,
			},
			{
				ID:          "email",
				Type:        models.FieldTypeEmail,
				Label:       "Email (optional)",
				Placeholder: "your@email.com",
				Required:    false,
				Order:       1,
			},
			{
				ID:       "category",
				Type:     models.FieldTypeSelect,
				Label:    "Category",
				Required: true,
				Order:    2,
				Options:  []string{"Product", "Service", "Website", "Processes", "Other"},
			},
			{
				ID:          "suggestion_title",
				Type:        models.FieldTypeText,
				Label:       "Suggestion Title",
				Placeholder: "Brief summary",
				Required:    true,
				Order:       3,
			},
			{
				ID:          "suggestion_description",
				Type:        models.FieldTypeTextarea,
				Label:       "Detailed Description",
				Placeholder: "Describe your suggestion in detail...",
				Required:    true,
				Order:       4,
			},
			{
				ID:          "current_situation",
				Type:        models.FieldTypeTextarea,
				Label:       "Current Situation",
				Placeholder: "What would you like to improve?",
				Required:    false,
				Order:       5,
			},
			{
				ID:          "expected_benefit",
				Type:        models.FieldTypeTextarea,
				Label:       "Expected Benefit",
				Placeholder: "What improvements do you expect?",
				Required:    false,
				Order:       6,
			},
			{
				ID:       "priority",
				Type:     models.FieldTypeRadio,
				Label:    "Priority",
				Required: false,
				Order:    7,
				Options:  []string{"Low", "Medium", "High"},
			},
			{
				ID:       "willing_to_collaborate",
				Type:     models.FieldTypeCheckbox,
				Label:    "I'd like to help with implementation",
				Required: false,
				Order:    8,
				Options:  []string{"Yes, I'd like to be contacted"},
			},
		},
		Settings: models.FormSettings{
			SubmitButtonText:   "Submit Suggestion",
			SuccessMessage:     "Thank you for your suggestion! We carefully review all submissions.",
			AllowMultiple:      true,
			NotifyOnSubmission: true,
		},
	}
}
