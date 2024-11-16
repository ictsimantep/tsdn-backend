package services

import (
	"backend-school/config"
	"backend-school/helpers"
	"backend-school/models"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type GalleryService struct {
	minioClient *minio.Client
	bucketName  string
}

func NewGalleryService() (*GalleryService, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	bucketName := os.Getenv("MINIO_BUCKET")

	// Initialize the MinIO client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MinIO client: %v", err)
	}

	// Return a new GalleryService with the initialized MinIO client
	return &GalleryService{
		minioClient: minioClient,
		bucketName:  bucketName,
	}, nil
}

type PaginationResultGallery struct {
	Total       int64            `json:"total"`
	CurrentPage int              `json:"current_page"`
	PageSize    int              `json:"page_size"`
	Gallery     []models.Gallery `json:"data"`
}

// GetGalleryWithCategoryFilter fetches galleries with pagination, optional search, and category filter
func GetGalleryWithCategoryFilter(currentPage, pageSize int, search string, category string) (*PaginationResultGallery, error) {
	var galleries []models.Gallery
	var total int64

	offset := (currentPage) * pageSize
	query := config.DB.Model(&models.Gallery{})

	// Apply search filter if provided
	if search != "" {
		query = query.Where("LOWER(category) LIKE ?", "%"+search+"%").Or("LOWER(year) LIKE ?", "%"+search+"%")
	}

	// Apply category filter if provided
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// Count total galleries matching the query
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Fetch galleries with limit and offset
	if err := query.Limit(pageSize).Offset(offset).Find(&galleries).Error; err != nil {
		return nil, err
	}

	result := &PaginationResultGallery{
		Total:       total,
		CurrentPage: currentPage,
		PageSize:    pageSize,
		Gallery:     galleries,
	}
	return result, nil
}

func GroupByCategory(currentPage, pageSize int, search string, category string) (*PaginationResultGallery, error) {
	var galleries []models.Gallery
	var total int64

	offset := (currentPage) * pageSize
	query := config.DB.Model(&models.Gallery{})

	// Apply search filter if provided
	if search != "" {
		query = query.Where("LOWER(category) LIKE ?", "%"+search+"%").
			Or("LOWER(year) LIKE ?", "%"+search+"%")
	}

	// Apply category filter if provided
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// Group by category and select one gallery per category (e.g., the latest one)
	query = query.Select("DISTINCT ON (category) *").Order("category, year DESC")

	// Count total galleries matching the query
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Fetch galleries with limit and offset
	if err := query.Limit(pageSize).Offset(offset).Find(&galleries).Error; err != nil {
		return nil, err
	}

	result := &PaginationResultGallery{
		Total:       total,
		CurrentPage: currentPage,
		PageSize:    pageSize,
		Gallery:     galleries,
	}
	return result, nil
}

// Paginated fetches paginated gallerys with sorting
func (s *GalleryService) GetGallerysPaginated(perPage, page int, sortBy string, sortDesc bool) (map[string]interface{}, error) {
	var gallerys []models.Gallery
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
	config.DB.Model(&models.Gallery{}).Count(&totalRecords)

	// Fetch the paginated data with sorting
	if err := config.DB.Where("deleted_at", nil).Order(sortOrder).Limit(perPage).Offset(offset).Find(&gallerys).Error; err != nil {
		return nil, errors.New("failed to fetch gallerys")
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalRecords) / float64(perPage)))

	// Return the data in a paginated format
	result := map[string]interface{}{
		"data":          gallerys,
		"current_page":  page,
		"per_page":      perPage,
		"total_pages":   totalPages,
		"total_records": totalRecords,
	}

	return result, nil
}

// CreateGallery handles the creation of a banner along with the image upload
func (s *GalleryService) CreateGallery(banner *models.Gallery, img *multipart.FileHeader) error {
	db, err := config.DB.DB()
	if err != nil {
		log.Printf("Failed to get database connection: %v", err)
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Ensure the sequence is correctly set to avoid conflicts
	if err := helpers.ResetSequenceToMax(db, "galleries", "id", "galleries_id_seq"); err != nil {
		log.Printf("Failed to reset sequence: %v", err)
		return fmt.Errorf("failed to reset sequence: %w", err)
	}

	newUUID := uuid.New().String()
	banner.UUID = newUUID

	if img != nil {
		// Generate a random string for the filename
		randomString, err := GenerateRandomString(8)
		// Generate a unique filename for the image to avoid name clashes
		fileName := fmt.Sprintf("%s%s", randomString, filepath.Ext(img.Filename))

		// Upload the file to MinIO
		filePath, err := s.saveFileToMinio(img, fileName)
		if err != nil {
			return fmt.Errorf("failed to upload file to MinIO: %v", err)
		}

		// Set the image URL to the banner model
		banner.ImageURL = filePath
	}

	// Save the banner to the database
	if err := config.DB.Create(banner).Error; err != nil {
		return fmt.Errorf("failed to create banner: %v", err)
	}

	return nil
}

func (s *GalleryService) GetGalleryBySlug(slug string) (*models.Gallery, error) {
	var gallery models.Gallery
	if err := config.DB.Where("deleted_at", nil).First(&gallery).Error; err != nil {
		return nil, err
	}
	return &gallery, nil
}

func (s *GalleryService) GetGalleryByUUID(uuid string) (*models.Gallery, error) {
	var gallery models.Gallery
	if err := config.DB.Where("uuid = ?", uuid).Where("deleted_at", nil).First(&gallery).Error; err != nil {
		return nil, err
	}
	return &gallery, nil
}

func (s *GalleryService) saveFileToMinio(file *multipart.FileHeader, fileName string) (string, error) {
	// Open the file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	// Upload the file to MinIO
	ctx := context.Background()
	_, err = s.minioClient.PutObject(ctx, s.bucketName, fileName, src, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to MinIO: %v", err)
	}

	// Generate the file URL
	fileURL := fmt.Sprintf("https://%s/%s/%s", os.Getenv("MINIO_ENDPOINT"), s.bucketName, fileName)
	return fileURL, nil

}

// UpdateGallery updates an existing banner by its UUID and handles image uploads
func (s *GalleryService) UpdateGallery(uuid string, updatedGallery *models.Gallery, img *multipart.FileHeader) error {
	var banner models.Gallery

	// Fetch the existing banner by UUID
	if err := config.DB.Where("uuid = ? OR id = ?", uuid, uuid).First(&banner).Error; err != nil {
		return errors.New("banner not found")
	}

	// If an image is uploaded, handle the image update
	if img != nil {
		randomString, err := GenerateRandomString(8)
		// Generate a unique filename for the image to avoid name clashes
		fileName := fmt.Sprintf("%s%s", randomString, filepath.Ext(img.Filename))

		// Upload the file to MinIO
		filePath, err := s.saveFileToMinio(img, fileName)
		if err != nil {
			log.Printf("Error uploading file to MinIO: %v", err)
			return fmt.Errorf("failed to upload file to MinIO: %v", err)
		}

		// Update the image field in the banner object
		updatedGallery.ImageURL = filePath
	}

	// Update the banner record with the new data
	if err := config.DB.Model(&banner).Updates(updatedGallery).Error; err != nil {
		return errors.New("failed to update banner")
	}

	return nil
}

// DeleteGallery deletes a banner by its slug
func (s *GalleryService) DeleteGallery(uuid string) error {
	// Get the current time for the soft delete
	currentTime := time.Now()

	// Update the DeletedAt field instead of deleting the record
	if err := config.DB.Model(&models.Gallery{}).Where("uuid = ? OR id = ?", uuid, uuid).Update("deleted_at", currentTime).Error; err != nil {
		return errors.New("failed to soft delete banner")
	}
	return nil
}
