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

type EventService struct {
	minioClient  *minio.Client
	minioService *MinioService
	bucketName   string
}

func NewEventService() (*EventService, error) {
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
	return &EventService{
		minioClient:  minioClient,
		minioService: minioService,
		bucketName:   bucketName,
	}, nil
}

type PaginationResultEvent struct {
	Total       int64          `json:"total"`
	CurrentPage int            `json:"current_page"`
	PageSize    int            `json:"page_size"`
	Events      []models.Event `json:"data"`
}

func (s *EventService) GetEvents(currentPage, pageSize int, search string) (*PaginationResultEvent, error) {
	var events []models.Event
	var total int64

	offset := (currentPage) * pageSize
	query := config.DB.Model(&models.Event{})

	if search != "" {
		query = query.Where("LOWER(title) LIKE ?", "%"+search+"%").Where("deleted_at", nil)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	err = query.Where("deleted_at", nil).Order("id DESC").Limit(pageSize).Offset(offset).Find(&events).Error
	if err != nil {
		return nil, err
	}

	return &PaginationResultEvent{
		Total:       total,
		CurrentPage: currentPage,
		PageSize:    pageSize,
		Events:      events,
	}, nil
}

func (s *EventService) SearchBySlug(search string) (models.Event, error) {
	var event models.Event
	err := config.DB.First(&event, "slug = ?", search).Where("deleted_at", nil).Error
	return event, err
}

// Admin Function
func (s *EventService) GetEventBySlug(slug string) (*models.Event, error) {
	var event models.Event
	if err := config.DB.Where("deleted_at", nil).Order("id DESC").Where("slug = ?", slug).First(&event).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (s *EventService) GetEventByUUID(uuidOrID string) (*models.Event, error) {
	var event models.Event
	// Attempt to query by UUID or ID
	if err := config.DB.Where("deleted_at", nil).Where("uuid = ?", uuidOrID).First(&event).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

// GetEventPaginated fetches paginated events with sorting
func (s *EventService) GetEventsPaginated(perPage, page int, sortBy string, sortDesc bool) (map[string]interface{}, error) {
	var events []models.Event
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
	config.DB.Model(&models.Event{}).Count(&totalRecords)

	// Fetch the paginated data with sorting
	if err := config.DB.Where("deleted_at", nil).Order(sortOrder).Limit(perPage).Offset(offset).Find(&events).Error; err != nil {
		return nil, errors.New("failed to fetch events")
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalRecords) / float64(perPage)))

	// Return the data in a paginated format
	result := map[string]interface{}{
		"data":          events,
		"current_page":  page,
		"per_page":      perPage,
		"total_pages":   totalPages,
		"total_records": totalRecords,
	}

	return result, nil
}

// CreateEvent handles the creation of a event along with the image upload
func (s *EventService) CreateEvent(event *models.Event, img *multipart.FileHeader) error {
	db, err := config.DB.DB()
	if err != nil {
		log.Printf("Failed to get database connection: %v", err)
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Ensure the sequence is correctly set to avoid conflicts
	if err := helpers.ResetSequenceToMax(db, "events", "id", "events_id_seq"); err != nil {
		log.Printf("Failed to reset sequence: %v", err)
		return fmt.Errorf("failed to reset sequence: %w", err)
	}

	newUUID := uuid.New().String()
	event.UUID = newUUID

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

		// Set the image URL to the event model
		event.ImageURL = filePath
	}

	// Save the event to the database
	if err := config.DB.Create(event).Error; err != nil {
		return fmt.Errorf("failed to create event: %v", err)
	}

	return nil
}

func (s *EventService) saveFileToMinio(file *multipart.FileHeader, fileName string) (string, error) {
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

func (s *EventService) UpdateEvent(uuid string, updatedEvent *models.Event, img *multipart.FileHeader) (*models.Event, error) {
	var event models.Event

	// Cari event yang ada berdasarkan UUID
	if err := config.DB.Where("uuid = ?", uuid).Where("deleted_at", nil).First(&event).Error; err != nil {
		return nil, errors.New("event not found")
	}

	// Jika ada gambar baru yang diupload, perbarui gambar
	if img != nil {
		randomString, err := GenerateRandomString(8)
		fileName := fmt.Sprintf("%s%s", randomString, filepath.Ext(img.Filename))

		// Upload file ke MinIO
		filePath, err := s.saveFileToMinio(img, fileName)
		if err != nil {
			log.Printf("Error uploading file to MinIO: %v", err)
			return nil, fmt.Errorf("failed to upload file to MinIO: %v", err)
		}

		// Set URL gambar baru di objek event
		updatedEvent.ImageURL = filePath
	}

	// Lakukan update data event
	if err := config.DB.Model(&event).Updates(updatedEvent).Error; err != nil {
		return nil, errors.New("failed to update event")
	}

	// Ambil ulang data event dari database untuk memastikan semua field ter-update
	if err := config.DB.Where("uuid = ?", uuid).First(&event).Where("deleted_at", nil).Error; err != nil {
		return nil, errors.New("failed to reload updated event")
	}

	return &event, nil
}

// DeleteEvent deletes a event by its slug
func (s *EventService) DeleteEvent(uuid string) error {
	// Get the current time for the soft delete
	currentTime := time.Now()

	// Update the DeletedAt field instead of deleting the record
	if err := config.DB.Model(&models.Event{}).Where("uuid = ?", uuid).Update("deleted_at", currentTime).Error; err != nil {
		return errors.New("failed to soft delete event")
	}
	return nil
}
