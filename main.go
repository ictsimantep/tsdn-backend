package main

import (
	"backend-school/config"
	"backend-school/routes"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func init() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func main() {
	// Initialize MinIO client
	if err := config.InitializeMinioClient(); err != nil {
		log.Fatalf("Failed to initialize MinIO client: %v", err)
	}

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Initialize database and Casbin
	config.ConnectDatabase()
	config.ConnectCasbin(config.DB)

	allowedOrigins := os.Getenv("FRONTEND_ORIGINS")

	// Split the comma-separated origins into a slice
	allowedOriginsArray := strings.Split(allowedOrigins, ",")

	// Create Fiber app
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(allowedOriginsArray, ","),   // Allow requests from Next.js frontend
		AllowCredentials: true,                                     // Allow cookies and credentials
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS", // Specify allowed methods
	}))

	// Setup routes
	routes.SetupRoutes(app)

	// Start server
	app.Listen(":3000")
}
