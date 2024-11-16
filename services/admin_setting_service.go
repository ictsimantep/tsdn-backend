package services

import (
	"backend-school/config"
	"backend-school/helpers"
	"backend-school/models"
	"errors"
	"fmt"
	"log"
	"math"
	"mime/multipart"

	"crypto/rand"
	"encoding/hex"
)

type AdminSettingsService struct{}

// GenerateRandomString generates a random string of a given length
func GenerateRandomStringAdminSetting(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func NewAdminSettingsService() (*AdminSettingsService, error) {
	return &AdminSettingsService{}, nil
}

func (s *AdminSettingsService) GetAdminSetting(key string) (string, error) {
	var setting struct {
		Value string
	}
	if err := config.DB.Table("settings").Select("value").Where("key = ?", key).First(&setting).Error; err != nil {
		return "", errors.New("setting not found")
	}
	return setting.Value, nil
}

func (s *AdminSettingsService) GetAdminSettingBySlug(slug string) (*models.Setting, error) {
	var setting models.Setting
	if err := config.DB.Where("slug = ?", slug).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

func (s *AdminSettingsService) GetAdminSettingByUUID(uuid string) (*models.Setting, error) {
	var setting models.Setting
	if err := config.DB.Where("uuid = ?", uuid).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

// GetAdminSettingsPaginated fetches paginated settings with sorting
func (s *AdminSettingsService) GetAdminSettingsPaginated(perPage, page int, sortBy string, sortDesc bool) (map[string]interface{}, error) {
	var settings []models.Setting
	var totalRecords int64
	var sortOrder string

	// Calculate offset for pagination
	offset := (page - 1) * perPage

	// Set the sort order based on the `sortDesc` flag
	if sortDesc {
		sortOrder = sortBy + " DESC"
	} else {
		sortOrder = sortBy + " ASC"
	}

	// Get the total number of records
	config.DB.Model(&models.Setting{}).Count(&totalRecords)

	// Fetch the paginated data with sorting
	if err := config.DB.Order(sortOrder).Limit(perPage).Offset(offset).Find(&settings).Error; err != nil {
		return nil, errors.New("failed to fetch settings")
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalRecords) / float64(perPage)))

	// Return the data in a paginated format
	result := map[string]interface{}{
		"data":          settings,
		"current_page":  page,
		"per_page":      perPage,
		"total_pages":   totalPages,
		"total_records": totalRecords,
	}

	return result, nil
}

// CreateSetting handles the creation of a setting along with the image upload
func (s *AdminSettingsService) CreateSetting(setting *models.Setting, img *multipart.FileHeader) error {
	db, err := config.DB.DB()
	if err != nil {
		log.Printf("Failed to get database connection: %v", err)
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Ensure the sequence is correctly set to avoid conflicts
	if err := helpers.ResetSequenceToMax(db, "settings", "id", "settings_id_seq"); err != nil {
		log.Printf("Failed to reset sequence: %v", err)
		return fmt.Errorf("failed to reset sequence: %w", err)
	}
	// Save the setting to the database
	if err := config.DB.Create(setting).Error; err != nil {
		return fmt.Errorf("failed to create setting: %v", err)
	}

	return nil
}

// UpdateSetting updates an existing setting by its UUID and handles image uploads
func (s *AdminSettingsService) UpdateSetting(uuid string, updatedSetting *models.Setting, img *multipart.FileHeader) error {
	var setting models.Setting

	// Fetch the existing setting by UUID
	if err := config.DB.Where("uuid = ?", uuid).First(&setting).Error; err != nil {
		return errors.New("setting not found")
	}

	// Update the setting record with the new data
	if err := config.DB.Model(&setting).Updates(updatedSetting).Error; err != nil {
		return errors.New("failed to update setting")
	}

	return nil
}

// DeleteSetting deletes a setting by its slug
func (s *AdminSettingsService) DeleteSetting(uuid string) error {
	if err := config.DB.Where("uuid = ?", uuid).Delete(&models.Setting{}).Error; err != nil {
		return errors.New("failed to delete setting")
	}
	return nil
}
