package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var MinioClient *minio.Client

// InitializeMinioClient sets up a connection to the MinIO server and initializes the MinioClient variable
func InitializeMinioClient() error {
	// Retrieve MinIO configuration from environment variables
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	// Validate required environment variables
	if endpoint == "" || accessKeyID == "" || secretAccessKey == "" {
		return fmt.Errorf("missing required MinIO configuration in environment variables")
	}

	// Initialize MinIO client
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize MinIO client: %v", err)
	}

	MinioClient = client
	log.Println("MinIO client initialized successfully")
	return nil
}

// EnsureBucket ensures that the specified bucket exists; if it doesn't, the function attempts to create it
func EnsureBucket(bucketName, location string) error {
	// Check if the bucket already exists
	exists, err := MinioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		return fmt.Errorf("failed to check if bucket exists: %v", err)
	}

	if exists {
		log.Printf("Bucket %s already exists.\n", bucketName)
		return nil
	}

	// Try to create the bucket if it doesn't exist
	err = MinioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		return fmt.Errorf("failed to create bucket %s: %v", bucketName, err)
	}

	log.Printf("Bucket %s created successfully.\n", bucketName)
	return nil
}
