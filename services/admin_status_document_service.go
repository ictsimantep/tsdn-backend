package services

import (
	"backend-school/config"
	"backend-school/helpers"
	"backend-school/models"
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StatusDocumentService struct{}

// Initialize StatusDocumentService
func NewStatusDocumentService() *StatusDocumentService {
	return &StatusDocumentService{}
}

// GetStatusDocumentByUUID fetches a StatusDocument based on UUID
func (s *StatusDocumentService) GetStatusDocumentByUUID(uuid string) (*models.StatusDocument, error) {
	var statusDocument models.StatusDocument
	if err := config.DB.Where("uuid = ?", uuid).Where("deleted_at IS NULL").First(&statusDocument).Error; err != nil {
		return nil, err
	}
	return &statusDocument, nil
}

// GetStatusDocumentsPaginated fetches paginated status documents with optional filtering by name
func (s *StatusDocumentService) GetStatusDocumentsPaginated(currentPage, pageSize int, search string) (map[string]interface{}, error) {
	var statusDocuments []models.StatusDocument
	var totalRecords int64

	offset := (currentPage - 1) * pageSize

	query := config.DB.Model(&models.StatusDocument{}).Where("deleted_at IS NULL")
	if search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, errors.New("failed to count status documents")
	}

	if err := query.Offset(offset).Limit(pageSize).Find(&statusDocuments).Error; err != nil {
		return nil, errors.New("failed to fetch status documents")
	}

	totalPages := int(math.Ceil(float64(totalRecords) / float64(pageSize)))

	result := map[string]interface{}{
		"data":          statusDocuments,
		"current_page":  currentPage,
		"per_page":      pageSize,
		"total_pages":   totalPages,
		"total_records": totalRecords,
	}

	return result, nil
}

// AddStatusDocument creates a new status document
func (s *StatusDocumentService) AddStatusDocument(statusDocument *models.StatusDocument) error {
	db, err := config.DB.DB()
	if err != nil {
		log.Printf("Failed to get database connection: %v", err)
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Ensure the sequence is correctly set to avoid conflicts
	if err := helpers.ResetSequenceToMax(db, "status_document", "id", "status_document_id_seq"); err != nil {
		log.Printf("Failed to reset sequence: %v", err)
		return fmt.Errorf("failed to reset sequence: %w", err)
	}
	// Ensure UUID is set
	statusDocument.UUID = uuid.New()

	// Insert the new status document into the database
	if err := config.DB.Create(statusDocument).Error; err != nil {
		return fmt.Errorf("failed to create status document in database: %w", err)
	}

	return nil
}

// UpdateStatusDocumentByUUID updates a status document's information by UUID
func (s *StatusDocumentService) UpdateStatusDocumentByUUID(uuid string, updatedStatusDocument *models.StatusDocument) (*models.StatusDocument, error) {
	var statusDocument models.StatusDocument

	// Find status document by UUID
	if err := config.DB.Where("uuid = ?", uuid).First(&statusDocument).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("status document not found")
		}
		return nil, fmt.Errorf("failed to find status document: %w", err)
	}

	// Update fields of the existing status document
	statusDocument.Name = updatedStatusDocument.Name
	statusDocument.UpdatedAt = time.Now()

	if err := config.DB.Save(&statusDocument).Error; err != nil {
		return nil, fmt.Errorf("failed to update status document: %w", err)
	}

	return &statusDocument, nil
}

// DeleteStatusDocument soft deletes a status document by setting the deleted_at timestamp
func (s *StatusDocumentService) DeleteStatusDocument(uuid string) error {
	currentTime := time.Now()

	// Check if status document with UUID exists
	var statusDocument models.StatusDocument
	if err := config.DB.Where("uuid = ?", uuid).First(&statusDocument).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("status document not found")
		}
		return errors.New("failed to check status document")
	}

	// Soft delete by setting deleted_at timestamp
	if err := config.DB.Model(&statusDocument).Update("deleted_at", currentTime).Error; err != nil {
		return errors.New("failed to soft delete status document")
	}

	return nil
}
