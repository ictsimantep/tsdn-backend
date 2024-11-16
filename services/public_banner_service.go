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

type BannerService struct {
	minioClient  *minio.Client
	minioService *MinioService
	bucketName   string
}

func NewBannerService() (*BannerService, error) {
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

	minioService := NewMinioService(minioClient)

	// Return a new PublicationService with the initialized MinIO client
	return &BannerService{
		minioClient:  minioClient,
		minioService: minioService,
		bucketName:   bucketName,
	}, nil
}

type PaginationResultBanner struct {
	Total       int64           `json:"total"`
	CurrentPage int             `json:"current_page"`
	PageSize    int             `json:"page_size"`
	Banners     []models.Banner `json:"data"`
}

func (s *BannerService) GetBanners(currentPage, pageSize int, search string) (*PaginationResultBanner, error) {
	var banners []models.Banner
	var total int64

	offset := (currentPage - 1) * pageSize
	query := config.DB.Model(&models.Banner{})

	if search != "" {
		query = query.Where("LOWER(title) LIKE ?", "%"+search+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	// Tambahkan pengurutan berdasarkan id secara descending
	err = query.Where("deleted_at", nil).Order("id DESC").Limit(pageSize).Offset(offset).Find(&banners).Error
	if err != nil {
		return nil, err
	}

	return &PaginationResultBanner{
		Total:       total,
		CurrentPage: currentPage,
		PageSize:    pageSize,
		Banners:     banners,
	}, nil
}

// Where("deleted_at", nil).Order("id DESC")
// Admin Function
func (s *BannerService) GetBannerBySlug(slug string) (*models.Banner, error) {
	var banner models.Banner
	if err := config.DB.Where("deleted_at", nil).Order("id DESC").Where("slug = ?", slug).First(&banner).Error; err != nil {
		return nil, err
	}
	return &banner, nil
}

func (s *BannerService) GetBannerByUUID(uuidOrID string) (*models.Banner, error) {
	var banner models.Banner
	// Attempt to query by UUID or ID
	if err := config.DB.Where("deleted_at", nil).Where("uuid = ? OR id = ?", uuidOrID, uuidOrID).First(&banner).Error; err != nil {
		return nil, err
	}
	return &banner, nil
}

// GetBannerPaginated fetches paginated banners with sorting
func (s *BannerService) GetBannersPaginated(perPage, page int, sortBy string, sortDesc bool) (map[string]interface{}, error) {
	var banners []models.Banner
	var totalRecords int64
	var sortOrder string

	// Calculate offset for pagination
	offset := (page) * perPage

	// Set the sort order based on the `sortDesc` flag
	if sortDesc {
		sortOrder = sortBy + " ASC"
	} else {
		sortOrder = sortBy + " DESC"
	}

	// Get the total number of records
	config.DB.Model(&models.Banner{}).Count(&totalRecords)

	// Fetch the paginated data with sorting
	if err := config.DB.Where("deleted_at", nil).Order(sortOrder).Limit(perPage).Offset(offset).Find(&banners).Error; err != nil {
		return nil, errors.New("failed to fetch banners")
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalRecords) / float64(perPage)))

	// Return the data in a paginated format
	result := map[string]interface{}{
		"data":          banners,
		"current_page":  page,
		"per_page":      perPage,
		"total_pages":   totalPages,
		"total_records": totalRecords,
	}

	return result, nil
}

// CreateBanner handles the creation of a banner along with the image upload
func (s *BannerService) CreateBanner(banner *models.Banner, img *multipart.FileHeader) error {
	db, err := config.DB.DB()
	if err != nil {
		log.Printf("Failed to get database connection: %v", err)
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Ensure the sequence is correctly set to avoid conflicts
	if err := helpers.ResetSequenceToMax(db, "banners", "id", "banners_id_seq"); err != nil {
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
		banner.ImageURL = &filePath
	}

	// Save the banner to the database
	if err := config.DB.Create(banner).Error; err != nil {
		return fmt.Errorf("failed to create banner: %v", err)
	}

	return nil
}

func (s *BannerService) saveFileToMinio(file *multipart.FileHeader, fileName string) (string, error) {
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

	// Set file to public-read using MinioService
	if err := s.minioService.SetFilePublicRead(fileName); err != nil {
		return "", fmt.Errorf("failed to set public-read policy: %v", err)
	}

	// Generate the file URL
	fileURL := fmt.Sprintf("https://%s/%s/%s", os.Getenv("MINIO_ENDPOINT"), s.bucketName, fileName)
	return fileURL, nil

}

// UpdateBanner updates an existing banner by its UUID and handles image uploads
func (s *BannerService) UpdateBanner(uuid string, updatedBanner *models.Banner, img *multipart.FileHeader) error {
	var banner models.Banner

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
		updatedBanner.ImageURL = &filePath
	}

	// Update the banner record with the new data
	if err := config.DB.Model(&banner).Updates(updatedBanner).Error; err != nil {
		return errors.New("failed to update banner")
	}

	return nil
}

// DeleteBanner deletes a banner by its slug
func (s *BannerService) DeleteBanner(uuid string) error {
	// Get the current time for the soft delete
	currentTime := time.Now()

	// Update the DeletedAt field instead of deleting the record
	if err := config.DB.Model(&models.Banner{}).Where("uuid = ? OR id = ?", uuid, uuid).Update("deleted_at", currentTime).Error; err != nil {
		return errors.New("failed to soft delete banner")
	}
	return nil
}
