// controllers/page_controller.go
package controllers

import (
	"backend-school/services"

	"github.com/gofiber/fiber/v2"
)

// PageController handles HTTP requests for pages
type PageController struct {
	PageService *services.PageService
}

// NewPageController creates a new instance of PageController with internally initialized PageService
func NewPageController() *PageController {
	pageService := services.NewPageService()

	return &PageController{PageService: pageService}
}

// GetAllPages returns all pages by calling the PageService
func (c *PageController) GetAllPages(ctx *fiber.Ctx) error {
	pages, err := c.PageService.GetAllPages() // Calling the service
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not fetch pages",
		})
	}
	return ctx.JSON(pages) // Return the response as JSON
}

// GetPageBySlug returns a single page by its slug by calling the PageService
func (c *PageController) GetPageBySlug(ctx *fiber.Ctx) error {
	slug := ctx.Query("slug")                      // Get slug from URL parameter
	page, err := c.PageService.GetPageBySlug(slug) // Calling the service
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Page not found",
		})
	}
	return ctx.JSON(page) // Return the page data as JSON
}
