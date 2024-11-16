// services/footer_service.go
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
	"time"

	"github.com/google/uuid"
)

type FooterService struct {
}

// NewFooterService creates a new FooterService
func NewFooterService() (*FooterService, error) {
	return &FooterService{}, nil
}

// GetAllFooters retrieves all footer data from the database
func (s *FooterService) GetAllFooters() ([]map[string]interface{}, error) {
	var footers []models.Footer

	// Fetch footers and preload their links
	err := config.DB.Preload("FooterLinks").Find(&footers).Error
	if err != nil {
		return nil, err
	}

	// Prepare the formatted data
	var formattedData []map[string]interface{}
	for _, footer := range footers {
		usefulLinks := []map[string]string{}
		ourCompany := []map[string]string{}
		socialLinks := []map[string]string{}

		for _, link := range footer.FooterLinks {
			if link.LinkType == "useful" {
				usefulLinks = append(usefulLinks, map[string]string{
					"text": link.Text,
					"link": link.Link,
				})
			} else if link.LinkType == "company" {
				ourCompany = append(ourCompany, map[string]string{
					"text": link.Text,
					"link": link.Link,
				})
			} else if link.LinkType == "social" {
				socialLinks = append(socialLinks, map[string]string{
					"icon": link.Icon,
					"link": link.Link,
				})
			}
		}

		// Format the data to the required structure
		formattedData = append(formattedData, map[string]interface{}{
			"description": footer.Description,
			"phone":       footer.Phone,
			"mail":        footer.Mail,
			"website":     footer.Website,
			"address":     footer.Address,
			"usefulLinks": usefulLinks,
			"ourCompany":  ourCompany,
			"socialLink":  socialLinks,
		})
	}

	return formattedData, nil
}

// GetFooterByID retrieves a footer by its ID
func (s *FooterService) GetFooterByID(id uint) (models.Footer, error) {
	var footer models.Footer
	if err := config.DB.Preload("FooterLinks").First(&footer, id).Error; err != nil {
		return footer, err
	}
	return footer, nil
}

// Admin Function
func (s *FooterService) GetFooterBySlug(slug string) (*models.Footer, error) {
	var footer models.Footer
	if err := config.DB.Where("deleted_at", nil).Where("slug = ?", slug).First(&footer).Error; err != nil {
		return nil, err
	}
	return &footer, nil
}

func (s *FooterService) GetFooterByUUID(uuidOrID string) (*models.FooterLink, error) {
	var footer models.FooterLink
	// Attempt to query by UUID or ID
	if err := config.DB.Where("deleted_at", nil).Where("uuid = ?", uuidOrID).First(&footer).Error; err != nil {
		return nil, err
	}
	return &footer, nil
}

// GetFooterPaginated fetches paginated footers with sorting
func (s *FooterService) GetFootersPaginated(perPage, page int, sortBy string, sortDesc bool) (map[string]interface{}, error) {
	var footers []models.FooterLink
	var totalRecords int64
	var sortOrder string

	// Calculate offset for pagination
	offset := (page - 1) * perPage

	// Set the sort order based on the `sortDesc` flag
	if sortDesc {
		sortOrder = sortBy + " ASC"
	} else {
		sortOrder = sortBy + " DESC"
	}

	// Get the total number of records
	config.DB.Model(&models.FooterLink{}).Count(&totalRecords)

	// Fetch the paginated data with sorting
	if err := config.DB.Where("deleted_at", nil).Order(sortOrder).Limit(perPage).Offset(offset).Find(&footers).Error; err != nil {
		return nil, errors.New("failed to fetch footers")
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalRecords) / float64(perPage)))

	// Return the data in a paginated format
	result := map[string]interface{}{
		"data":          footers,
		"current_page":  page,
		"per_page":      perPage,
		"total_pages":   totalPages,
		"total_records": totalRecords,
	}

	return result, nil
}

// CreateFooter handles the creation of a footer along with the image upload
func (s *FooterService) CreateFooter(footer *models.FooterLink, img *multipart.FileHeader) error {
	db, err := config.DB.DB()
	if err != nil {
		log.Printf("Failed to get database connection: %v", err)
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Ensure the sequence is correctly set to avoid conflicts
	if err := helpers.ResetSequenceToMax(db, "footers", "id", "footers_id_seq"); err != nil {
		log.Printf("Failed to reset sequence: %v", err)
		return fmt.Errorf("failed to reset sequence: %w", err)
	}

	newUUID := uuid.New().String()
	footer.UUID = newUUID

	// Save the footer to the database
	if err := config.DB.Create(footer).Error; err != nil {
		return fmt.Errorf("failed to create footer: %v", err)
	}

	return nil
}

// UpdateFooter updates an existing footer by its UUID and handles image uploads
func (s *FooterService) UpdateFooter(uuid string, updatedFooter *models.FooterLink, img *multipart.FileHeader) error {
	var footer models.FooterLink

	// Fetch the existing footer by UUID
	if err := config.DB.Where("uuid = ?", uuid).First(&footer).Error; err != nil {
		return errors.New("footer not found")
	}

	// Update the footer record with the new data
	if err := config.DB.Model(&footer).Updates(updatedFooter).Error; err != nil {
		return errors.New("failed to update footer")
	}

	return nil
}

// DeleteFooter deletes a footer by its slug
func (s *FooterService) DeleteFooter(uuid string) error {
	// Get the current time for the soft delete
	currentTime := time.Now()

	// Update the DeletedAt field instead of deleting the record
	if err := config.DB.Model(&models.FooterLink{}).Where("uuid = ?", uuid).Update("deleted_at", currentTime).Error; err != nil {
		return errors.New("failed to soft delete footer")
	}
	return nil
}
