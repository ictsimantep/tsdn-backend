package handlers

import (
	"backend-school/config"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"
)

func AuthMiddleware() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey:  []byte(config.SecretKey),
		TokenLookup: "header:Authorization", // Ensure token is looked up in the Authorization header
		// Default is Bearer token scheme
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized",
			})
		},
	})
}

func Authorize() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// token, ok := c.Locals("user").(*jwt.Token)
		// if !ok {
		// 	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		// 		"message": "Unauthorized",
		// 	})
		// }

		// claims, ok := token.Claims.(jwt.MapClaims)
		// if !ok {
		// 	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		// 		"message": "Unauthorized",
		// 	})
		// }

		// username, ok := claims["username"].(string)
		// if !ok {
		// 	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		// 		"message": "Unauthorized",
		// 	})
		// }

		// // Enforce policy and handle both values
		// allowed, err := config.Enforcer.Enforce(username, c.Path(), c.Method())
		// if err != nil {
		// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		// 		"message": "Error authorizing request",
		// 	})
		// }

		// if !allowed {
		// 	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
		// 		"message": "Forbidden",
		// 	})
		// }

		return c.Next()
	}
}
