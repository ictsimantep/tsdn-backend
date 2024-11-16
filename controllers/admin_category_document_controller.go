package controllers

import (
	"backend-school/config"
	"backend-school/helpers"
	"backend-school/models"
	"backend-school/services"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryDocumentController struct {
	Service *services.CategoryDocumentService
}

func NewCategoryDocumentController() *CategoryDocumentController {
	return &CategoryDocumentController{
		Service: services.NewCategoryDocumentService(),
	}
}

type CategoryDocumentRequest struct {
	Name         string                        `json:"name"`
	Prefix       string                        `json:"prefix"`
	RoleHasRules []services.RoleHasRulePayload `json:"role_has_rules"`
}

// GetCategoryDocuments retrieves a paginated list of category documents.
func (c *CategoryDocumentController) GetCategoryDocuments(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to the "category-document" resource using the "read" action
	hasAccess, err := enforcer.Enforce(username, "category-document", "read", "none", "none", "none")
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

	// Parse pagination and search query parameters
	pageStr := ctx.Query("currentPage", "1")
	pageSizeStr := ctx.Query("pageSize", "10")
	search := ctx.Query("search", "")
	showAll := ctx.Query("showAll", "")

	currentPage, err := strconv.Atoi(pageStr)
	if err != nil || currentPage < 1 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid currentPage",
			"data":       nil,
		})
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid pageSize",
			"data":       nil,
		})
	}

	// Call service to get paginated category documents
	result, err := c.Service.GetCategoryDocumentsPaginated(currentPage, pageSize, search, showAll)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to fetch category documents",
			"data":       nil,
		})
	}

	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Success",
		"data":       result,
	})
}

func (c *CategoryDocumentController) CreateCategoryDocument(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to create a "category-document"
	hasAccess, err := enforcer.Enforce(username, "category-document", "create", "none", "none", "none")
	if err != nil {
		log.Printf("Casbin enforcement error: %v", err)
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

	// Parse body from JSON into CategoryDocumentRequest struct
	var req CategoryDocumentRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request payload",
			"data":       nil,
		})
	}

	// Check if a CategoryDocument with the same Prefix already exists
	var existingCategoryDocument models.CategoryDocument
	if err := config.DB.Where("prefix = ?", req.Prefix).First(&existingCategoryDocument).Error; err == nil {
		// If a document with the same Prefix exists, return conflict response
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"statusCode": fiber.StatusConflict,
			"message":    fmt.Sprintf("Category document with prefix '%s' already exists", req.Prefix),
			"data":       nil,
		})
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// If there's an unexpected error in querying, return an internal server error
		log.Printf("Error checking existing category document: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Could not check for existing category document",
			"data":       nil,
		})
	}

	// Convert CategoryDocumentRequest to CategoryDocumentPayload
	payload := services.CategoryDocumentPayload{
		Name:         req.Name,
		Prefix:       req.Prefix,
		RoleHasRules: req.RoleHasRules,
	}

	// Call service to create the category document and related entries
	categoryDocument, err := c.Service.AddCategoryDocument(&payload)
	if err != nil {
		log.Printf("Error creating category document: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Could not create category document",
			"data":       nil,
		})
	}

	// Return success response if no error
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"statusCode": fiber.StatusCreated,
		"message":    "Category document created successfully",
		"data": fiber.Map{
			"name":           categoryDocument.Name,
			"prefix":         categoryDocument.Prefix,
			"uuid":           categoryDocument.UUID, // Include UUID in response
			"role_has_rules": req.RoleHasRules,
		},
	})
}

// GetCategoryDocumentByUUID retrieves a category document by its UUID.
func (c *CategoryDocumentController) GetCategoryDocumentByUUID(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to read a "category-document"
	hasAccess, err := enforcer.Enforce(username, "category-document", "read", "none", "none", "none")
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

	uuidStr := ctx.Params("uuid")
	if uuidStr == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "UUID is required",
			"data":       nil,
		})
	}

	// Parse UUID
	uuid, err := uuid.Parse(uuidStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid UUID format",
			"data":       nil,
		})
	}

	// Fetch category document by UUID
	categoryDocument, err := c.Service.GetCategoryDocumentByUUID(uuid.String())
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"statusCode": fiber.StatusNotFound,
			"message":    "Category document not found",
			"data":       nil,
		})
	}

	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Success",
		"data":       categoryDocument,
	})
}

func (c *CategoryDocumentController) UpdateCategoryDocument(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to update a "category-document"
	hasAccess, err := enforcer.Enforce(username, "category-document", "update", "none", "none", "none")
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

	uuidStr := ctx.Params("uuid")

	// Parse body into CategoryDocumentRequest for easier handling
	var req CategoryDocumentRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request payload",
			"data":       nil,
		})
	}

	// Convert CategoryDocumentRequest to CategoryDocumentPayload for the service layer
	payload := services.CategoryDocumentPayload{
		Name:         req.Name,
		Prefix:       req.Prefix,
		RoleHasRules: req.RoleHasRules,
	}

	// Use service to update the category document with the given UUID
	updatedCategoryDocumentData, err := c.Service.UpdateCategoryDocumentByUUID(uuidStr, &payload)
	if err != nil {
		// Handle specific errors, such as if the document was not found or other issues
		if strings.Contains(err.Error(), "category document not found") {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": fiber.StatusNotFound,
				"message":    "Category document not found",
				"data":       nil,
			})
		}

		// Handle other errors as internal server errors
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    fmt.Sprintf("Could not update category document: %v", err),
			"data":       nil,
		})
	}

	// Return a successful response with the updated document data
	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Category document updated successfully",
		"data": fiber.Map{
			"name":           updatedCategoryDocumentData.Name,
			"prefix":         updatedCategoryDocumentData.Prefix,
			"uuid":           updatedCategoryDocumentData.UUID,
			"role_has_rules": req.RoleHasRules, // returning the updated role_has_rules from the request payload
		},
	})
}

// DeleteCategoryDocument deletes a category document by its UUID.
func (c *CategoryDocumentController) DeleteCategoryDocument(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to delete a "category-document"
	hasAccess, err := enforcer.Enforce(username, "category-document", "delete", "none", "none", "none")
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

	// Get the UUID from the URL parameters
	uuidStr := ctx.Params("uuid")

	// Call the service to delete the category document by UUID
	if err := c.Service.DeleteCategoryDocument(uuidStr); err != nil {
		if err.Error() == "category document not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": fiber.StatusNotFound,
				"message":    "Category document not found",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Could not delete category document",
		})
	}

	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusNoContent,
		"message":    "Category document deleted successfully",
	})
}

func (c *CategoryDocumentController) GetRolesAndActions(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to delete a "category-document"
	hasAccess, err := enforcer.Enforce(username, "category-document", "read", "none", "none", "none")
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

	var roles []string
	var actions []string

	// Query distinct roles
	if err := config.DB.Model(&models.RoleHasRule{}).
		Distinct("role_guard_name").
		Pluck("role_guard_name", &roles).Error; err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to retrieve roles.",
			"data":       nil,
		})
	}

	// Query distinct actions
	if err := config.DB.Model(&models.RoleHasRule{}).
		Distinct("action").
		Pluck("action", &actions).Error; err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to retrieve actions.",
			"data":       nil,
		})
	}

	// Construct the response
	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Success",
		"data": fiber.Map{
			"role":   roles,
			"action": actions,
		},
	})
}
