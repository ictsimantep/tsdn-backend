package controllers

import (
	"backend-school/services"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type BannerController struct {
	BannerService *services.BannerService
}

func NewBannerController() *BannerController {
	bannerService, err := services.NewBannerService()
	if err != nil {
		log.Fatalf("Failed to initialize publication service: %v", err) // Log and stop execution if there's an error
	}
	return &BannerController{BannerService: bannerService}
}

func (c *BannerController) GetBanners(ctx *fiber.Ctx) error {
	pageStr := ctx.Query("currentPage", "1")
	pageSizeStr := ctx.Query("pageSize", "10")
	search := ctx.Query("search", "")

	currentPage, err := strconv.Atoi(pageStr)
	if err != nil || currentPage < 1 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid currentPage"})
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid pageSize"})
	}

	result, err := c.BannerService.GetBanners(currentPage, pageSize, search)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch banners"})
	}

	return ctx.JSON(result)
}
