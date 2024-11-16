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
var validate = validator.New()

var customErrorMessages = map[string]string{
	"Title.required":       "Title is required",
	"Title.min":            "Title must be at least 3 characters long",
	"Slug.required":        "Slug is required",
	"Description.required": "Description is required",
	"Description.min":      "Description must be at least 10 characters long",
	"CategoryID.required":  "Category is required",
	"CategoryID.numeric":   "Category must be a number",
	"Like.gte":             "Like count cannot be negative",
	"View.gte":             "View count cannot be negative",
	"Status.required":      "Status is required",
	"Status.oneof":         "Status must be either 1 or 2",
}

type PublicationController struct {
	PublicationService *services.PublicationService
}

// NewPublicationController initializes the controller and handles any errors from NewPublicationService
func NewPublicationController() *PublicationController {
	publicationService, err := services.NewPublicationService()
	if err != nil {
		log.Fatalf("Failed to initialize publication service: %v", err) // Log and stop execution if there's an error
	}
	return &PublicationController{PublicationService: publicationService}
}

// GetPublication handles fetching a publication by slug
func (c *PublicationController) GetPublication(ctx *fiber.Ctx) error {
	slug := ctx.Params("slug")
	publication, err := c.PublicationService.GetPublicationBySlug(slug)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{ // Use ctx.Status
			"statusCode": fiber.StatusNotFound,
			"message":    "Publication not found",
			"data":       nil,
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{ // Use ctx.Status
		"statusCode": fiber.StatusOK,
		"message":    "Publication fetched successfully",
		"data":       publication,
	})
}

// GetPublicationUUID handles fetching a publication by UUID
func (c *PublicationController) GetPublicationUUID(ctx *fiber.Ctx) error {
	// Get the username from the context (set by the JWT middleware)
	username := ctx.Locals("username").(string) // Use ctx.Locals instead of c.Locals

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to the "/admin" resource using the "GET" action
	hasAccess, err := enforcer.Enforce(username, "publication", "read")
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
	publication, err := c.PublicationService.GetPublicationByUUID(uuidParam)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"statusCode": fiber.StatusNotFound,
			"message":    "Publication not found",
			"data":       nil,
		})
	}

	// Jika berhasil, kembalikan respon dengan data
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Publication fetched successfully",
		"data":       publication,
	})
}

// Other methods follow the same pattern of using ctx for handling status and locals

// GetPublicationByCategory handles fetching publications by category slug
func (c *PublicationController) GetPublicationByCategory(ctx *fiber.Ctx) error {
	slug := ctx.Params("slug")
	publications, err := c.PublicationService.GetPublicationsByCategory(slug)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{ // Use ctx.Status
			"statusCode": fiber.StatusNotFound,
			"message":    "No publications found in this category",
			"data":       nil,
		})
	}
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{ // Use ctx.Status
		"statusCode": fiber.StatusOK,
		"message":    "Publications fetched successfully",
		"data":       publications,
	})
}

// GetPublicationsPaginated handles fetching paginated publications with sorting
func (c *PublicationController) GetPublicationsCategory(ctx *fiber.Ctx) error {
	// Get the username from the context (set by the JWT middleware)
	username := ctx.Locals("username").(string) // Use ctx.Locals instead of c.Locals

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to the "/admin" resource using the "GET" action
	hasAccess, err := enforcer.Enforce(username, "publication", "read")
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
	pageStr := ctx.Query("currentPage", ctx.Query("page", "0"))
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

	publications, err := c.PublicationService.GetPublicationsCategoryPaginated(perPage, page, sortBy, sortDesc)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{ // Use ctx.Status
			"statusCode": fiber.StatusInternalServerError,
			"message":    err.Error(),
			"data":       nil,
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{ // Use ctx.Status
		"statusCode": fiber.StatusOK,
		"message":    "Category Publications fetched successfully",
		"data":       publications,
	})
}

// GetPublicationsPaginated handles fetching paginated publications with sorting
func (c *PublicationController) GetPublicationsPaginated(ctx *fiber.Ctx) error {
	// Get the username from the context (set by the JWT middleware)
	username := ctx.Locals("username").(string) // Use ctx.Locals instead of c.Locals

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to the "/admin" resource using the "GET" action
	hasAccess, err := enforcer.Enforce(username, "publication", "read")
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

	perPageStr := ctx.Query("pageSize", ctx.Query("pageSize", "10"))
	pageStr := ctx.Query("currentPage", "1")
	sortBy := ctx.Query("sortBy", "id")
	sortDescStr := ctx.Query("sortDesc", "false")

	perPage, err := strconv.Atoi(perPageStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{ // Use ctx.Status
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid perPage values",
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

	publications, err := c.PublicationService.GetPublicationsPaginated(perPage, page, sortBy, sortDesc)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{ // Use ctx.Status
			"statusCode": fiber.StatusInternalServerError,
			"message":    err.Error(),
			"data":       nil,
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{ // Use ctx.Status
		"statusCode": fiber.StatusOK,
		"message":    "Publications fetched successfully",
		"data":       publications,
	})
}

func (c *PublicationController) CreatePublication(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string) // Use ctx.Locals instead of c.Locals

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to the "/admin" resource using the "GET" action
	hasAccess, err := enforcer.Enforce(username, "publication", "create")
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
	// Parse request body into the Publication model
	publication := new(models.Publication)
	if err := ctx.BodyParser(publication); err != nil {
		// Handle empty or invalid input format
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"status":     "error",
			"message":    "Invalid or empty input",
		})
	}

	// Validate the input data
	if err := validate.Struct(publication); err != nil {
		// If validation fails, return the validation error messages
		validationErrors := err.(validator.ValidationErrors)
		errors := make(map[string]string)

		for _, err := range validationErrors {
			field := err.Field() // Field that failed validation
			tag := err.Tag()     // The validation rule (e.g., 'required', 'min', 'numeric')

			// Construct error key using field and tag, e.g., "Title.required"
			errorKey := fmt.Sprintf("%s.%s", field, tag)

			// Get custom error message from customErrorMessages or use a default message
			if customMsg, exists := customErrorMessages[errorKey]; exists {
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

	// Create the publication with the uploaded image
	if err := c.PublicationService.CreatePublication(publication, img); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"status":     "error",
			"message":    err.Error(),
		})
	}

	// Return a success response when the publication is created
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"status":     "success",
		"message":    "Publication created successfully",
		"data":       publication,
	})
}

func (c *PublicationController) UpdatePublication(ctx *fiber.Ctx) error {
	// Get the username from the context (set by the JWT middleware)
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to update
	hasAccess, err := enforcer.Enforce(username, "publication", "update")
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
	updatedPublication := new(models.Publication)

	// Parse request body into the Publication model
	if err := ctx.BodyParser(updatedPublication); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"status":     "error",
			"message":    "Invalid or empty input",
		})
	}
	fmt.Printf("updatedPublication", updatedPublication)

	// Validate the input data
	if err := validate.Struct(updatedPublication); err != nil {
		// If validation fails, return the validation error messages
		validationErrors := err.(validator.ValidationErrors)
		errors := make(map[string]string)

		for _, err := range validationErrors {
			field := err.Field() // Field that failed validation
			tag := err.Tag()     // The validation rule (e.g., 'required', 'min', 'numeric')
			// Debugging output
			fmt.Printf("Validation failed for field: %s with rule: %s\n", field, tag)

			// Construct error key using field and tag, e.g., "Title.required"
			errorKey := fmt.Sprintf("%s.%s", field, tag)

			// Get custom error message from customErrorMessages or use a default message
			if customMsg, exists := customErrorMessages[errorKey]; exists {
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

	// Call the service layer to update the publication with the provided data and image
	if err := c.PublicationService.UpdatePublication(uuid, updatedPublication, img); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"status":     "error",
			"message":    err.Error(),
		})
	}

	// Return a success response when the publication is updated
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"status":     "success",
		"message":    "Publication updated successfully",
		"data":       updatedPublication,
	})
}

// DeletePublication handles deleting a publication by slug
func (c *PublicationController) DeletePublication(ctx *fiber.Ctx) error {
	// Get the username from the context (set by the JWT middleware)
	username := ctx.Locals("username").(string) // Use ctx.Locals instead of c.Locals

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to the "/admin" resource using the "GET" action
	hasAccess, err := enforcer.Enforce(username, "publication", "delete")
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

	if err := c.PublicationService.DeletePublication(uuid); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{ // Use ctx.Status
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to delete publication",
			"data":       nil,
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{ // Use ctx.Status
		"statusCode": fiber.StatusOK,
		"message":    "Publication deleted successfully",
		"data":       nil,
	})
}
