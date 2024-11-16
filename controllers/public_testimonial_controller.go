package controllers

import (
	"backend-school/models"
	"backend-school/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TestimonialController struct {
	Service *services.TestimonialService
}

func NewTestimonialController() *TestimonialController {
	return &TestimonialController{
		Service: services.NewTestimonialService(),
	}
}

func (c *TestimonialController) GetTestimonials(ctx *fiber.Ctx) error {
	// Get the username from the context (set by the JWT middleware)

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

	result, err := c.Service.GetTestimonials(currentPage, pageSize, search)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch testimonials"})
	}

	return ctx.JSON(result)
}

func (c *TestimonialController) CreateTestimonial(ctx *fiber.Ctx) error {

	var testimonial models.Testimonial
	if err := ctx.BodyParser(&testimonial); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := c.Service.CreateTestimonial(testimonial); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create testimonial"})
	}

	return ctx.Status(fiber.StatusCreated).JSON(testimonial)
}

func (c *TestimonialController) GetTestimonialByID(ctx *fiber.Ctx) error {

	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	testimonial, err := c.Service.GetTestimonialByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Testimonial not found"})
	}

	return ctx.JSON(testimonial)
}

func (c *TestimonialController) UpdateTestimonial(ctx *fiber.Ctx) error {

	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	var testimonial models.Testimonial
	if err := ctx.BodyParser(&testimonial); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	testimonial.UUID = id
	if err := c.Service.UpdateTestimonial(testimonial); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update testimonial"})
	}

	return ctx.JSON(testimonial)
}

func (c *TestimonialController) DeleteTestimonial(ctx *fiber.Ctx) error {

	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	if err := c.Service.DeleteTestimonial(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not delete testimonial"})
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
