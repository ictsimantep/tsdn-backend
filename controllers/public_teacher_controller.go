package controllers

import (
	"backend-school/models"
	"backend-school/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TeacherController struct {
	Service *services.TeacherService
}

func NewTeacherController() *TeacherController {
	return &TeacherController{
		Service: services.NewTeacherService(),
	}
}

func (c *TeacherController) GetTeachers(ctx *fiber.Ctx) error {
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

	result, err := c.Service.GetTeachers(currentPage, pageSize, search)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch teachers"})
	}

	return ctx.JSON(result)
}

func (c *TeacherController) CreateTeacher(ctx *fiber.Ctx) error {
	var teacher models.Teacher
	if err := ctx.BodyParser(&teacher); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := c.Service.CreateTeacher(teacher); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create teacher"})
	}

	return ctx.Status(fiber.StatusCreated).JSON(teacher)
}

func (c *TeacherController) GetTeacherByID(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	teacher, err := c.Service.GetTeacherByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Teacher not found"})
	}

	return ctx.JSON(teacher)
}

func (c *TeacherController) UpdateTeacher(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")

	// Konversi id dari string ke int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	var teacher models.Teacher
	if err := ctx.BodyParser(&teacher); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	teacher.ID = id
	if err := c.Service.UpdateTeacher(teacher); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update teacher"})
	}

	return ctx.JSON(teacher)
}

func (c *TeacherController) DeleteTeacher(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	if err := c.Service.DeleteTeacher(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not delete teacher"})
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
