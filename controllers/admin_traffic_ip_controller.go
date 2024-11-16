package controllers

import (
	"backend-school/middleware"
	"backend-school/services"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mileusna/useragent"
)

// TrafficIPController handles HTTP requests for logging traffic IP data.
type TrafficIPController struct {
	Service *services.TrafficIPService
}

// NewTrafficIPController creates a new instance of TrafficIPController.
func NewTrafficIPController(service *services.TrafficIPService) *TrafficIPController {
	return &TrafficIPController{Service: service}
}

// LogTrafficIP handles the POST request to log traffic IP details.
func (controller *TrafficIPController) LogTrafficIP(c *fiber.Ctx) error {
	// Use the middleware function to extract user_id from the token
	userID, err := middleware.ExtractUserIDFromToken(c) // Call the refactored helper function
	if err != nil {
		// If there is an error (e.g., token invalid), return Unauthorized
		log.Printf("User ID tidak ditemukan")
	}

	// Define a struct to bind the incoming JSON payload
	type TrafficIPPayload struct {
		URL       *string  `json:"url"`
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
	}

	var payload TrafficIPPayload
	// Parse the incoming JSON payload
	if err := c.BodyParser(&payload); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Check if the Url-Page header is present and set payload.URL if it exists
	urlPageHeader := c.Get("Url-Page")
	if urlPageHeader != "" {
		payload.URL = &urlPageHeader
	}

	// Get the client's IP address from the request headers
	ip := c.Get("X-Forwarded-For")
	if ip == "" {
		ip = c.IP() // Fallback to the default method to retrieve the IP
	}

	// Extract User-Agent to determine OS and Browser
	userAgent := c.Get("User-Agent")
	ua := useragent.Parse(userAgent)

	// Retrieve OS and Browser information
	os := ua.OS
	browser := ua.Name

	// Use the GetCountry method from the TrafficIPService instance to fetch the location details
	country, city, latitude, longitude, err := controller.Service.GetCountry(ip)
	if err != nil {
		// log.Printf("Failed to get location for IP %s: %v", ip, err)
		// return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get location details"})
	}

	// Dereference the URL pointer or use an empty string if it's nil
	var url string
	if payload.URL != nil {
		url = *payload.URL
	}

	if payload.Latitude != nil {
		latitude = strconv.FormatFloat(*payload.Latitude, 'f', -1, 64)
	}

	if payload.Longitude != nil {
		longitude = strconv.FormatFloat(*payload.Longitude, 'f', -1, 64)
	}

	// Use the TrafficIPService to log the traffic IP, including OS, Browser, and user_id
	if err := controller.Service.LogTrafficIP(ip, country, city, latitude, longitude, url, os, browser, &userID); err != nil {
		log.Printf("Error logging traffic IP: %v", err)
		if err.Error() == "IP has already been logged today" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()}) // 409 Conflict
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to log traffic IP"})
	}

	// Return a success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Traffic IP logged successfully"})
}
