package middleware

import (
	"backend-school/config"
	"backend-school/models"
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

// LoadCasbinPolicy memuat ulang kebijakan Casbin
func LoadCasbinPolicy() error {
	err := config.Enforcer.LoadPolicy()
	if err != nil {
		return errors.New("failed to load Casbin policy: " + err.Error())
	}
	return nil
}

// GetUsernameFromToken extracts the username from the JWT token in the Authorization header.
func GetUsernameFromToken(c *fiber.Ctx) (string, error) {
	// Retrieve the Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is missing")
	}

	// Split the header to get the token part
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	tokenString := parts[1]

	// Parse the token using the secret key
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure that the signing method is HMAC before continuing
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(config.SecretKey), nil
	})

	if err != nil {
		return "", err
	}

	// Extract the claims from the token and validate them
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token")
	}

	// Extract the username from the claims
	username, ok := claims["username"].(string)
	if !ok {
		return "", errors.New("username not found in token")
	}

	return username, nil
}

// JWTMiddleware validasi JWT, menyimpan username, user_id, dan role_guard_name di konteks
func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Muat ulang kebijakan Casbin sebelum mendapatkan role_guard_name
		if err := LoadCasbinPolicy(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"statusCode": fiber.StatusInternalServerError,
				"message":    "Could not load Casbin policy: " + err.Error(),
			})
		}

		// Extract username from the JWT token
		username, err := GetUsernameFromToken(c)
		if err != nil {
			// If there is an error (e.g., token invalid), return Unauthorized
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"statusCode": fiber.StatusUnauthorized,
				"message":    "Unauthorized: " + err.Error(),
			})
		}

		// Retrieve the user from the database based on the username
		var user models.User
		if err := config.DB.Where("username = ?", username).Where("deleted_at", nil).First(&user).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"statusCode": fiber.StatusInternalServerError,
				"message":    "Could not find user in database",
			})
		}

		// Dapatkan role_guard_name dari tabel casbin_rule berdasarkan username
		roleGuardName, err := GetRoleGuardNameFromCasbinRule(username)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"statusCode": fiber.StatusInternalServerError,
				"message":    "Could not retrieve role_guard_name: " + err.Error(),
			})
		}

		// Simpan username, user_id, dan role_guard_name di context
		c.Locals("username", username)
		c.Locals("user_id", int(user.ID)) // Convert uint to int
		c.Locals("role_guard_name", roleGuardName)

		// Lanjutkan ke middleware atau handler berikutnya
		return c.Next()
	}
}

// ExtractUserIDFromToken extracts the user_id from the JWT token in the Authorization header.
func ExtractUserIDFromToken(c *fiber.Ctx) (int, error) {
	// Retrieve the Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return 0, errors.New("authorization header is missing")
	}

	// Split the header to get the token part
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, errors.New("invalid authorization header format")
	}

	tokenString := parts[1]

	// Parse the token using the secret key
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure that the signing method is HMAC before continuing
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(config.SecretKey), nil
	})

	if err != nil {
		return 0, err
	}

	// Extract the claims from the token and validate them
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}

	// Extract the username and fetch user_id from database based on username
	username, ok := claims["username"].(string)
	if !ok {
		return 0, errors.New("username not found in token")
	}

	// Retrieve the user from the database based on the username
	var user models.User
	if err := config.DB.Where("username = ?", username).Where("deleted_at", nil).First(&user).Error; err != nil {
		return 0, errors.New("could not find user in database")
	}

	return int(user.ID), nil // Convert user.ID (uint) to int and return
}

func GetRoleGuardNameFromCasbinRule(username string) (string, error) {
	var rule models.CasbinRule

	// Cetak pesan log untuk memastikan query yang dijalankan benar
	// fmt.Println("Mencari role_guard_name untuk username:", username)

	// Query untuk menemukan entri di mana v0 sesuai dengan username dan ptype adalah "g"
	err := config.DB.Where("v0 = ? AND ptype = ?", username, "g").First(&rule).Error
	if err != nil {
		fmt.Println("Kesalahan saat mencari casbin_rule:", err)
		return "", errors.New("role_guard_name tidak ditemukan untuk user")
	}

	// Memastikan nilai v1 yang seharusnya berisi role_guard_name
	// fmt.Println("Role_guard_name ditemukan:", rule.V1)

	// Mengembalikan role_guard_name
	return rule.V1, nil
}
