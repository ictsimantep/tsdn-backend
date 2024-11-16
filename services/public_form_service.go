// services/form_service.go
package services

import (
	"backend-school/config"
	"backend-school/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// PageService handles all page-related logic
type FormService struct {
	EmailService    *EmailService
	SettingsService *SettingsService
}

// NewPageService creates a new instance of PageService
func NewFormService(emailService *EmailService, settingsService *SettingsService) *FormService {
	return &FormService{
		EmailService:    emailService,
		SettingsService: settingsService,
	}
}

// GetFormBySlug retrieves a single page by its slug
func (s *FormService) GetFormBySlug(slug string) (*models.Form, error) {
	var form models.Form
	if err := config.DB.
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("step_order ASC")
		}).
		Preload("Steps.FormItems", func(db *gorm.DB) *gorm.DB {
			return db.Order("input_order ASC")
		}).
		Preload("Steps.FormItems.FormItemOptions").
		Where("slug = ?", slug).First(&form).Error; err != nil {
		return nil, err
	}
	return &form, nil
}

// SubmitForm
func (s *FormService) SubmitForm(form *models.Form, submittedData map[string]string) (*models.Form, error) {
	// Extract the email from submittedData. Ensure the form has an "email" field.
	email, ok := submittedData["email"]
	if !ok {
		return nil, fmt.Errorf("email field is required")
	}
	for fieldName, fieldValue := range submittedData {
		var submission models.FormSubmission

		// Check if a submission already exists for this FormID and FieldName
		err := config.DB.Where("form_id = ? AND field_name = ? AND email = ?", form.ID, fieldName, email).First(&submission).Error

		if err != nil {
			// If no record found, err will be 'record not found', so we insert a new one
			if err == gorm.ErrRecordNotFound {
				// Create a new submission
				newSubmission := models.FormSubmission{
					FormID:      form.ID,
					FieldName:   fieldName,
					FieldValue:  fieldValue,
					Email:       email,
					SubmittedAt: time.Now(),
				}

				// Save the new submission to the database
				if err := config.DB.Create(&newSubmission).Error; err != nil {
					return nil, err // Return error if saving the new submission failed
				}
			} else {
				// Return error if a different kind of error occurred
				return nil, err
			}
		} else {
			// If submission found, update the FieldValue and SubmittedAt fields
			submission.FieldValue = fieldValue
			submission.SubmittedAt = time.Now()

			// Save the updated submission to the database
			if err := config.DB.Save(&submission).Error; err != nil {
				return nil, err // Return error if saving the updated submission failed
			}
		}
	}

	subject := fmt.Sprintf("New form submission for %s", form.FormTitle)
	body := "A new form has been submitted with the following details:\n"
	for fieldName, fieldValue := range submittedData {
		body += fmt.Sprintf("%s: %s\n", fieldName, fieldValue)
	}

	// Fetch the admin email address from settings (or hard-code it here if preferred)
	adminEmail, err := s.SettingsService.GetSetting("admin_email")
	if err != nil {
		return nil, fmt.Errorf("could not retrieve admin email: %v", err)
	}

	// Send email to the admin
	if err := s.EmailService.SendEmail(adminEmail, subject, body); err != nil {
		return nil, fmt.Errorf("failed to send email notification: %v", err)
	}

	// Return the form after successful submission
	return form, nil
}
