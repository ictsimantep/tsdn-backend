package controllers

import (
	"backend-school/helpers"
	"backend-school/services"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type HealthController struct {
	Service *services.HealthService
}

func NewHealthController() *HealthController {
	return &HealthController{
		Service: services.NewHealthService(),
	}
}

type HealthRequest struct {
	Nama         string `json:"nama"`
	JenisKelamin string `json:"jenis_kelamin"`
	Umur         string `json:"umur"`
	Bb           string `json:"bb"`
	Tb           string `json:"tb"`
	Systol       string `json:"systol"`
	Diastol      string `json:"diastol"`
	HeartRate    string `json:"heart_rate"`
	Profesi      string `json:"profesi"`
}

// GetHealths retrieves a paginated list of health data.
func (c *HealthController) GetHealths(ctx *fiber.Ctx) error {
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
	// if err != nil || currentPage < 1 {
	// 	return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"statusCode": fiber.StatusBadRequest,
	// 		"message":    "Invalid currentPage",
	// 		"data":       nil,
	// 	})
	// }

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid pageSize",
			"data":       nil,
		})
	}

	// Call service to get paginated health data
	result, err := c.Service.GetHealthsPaginated(currentPage, pageSize, search, showAll)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Failed to fetch health data",
			"data":       nil,
		})
	}

	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Success",
		"data":       result,
	})
}

func (c *HealthController) CreateHealth(ctx *fiber.Ctx) error {
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

	// Parse body from JSON into HealthRequest struct
	var req HealthRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"statusCode": fiber.StatusBadRequest,
			"message":    "Invalid request payload",
			"data":       nil,
		})
	}

	// Convert HealthRequest to Health model
	Health := services.HealthPayload{
		Nama:         ctx.FormValue("nama"),
		JenisKelamin: ctx.FormValue("jenis_kelamin"),
		Umur:         ctx.FormValue("umur"),
		Bb:           ctx.FormValue("bb"),
		Tb:           ctx.FormValue("tb"),
		Systol:       ctx.FormValue("systol"),
		Diastol:      ctx.FormValue("diastol"),
		HeartRate:    ctx.FormValue("heart_rate"),
		Profesi:      ctx.FormValue("profesi"),
	}
	log.Printf("Payload: %v", Health)
	// Call service to create the document type
	createdHealth, err := c.Service.AddHealth(&Health)
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
		"message":    "Health Data created successfully",
		"data": fiber.Map{
			"name": createdHealth.Nama,
			"uuid": createdHealth.UUID,
		},
	})
}

// GetHealthByUUID retrieves a Health Data by its UUID.
func (c *HealthController) GetHealthByUUID(ctx *fiber.Ctx) error {
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

	Health, err := c.Service.GetHealthByUUID(uuid.String())
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"statusCode": fiber.StatusNotFound,
			"message":    "Health Data not found",
			"data":       nil,
		})
	}

	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusOK,
		"message":    "Success",
		"data":       Health,
	})
}

// DeleteHealth deletes a Health Data by its UUID.
func (c *HealthController) DeleteHealth(ctx *fiber.Ctx) error {
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

	if err := c.Service.DeleteHealth(uuidStr); err != nil {
		if err.Error() == "Health Data not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"statusCode": fiber.StatusNotFound,
				"message":    "Health Data not found",
			})
		}
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"statusCode": fiber.StatusInternalServerError,
			"message":    "Could not delete document type",
		})
	}

	return ctx.JSON(fiber.Map{
		"statusCode": fiber.StatusNoContent,
		"message":    "Health Data deleted successfully",
	})
}
