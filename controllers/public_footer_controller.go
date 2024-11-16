// controllers/footer_controller.go
package controllers

import (
	"backend-school/services"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type FooterController struct {
	Service *services.FooterService
}

// NewFooterController creates a new FooterController
func NewFooterController() *FooterController {
	footerService, err := services.NewFooterService()
	if err != nil {
		log.Fatalf("Failed to initialize footer service: %v", err) // Log and stop execution if there's an error
	}
	return &FooterController{Service: footerService}
}

// GetFooters returns all footers
func (c *FooterController) GetFooters(ctx *fiber.Ctx) error {
	footers, err := c.Service.GetAllFooters()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return ctx.JSON(footers)
}

// GetFooterByID returns a footer by its ID
func (c *FooterController) GetFooterByID(ctx *fiber.Ctx) error {
	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	footer, err := c.Service.GetFooterByID(uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return ctx.JSON(footer)
}
