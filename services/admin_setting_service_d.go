package services

import (
	"backend-school/config"
	"backend-school/models"
	"errors"
	"math"

	"crypto/rand"
	"encoding/hex"
)

type SettingsService struct{}

// GenerateRandomString generates a random string of a given length
func GenerateRandomStringSetting(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func NewSettingsService() *SettingsService {
	return &SettingsService{}
}

func (s *SettingsService) GetSetting(key string) (string, error) {
	var setting struct {
		Value string
	}
	if err := config.DB.Table("settings").Select("value").Where("key = ?", key).First(&setting).Error; err != nil {
		return "", errors.New("setting not found")
	}
	return setting.Value, nil
}

func (s *SettingsService) GetSettingBySlug(slug string) (*models.Setting, error) {
	var setting models.Setting
	if err := config.DB.Preload("Category").Where("slug = ?", slug).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

func (s *SettingsService) GetSettingByUUID(uuid string) (*models.Setting, error) {
	var setting models.Setting
	if err := config.DB.Preload("Category").Where("uuid = ?", uuid).First(&setting).Error; err != nil {
		return nil, err
	}
	return &setting, nil
}

// GetSettingsPaginated fetches paginated settings with sorting
func (s *SettingsService) GetSettingsPaginated(perPage, page int, sortBy string, sortDesc bool) (map[string]interface{}, error) {
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
