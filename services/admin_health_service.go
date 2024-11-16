package services

import (
	"backend-school/config"
	"backend-school/models"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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

	offset := (currentPage - 1) * pageSize

	// Query dasar untuk document type
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

func (s *HealthService) AddHealth(payload *HealthPayload) (*models.Health, error) {
	var Health models.Health

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// Create the  entry
		const Risk = ""
		const Bmi = ""
		const RecommendationFood = ""
		const RecommendationSport = ""
		const RecommendationMedicine = ""

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
			Risk:                   Risk,
			Bmi:                    Bmi,
			RecommendationFood:     RecommendationFood,
			RecommendationSport:    RecommendationSport,
			RecommendationMedicine: RecommendationMedicine,
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

	// Fetch the document type with RoleHasRules
	if err := config.DB.
		Preload("RoleHasRules").
		Where("uuid = ?", uuid).
		First(&Health).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("document type not found")
		}
		return nil, fmt.Errorf("failed to fetch document type: %w", err)
	}

	return &Health, nil
}

// DeleteHealth soft deletes a document type by setting the deleted_at timestamp
func (s *HealthService) DeleteHealth(uuid string) error {
	currentTime := time.Now()

	// Check if document type with UUID exists
	var Health models.Health
	if err := config.DB.Where("uuid = ?", uuid).First(&Health).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("document type not found")
		}
		return errors.New("failed to check document type")
	}

	// Soft delete by setting deleted_at timestamp
	if err := config.DB.Model(&Health).Update("deleted_at", currentTime).Error; err != nil {
		return errors.New("failed to soft delete document type")
	}

	return nil
}
