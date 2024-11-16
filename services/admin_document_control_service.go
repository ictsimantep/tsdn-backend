package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"backend-school/config"
	"backend-school/models"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type DocumentControlService struct {
	minioClient  *minio.Client
	bucketName   string
	minioService *MinioService
}

// NewDocumentControlService initializes and returns a new instance of DocumentControlService
func NewDocumentControlService(minioClient *minio.Client, bucketName string, minioService *MinioService) *DocumentControlService {
	return &DocumentControlService{
		minioClient:  minioClient,
		bucketName:   bucketName,
		minioService: minioService,
	}
}

// Helper function to convert int to *int
func IntPtr(i int) *int {
	return &i
}

var validate = validator.New() // Initialize validator instance

// DocumentControlPayload defines the structure for the create and update request payload
type DocumentControlPayload struct {
	DocumentName       string `json:"document_name" form:"document_name" validate:"required"`
	Description        string `json:"description" form:"description" validate:"required"`
	DocumentNumber     string `json:"document_number" form:"document_number" validate:"required"`
	ClauseNumber       string `json:"clause_number" form:"clause_number"`
	RevisionNumber     int    `json:"revision_number" form:"revision_number"`
	PublishDate        string `json:"publish_date" form:"publish_date" validate:"required,datetime=2006-01-02"`
	PageCount          int    `json:"page_count" form:"page_count"`
	DocumentTypeID     int    `json:"document_type_id" form:"document_type_id" validate:"required"`
	DocumentCategoryID int    `json:"document_category_id" form:"document_category_id" validate:"required"`
	SequenceNumber     int    `json:"sequence_number" form:"sequence_number"`
	StatusDocumentID   int    `json:"status_document_id" form:"status_document_id" validate:"required"`
	Version            int    `json:"version" form:"version"`
	CreatedBy          int    `json:"created_by"`
}

// PaginatedResult is a struct to hold paginated data
type PaginatedResult struct {
	Data         interface{} `json:"data"`
	CurrentPage  int         `json:"current_page"`
	PerPage      int         `json:"per_page"`
	TotalPages   int         `json:"total_pages"`
	TotalRecords int64       `json:"total_records"`
}

type DocumentControlResponse struct {
	ID                 int        `json:"id"`
	UUID               uuid.UUID  `json:"uuid"`
	DocumentName       string     `json:"document_name"`
	Description        string     `json:"description"`
	DocumentNumber     string     `json:"document_number"`
	ClauseNumber       string     `json:"clause_number"`
	RevisionNumber     int        `json:"revision_number"`
	PublishDate        string     `json:"publish_date"` // Format this as Y-m-d
	PageCount          int        `json:"page_count"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at"`
	DocumentTypeID     *int       `json:"document_type_id"`
	DocumentCategoryID *int       `json:"document_category_id"`
	SequenceNumber     *int       `json:"sequence_number"`
	StatusDocumentID   *int       `json:"status_document_id"`
	CreatedBy          *int       `json:"created_by"`

	// Fields for joined data
	CategoryName   string `json:"category_name"`
	CategoryPrefix string `json:"category_prefix"`
	TypeName       string `json:"type_name"`
	TypePrefix     string `json:"type_prefix"`
	StatusName     string `json:"status_name"`
}

// Function to convert DocumentControl to DocumentControlResponse
func formatDocumentControl(doc models.DocumentControlJoined) DocumentControlResponse {
	return DocumentControlResponse{
		ID:                 doc.ID,
		UUID:               doc.UUID,
		DocumentName:       doc.DocumentName,
		Description:        doc.Description,
		DocumentNumber:     doc.DocumentNumber,
		ClauseNumber:       doc.ClauseNumber,
		RevisionNumber:     doc.RevisionNumber,
		PublishDate:        doc.PublishDate.Format("2006-01-02"), // Format as Y-m-d
		PageCount:          doc.PageCount,
		CreatedAt:          doc.CreatedAt,
		UpdatedAt:          doc.UpdatedAt,
		DeletedAt:          doc.DeletedAt,
		DocumentTypeID:     doc.DocumentTypeID,
		DocumentCategoryID: doc.DocumentCategoryID,
		SequenceNumber:     doc.SequenceNumber,
		StatusDocumentID:   doc.StatusDocumentID,
		CreatedBy:          doc.CreatedBy,
		CategoryName:       doc.CategoryName,
		CategoryPrefix:     doc.CategoryPrefix,
		TypeName:           doc.TypeName,
		TypePrefix:         doc.TypePrefix,
		StatusName:         doc.StatusName,
	}
}

func (s *DocumentControlService) AddDocumentControlWithVersion(payload *DocumentControlPayload, fileHeader *multipart.FileHeader) (*models.DocumentControl, *models.DocumentVersion, error) {
	// Validate and parse publish date
	if payload.PublishDate == "" {
		return nil, nil, fmt.Errorf("publish_date is required and cannot be empty")
	}

	publishDate, err := time.Parse("2006-01-02", payload.PublishDate)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid date format for publish_date: expected format is YYYY-MM-DD")
	}

	// Initialize DocumentControl instance
	documentControl := models.DocumentControl{
		UUID:               uuid.New(),
		DocumentName:       payload.DocumentName,
		Description:        payload.Description,
		DocumentNumber:     payload.DocumentNumber,
		ClauseNumber:       payload.ClauseNumber,
		RevisionNumber:     payload.RevisionNumber,
		PublishDate:        publishDate,
		PageCount:          payload.PageCount,
		DocumentTypeID:     IntPtr(payload.DocumentTypeID),
		DocumentCategoryID: IntPtr(payload.DocumentCategoryID),
		SequenceNumber:     IntPtr(payload.SequenceNumber),
		StatusDocumentID:   IntPtr(1),
		CreatedBy:          IntPtr(payload.CreatedBy),
	}

	// Initialize an empty DocumentVersion instance outside the transaction block
	var documentVersion models.DocumentVersion

	// Start a transaction to ensure atomicity
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		// Step 1: Save DocumentControl to the database
		if err := tx.Create(&documentControl).Error; err != nil {
			return fmt.Errorf("failed to create document control: %w", err)
		}

		// Step 2: Upload file to MinIO
		filePath, err := s.uploadFileToMinio(fileHeader, "document-versions")
		if err != nil {
			// Capture the detailed error from uploadToMinio
			return fmt.Errorf("failed to upload file to MinIO: %w", err)
		}

		// Step 3: Initialize DocumentVersion for the initial version
		documentVersion = models.DocumentVersion{
			UUID:              uuid.New(),
			File:              filePath,
			DocumentControlID: &documentControl.ID,
			Version:           IntPtr(payload.Version), // Initial version number
			StatusDocumentID:  IntPtr(1),
			Note:              "Initial version",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		// Step 4: Save DocumentVersion to the database
		if err := tx.Create(&documentVersion).Error; err != nil {
			return fmt.Errorf("failed to create document version: %w", err)
		}

		return nil
	})

	// If the transaction fails, return the detailed error to the caller
	if err != nil {
		return nil, nil, fmt.Errorf("transaction error: %w", err)
	}

	return &documentControl, &documentVersion, nil
}

func (s *DocumentControlService) uploadFileToMinio(file *multipart.FileHeader, directory string) (string, error) {
	// Check if MinIO client is initialized
	if s.minioClient == nil {
		log.Println("Error: MinIO client is not initialized")
		return "", fmt.Errorf("MinIO client is not initialized")
	}

	// Validate file
	const maxFileSize = 10 * 1024 * 1024 // 10 MB limit
	if file.Size > maxFileSize {
		return "", fmt.Errorf("file size exceeds maximum limit of %d bytes", maxFileSize)
	}

	allowedMimeTypes := map[string]bool{
		"application/pdf": true,
		"image/jpeg":      true,
		"image/png":       true,
	}
	contentType := file.Header.Get("Content-Type")
	if !allowedMimeTypes[contentType] {
		return "", fmt.Errorf("unsupported file type: %s", contentType)
	}

	// Open file and prepare unique filename
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()
	fileName := fmt.Sprintf("%s/%s%s", directory, uuid.New().String(), filepath.Ext(file.Filename))

	// Upload to MinIO with a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = s.minioClient.PutObject(ctx, s.bucketName, fileName, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to MinIO: %v", err)
	}

	// Set file to public-read
	if err := s.minioService.SetFilePublicRead(fileName); err != nil {
		return "", fmt.Errorf("failed to set public-read policy: %v", err)
	}

	// Return public URL
	fileURL := fmt.Sprintf("https://%s/%s/%s", os.Getenv("MINIO_ENDPOINT"), s.bucketName, fileName)
	return fileURL, nil
}

func (s *DocumentControlService) GetDocumentControlsInternalPaginated(currentPage, pageSize int, search string) (*PaginatedResult, error) {
	var documentControls []models.DocumentControlJoined
	var totalRecords int64

	offset := (currentPage - 1) * pageSize

	// Base query for fetching document controls with joins
	query := config.DB.Model(&models.DocumentControlJoined{}).
		Select("document_control.*, category_document.name AS category_name, category_document.prefix AS category_prefix, "+
			"document_type.name AS type_name, document_type.prefix AS type_prefix, "+
			"status_document.name AS status_name").
		Joins("LEFT JOIN category_document ON category_document.id = document_control.document_category_id").
		Joins("LEFT JOIN document_type ON document_type.id = document_control.document_type_id").
		Joins("LEFT JOIN status_document ON status_document.id = document_control.status_document_id").
		Where("document_control.document_category_id = ?", 1).
		Where("document_control.deleted_at IS NULL")

	// Apply search filter if provided
	if search != "" {
		query = query.Where("LOWER(document_control.document_name) LIKE ?", "%"+search+"%")
	}

	// Get total record count with applied filters
	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, errors.New("failed to count document controls")
	}

	// Fetch paginated records
	if err := query.Offset(offset).Limit(pageSize).Find(&documentControls).Error; err != nil {
		return nil, errors.New("failed to fetch document controls")
	}

	// Convert each DocumentControl to DocumentControlResponse with formatted date
	var formattedControls []DocumentControlResponse
	for _, doc := range documentControls {
		formattedControls = append(formattedControls, formatDocumentControl(doc))
	}

	// Calculate total pages based on the record count and page size
	totalPages := int((totalRecords + int64(pageSize) - 1) / int64(pageSize))

	// Prepare the paginated result
	result := &PaginatedResult{
		Data:         formattedControls,
		CurrentPage:  currentPage,
		PerPage:      pageSize,
		TotalPages:   totalPages,
		TotalRecords: totalRecords,
	}

	return result, nil
}

func (s *DocumentControlService) GetDocumentControlsExternalPaginated(currentPage, pageSize int, search string) (*PaginatedResult, error) {
	var documentControls []models.DocumentControlJoined
	var totalRecords int64

	offset := (currentPage - 1) * pageSize

	// Base query for fetching document controls with joins
	query := config.DB.Model(&models.DocumentControlJoined{}).
		Select("document_control.*, category_document.name AS category_name, category_document.prefix AS category_prefix, "+
			"document_type.name AS type_name, document_type.prefix AS type_prefix, "+
			"status_document.name AS status_name").
		Joins("LEFT JOIN category_document ON category_document.id = document_control.document_category_id").
		Joins("LEFT JOIN document_type ON document_type.id = document_control.document_type_id").
		Joins("LEFT JOIN status_document ON status_document.id = document_control.status_document_id").
		Where("document_control.document_category_id = ?", 2).
		Where("document_control.deleted_at IS NULL")

	// Apply search filter if provided
	if search != "" {
		query = query.Where("LOWER(document_control.document_name) LIKE ?", "%"+search+"%")
	}

	// Get total record count with applied filters
	if err := query.Count(&totalRecords).Error; err != nil {
		return nil, errors.New("failed to count document controls")
	}

	// Fetch paginated records
	if err := query.Offset(offset).Limit(pageSize).Find(&documentControls).Error; err != nil {
		return nil, errors.New("failed to fetch document controls")
	}

	// Convert each DocumentControl to DocumentControlResponse with formatted date
	var formattedControls []DocumentControlResponse
	for _, doc := range documentControls {
		formattedControls = append(formattedControls, formatDocumentControl(doc))
	}

	// Calculate total pages based on the record count and page size
	totalPages := int((totalRecords + int64(pageSize) - 1) / int64(pageSize))

	// Prepare the paginated result
	result := &PaginatedResult{
		Data:         formattedControls,
		CurrentPage:  currentPage,
		PerPage:      pageSize,
		TotalPages:   totalPages,
		TotalRecords: totalRecords,
	}

	return result, nil
}

// GetDocumentControlByUUID retrieves a DocumentControl by its UUID
func (s *DocumentControlService) GetDocumentControlByUUID(uuid string) (*models.DocumentControl, error) {
	var documentControl models.DocumentControl

	if err := config.DB.Where("uuid = ?", uuid).Where("deleted_at IS NULL").First(&documentControl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("document control not found")
		}
		return nil, err
	}

	return &documentControl, nil
}

// DeleteDocumentControl deletes a DocumentControl and its associated initial version file from MinIO
func (s *DocumentControlService) DeleteDocumentControl(uuid string) error {
	var documentControl models.DocumentControl
	var documentVersion models.DocumentVersion

	// Start a transaction to delete control and version data atomically
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// Retrieve DocumentControl entry by UUID
		if err := tx.Where("uuid = ?", uuid).First(&documentControl).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("document control not found")
			}
			return fmt.Errorf("failed to retrieve document control: %w", err)
		}

		// Retrieve the related DocumentVersion
		if err := tx.Where("document_control_id = ?", documentControl.ID).First(&documentVersion).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("document version not found")
			}
			return fmt.Errorf("failed to retrieve document version: %w", err)
		}

		// Delete file from MinIO
		if documentVersion.File != "" {
			err := config.MinioClient.RemoveObject(context.Background(), "document-versions", documentVersion.File, minio.RemoveObjectOptions{})
			if err != nil {
				return fmt.Errorf("failed to delete file from MinIO: %w", err)
			}
		}

		// Soft delete DocumentVersion entry
		if err := tx.Delete(&documentVersion).Error; err != nil {
			return fmt.Errorf("failed to delete document version: %w", err)
		}

		// Soft delete DocumentControl entry
		if err := tx.Delete(&documentControl).Error; err != nil {
			return fmt.Errorf("failed to delete document control: %w", err)
		}

		return nil
	})

	return err
}

func (s *DocumentControlService) UpdateDocumentControl(uuid string, payload *DocumentControlPayload, fileHeader *multipart.FileHeader) (*models.DocumentControl, error) {
	var documentControl models.DocumentControl

	// Find the document control by UUID
	if err := config.DB.Where("uuid = ?", uuid).First(&documentControl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("document control not found")
		}
		return nil, fmt.Errorf("failed to find document control: %w", err)
	}

	// Parse the publish date
	publishDate, err := time.Parse("2006-01-02", payload.PublishDate)
	if err != nil {
		return nil, fmt.Errorf("invalid date format for publish_date: %w", err)
	}

	// Update document control fields
	documentControl.DocumentName = payload.DocumentName
	documentControl.Description = payload.Description
	documentControl.DocumentNumber = payload.DocumentNumber
	documentControl.ClauseNumber = payload.ClauseNumber
	documentControl.RevisionNumber = payload.RevisionNumber
	documentControl.PublishDate = publishDate
	documentControl.PageCount = payload.PageCount
	documentControl.DocumentTypeID = IntPtr(payload.DocumentTypeID)
	documentControl.DocumentCategoryID = IntPtr(payload.DocumentCategoryID)
	documentControl.SequenceNumber = IntPtr(payload.SequenceNumber)
	documentControl.StatusDocumentID = IntPtr(payload.StatusDocumentID)
	documentControl.UpdatedAt = time.Now()

	// Begin transaction to save updates and manage file versioning
	err = config.DB.Transaction(func(tx *gorm.DB) error {
		// Update document control in the database
		if err := tx.Save(&documentControl).Error; err != nil {
			return fmt.Errorf("failed to update document control: %w", err)
		}

		// Check if a file is provided for version update
		if fileHeader != nil {
			// Upload the new file to MinIO
			filePath, err := s.uploadFileToMinio(fileHeader, "document-versions")
			if err != nil {
				return fmt.Errorf("failed to upload file: %w", err)
			}

			// Create a new document version entry
			newDocumentVersion := models.DocumentVersion{
				File:              filePath,
				DocumentControlID: &documentControl.ID,
				Version:           IntPtr(documentControl.RevisionNumber),
				StatusDocumentID:  IntPtr(payload.StatusDocumentID),
				Note:              "Updated version",
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
			}

			// Save the new document version to the database
			if err := tx.Create(&newDocumentVersion).Error; err != nil {
				return fmt.Errorf("failed to create new document version: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &documentControl, nil
}
