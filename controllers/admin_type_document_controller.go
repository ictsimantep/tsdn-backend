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

type DocumentTypeController struct {
	Service *services.DocumentTypeService
}

func NewDocumentTypeController() *DocumentTypeController {
	return &DocumentTypeController{
		Service: services.NewDocumentTypeService(),
	}
}

type DocumentTypeRequest struct {
	Name               string                        `json:"name"`
	Prefix             string                        `json:"prefix"`
	RoleHasRules       []services.RoleHasRulePayload `json:"role_has_rules"`
	DocumentCategoryID int                           `json:"document_category_id"`
}

// GetDocumentTypes retrieves a paginated list of document types.
func (c *DocumentTypeController) GetDocumentTypes(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to the "document-type" resource with "read" action
	hasAccess, err := enforcer.Enforce(username, "document-type", "read", "none", "none", "none")
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

	// Call service to get paginated document types
	result, err := c.Service.GetDocumentTypesPaginated(currentPage, pageSize, search, showAll)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to fetch document types",
			"data":       nil,
		})
	}

	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Success",
		"data":       result,
	})
}

func (c *DocumentTypeController) CreateDocumentType(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to create a "document-type"
	hasAccess, err := enforcer.Enforce(username, "document-type", "create", "none", "none", "none")
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

	// Parse body from JSON into DocumentTypeRequest struct
	var req DocumentTypeRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request payload",
			"data":       nil,
		})
	}

	// Check if a DocumentType with the same Prefix already exists
	var existingDocumentType models.DocumentType
	if err := config.DB.Where("prefix = ?", req.Prefix).First(&existingDocumentType).Error; err == nil {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"statusCode": fiber.StatusConflict,
			"message":    fmt.Sprintf("Document type with prefix '%s' already exists", req.Prefix),
			"data":       nil,
		})
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Error checking existing document type: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Could not check for existing document type",
			"data":       nil,
		})
	}

	// Convert DocumentTypeRequest to DocumentType model
	documentType := services.DocumentTypePayload{
		Name:               req.Name,
		Prefix:             req.Prefix,
		RoleHasRules:       req.RoleHasRules,
		DocumentCategoryID: req.DocumentCategoryID,
	}

	// Call service to create the document type
	createdDocumentType, err := c.Service.AddDocumentType(&documentType)
	if err != nil {
		log.Printf("Error creating document type: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Could not create document type",
			"data":       nil,
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"statusCode": fiber.StatusCreated,
		"message":    "Document type created successfully",
		"data": fiber.Map{
			"name":   createdDocumentType.Name,
			"prefix": createdDocumentType.Prefix,
			"uuid":   createdDocumentType.UUID,
		},
	})
}

// GetDocumentTypeByUUID retrieves a document type by its UUID.
func (c *DocumentTypeController) GetDocumentTypeByUUID(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to read a "document-type"
	hasAccess, err := enforcer.Enforce(username, "document-type", "read", "none", "none", "none")
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

	uuid, err := uuid.Parse(uuidStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid UUID format",
			"data":       nil,
		})
	}

	documentType, err := c.Service.GetDocumentTypeByUUID(uuid.String())
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"statusCode": fiber.StatusNotFound,
			"message":    "Document type not found",
			"data":       nil,
		})
	}

	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Success",
		"data":       documentType,
	})
}

func (c *DocumentTypeController) UpdateDocumentType(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to update a "category-document"
	hasAccess, err := enforcer.Enforce(username, "document-type", "update", "none", "none", "none")
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
	var req DocumentTypeRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request payload",
			"data":       nil,
		})
	}

	// Convert CategoryDocumentRequest to CategoryDocumentPayload for the service layer
	payload := services.DocumentTypePayload{
		Name:               req.Name,
		Prefix:             req.Prefix,
		RoleHasRules:       req.RoleHasRules,
		DocumentCategoryID: req.DocumentCategoryID,
	}

	// Use service to update the category document with the given UUID
	updatedDocumentTypeData, err := c.Service.UpdateDocumentTypeByUUID(uuidStr, &payload)
	if err != nil {
		// Handle specific errors, such as if the document was not found or other issues
		if strings.Contains(err.Error(), "type document not found") {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": fiber.StatusNotFound,
				"message":    "type document not found",
				"data":       nil,
			})
		}

		// Handle other errors as internal server errors
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    fmt.Sprintf("Could not update type document: %v", err),
			"data":       nil,
		})
	}

	// Return a successful response with the updated document data
	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "type document updated successfully",
		"data": fiber.Map{
			"name":           updatedDocumentTypeData.Name,
			"prefix":         updatedDocumentTypeData.Prefix,
			"uuid":           updatedDocumentTypeData.UUID,
			"role_has_rules": req.RoleHasRules, // returning the updated role_has_rules from the request payload
		},
	})
}

// DeleteDocumentType deletes a document type by its UUID.
func (c *DocumentTypeController) DeleteDocumentType(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to delete a "document-type"
	hasAccess, err := enforcer.Enforce(username, "document-type", "delete", "none", "none", "none")
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

	if err := c.Service.DeleteDocumentType(uuidStr); err != nil {
		if err.Error() == "document type not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": fiber.StatusNotFound,
				"message":    "Document type not found",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Could not delete document type",
		})
	}

	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusNoContent,
		"message":    "Document type deleted successfully",
	})
}
