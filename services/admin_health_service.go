package services

import (
	"backend-school/config"
	"backend-school/models"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RiskResponse struct {
	Risk   string `json:"risk"`
	Sample int    `json:"sample"`
}

type HealthService struct{}

// NewHealthService initializes a new HealthService
func NewHealthService() *HealthService {
	return &HealthService{}
}

// defines the structure for the create request payload
type HealthPayload struct {
	Nama         string `json:"nama"`
	JenisKelamin string `json:"jenis_kelamin"`
	Umur         string `json:"umur"`
	Bb           string `json:"bb"`
	Tb           string `json:"tb"`
	Systol       string `json:"systol"`
	Diastol      string `json:"diastol"`
	HeartRate    string `json:"heart_rate"`
	Profesi      string `json:"profesi"`
}

func (s *HealthService) GetHealthsPaginated(currentPage, pageSize int, search string, showAll string) (map[string]interface{}, error) {
	var Healths []models.Health
	var totalRecords int64

	offset := (currentPage) * pageSize

	// Query dasar untuk health data
	query := config.DB.Model(&models.Health{}).Where("deleted_at IS NULL")
	if search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}

	// Hitung total dokumen yang sesuai dengan query
	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, errors.New("failed to count health data")
	}

	// Ambil data health data dengan paginasi
	if showAll == "" {
		if err := query.Offset(offset).Limit(pageSize).Find(&Healths).Error; err != nil {
			return nil, errors.New("failed to fetch health data")
		}
	} else {
		if err := query.Find(&Healths).Error; err != nil {
			return nil, errors.New("failed to fetch health data")
		}
	}

	// Hitung total halaman
	totalPages := int(math.Ceil(float64(totalRecords) / float64(pageSize)))

	// Menyiapkan hasil dalam bentuk map untuk JSON response
	result := map[string]interface{}{
		"data":          Healths,
		"current_page":  currentPage,
		"per_page":      pageSize,
		"total_pages":   totalPages,
		"total_records": totalRecords,
	}

	return result, nil
}

func CalculateBMI(weight, height string) (float64, error) {
	// Convert weight and height to float64
	w, err := strconv.ParseFloat(weight, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid weight: %v", err)
	}

	h, err := strconv.ParseFloat(height, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid height: %v", err)
	}

	if h <= 0 {
		return 0, fmt.Errorf("height must be greater than zero")
	}

	// Calculate BMI (Height is converted to meters)
	hMeters := h / 100
	bmi := w / math.Pow(hMeters, 2)
	return bmi, nil
}

// GenerateRecommendations generates food, sport, and medicine recommendations based on risk and BMI.
func GenerateRecommendations(risk string, bmi float64) (string, string, string) {
	var food, sport, medicine string

	// Generate recommendations based on BMI and risk
	switch {
	case bmi < 18.5:
		food = "High-calorie meals, including nuts, dairy, and lean proteins."
		sport = "Light strength training or yoga."
		medicine = "Consult a doctor for potential supplements."
	case bmi >= 18.5 && bmi < 24.9:
		food = "Balanced diet with fruits, vegetables, and lean proteins."
		sport = "Regular cardio like running or cycling."
		medicine = "No specific medication needed."
	case bmi >= 25 && bmi < 29.9:
		food = "Low-carb meals with controlled portion sizes."
		sport = "Cardio combined with strength training."
		medicine = "Consider checking for blood sugar levels."
	case bmi >= 30:
		food = "Low-fat, high-fiber diet; avoid sugar and processed foods."
		sport = "Low-impact exercises like swimming or brisk walking."
		medicine = "Consult a doctor for weight management options."
	}

	// Adjust recommendations based on risk
	switch risk {
	case "Low Risk":
		// Keep general advice
	case "Medium Risk":
		food += " Reduce sodium intake and avoid junk food."
		sport += " Include flexibility exercises."
		medicine += " Monitor health parameters regularly."
	case "High Risk":
		food += " Strictly avoid salty and sugary foods."
		sport = "Consult a physician before starting an exercise routine."
		medicine = "Medical intervention may be necessary."
	}

	return food, sport, medicine
}

// CalculateRisk sends a POST request to the external API to get the risk level.
func CalculateRisk(systol, diastol, heartRate string) (string, error) {
	// Get the URL from .env
	apiURL := os.Getenv("RISK_API_URL")
	if apiURL == "" {
		return "", errors.New("RISK_API_URL is not set in the environment variables")
	}

	// Prepare the request payload
	payload := map[string]interface{}{
		"systol":     systol,
		"diastol":    diastol,
		"heart_rate": heartRate,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %v", err)
	}

	// Send the POST request
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to send request to %s: %v", apiURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("received non-200 response: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var riskResponse []RiskResponse
	if err := json.NewDecoder(resp.Body).Decode(&riskResponse); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	if len(riskResponse) > 0 {
		return riskResponse[0].Risk, nil
	}

	return "", errors.New("no risk data found in response")
}

func (s *HealthService) AddHealth(payload *HealthPayload) (*models.Health, error) {
	var Health models.Health

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// Create the  entry
		risk, _ := CalculateRisk(payload.Systol, payload.Diastol, payload.HeartRate)
		bmi, _ := CalculateBMI(payload.Bb, payload.Tb)

		// Generate recommendations
		food, sport, medicine := GenerateRecommendations(risk, bmi)

		Health = models.Health{
			UUID:                   uuid.New(),
			Nama:                   payload.Nama,
			JenisKelamin:           payload.JenisKelamin,
			Umur:                   payload.Umur,
			Bb:                     payload.Bb,
			Tb:                     payload.Tb,
			Systol:                 payload.Systol,
			Diastol:                payload.Diastol,
			HeartRate:              payload.HeartRate,
			Profesi:                payload.Profesi,
			Risk:                   risk,
			Bmi:                    fmt.Sprintf("%.2f", bmi),
			RecommendationFood:     food,
			RecommendationSport:    sport,
			RecommendationMedicine: medicine,
		}

		if err := tx.Create(&Health).Error; err != nil {
			return fmt.Errorf("failed to create health data: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	return &Health, nil
}

// GetHealthByUUID fetches a Health by UUID
func (s *HealthService) GetHealthByUUID(uuid string) (*models.Health, error) {
	var Health models.Health

	// Fetch the health data with RoleHasRules
	if err := config.DB.
		Preload("RoleHasRules").
		Where("uuid = ?", uuid).
		First(&Health).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("health data not found")
		}
		return nil, fmt.Errorf("failed to fetch health data: %w", err)
	}

	return &Health, nil
}

// DeleteHealth soft deletes a health data by setting the deleted_at timestamp
func (s *HealthService) DeleteHealth(uuid string) error {
	currentTime := time.Now()

	// Check if health data with UUID exists
	var Health models.Health
	if err := config.DB.Where("uuid = ?", uuid).First(&Health).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("health data not found")
		}
		return errors.New("failed to check health data")
	}

	// Soft delete by setting deleted_at timestamp
	if err := config.DB.Model(&Health).Update("deleted_at", currentTime).Error; err != nil {
		return errors.New("failed to soft delete health data")
	}

	return nil
}
