package controllers

import (
	"strconv"

	"backend-school/services"

	"github.com/gofiber/fiber/v2"
)

type MenuController struct {
	Service *services.MenuService
}

func NewMenuController() *MenuController {
	service := services.NewMenuService()
	return &MenuController{Service: service}
}

func (c *MenuController) GetMenus(ctx *fiber.Ctx) error {
	menus, err := c.Service.GetAllMenus()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return ctx.JSON(menus)
}

func (c *MenuController) GetMenuByID(ctx *fiber.Ctx) error {
	id, err := strconv.Atoi(ctx.Params("id"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID",
		})
	}

	menu, err := c.Service.GetMenuByID(uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return ctx.JSON(menu)
}
