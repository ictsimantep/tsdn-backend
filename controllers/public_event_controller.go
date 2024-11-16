package controllers

import (
	"backend-school/services"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type EventController struct {
	Service *services.EventService
}

func NewEventController() *EventController {
	eventService, err := services.NewEventService()
	if err != nil {
		log.Fatalf("Failed to initialize event service: %v", err) // Log and stop execution if there's an error
	}
	return &EventController{
		Service: eventService,
	}
}

func (c *EventController) GetEvents(ctx *fiber.Ctx) error {
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

	result, err := c.Service.GetEvents(currentPage, pageSize, search)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch events"})
	}

	return ctx.JSON(result)
}

func (c *EventController) Search(ctx *fiber.Ctx) error {
	search := ctx.Query("slug", "")
	event, err := c.Service.SearchBySlug(search)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Event not found"})
	}

	return ctx.JSON(event)
}
