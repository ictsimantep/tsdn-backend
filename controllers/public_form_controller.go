// controllers/Form_controller.go
package controllers

import (
	"backend-school/services"

	"github.com/gofiber/fiber/v2"
)

var (
	// Initialize settings service globally (assumes you have a DB connection or config ready)
	settingService = services.NewSettingsService()

	// Initialize email service with settings
	emailService = services.NewEmailService(settingService)

	// Initialize form service using email and settings services
	FormService = services.NewFormService(emailService, settingService)
)

// FormController handles HTTP requests for Forms
type FormController struct {
	FormService *services.FormService
}

// NewFormController creates a new instance of FormController with internally initialized FormService
func NewFormController() *FormController {
	return &FormController{FormService: FormService}
}

// GetFormBySlug returns a single Form by its slug by calling the FormService
func (c *FormController) GetFormBySlug(ctx *fiber.Ctx) error {
	slug := ctx.Params("slug")                     // Get slug from URL parameter
	Form, err := c.FormService.GetFormBySlug(slug) // Calling the service
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Form not found",
		})
	}
	return ctx.JSON(Form) // Return the Form data as JSON
}

func (c *FormController) SubmitFormHandler(ctx *fiber.Ctx) error {
	slug := ctx.Params("slug")

	Form, err := c.FormService.GetFormBySlug(slug) // Calling the service
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Form not found",
		})
	}

	// var form Form

	// Parse the submitted data
	var submittedData map[string]string
	if err := ctx.BodyParser(&submittedData); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid form data",
		})
	}

	c.FormService.SubmitForm(Form, submittedData)

	return ctx.JSON(fiber.Map{
		"message": "Form submitted successfully",
	})
}
