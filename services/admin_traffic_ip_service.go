package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"backend-school/config"
	"backend-school/models"
)

// TrafficIPService provides methods for managing traffic IP data.
type TrafficIPService struct{}

// NewTrafficIPService creates a new instance of TrafficIPService.
func NewTrafficIPService() *TrafficIPService {
	return &TrafficIPService{}
}

// GetCountry retrieves the country, city, latitude, and longitude based on the IP address,
// checking the database first to avoid unnecessary API calls.
func (s *TrafficIPService) GetCountry(ip string) (string, string, string, string, error) {
	var existingIP models.TrafficIP

	// Check if the IP already exists in the database
	if err := config.DB.Where("ip = ?", ip).First(&existingIP).Error; err == nil {
		// IP exists, return the details from the database
		return existingIP.Country, existingIP.City, existingIP.Latitude, existingIP.Longitude, nil
	}

	apiKey := os.Getenv("API_KEY_FINDIP")
	baseURL := os.Getenv("URL_FINDIP")
	url := fmt.Sprintf("%s/IP_Address/?token=%s&ip=%s", baseURL, apiKey, ip)

	// Create a request to the FindIP.net API
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", "", "", err
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", "", errors.New("failed to get country information")
	}

	var response struct {
		Country struct {
			Names map[string]string `json:"names"`
		} `json:"country"`
		City struct {
			Names map[string]string `json:"names"`
		} `json:"city"`
		Location struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Error decoding response: %v", err)
		return "", "", "", "", err
	}

	countryName := response.Country.Names["en"]
	if countryName == "" {
		for _, name := range response.Country.Names {
			countryName = name
			break
		}
	}

	cityName := response.City.Names["en"]
	if cityName == "" {
		for _, name := range response.City.Names {
			cityName = name
			break
		}
	}

	return countryName, cityName, fmt.Sprintf("%.6f", response.Location.Latitude), fmt.Sprintf("%.6f", response.Location.Longitude), nil
}

// LogTrafficIP logs traffic IP details to the database.
func (s *TrafficIPService) LogTrafficIP(ip, country, city, latitude, longitude, url, os, browser string, userID *int) error {
	// Check if the IP already exists for today with the same URL, OS, and Browser
	var existingIP models.TrafficIP
	currentDate := time.Now().Format("2006-01-02")

	query := config.DB.
		Where("url = ?", url).
		Where("os = ?", os).
		Where("browser = ?", browser).
		Where("ip = ? AND DATE(created_at) = ?", ip, currentDate)

	// If userID is provided, include it in the query to check for a unique log per user
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	if err := query.First(&existingIP).Error; err == nil {
		log.Printf("IP %s has already been logged today for this user.", ip)
		return errors.New("IP has already been logged today")
	}

	// Create new record for traffic IP logging
	newTrafficIP := models.TrafficIP{
		IP:        ip,
		Country:   country,
		City:      city,
		Latitude:  latitude,
		Longitude: longitude,
		Url:       url,
		Os:        os,
		Browser:   browser,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set userID if provided
	if userID != nil {
		newTrafficIP.UserID = *userID
	}

	// Insert the new record into the database
	if err := config.DB.Create(&newTrafficIP).Error; err != nil {
		log.Printf("Failed to insert IP %s into the database: %v", ip, err)
		return err
	}

	log.Printf("Successfully inserted IP %s with country %s, city %s, latitude %s, longitude %s for user %v into the database.", ip, country, city, latitude, longitude, userID)
	return nil
}
