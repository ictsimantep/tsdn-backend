package controllers

import (
	"backend-school/helpers"
	"backend-school/models"
	"backend-school/services"
	"fmt"
	"log"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Inisialisasi validator
var validateAdminSetting = validator.New()

var customErrorMessagesAdminSetting = map[string]string{
	"Key.required":   "key is required",
	"Key.min":        "key must be at least 3 characters long",
	"Value.required": "Value is required",
	"Value.min":      "Value must be at least 3 characters long",
}

type AdminSettingController struct {
	AdminSettingsService *services.AdminSettingsService
}

// NewAdminSettingController initializes the controller and handles any errors from NewAdminSettingsService
func NewAdminSettingController() *AdminSettingController {
	settingService, err := services.NewAdminSettingsService()
	if err != nil {
		log.Fatalf("Failed to initialize setting service: %v", err) // Log and stop execution if there's an error
	}
	return &AdminSettingController{AdminSettingsService: settingService}
}

// GetAdminSetting handles fetching a setting by slug
func (c *AdminSettingController) GetAdminSetting(ctx *fiber.Ctx) error {
	slug := ctx.Params("slug")
	setting, err := c.AdminSettingsService.GetAdminSettingBySlug(slug)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{ // Use ctx.Status
			"statusCode": fiber.StatusNotFound,
			"message":    "Setting not found",
			"data":       nil,
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{ // Use ctx.Status
		"statusCode": fiber.StatusOK,
		"message":    "Setting fetched successfully",
		"data":       setting,
	})
}

// GetAdminSettingUUID handles fetching a setting by UUID
func (c *AdminSettingController) GetAdminSettingUUID(ctx *fiber.Ctx) error {
	// Get the username from the context (set by the JWT middleware)
	username := ctx.Locals("username").(string) // Use ctx.Locals instead of c.Locals

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to the "/admin" resource using the "GET" action
	hasAccess, err := enforcer.Enforce(username, "settings", "read", "none", "none", "none")
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{ // Use ctx.Status
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to check access.",
		})
	}

	if !hasAccess {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{ // Use ctx.Status
			"statusCode": fiber.StatusForbidden,
			"message":    "Forbidden: You don't have access to this resource",
		})
	}
	// Ambil parameter UUID
	uuidParam := ctx.Params("uuid")

	// Validasi apakah UUID memiliki format yang benar
	if _, err := uuid.Parse(uuidParam); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid UUID format",
			"data":       nil,
		})
	}

	// Memanggil service untuk mendapatkan data publikasi
	setting, err := c.AdminSettingsService.GetAdminSettingByUUID(uuidParam)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"statusCode": fiber.StatusNotFound,
			"message":    "Setting not found",
			"data":       nil,
		})
	}

	// Jika berhasil, kembalikan respon dengan data
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Setting fetched successfully",
		"data":       setting,
	})
}

// Other methods follow the same pattern of using ctx for handling status and locals
// GetAdminSettingsPaginated handles fetching paginated settings with sorting
func (c *AdminSettingController) GetAdminSettingsPaginated(ctx *fiber.Ctx) error {
	// Get the username from the context (set by the JWT middleware)
	username := ctx.Locals("username").(string) // Use ctx.Locals instead of c.Locals

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to the "/admin" resource using the "GET" action
	hasAccess, err := enforcer.Enforce(username, "settings", "read", "none", "none", "none")
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{ // Use ctx.Status
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to check access.",
		})
	}

	if !hasAccess {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{ // Use ctx.Status
			"statusCode": fiber.StatusForbidden,
			"message":    "Forbidden: You don't have access to this resource",
		})
	}

	perPageStr := ctx.Query("perPage", ctx.Query("pageSize", "10"))
	pageStr := ctx.Query("currentPage", ctx.Query("page", "1"))
	sortBy := ctx.Query("sortBy", "id")
	sortDescStr := ctx.Query("sortDesc", "false")

	perPage, err := strconv.Atoi(perPageStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{ // Use ctx.Status
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid perPage value",
			"data":       nil,
		})
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{ // Use ctx.Status
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid page value",
			"data":       nil,
		})
	}

	sortDesc, err := strconv.ParseBool(sortDescStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{ // Use ctx.Status
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid sortDesc value",
			"data":       nil,
		})
	}

	settings, err := c.AdminSettingsService.GetAdminSettingsPaginated(perPage, page, sortBy, sortDesc)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{ // Use ctx.Status
			"statusCode": fiber.StatusInternalServerError,
			"message":    err.Error(),
			"data":       nil,
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{ // Use ctx.Status
		"statusCode": fiber.StatusOK,
		"message":    "AdminSettings fetched successfully",
		"data":       settings,
	})
}

func (c *AdminSettingController) CreateSetting(ctx *fiber.Ctx) error {
	// Parse request body into the AdminSetting model
	setting := new(models.Setting)
	if err := ctx.BodyParser(setting); err != nil {
		// Handle empty or invalid input format
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"status":     "error",
			"message":    "Invalid or empty input",
		})
	}

	// Validate the input data
	if err := validateAdminSetting.Struct(setting); err != nil {
		// If validation fails, return the validation error messages
		validationErrors := err.(validator.ValidationErrors)
		errors := make(map[string]string)

		for _, err := range validationErrors {
			field := err.Field() // Field that failed validation
			tag := err.Tag()     // The validation rule (e.g., 'required', 'min', 'numeric')

			// Construct error key using field and tag, e.g., "Title.required"
			errorKey := fmt.Sprintf("%s.%s", field, tag)

			// Get custom error message from customErrorMessagesAdminSetting or use a default message
			if customMsg, exists := customErrorMessagesAdminSetting[errorKey]; exists {
				errors[field] = customMsg
			} else {
				// Provide a default message if no custom message is found
				switch tag {
				case "required":
					errors[field] = fmt.Sprintf("%s is required", field)
				case "min":
					errors[field] = fmt.Sprintf("%s must be at least %s characters", field, err.Param())
				case "numeric":
					errors[field] = fmt.Sprintf("%s must be a number", field)
				case "gte":
					errors[field] = fmt.Sprintf("%s must be greater than or equal to %s", field, err.Param())
				case "oneof":
					errors[field] = fmt.Sprintf("%s must be one of the following: %s", field, err.Param())
				default:
					errors[field] = fmt.Sprintf("%s is invalid", field)
				}
			}
		}

		// Return the validation errors with a 400 status code
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"status":     "error",
			"message":    "Validation failed",
			"errors":     errors,
		})
	}

	// Handle file upload (if present)
	img, err := ctx.FormFile("image_url")
	if err != nil && err != fiber.ErrUnprocessableEntity {
		img = nil
	}

	// Create the setting with the uploaded image
	if err := c.AdminSettingsService.CreateSetting(setting, img); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"status":     "error",
			"message":    err.Error(),
		})
	}

	// Return a success response when the setting is created
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"status":     "success",
		"message":    "Setting created successfully",
		"data":       setting,
	})
}

func (c *AdminSettingController) UpdateSetting(ctx *fiber.Ctx) error {
	// Get the username from the context (set by the JWT middleware)
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to update
	hasAccess, err := enforcer.Enforce(username, "settings", "update", "none", "none", "none")
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to check access.",
		})
	}

	if !hasAccess {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"statusCode": fiber.StatusForbidden,
			"message":    "Forbidden: You don't have access to this resource",
		})
	}

	uuid := ctx.Params("uuid")
	updatedAdminSetting := new(models.Setting)

	// Parse request body into the AdminSetting model
	if err := ctx.BodyParser(updatedAdminSetting); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"status":     "error",
			"message":    "Invalid or empty input",
		})
	}

	// Validate the input data
	if err := validateAdminSetting.Struct(updatedAdminSetting); err != nil {
		// If validation fails, return the validation error messages
		validationErrors := err.(validator.ValidationErrors)
		errors := make(map[string]string)

		for _, err := range validationErrors {
			field := err.Field() // Field that failed validation
			tag := err.Tag()     // The validation rule (e.g., 'required', 'min', 'numeric')

			// Construct error key using field and tag, e.g., "Title.required"
			errorKey := fmt.Sprintf("%s.%s", field, tag)

			// Get custom error message from customErrorMessagesAdminSetting or use a default message
			if customMsg, exists := customErrorMessagesAdminSetting[errorKey]; exists {
				errors[field] = customMsg
			} else {
				// Provide a default message if no custom message is found
				switch tag {
				case "required":
					errors[field] = fmt.Sprintf("%s is required", field)
				case "min":
					errors[field] = fmt.Sprintf("%s must be at least %s characters", field, err.Param())
				case "numeric":
					errors[field] = fmt.Sprintf("%s must be a number", field)
				case "gte":
					errors[field] = fmt.Sprintf("%s must be greater than or equal to %s", field, err.Param())
				case "oneof":
					errors[field] = fmt.Sprintf("%s must be one of the following: %s", field, err.Param())
				default:
					errors[field] = fmt.Sprintf("%s is invalid", field)
				}
			}
		}

		// Return the validation errors with a 400 status code
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"status":     "error",
			"message":    "Validation failed",
			"errors":     errors,
		})
	}

	// Handle file upload (if present)
	img, err := ctx.FormFile("img")
	if err != nil && err != fiber.ErrUnprocessableEntity {
		img = nil
	}

	// Call the service layer to update the setting with the provided data and image
	if err := c.AdminSettingsService.UpdateSetting(uuid, updatedAdminSetting, img); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"status":     "error",
			"message":    err.Error(),
		})
	}

	// Return a success response when the setting is updated
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"status":     "success",
		"message":    "Setting updated successfully",
		"data":       updatedAdminSetting,
	})
}

// DeleteAdminSetting handles deleting a setting by slug
func (c *AdminSettingController) DeleteSetting(ctx *fiber.Ctx) error {
	// Get the username from the context (set by the JWT middleware)
	username := ctx.Locals("username").(string) // Use ctx.Locals instead of c.Locals

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to the "/admin" resource using the "GET" action
	hasAccess, err := enforcer.Enforce(username, "settings", "delete", "none", "none", "none")
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{ // Use ctx.Status
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to check access.",
		})
	}

	if !hasAccess {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{ // Use ctx.Status
			"statusCode": fiber.StatusForbidden,
			"message":    "Forbidden: You don't have access to this resource",
		})
	}

	uuid := ctx.Params("uuid")

	if err := c.AdminSettingsService.DeleteSetting(uuid); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{ // Use ctx.Status
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to delete setting",
			"data":       nil,
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{ // Use ctx.Status
		"statusCode": fiber.StatusOK,
		"message":    "Setting deleted successfully",
		"data":       nil,
	})
}
