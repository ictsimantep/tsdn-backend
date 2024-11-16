package controllers

import (
	"backend-school/config"
	"backend-school/helpers"
	"backend-school/models"
	"backend-school/services"
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type DocumentControlController struct {
	Service *services.DocumentControlService
}

// NewDocumentControlController initializes and returns a new DocumentControlController
func NewDocumentControlController() *DocumentControlController {
	minioClient := config.MinioClient // Assumes MinioClient is initialized
	bucketName := os.Getenv("MINIO_BUCKET")
	minioService := services.NewMinioService(minioClient)

	documentControlService := services.NewDocumentControlService(minioClient, bucketName, minioService)

	return &DocumentControlController{Service: documentControlService}
}

func (c *DocumentControlController) CreateDocumentControl(ctx *fiber.Ctx) error {
	// Get the username of the requester from the context (set by JWT middleware)
	requesterUsername := ctx.Locals("username").(string)
	userID := ctx.Locals("user_id").(string)

	var req services.DocumentControlPayload
	err := ctx.BodyParser(&req)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to parse request body",
			"error":      err.Error(),
		})
	}

	//add user_id
	id, err := strconv.Atoi(userID)
	if err != nil {
		// handle the error, e.g., log it or return an error response
		log.Println("Error converting userID:", err)
	}
	req.CreatedBy = id

	// Get Casbin enforcer
	enforcer := helpers.GetCasbinEnforcer()

	// Check if the requester has access to view the specific user's details
	hasAccess, err := enforcer.Enforce(requesterUsername, "document", "create", "none", "none", "none")
	if err != nil {
		log.Printf("Error checking Casbin permissions: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to check access permissions.",
		})
	}

	// If the requester doesn't have access, return a forbidden status
	if !hasAccess {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"statusCode": fiber.StatusForbidden,
			"message":    "Forbidden: You don't have permission to access this resource.",
		})
	}

	// Process file from form-data
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "File upload failed",
			"detail":     err.Error(), // Provide file-specific error details
		})
	}

	// Call service to create DocumentControl and save the initial document version
	documentControl, documentVersion, err := c.Service.AddDocumentControlWithVersion(&req, fileHeader)
	if err != nil {
		// Capture the detailed error from the service and return it to the client
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Could not create document control and version",
			"detail":     err.Error(), // Send detailed error back in the response
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"statusCode": fiber.StatusCreated,
		"message":    "Document control and initial version created successfully",
		"data": fiber.Map{
			"document_control": documentControl,
			"document_version": documentVersion,
		},
	})
}

// GetDocumentControls retrieves a paginated list of document controls with optional search filter
func (c *DocumentControlController) GetDocumentInternalControls(ctx *fiber.Ctx) error {
	pageStr := ctx.Query("currentPage", "1")
	pageSizeStr := ctx.Query("pageSize", "10")
	search := ctx.Query("search", "")

	currentPage, err := strconv.Atoi(pageStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid currentPage",
		})
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid pageSize",
		})
	}

	// Call service to get paginated document controls
	result, err := c.Service.GetDocumentControlsInternalPaginated(currentPage, pageSize, search)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to fetch document controls",
		})
	}

	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Success",
		"data":       result,
	})
}

// GetDocumentControls retrieves a paginated list of document controls with optional search filter
func (c *DocumentControlController) GetDocumentExternalControls(ctx *fiber.Ctx) error {
	pageStr := ctx.Query("currentPage", "1")
	pageSizeStr := ctx.Query("pageSize", "10")
	search := ctx.Query("search", "")

	currentPage, err := strconv.Atoi(pageStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid currentPage",
		})
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid pageSize",
		})
	}

	// Call service to get paginated document controls
	result, err := c.Service.GetDocumentControlsExternalPaginated(currentPage, pageSize, search)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to fetch document controls",
		})
	}

	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Success",
		"data":       result,
	})
}

// GetDocumentControlByUUID retrieves a single document control by its UUID
func (c *DocumentControlController) GetDocumentControlByUUID(ctx *fiber.Ctx) error {
	uuidStr := ctx.Params("uuid")
	Username := ctx.Locals("username").(string)
	userID := ctx.Locals("user_id").(int)

	// Call service to get document control by UUID
	documentControl, err := c.Service.GetDocumentControlByUUID(uuidStr)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"statusCode": fiber.StatusNotFound,
			"message":    "Document control not found",
		})
	}

	// Deklarasi variabel untuk menyimpan hasil query
	var documentCategory models.CategoryDocument
	var documentType models.DocumentType
	var statusDocument models.StatusDocument
	var documentControlCB models.DocumentControl

	// Query untuk mencari data berdasarkan `status_document_id` dari request
	if err := config.DB.Where("id = ?", documentControl.DocumentCategoryID).First(&documentCategory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Jika data tidak ditemukan
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": fiber.StatusNotFound,
				"message":    "Category document not found",
			})
		}
		// Jika ada error lain saat query
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to retrieve document category",
			"error":      err.Error(),
		})
	}

	// Query untuk mencari data berdasarkan `status_document_id` dari request
	if err := config.DB.Where("id = ?", documentControl.StatusDocumentID).First(&statusDocument).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Jika data tidak ditemukan
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": fiber.StatusNotFound,
				"message":    "Status document not found",
			})
		}
		// Jika ada error lain saat query
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to retrieve document status",
			"error":      err.Error(),
		})
	}

	// Query untuk mencari data berdasarkan `status_document_id` dari request
	if err := config.DB.Where("id = ?", documentControl.DocumentTypeID).First(&documentType).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Jika data tidak ditemukan
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": fiber.StatusNotFound,
				"message":    "Type document not found",
			})
		}
		// Jika ada error lain saat query
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to retrieve document type",
			"error":      err.Error(),
		})
	}

	// Query untuk mencari data berdasarkan `document_type_id` dari request
	if err := config.DB.Where("created_by = ?", userID).Where("id = ?", documentControl.ID).First(&documentControlCB).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Jika data tidak ditemukan
			// Get Casbin enforcer
			enforcer := helpers.GetCasbinEnforcer()

			// Check if the requester has access to view the specific details
			hasAccess, err := enforcer.Enforce(Username, "document", strings.ToLower(statusDocument.Name), documentCategory.Prefix, documentType.Prefix, "none")
			if err != nil {
				log.Printf("Error checking Casbin permissions: %v", err)
				return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"statusCode": fiber.StatusInternalServerError,
					"message":    "Failed to check access permissions.",
				})
			}

			// If the requester doesn't have access, return a forbidden status
			if !hasAccess {
				return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"statusCode": fiber.StatusForbidden,
					"message":    "Forbidden: You don't have permission to access this resource.",
				})
			}
		}
	}

	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Success",
		"data":       documentControl,
	})
}

// DeleteDocumentControl deletes a document control by its UUID
func (c *DocumentControlController) DeleteDocumentControl(ctx *fiber.Ctx) error {
	uuidStr := ctx.Params("uuid")

	// Call service to delete the document control and its related version file
	if err := c.Service.DeleteDocumentControl(uuidStr); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Could not delete document control",
		})
	}

	return ctx.Status(fiber.StatusNoContent).JSON(fiber.Map{
		"statusCode": fiber.StatusNoContent,
		"message":    "Document control deleted successfully",
	})
}

// UpdateDocumentControl updates a document control by UUID
func (c *DocumentControlController) UpdateDocumentControl(ctx *fiber.Ctx) error {
	uuidStr := ctx.Params("uuid")

	// Parse the request body
	var req services.DocumentControlPayload
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request payload",
		})
	}

	// Handle optional file update
	fileHeader, err := ctx.FormFile("file")
	if err != nil && err.Error() != "http: no such file" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "File upload failed",
		})
	}

	// Call the service to update the document control
	documentControl, err := c.Service.UpdateDocumentControl(uuidStr, &req, fileHeader)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Could not update document control",
		})
	}

	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Document control updated successfully",
		"data":       documentControl,
	})
}
