package controllers

import (
	"backend-school/models"
	"backend-school/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type PersonController struct {
	Service *services.PersonService
}

func NewPersonController() *PersonController {
	return &PersonController{
		Service: services.NewPersonService(),
	}
}

func (c *PersonController) GetPeople(ctx *fiber.Ctx) error {
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

	result, err := c.Service.GetPeople(currentPage, pageSize, search)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch people"})
	}

	return ctx.JSON(result)
}

func (c *PersonController) CreatePerson(ctx *fiber.Ctx) error {
	var person models.Person
	if err := ctx.BodyParser(&person); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := c.Service.CreatePerson(person); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create person"})
	}

	return ctx.Status(fiber.StatusCreated).JSON(person)
}

func (c *PersonController) GetPersonByID(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	person, err := c.Service.GetPersonByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Person not found"})
	}

	return ctx.JSON(person)
}

func (c *PersonController) UpdatePerson(ctx *fiber.Ctx) error {
	idStr := ctx.Params("id")

	// Konversi id dari string ke int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	var person models.Person
	if err := ctx.BodyParser(&person); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	person.ID = id // pastikan `person.ID` adalah tipe int
	if err := c.Service.UpdatePerson(person); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not update person"})
	}

	return ctx.JSON(person)
}

func (c *PersonController) DeletePerson(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	if err := c.Service.DeletePerson(id); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not delete person"})
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}
