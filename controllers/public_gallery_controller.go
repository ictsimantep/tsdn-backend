package controllers

import (
	"backend-school/services"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type GalleryController struct {
	GalleryService *services.GalleryService
}

func NewGalleryController() *GalleryController {
	galleryService, err := services.NewGalleryService()
	if err != nil {
		log.Fatalf("Failed to initialize gallery service: %v", err) // Log and stop execution if there's an error
	}
	return &GalleryController{GalleryService: galleryService}
}

// GetGalleries handles fetching paginated galleries with optional category filter
func (c *GalleryController) GetGalleries(ctx *fiber.Ctx) error {
	// Get pagination parameters
	pageStr := ctx.Query("currentPage", "1")
	pageSizeStr := ctx.Query("pageSize", "10")
	search := ctx.Query("search", "")
	category := ctx.Query("category", "") // new category filter parameter

	// Convert pagination parameters to integers
	currentPage, err := strconv.Atoi(pageStr)
	if err != nil || currentPage < 1 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid currentPage"})
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid pageSize"})
	}

	// Fetch paginated galleries from service with optional category filter
	result, err := services.GetGalleryWithCategoryFilter(currentPage, pageSize, search, category)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch galleries"})
	}

	// Return paginated result
	return ctx.JSON(result)
}

// GroupByCategory handles grouping galleries by category
func (c *GalleryController) GroupByCategory(ctx *fiber.Ctx) error {
	// Get pagination parameters
	pageStr := ctx.Query("currentPage", "1")
	pageSizeStr := ctx.Query("pageSize", "10")
	search := ctx.Query("search", "")
	category := ctx.Query("category", "") // new category filter parameter

	// Convert pagination parameters to integers
	currentPage, err := strconv.Atoi(pageStr)
	if err != nil || currentPage < 1 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid currentPage"})
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid pageSize"})
	}

	// Fetch paginated galleries from service with optional category filter
	result, err := services.GroupByCategory(currentPage, pageSize, search, category)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch galleries"})
	}

	// Return paginated result
	return ctx.JSON(result)
}
