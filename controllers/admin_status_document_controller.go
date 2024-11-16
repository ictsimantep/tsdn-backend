package controllers

import (
	"backend-school/helpers"
	"backend-school/models"
	"backend-school/services"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type StatusDocumentController struct {
	Service *services.StatusDocumentService
}

func NewStatusDocumentController() *StatusDocumentController {
	return &StatusDocumentController{
		Service: services.NewStatusDocumentService(),
	}
}

// GetStatusDocuments retrieves a paginated list of status documents.
func (c *StatusDocumentController) GetStatusDocuments(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to the "status-document" resource using the "read" action
	hasAccess, err := enforcer.Enforce(username, "status-document", "read", "none", "none", "none")
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

	// Call service to get paginated status documents
	result, err := c.Service.GetStatusDocumentsPaginated(currentPage, pageSize, search)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to fetch status documents",
			"data":       nil,
		})
	}

	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Success",
		"data":       result,
	})
}

// CreateStatusDocument creates a new status document.
func (c *StatusDocumentController) CreateStatusDocument(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to create a "status-document"
	hasAccess, err := enforcer.Enforce(username, "status-document", "create", "none", "none", "none")
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

	var statusDocument models.StatusDocument

	// Parse body from JSON
	if err := ctx.BodyParser(&statusDocument); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    err.Error(),
			"data":       nil,
		})
	}

	// Call service to create the status document
	if err := c.Service.AddStatusDocument(&statusDocument); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Could not create status document",
			"data":       nil,
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"statusCode": fiber.StatusCreated,
		"message":    "Status document created successfully",
		"data":       statusDocument,
	})
}

// GetStatusDocumentByUUID retrieves a status document by its UUID.
func (c *StatusDocumentController) GetStatusDocumentByUUID(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to read a "status-document"
	hasAccess, err := enforcer.Enforce(username, "status-document", "read", "none", "none", "none")
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

	// Fetch status document by UUID
	statusDocument, err := c.Service.GetStatusDocumentByUUID(uuid.String())
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"statusCode": fiber.StatusNotFound,
			"message":    "Status document not found",
			"data":       nil,
		})
	}

	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Success",
		"data":       statusDocument,
	})
}

// UpdateStatusDocument updates an existing status document by its UUID.
func (c *StatusDocumentController) UpdateStatusDocument(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to update a "status-document"
	hasAccess, err := enforcer.Enforce(username, "status-document", "update", "none", "none", "none")
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

	var updatedStatusDocument models.StatusDocument

	// Parse body for updated status document data
	if err := ctx.BodyParser(&updatedStatusDocument); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    err.Error(),
			"data":       nil,
		})
	}

	// Update status document using service
	updatedStatusDocumentData, err := c.Service.UpdateStatusDocumentByUUID(uuidStr, &updatedStatusDocument)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    fmt.Sprintf("Could not update status document: %v", err),
			"data":       nil,
		})
	}

	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Status document updated successfully",
		"data":       updatedStatusDocumentData,
	})
}

// DeleteStatusDocument deletes a status document by its UUID.
func (c *StatusDocumentController) DeleteStatusDocument(ctx *fiber.Ctx) error {
	username := ctx.Locals("username").(string)

	// Get the Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the user has access to delete a "status-document"
	hasAccess, err := enforcer.Enforce(username, "status-document", "delete", "none", "none", "none")
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

	// Call the service to delete the status document by UUID
	if err := c.Service.DeleteStatusDocument(uuidStr); err != nil {
		if err.Error() == "status document not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": fiber.StatusNotFound,
				"message":    "Status document not found",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Could not delete status document",
		})
	}

	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusNoContent,
		"message":    "Status document deleted successfully",
	})
}
